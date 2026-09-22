package main

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type NumericRange struct {
	Min, Max, Step, Default float64
	Unit                    string
}

type Operation struct {
	Ranges      map[string]NumericRange
	Family      string
	Targets     string
	Properties  []string
	Values      []string
	Breakpoints bool
	Reversible  bool
	Generate    func(EditIntent, Target) (map[string]string, error)
}

var registry = map[string]Operation{}

func init() {
	add := func(name, family, targets string, props, values []string) {
		op := Operation{Family: family, Targets: targets, Properties: props, Values: values, Breakpoints: true, Reversible: true, Generate: generate, Ranges: map[string]NumericRange{}}
		if len(values) == 0 {
			for _, p := range props {
				r := NumericRange{0, 128, 4, 0, "px"}
				switch p {
				case "gap":
					r.Max = 96
				case "font-size":
					r = NumericRange{10, 120, 4, 16, "px"}
				case "font-weight":
					r = NumericRange{100, 900, 100, 400, ""}
				}
				op.Ranges[p] = r
			}
		}
		registry[name] = op
	}
	add("density", "spacing", "container", []string{"padding-top", "padding-right", "padding-bottom", "padding-left", "gap"}, nil)
	add("padding", "spacing", "any", []string{"padding-top", "padding-right", "padding-bottom", "padding-left"}, nil)
	add("gap", "spacing", "container", []string{"gap"}, nil)
	add("fontSize", "typography", "text", []string{"font-size"}, nil)
	add("fontWeight", "typography", "text", []string{"font-weight"}, nil)
	add("textAlign", "typography", "any", []string{"text-align"}, []string{"left", "center", "right"})
	add("center", "layout", "container", []string{"display", "align-items", "justify-content", "text-align"}, []string{"center"})
	add("columns", "layout", "container", []string{"display", "grid-template-columns", "min-width"}, []string{"1", "2", "3", "4"})
	add("axis", "layout", "container", []string{"display", "flex-direction", "flex-wrap"}, []string{"row", "column"})
	add("width", "layout", "any", []string{"width", "max-width", "box-sizing"}, []string{"narrow", "normal", "wide", "full"})
	add("radius", "appearance", "any", []string{"border-radius"}, []string{"sharp", "medium", "round"})
	add("background", "appearance", "any", []string{"background-color", "color"}, []string{"stronger", "softer", "transparent"})
	add("border", "appearance", "any", []string{"border"}, []string{"show", "hide"})
	add("emphasis", "appearance", "any", []string{"background-color", "color", "border", "font-weight", "box-shadow", "opacity"}, []string{"primary", "normal", "muted"})
	add("visibility", "visibility", "any", []string{"display"}, []string{"hidden", "visible"})
}
func contains(xs []string, x string) bool {
	for _, s := range xs {
		if s == x {
			return true
		}
	}
	return false
}

var idRE = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,100}$`)

func validateIntent(in EditIntent, t Target) error {
	op, ok := registry[in.Operation]
	if !ok || op.Family != in.Family {
		return unsupported
	}
	if !idRE.MatchString(t.ID) || t.Count != 1 {
		return fmt.Errorf("目标 ID 不存在或重复")
	}
	if !contains([]string{"selected", "siblings", "container"}, in.Scope) || !contains([]string{"all", "mobile", "desktop", "sm", "md", "lg"}, in.Breakpoint) || !contains([]string{"subtle", "medium", "strong"}, in.Magnitude) || math.IsNaN(in.Confidence) || in.Confidence < .75 || in.Confidence > 1 {
		return unsupported
	}
	if !contains(in.Preserve, "content") || !contains(in.Preserve, "responsive") {
		return unsupported
	}
	if in.Operation == "insert_child" {
		if in.Direction != "set" || in.Arguments.Element != "button" || in.Arguments.Placement != "center" || in.Breakpoint != "all" || in.State != "" {
			return unsupported
		}
		return nil
	}
	if in.Operation == "set_color" {
		if in.Direction != "set" || in.Arguments.Color == nil || !contains([]string{"background", "text", "border"}, in.Arguments.Channel) {
			return unsupported
		}
		return nil
	}
	if in.Operation == "set_utility" {
		allowed := map[string][]string{"shadow": {"shadow-sm", "shadow-lg", "shadow-none"}, "opacity": {"opacity-50", "opacity-100"}, "items": {"items-start", "items-center", "items-end"}, "justify": {"justify-center", "justify-between"}, "display": {"block", "flex", "grid"}}
		if !contains(allowed[in.Arguments.Channel], in.Arguments.Utility) {
			return unsupported
		}
		return nil
	}
	if len(op.Values) > 0 {
		if in.Direction != "set" || !contains(op.Values, in.Value) {
			return unsupported
		}
	} else if !contains([]string{"increase", "decrease"}, in.Direction) || in.Value != "" {
		return unsupported
	}
	if op.Targets == "container" && !contains([]string{"div", "section", "main", "article", "aside", "header", "footer", "nav", "ul", "ol", "li", "form"}, t.Tag) {
		return fmt.Errorf("不支持：该操作需要容器，请选择卡片组或其父容器")
	}
	if op.Targets == "text" && !contains([]string{"h1", "h2", "h3", "h4", "h5", "h6", "p", "span", "a", "button", "label", "li", "strong", "em", "small"}, t.Tag) {
		return fmt.Errorf("不支持：请选择标题、文字或按钮")
	}
	for _, p := range t.Important {
		conflict := p == "all"
		for _, edited := range op.Properties {
			if p == edited || strings.HasPrefix(p, edited+"-") || strings.HasPrefix(edited, p+"-") {
				conflict = true
			}
		}
		if conflict {
			return fmt.Errorf("不支持：目标含内联 !important 样式 %s", p)
		}
	}
	return nil
}
func numeric(t Target, p string, defaultValue float64) (float64, error) {
	s := t.Styles[p]
	if s == "" || s == "normal" {
		return defaultValue, nil
	}
	s = strings.TrimSuffix(s, "px")
	v, e := strconv.ParseFloat(s, 64)
	if e != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("不支持：无法确定 %s 当前值", p)
	}
	return v, nil
}
func generate(in EditIntent, t Target) (map[string]string, error) {
	if e := validateIntent(in, t); e != nil {
		return nil, e
	}
	out := map[string]string{}
	adjust := func(p string) error {
		r, ok := registry[in.Operation].Ranges[p]
		if !ok {
			return unsupported
		}
		delta := r.Step * map[string]float64{"subtle": 1, "medium": 2, "strong": 4}[in.Magnitude]
		if in.Direction == "decrease" {
			delta = -delta
		}
		v, e := numeric(t, p, r.Default)
		if e != nil {
			return e
		}
		v = math.Max(r.Min, math.Min(r.Max, v+delta))
		out[p] = strconv.FormatFloat(v, 'f', -1, 64) + r.Unit
		return nil
	}
	switch in.Operation {
	case "density", "padding", "gap":
		if in.Operation != "gap" {
			for _, p := range []string{"padding-top", "padding-right", "padding-bottom", "padding-left"} {
				if e := adjust(p); e != nil {
					return nil, e
				}
			}
		}
		if in.Operation != "padding" {
			if e := adjust("gap"); e != nil {
				return nil, e
			}
		}
	case "fontSize":
		if e := adjust("font-size"); e != nil {
			return nil, e
		}
	case "fontWeight":
		if e := adjust("font-weight"); e != nil {
			return nil, e
		}
	case "textAlign":
		out["text-align"] = in.Value
	case "center":
		out["display"] = "flex"
		out["align-items"] = "center"
		out["justify-content"] = "center"
		out["text-align"] = "center"
	case "columns":
		out["display"] = "grid"
		out["grid-template-columns"] = "repeat(" + in.Value + ", minmax(0, 1fr))"
		out["min-width"] = "0"
	case "axis":
		out["display"] = "flex"
		out["flex-direction"] = in.Value
		out["flex-wrap"] = "wrap"
	case "width":
		out["width"] = "100%"
		out["max-width"] = map[string]string{"narrow": "36rem", "normal": "64rem", "wide": "80rem", "full": "100%"}[in.Value]
		out["box-sizing"] = "border-box"
	case "radius":
		out["border-radius"] = map[string]string{"sharp": "2px", "medium": "10px", "round": "24px"}[in.Value]
	case "background":
		out["background-color"] = map[string]string{"stronger": "#164e63", "softer": "#e8f1f3", "transparent": "transparent"}[in.Value]
		if in.Value == "stronger" {
			out["color"] = "#ffffff"
		} else if in.Value == "softer" {
			out["color"] = "#16323b"
		}
	case "border":
		out["border"] = map[string]string{"show": "1px solid #64748b", "hide": "0 solid transparent"}[in.Value]
	case "emphasis":
		out["opacity"] = "1"
		out["box-shadow"] = "none"
		out["font-weight"] = "500"
		out["border"] = "1px solid #94a3b8"
		out["background-color"] = "#f1f5f9"
		out["color"] = "#334155"
		if in.Value == "primary" {
			out["background-color"] = "#0e5268"
			out["color"] = "#ffffff"
			out["font-weight"] = "700"
			out["border"] = "1px solid #0e5268"
			out["box-shadow"] = "0 3px 0 #083344"
		}
		if in.Value == "muted" {
			out["background-color"] = "#e2e8f0"
			out["color"] = "#475569"
			out["border"] = "1px solid #cbd5e1"
		}
	case "visibility":
		if in.Value == "hidden" {
			out["display"] = "none"
		}
	}
	for p := range out {
		if !contains(registry[in.Operation].Properties, p) {
			return nil, unsupported
		}
	}
	return out, nil
}
func makePatch(before []byte, change string, intents []EditIntent, targets []Target) ([]byte, error) {
	css := string(before)
	for _, in := range intents {
		for _, t := range targets {
			op, ok := registry[in.Operation]
			if !ok {
				return nil, unsupported
			}
			props, err := op.Generate(in, t)
			if err != nil {
				return nil, err
			}
			if in.Operation == "visibility" {
				pattern := `(?s)/\* jev-edit:[a-zA-Z0-9_-]+ target:` + regexp.QuoteMeta(t.ID) + ` operation:visibility breakpoint:` + in.Breakpoint + ` \*/.*?/\* /jev-edit \*/\n?`
				css = regexp.MustCompile(pattern).ReplaceAllString(css, "")
			}
			if len(props) == 0 {
				continue
			}
			var b strings.Builder
			fmt.Fprintf(&b, "\n/* jev-edit:%s target:%s operation:%s breakpoint:%s */\n", change, t.ID, in.Operation, in.Breakpoint)
			if in.Breakpoint != "all" {
				query := "min-width: 768px"
				if in.Breakpoint == "mobile" {
					query = "max-width: 767.98px"
				}
				fmt.Fprintf(&b, "@media (%s) {\n", query)
			}
			fmt.Fprintf(&b, "[data-jev-id=\"%s\"] {\n", t.ID)
			keys := make([]string, 0, len(props))
			for k := range props {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Fprintf(&b, "  %s: %s !important;\n", k, props[k])
			}
			b.WriteString("}\n")
			if in.Breakpoint != "all" {
				b.WriteString("}\n")
			}
			b.WriteString("/* /jev-edit */\n")
			css += b.String()
		}
	}
	return []byte(css), nil
}
