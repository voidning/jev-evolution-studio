package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type OperationDescriptor struct {
	ID, Family, Description                            string
	Allowed, Forbidden, Parameters, Affects, Executors []string
	Defaults                                           map[string]string
	Validator, Risk                                    string
	Reversible                                         bool
}

func operationCatalog(t Target, m ProjectMeta) map[string][]OperationDescriptor {
	out := map[string][]OperationDescriptor{"appearance": {}, "typography": {}, "layout": {}, "content": {}, "structure": {}, "behavior": {}}
	for id, op := range registry {
		if op.Targets == "container" && !contains([]string{"div", "section", "main", "article", "aside", "header", "footer", "nav", "ul", "ol", "li", "form"}, t.Tag) {
			continue
		}
		if id == "insert_child" && !m.Tailwind {
			continue
		}
		if id == "set_color" && !m.Tailwind {
			continue
		}
		executors := []string{"css", "tailwind"}
		risk := "low"
		if id == "insert_child" {
			executors = []string{"jsx"}
			risk = "medium"
		} else if contains([]string{"set_color", "set_utility", "margin"}, id) {
			executors = []string{"tailwind"}
		}
		if t.ClassKind == "dynamic" || t.Repeated {
			continue
		}
		out[op.Family] = append(out[op.Family], OperationDescriptor{ID: id, Family: op.Family, Description: descriptions[id], Allowed: []string{op.Targets}, Forbidden: []string{"repeated", "unresolved-dynamic", "third-party-internal"}, Parameters: []string{"scope", "value", "breakpoint", "preserve"}, Affects: op.Properties, Executors: executors, Defaults: map[string]string{"scope": "selected", "breakpoint": "all"}, Validator: "registry + source revision + AST parse + build + HMR", Risk: risk, Reversible: true})
	}
	return out
}

var descriptions = map[string]string{"set_color": "修改指定通道颜色，保留其他状态和断点", "insert_child": "仅在静态空容器中插入注册按钮", "set_utility": "修改注册的对齐、阴影或透明度档位", "density": "调整内边距和间距", "padding": "调整内边距", "margin": "调整外边距", "gap": "调整子项间距", "fontSize": "调整文字大小", "fontWeight": "调整字重", "textAlign": "文字对齐", "columns": "网格列数", "axis": "横纵布局", "center": "容器内容居中", "width": "宽度档位", "radius": "圆角档位", "border": "边框可见性", "background": "背景强调程度", "emphasis": "元素视觉强调", "visibility": "显示或隐藏元素"}

func init() {
	for _, v := range []struct {
		id, family, target string
		props              []string
	}{{"set_color", "appearance", "any", []string{"className:color"}}, {"insert_child", "structure", "container", []string{"JSXElement.children", "className", "ImportDeclaration"}}, {"set_utility", "layout", "any", []string{"className:utility"}}, {"margin", "spacing", "any", []string{"margin"}}} {
		registry[v.id] = Operation{Family: v.family, Targets: v.target, Properties: v.props, Breakpoints: true, Reversible: true}
	}
}

var colorCommand = regexp.MustCompile(`^(?:把)?(?:这个按钮|按钮|这个元素|背景|文字颜色|文字|文本|边框颜色|边框)?(?:的)?(?:背景|颜色)?(?:换成|改成|设为|变成)(.+)$`)
var insertCommand = regexp.MustCompile(`^在(?:这个)?(?:空框|框|空容器|容器)?中间加一个(?:叫)?(?:[“‘"']([^”’"']+)[”’"'](?:的)?)?按钮$`)
var hexColor = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)
var functionColor = regexp.MustCompile(`^(rgb|hsl|oklch)\([0-9.,%/ +\-]+\)$`)

func parseColor(value string, m ProjectMeta) (*ColorValue, bool) {
	value = strings.ToLower(value)
	names := map[string]string{"红色": "red", "红": "red", "蓝色": "blue", "蓝": "blue", "绿色": "green", "黄色": "yellow", "橙色": "orange", "紫色": "purple", "粉色": "pink", "白色": "white", "黑色": "black", "灰色": "gray", "透明": "transparent"}
	if v, ok := names[value]; ok {
		value = v
	}
	if contains([]string{"red", "blue", "green", "yellow", "orange", "purple", "pink", "slate", "gray", "black", "white", "transparent"}, value) {
		return &ColorValue{Kind: "literal", Value: value, Shade: "600"}, true
	}
	if hexColor.MatchString(value) {
		return &ColorValue{Kind: "literal", Value: value}, true
	}
	if functionColor.MatchString(value) {
		parts := regexp.MustCompile(`[0-9]+(?:\.[0-9]+)?`).FindAllString(value, -1)
		if len(parts) == 3 || len(parts) == 4 {
			return &ColorValue{Kind: "literal", Value: value}, true
		}
	}
	tokens := map[string]string{"品牌色": "brand", "主色": "primary", "危险色": "danger", "警告色": "warning", "成功色": "success", "弱化色": "muted"}
	key := tokens[value]
	if key == "" && m.Tokens[value] != "" {
		key = value
	}
	if key != "" && m.Tokens[key] != "" {
		return &ColorValue{Kind: "token", Value: key}, true
	}
	return nil, false
}
func localSourceIntent(prompt string, m ProjectMeta) ([]EditIntent, error) {
	p := strings.TrimSpace(prompt)
	p = strings.TrimRight(p, "。！!")
	scope, bp, state := "selected", "all", ""
	for _, v := range [][2]string{{"父容器", "container"}, {"同级元素", "siblings"}} {
		if strings.HasPrefix(p, v[0]) {
			scope = v[1]
			p = strings.TrimPrefix(p, v[0])
		}
	}
	for n := 0; n < 2; n++ {
		for _, v := range [][2]string{{"桌面端", "desktop"}, {"手机端", "mobile"}, {"移动端", "mobile"}, {"sm:", "sm"}, {"md:", "md"}, {"lg:", "lg"}} {
			if strings.HasPrefix(p, v[0]) {
				bp = v[1]
				p = strings.TrimPrefix(p, v[0])
			}
		}
		for _, v := range [][2]string{{"悬停时", "hover"}, {"hover:", "hover"}, {"聚焦时", "focus"}, {"focus:", "focus"}} {
			if strings.HasPrefix(p, v[0]) {
				state = v[1]
				p = strings.TrimPrefix(p, v[0])
			}
		}
	}
	makeIntent := func(op string, args IntentArguments) []EditIntent {
		i := intent(op, "set", "", bp, scope)
		i.State = state
		i.Arguments = args
		i.Preserve = []string{"content", "behavior", "responsive"}
		return []EditIntent{i}
	}
	if match := insertCommand.FindStringSubmatch(p); match != nil {
		if !m.Tailwind || bp != "all" || state != "" || scope != "selected" {
			return nil, fmt.Errorf("按钮插入只支持 Tailwind 静态空容器的 selected/all 作用域")
		}
		i := makeIntent("insert_child", IntentArguments{Element: "button", Placement: "center", Label: match[1]})
		i[0].Preserve = append(i[0].Preserve, "existing_children")
		return i, nil
	}
	if match := colorCommand.FindStringSubmatch(p); match != nil {
		channel := "background"
		if strings.Contains(p, "文字") || strings.Contains(p, "文本") {
			channel = "text"
		}
		if strings.Contains(p, "边框") {
			channel = "border"
		}
		color, ok := parseColor(match[1], m)
		if !ok {
			return nil, fmt.Errorf("不支持：颜色不明确或项目没有该设计令牌")
		}
		return makeIntent("set_color", IntentArguments{Channel: channel, Color: color}), nil
	}
	utility := map[string][2]string{"阴影大一点": {"shadow", "shadow-lg"}, "阴影小一点": {"shadow", "shadow-sm"}, "隐藏阴影": {"shadow", "shadow-none"}, "透明度50%": {"opacity", "opacity-50"}, "透明度100%": {"opacity", "opacity-100"}, "交叉轴居中": {"items", "items-center"}, "交叉轴靠左": {"items", "items-start"}, "交叉轴靠右": {"items", "items-end"}, "主轴居中": {"justify", "justify-center"}, "两端对齐": {"justify", "justify-between"}, "显示为flex": {"display", "flex"}, "显示为grid": {"display", "grid"}, "显示为block": {"display", "block"}}
	if v, ok := utility[p]; ok {
		return makeIntent("set_utility", IntentArguments{Channel: v[0], Utility: v[1]}), nil
	}
	if p == "外边距大一点" || p == "外边距小一点" {
		i := makeIntent("margin", IntentArguments{})
		i[0].Direction = "increase"
		if strings.Contains(p, "小") {
			i[0].Direction = "decrease"
		}
		return i, nil
	}
	aliases := map[string]string{"文字大一点": "文字变大一点", "间距小一点": "间距减少", "间距大一点": "间距增加"}
	if v, ok := aliases[p]; ok {
		p = v
	}
	is, e := LocalIntent(p)
	if e != nil {
		return nil, e
	}
	for j := range is {
		is[j].Scope = scope
		is[j].State = state
		if bp != "all" {
			is[j].Breakpoint = bp
		}
		is[j].Preserve = append(is[j].Preserve, "behavior")
	}
	return is, nil
}
func understandEdit(ctx context.Context, prompt string, target Target, meta ProjectMeta, offline bool) ([]EditIntent, string, error) {
	local, e := localSourceIntent(prompt, meta)
	if e == nil {
		return local, "offline", nil
	} // Exact colors and quoted labels never need a network round trip.
	if offline || loadAPIKey() == "" {
		return nil, "offline", e
	}
	return semanticJev(ctx, prompt, target, meta)
}
func semanticJev(ctx context.Context, prompt string, target Target, meta ProjectMeta) ([]EditIntent, string, error) {
	if !meta.Tailwind {
		return nil, "jev", unsupported
	}
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	choice := func(description string, values ...string) map[string]any {
		c := map[string]string{}
		for _, v := range values {
			c[v] = v
		}
		return map[string]any{"type": "choice", "instructions": description, "criteria": c}
	}
	questions := map[string]any{"family": choice("Select appearance only for a color-only request. Any content, logic, or mixed request is unsupported.", "appearance", "typography", "layout", "unsupported"), "appearance_operation": choice("Only a fuzzy color adjustment is supported on this online path.", "set_color", "emphasis", "unsupported"), "scope": choice("Only the selected element is eligible.", "selected", "container", "unsupported"), "channel": choice("Choose the explicitly requested color channel; button color defaults to background.", "background", "text", "border", "unsupported"), "color": choice("Choose a constrained color meaning. Never invent classes or colors.", "warm", "cool", "vivid", "soft", "darker", "lighter", "danger", "brand", "unsupported"), "breakpoint": choice("Choose only an explicitly requested breakpoint.", "all", "desktop", "mobile", "unsupported"), "preserve": map[string]any{"type": "noul", "instructions": "All content, business behavior, and unspecified states/responsive variants must remain unchanged. Reject mixed or code-generation requests."}}
	colors := []string{"warm", "cool", "vivid", "soft", "unsupported"}
	if regexp.MustCompile(`(?:bg|text|border)-(?:red|blue|green|yellow|orange|purple|pink|slate|gray)-[1-9]00`).MatchString(target.ClassName) {
		colors = append(colors, "darker", "lighter")
	}
	for _, key := range []string{"brand", "danger"} {
		if meta.Tokens[key] != "" {
			colors = append(colors, key)
		}
	}
	questions["color"] = choice("Select only among the available contextual color meanings.", colors...)
	body, _ := json.Marshal(map[string]any{"model": "jev-1.13.0", "state": map[string]any{"prompt": prompt, "target": target, "tokens": meta.Tokens, "registry": operationCatalog(target, meta)}, "questions": questions})
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.typesafe.ai/v1/systemone", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+loadAPIKey())
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	res, e := client.Do(req)
	if e != nil {
		return nil, "offline-fallback", unsupported
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, "offline-fallback", unsupported
	}
	return decodeSemantic(io.LimitReader(res.Body, 65536), target, meta)
}
func decodeSemantic(reader io.Reader, target Target, meta ProjectMeta) ([]EditIntent, string, error) {
	var r struct {
		Answers map[string]struct {
			Choice     string  `json:"choice"`
			Confidence float64 `json:"confidence"`
			Noul       float64 `json:"noul"`
		} `json:"answers"`
	}
	if json.NewDecoder(reader).Decode(&r) != nil {
		return nil, "jev", unsupported
	}
	for _, k := range []string{"family", "appearance_operation", "scope", "channel", "color", "breakpoint"} {
		a := r.Answers[k]
		if a.Confidence < .8 || a.Confidence > 1 {
			return nil, "jev", unsupported
		}
	}
	if r.Answers["family"].Choice != "appearance" || r.Answers["appearance_operation"].Choice != "set_color" || r.Answers["scope"].Choice != "selected" || r.Answers["preserve"].Noul < .8 || r.Answers["preserve"].Noul > 1 {
		return nil, "jev", unsupported
	}
	channel, bp := r.Answers["channel"].Choice, r.Answers["breakpoint"].Choice
	if !contains([]string{"background", "text", "border"}, channel) || !contains([]string{"all", "desktop", "mobile"}, bp) {
		return nil, "jev", unsupported
	}
	semantic := r.Answers["color"].Choice
	color := ColorValue{Kind: "literal", Shade: "600"}
	switch semantic {
	case "warm":
		color.Value = "orange"
	case "cool":
		color.Value = "blue"
	case "vivid":
		color.Value = "pink"
	case "soft":
		color.Value = "slate"
		color.Shade = "100"
	case "danger", "brand":
		if meta.Tokens[semantic] == "" {
			return nil, "jev", fmt.Errorf("项目没有 %s 设计令牌", semantic)
		}
		color.Kind = "token"
		color.Value = semantic
	case "darker", "lighter":
		prefix := map[string]string{"background": "bg-", "text": "text-", "border": "border-"}[channel]
		prefix = map[string]string{"all": "", "desktop": "md:", "mobile": "max-md:"}[bp] + prefix
		re := regexp.MustCompile(`^` + prefix + `(red|blue|green|yellow|orange|purple|pink|slate|gray)-(100|200|300|400|500|600|700|800|900)$`)
		for _, c := range strings.Fields(target.ClassName) {
			if match := re.FindStringSubmatch(c); match != nil {
				color.Value = match[1]
				shades := []string{"100", "200", "300", "400", "500", "600", "700", "800", "900"}
				for idx, shade := range shades {
					if shade == match[2] {
						if semantic == "darker" {
							idx = min(idx+1, 8)
						} else {
							idx = max(idx-1, 0)
						}
						color.Shade = shades[idx]
						break
					}
				}
				break
			}
		}
		if color.Value == "" {
			return nil, "jev", fmt.Errorf("无法从当前基础颜色推导明暗，请指定颜色")
		}
	default:
		return nil, "jev", unsupported
	}
	if err := checkSemanticContrast(color, channel, target, meta); err != nil {
		return nil, "jev", err
	}
	i := intent("set_color", "set", "", bp, "selected")
	i.Arguments = IntentArguments{Channel: channel, Color: &color}
	i.Confidence = r.Answers["color"].Confidence
	i.Preserve = append(i.Preserve, "behavior")
	return []EditIntent{i}, "jev", nil
}
