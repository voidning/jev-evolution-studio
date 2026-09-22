package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Fixed Tailwind 4 palette; unknown colors are refused on the fuzzy online path.
// Exact user-specified colors retain their requested value and other classes.
//
//go:embed integration/colors.json
var paletteJSON []byte
var colorNumbers = regexp.MustCompile(`[-+]?(?:\d*\.)?\d+`)

func luminance(value string) (float64, bool) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "white" {
		value = "#ffffff"
	}
	if value == "black" {
		value = "#000000"
	}
	var rgb [3]float64
	linear := false
	if strings.HasPrefix(value, "#") {
		h := value[1:]
		if len(h) == 3 {
			h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
		}
		if len(h) != 6 {
			return 0, false
		}
		for i := range rgb {
			n, e := strconv.ParseUint(h[i*2:i*2+2], 16, 8)
			if e != nil {
				return 0, false
			}
			rgb[i] = float64(n) / 255
		}
	} else {
		parts := colorNumbers.FindAllString(value, -1)
		if len(parts) != 3 {
			return 0, false
		}
		v := [3]float64{}
		for i := range v {
			v[i], _ = strconv.ParseFloat(parts[i], 64)
		}
		switch {
		case strings.HasPrefix(value, "rgb(") && !strings.Contains(value, "%"):
			for i := range rgb {
				rgb[i] = v[i] / 255
			}
		case strings.HasPrefix(value, "oklch("):
			if strings.Contains(value, "%") {
				v[0] /= 100
			}
			a, b := v[1]*math.Cos(v[2]*math.Pi/180), v[1]*math.Sin(v[2]*math.Pi/180)
			l := math.Pow(v[0]+.3963377774*a+.2158037573*b, 3)
			m := math.Pow(v[0]-.1055613458*a-.0638541728*b, 3)
			s := math.Pow(v[0]-.0894841775*a-1.291485548*b, 3)
			rgb = [3]float64{4.0767416621*l - 3.3077115913*m + .2309699292*s, -1.2684380046*l + 2.6097574011*m - .3413193965*s, -.0041960863*l - .7034186147*m + 1.707614701*s}
			linear = true
		default:
			return 0, false
		}
	}
	for i, c := range rgb {
		c = math.Max(0, math.Min(1, c))
		if !linear {
			if c <= .04045 {
				c /= 12.92
			} else {
				c = math.Pow((c+.055)/1.055, 2.4)
			}
		}
		rgb[i] = c
	}
	return .2126*rgb[0] + .7152*rgb[1] + .0722*rgb[2], true
}
func checkSemanticContrast(c ColorValue, channel string, t Target, m ProjectMeta) error {
	if channel == "border" {
		return nil
	}
	var palette map[string]string
	_ = json.Unmarshal(paletteJSON, &palette)
	value := c.Value
	if c.Kind == "token" {
		value = m.Tokens[value]
	} else if !strings.HasPrefix(value, "#") {
		key := value
		if c.Shade != "" {
			key += "-" + c.Shade
		}
		value = palette[key]
	}
	against := t.Styles["color"]
	if channel == "text" {
		against = t.Styles["background-color"]
	}
	a, ok := luminance(value)
	b, other := luminance(against)
	if !ok || !other {
		return fmt.Errorf("不支持：无法验证模糊颜色的文字对比度，请明确指定颜色；透明背景或变量色需要项目提供确定颜色")
	}
	ratio := (math.Max(a, b) + .05) / (math.Min(a, b) + .05)
	if ratio < 4.5 {
		return fmt.Errorf("不支持：该模糊颜色与现有文字/背景的对比度 %.2f:1 低于 4.5:1；请明确指定颜色或先调整文字色", ratio)
	}
	return nil
}
