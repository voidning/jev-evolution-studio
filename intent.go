package main

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var unsupported = errors.New("不支持：请使用已支持的视觉或布局命令，不会猜测性修改代码")

func intent(op, direction, value, bp, scope string) EditIntent {
	return EditIntent{Scope: scope, Family: registry[op].Family, Operation: op, Direction: direction, Magnitude: "subtle", Breakpoint: bp, Value: value, Preserve: []string{"content", "responsive"}, Confidence: 1}
}

var responsiveRE = regexp.MustCompile(`^(?:卡片)?(?:在)?桌面端(?:改成|设为|设置为)?([一二三四1-4])列[，,、 ]*(?:手机端|移动端)(?:保持|改成|设为|设置为)?([一二三四1-4])列$`)
var columnsRE = regexp.MustCompile(`^(?:卡片)?(?:改成|设为|设置为)?([一二三四1-4])列$`)

func column(s string) string {
	for i, v := range []string{"一", "二", "三", "四"} {
		if s == v {
			return strconv.Itoa(i + 1)
		}
	}
	return s
}
func LocalIntent(prompt string) ([]EditIntent, error) {
	p := strings.TrimSpace(strings.ToLower(prompt))
	p = strings.TrimRight(p, "。！!.")
	scope := "selected"
	for _, pair := range [][2]string{{"同级元素", "siblings"}, {"所有同级", "siblings"}, {"父容器", "container"}} {
		if strings.HasPrefix(p, pair[0]) {
			scope = pair[1]
			p = strings.TrimPrefix(p, pair[0])
			break
		}
	}
	if m := responsiveRE.FindStringSubmatch(p); m != nil {
		return []EditIntent{intent("columns", "set", column(m[1]), "desktop", scope), intent("columns", "set", column(m[2]), "mobile", scope)}, nil
	}
	bp := "all"
	for _, pair := range [][2]string{{"桌面端", "desktop"}, {"手机端", "mobile"}, {"移动端", "mobile"}} {
		if strings.HasPrefix(p, pair[0]) {
			bp = pair[1]
			p = strings.TrimPrefix(p, pair[0])
			break
		}
	}
	if m := columnsRE.FindStringSubmatch(p); m != nil {
		return []EditIntent{intent("columns", "set", column(m[1]), bp, scope)}, nil
	}
	// Anchored phrases: never partially apply a recognized substring of a larger request.
	rules := []struct{ pattern, op, dir, value string }{
		{`^(?:这里|这块|这个容器)?(?:再|更)?紧凑(?:一点|一些)?$`, "density", "decrease", ""},
		{`^(?:这里|这块|这个容器)?(?:再|更)?宽松(?:一点|一些)?$`, "density", "increase", ""},
		{`^(?:padding|内边距)(?:增加|增大|大一点)$`, "padding", "increase", ""},
		{`^(?:padding|内边距)(?:减少|减小|小一点)$`, "padding", "decrease", ""},
		{`^(?:gap|间距)(?:增加|增大|大一点)$`, "gap", "increase", ""},
		{`^(?:gap|间距)(?:减少|减小|小一点)$`, "gap", "decrease", ""},
		{`^(?:(?:这个|这段)?(?:标题|文字|文本))?(?:再|更|变)?大(?:一点|一些)?$`, "fontSize", "increase", ""},
		{`^(?:(?:这个|这段)?(?:标题|文字|文本))?(?:再|更|变)?小(?:一点|一些)?$`, "fontSize", "decrease", ""},
		{`^(?:字重增强|字重增加|文字加粗|加粗|字重更强)$`, "fontWeight", "increase", ""},
		{`^(?:字重减弱|字重减少|文字变细|变细)$`, "fontWeight", "decrease", ""},
		{`^(?:文字|文本)?(?:左对齐|居左)$`, "textAlign", "set", "left"},
		{`^(?:文字|文本)?居中$`, "textAlign", "set", "center"},
		{`^(?:文字|文本)?(?:右对齐|居右)$`, "textAlign", "set", "right"},
		{`^(?:把这块内容居中|内容居中|容器内容居中)$`, "center", "set", "center"},
		{`^(?:改成|改为)?横向排列$`, "axis", "set", "row"},
		{`^(?:改成|改为)?纵向排列$`, "axis", "set", "column"},
		{`^(?:宽度)?(?:改成|设为)?(?:narrow|窄)$`, "width", "set", "narrow"},
		{`^宽度(?:改成|设为)?(?:normal|正常)$`, "width", "set", "normal"},
		{`^(?:宽度)?(?:改成|设为)?(?:wide|宽)$`, "width", "set", "wide"},
		{`^(?:宽度)?(?:改成|设为)?(?:full|全宽)$`, "width", "set", "full"},
		{`^(?:圆角小一点|圆角sharp|直角)$`, "radius", "set", "sharp"},
		{`^(?:圆角适中|圆角medium)$`, "radius", "set", "medium"},
		{`^(?:圆角大一点|圆角round|更圆润)$`, "radius", "set", "round"},
		{`^(?:背景更强|背景强调stronger|背景强调增强)$`, "background", "set", "stronger"},
		{`^(?:背景更柔和|背景强调softer|背景强调减弱)$`, "background", "set", "softer"},
		{`^(?:透明背景|背景透明|背景transparent)$`, "background", "set", "transparent"},
		{`^(?:显示边框|添加边框)$`, "border", "set", "show"},
		{`^(?:隐藏边框|移除边框)$`, "border", "set", "hide"},
		{`^(?:让这个按钮更突出|按钮更突出|元素更突出|强调primary)$`, "emphasis", "set", "primary"},
		{`^(?:普通强调|强调normal|恢复普通样式)$`, "emphasis", "set", "normal"},
		{`^(?:弱化元素|强调muted|弱化按钮)$`, "emphasis", "set", "muted"},
		{`^(?:隐藏|隐藏这个元素|隐藏元素)$`, "visibility", "set", "hidden"},
		{`^(?:恢复|恢复这个元素|恢复元素|显示元素)$`, "visibility", "set", "visible"},
	}
	for _, r := range rules {
		if regexp.MustCompile(r.pattern).MatchString(p) {
			return []EditIntent{intent(r.op, r.dir, r.value, bp, scope)}, nil
		}
	}
	return nil, unsupported
}
