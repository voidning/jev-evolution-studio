package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultGeneratorModel = "gpt-5.6-luna"

func GenerateBlueprints(prompt string, specs []PageSpec) ([]PageBlueprint, string, error) {
	key := loadGeneratorKey()
	if key == "" {
		return nil, "local-grammar", errors.New("OPENAI_API_KEY or GENERATOR_API_KEY is not configured")
	}
	model := strings.TrimSpace(os.Getenv("GENERATOR_MODEL"))
	if model == "" {
		model = defaultGeneratorModel
	}
	endpoint := strings.TrimSpace(os.Getenv("GENERATOR_BASE_URL"))
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/responses"
	}
	constitutions := []map[string]string{
		{"id": "signal", "purpose": "information clarity", "rule": "Make the product understandable through an unexpected but highly usable composition."},
		{"id": "pulse", "purpose": "immersive emotion", "rule": "Build a cinematic interactive journey; avoid conventional SaaS section rhythms."},
		{"id": "atlas", "purpose": "system credibility", "rule": "Turn evidence, operations, and technical reality into a distinctive visual world."},
	}
	state := map[string]any{"request": prompt, "constitutions": constitutions, "semantic_seed": specs}
	stateJSON, _ := json.Marshal(state)
	instructions := `You are a generative web art director. Create exactly three meaningfully different page blueprints from the request and constitutions. Return a compact page AST, not HTML, CSS, Tailwind classes, or implementation code. Each blueprint must contain 5-8 sections, start with a hero, end with a CTA, use a different narrative order, and include at least one unusual interactive or visual section. Use concise, product-specific copy. Children create nested compositions. Never repeat the same section sequence across blueprints.`
	payload := map[string]any{
		"model":             model,
		"instructions":      instructions,
		"input":             string(stateJSON),
		"reasoning":         map[string]any{"effort": "none"},
		"max_output_tokens": 5200,
		"text":              map[string]any{"format": map[string]any{"type": "json_schema", "name": "page_blueprints", "strict": true, "schema": blueprintSchema()}},
	}
	body, _ := json.Marshal(payload)
	ctx, cancel := context.WithTimeout(context.Background(), 14*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, model, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, model, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 768))
		return nil, model, fmt.Errorf("generator returned %s: %s", resp.Status, strings.TrimSpace(string(limited)))
	}
	var response struct {
		OutputText string `json:"output_text"`
		Output     []struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, model, err
	}
	text := response.OutputText
	if text == "" {
		for _, output := range response.Output {
			for _, content := range output.Content {
				if content.Type == "output_text" && content.Text != "" {
					text = content.Text
					break
				}
			}
		}
	}
	if text == "" {
		return nil, model, errors.New("generator returned no output text")
	}
	var decoded struct {
		Blueprints []PageBlueprint `json:"blueprints"`
	}
	if err := json.Unmarshal([]byte(text), &decoded); err != nil {
		return nil, model, fmt.Errorf("decode page blueprints: %w", err)
	}
	if len(decoded.Blueprints) != 3 {
		return nil, model, fmt.Errorf("generator returned %d blueprints, expected 3", len(decoded.Blueprints))
	}
	return normalizeBlueprints(decoded.Blueprints, specs), model, nil
}

func blueprintSchema() map[string]any {
	item := map[string]any{
		"type": "object", "additionalProperties": false,
		"properties": map[string]any{
			"label": map[string]any{"type": "string"}, "title": map[string]any{"type": "string"},
			"body": map[string]any{"type": "string"}, "value": map[string]any{"type": "string"},
		},
		"required": []string{"label", "title", "body", "value"},
	}
	node := map[string]any{
		"type": "object", "additionalProperties": false,
		"properties": map[string]any{
			"id":      map[string]any{"type": "string"},
			"kind":    map[string]any{"type": "string", "enum": []string{"hero", "metrics", "manifesto", "features", "timeline", "gallery", "terminal", "quote", "cta", "cluster"}},
			"layout":  map[string]any{"type": "string", "enum": []string{"split", "centered", "asymmetric", "fullbleed", "grid", "mosaic", "editorial", "horizontal", "sticky", "orbit", "console"}},
			"visual":  map[string]any{"type": "string", "enum": []string{"dashboard", "waveform", "constellation", "particles", "specimens", "telemetry", "code", "portal", "typography", "none"}},
			"eyebrow": map[string]any{"type": "string"}, "headline": map[string]any{"type": "string"}, "body": map[string]any{"type": "string"},
			"items":    map[string]any{"type": "array", "maxItems": 6, "items": item},
			"children": map[string]any{"type": "array", "maxItems": 4, "items": map[string]any{"$ref": "#/$defs/node"}},
		},
		"required": []string{"id", "kind", "layout", "visual", "eyebrow", "headline", "body", "items", "children"},
	}
	return map[string]any{
		"type": "object", "additionalProperties": false,
		"properties": map[string]any{"blueprints": map[string]any{
			"type": "array", "minItems": 3, "maxItems": 3,
			"items": map[string]any{
				"type": "object", "additionalProperties": false,
				"properties": map[string]any{
					"version":           map[string]any{"type": "integer", "enum": []int{2}},
					"id":                map[string]any{"type": "string", "enum": []string{"signal", "pulse", "atlas"}},
					"creativeDirection": map[string]any{"type": "string"},
					"sections":          map[string]any{"type": "array", "minItems": 5, "maxItems": 8, "items": map[string]any{"$ref": "#/$defs/node"}},
				},
				"required": []string{"version", "id", "creativeDirection", "sections"},
			},
		}},
		"required": []string{"blueprints"}, "$defs": map[string]any{"node": node},
	}
}

func loadGeneratorKey() string {
	for _, name := range []string{"GENERATOR_API_KEY", "OPENAI_API_KEY"} {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	paths := apiKeyPaths("")
	if executable, err := os.Executable(); err == nil {
		paths = apiKeyPaths(executable)
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			parts := strings.SplitN(strings.TrimSpace(line), "=", 2)
			if len(parts) == 2 && (parts[0] == "GENERATOR_API_KEY" || parts[0] == "OPENAI_API_KEY") {
				return strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			}
		}
	}
	return ""
}

func normalizeBlueprints(input []PageBlueprint, specs []PageSpec) []PageBlueprint {
	fallbacks := fallbackBlueprints(specs)
	byID := map[string]PageBlueprint{}
	for _, blueprint := range input {
		if blueprint.ID == "signal" || blueprint.ID == "pulse" || blueprint.ID == "atlas" {
			blueprint.Version = 2
			blueprint.Sections = sanitizeSections(blueprint.Sections, 0)
			if len(blueprint.Sections) >= 3 {
				byID[blueprint.ID] = blueprint
			}
		}
	}
	result := make([]PageBlueprint, 0, 3)
	for _, fallback := range fallbacks {
		if generated, ok := byID[fallback.ID]; ok {
			result = append(result, generated)
		} else {
			result = append(result, fallback)
		}
	}
	return result
}

func sanitizeSections(nodes []SectionNode, depth int) []SectionNode {
	if len(nodes) > 9 {
		nodes = nodes[:9]
	}
	allowedKind := set("hero", "metrics", "manifesto", "features", "timeline", "gallery", "terminal", "quote", "cta", "cluster")
	allowedLayout := set("split", "centered", "asymmetric", "fullbleed", "grid", "mosaic", "editorial", "horizontal", "sticky", "orbit", "console")
	allowedVisual := set("dashboard", "waveform", "constellation", "particles", "specimens", "telemetry", "code", "portal", "typography", "none")
	for index := range nodes {
		node := &nodes[index]
		if !allowedKind[node.Kind] {
			node.Kind = "features"
		}
		if !allowedLayout[node.Layout] {
			node.Layout = "grid"
		}
		if !allowedVisual[node.Visual] {
			node.Visual = "none"
		}
		if len(node.Items) > 6 {
			node.Items = node.Items[:6]
		}
		if depth >= 1 {
			node.Children = nil
		} else {
			node.Children = sanitizeSections(node.Children, depth+1)
		}
	}
	return nodes
}

func set(values ...string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func applyBlueprints(result DesignResult, blueprints []PageBlueprint, source, generator string) DesignResult {
	blueprints = normalizeBlueprints(blueprints, result.Specs)
	for index := range result.Specs {
		result.Specs[index].Blueprint = blueprints[index]
		result.Specs[index].Descriptor = blueprints[index].CreativeDirection
	}
	result.ASTSource = source
	result.Generator = generator
	return result
}

func fallbackBlueprints(specs []PageSpec) []PageBlueprint {
	result := make([]PageBlueprint, 0, len(specs))
	for index, spec := range specs {
		items := make([]BlueprintItem, 0, len(spec.FeaturesContent))
		for _, feature := range spec.FeaturesContent {
			items = append(items, BlueprintItem{Label: feature.Kicker, Title: feature.Title, Body: feature.Body})
		}
		metrics := make([]BlueprintItem, 0, len(spec.Metrics))
		for _, metric := range spec.Metrics {
			metrics = append(metrics, BlueprintItem{Label: metric.Label, Value: metric.Value + metric.Unit})
		}
		var direction string
		var sections []SectionNode
		switch index {
		case 1:
			direction = "Immersive living narrative"
			sections = []SectionNode{
				{ID: "arrival", Kind: "hero", Layout: "fullbleed", Visual: "portal", Eyebrow: spec.Eyebrow, Headline: spec.Title, Body: spec.Description},
				{ID: "signals", Kind: "gallery", Layout: "orbit", Visual: "particles", Eyebrow: "SIGNALS EMERGING", Headline: spec.SectionTitle, Body: "Move through a living field of evidence.", Items: items},
				{ID: "belief", Kind: "manifesto", Layout: "editorial", Visual: "typography", Eyebrow: "A DIFFERENT POSSIBILITY", Headline: "The interface should feel like entering the subject itself.", Body: spec.Description},
				{ID: "journey", Kind: "timeline", Layout: "sticky", Visual: "waveform", Eyebrow: "THE JOURNEY", Headline: "One continuous descent into understanding.", Items: items},
				{ID: "proof", Kind: "metrics", Layout: "horizontal", Visual: "telemetry", Eyebrow: "LIVE EVIDENCE", Headline: "The world is already moving.", Items: metrics},
				{ID: "join", Kind: "cta", Layout: "centered", Visual: "constellation", Eyebrow: "ENTER THE NEXT CHAPTER", Headline: "Come closer.", Body: "Join the people building what comes next."},
			}
		case 2:
			direction = "Operational evidence system"
			sections = []SectionNode{
				{ID: "command", Kind: "hero", Layout: "console", Visual: "code", Eyebrow: spec.Eyebrow, Headline: spec.Title, Body: spec.Description},
				{ID: "runtime", Kind: "terminal", Layout: "split", Visual: "telemetry", Eyebrow: "SYSTEM ONLINE", Headline: "Reality, instrumented.", Body: "Every signal enters an observable chain of evidence.", Items: metrics},
				{ID: "evidence", Kind: "cluster", Layout: "asymmetric", Visual: "dashboard", Eyebrow: "EVIDENCE GRAPH", Headline: spec.SectionTitle, Children: []SectionNode{{ID: "capabilities", Kind: "features", Layout: "grid", Visual: "none", Items: items}, {ID: "telemetry", Kind: "metrics", Layout: "horizontal", Visual: "waveform", Items: metrics}}},
				{ID: "sequence", Kind: "timeline", Layout: "horizontal", Visual: "code", Eyebrow: "VERIFIED SEQUENCE", Headline: "From first observation to confident action.", Items: items},
				{ID: "commit", Kind: "cta", Layout: "split", Visual: "dashboard", Eyebrow: "READY WHEN YOU ARE", Headline: "Put the system to work.", Body: "Start with a live, inspectable demonstration."},
			}
		default:
			direction = "Editorial intelligence landscape"
			sections = []SectionNode{
				{ID: "hero", Kind: "hero", Layout: "asymmetric", Visual: "dashboard", Eyebrow: spec.Eyebrow, Headline: spec.Title, Body: spec.Description},
				{ID: "pulse", Kind: "metrics", Layout: "horizontal", Visual: "waveform", Eyebrow: "LIVE PULSE", Headline: "A system you can read at a glance.", Items: metrics},
				{ID: "thesis", Kind: "manifesto", Layout: "editorial", Visual: "typography", Eyebrow: "THE CENTRAL IDEA", Headline: spec.SectionTitle, Body: spec.Description},
				{ID: "capabilities", Kind: "features", Layout: "mosaic", Visual: "specimens", Eyebrow: "CAPABILITIES", Headline: "Different tools. One coherent intelligence.", Items: items},
				{ID: "flow", Kind: "timeline", Layout: "split", Visual: "constellation", Eyebrow: "HOW IT MOVES", Headline: "From signal to outcome without the dead space.", Items: items},
				{ID: "begin", Kind: "cta", Layout: "asymmetric", Visual: "portal", Eyebrow: "BEGIN", Headline: "See the whole system live.", Body: "Bring one real question. Leave with a new direction."},
			}
		}
		result = append(result, PageBlueprint{Version: 2, ID: strings.Split(spec.ID, "-")[0], CreativeDirection: direction, Sections: sections})
	}
	return result
}

func evolveBlueprint(blueprint PageBlueprint, weakness string, generation int, spec PageSpec) PageBlueprint {
	if len(blueprint.Sections) == 0 {
		fallback := fallbackBlueprints([]PageSpec{spec})
		if len(fallback) > 0 {
			blueprint = fallback[0]
		}
	}
	blueprint.Version = 2
	blueprint.ID = strings.Split(spec.ID, "-")[0]
	blueprint.CreativeDirection = fmt.Sprintf("%s · generation %02d %s mutation", blueprint.CreativeDirection, generation, weakness)
	visuals := []string{"particles", "waveform", "constellation", "specimens", "telemetry", "portal"}
	switch weakness {
	case "originality":
		index := generation % len(visuals)
		insertAt := len(blueprint.Sections) - 1
		node := SectionNode{ID: fmt.Sprintf("artifact-g%d", generation), Kind: "gallery", Layout: "orbit", Visual: visuals[index], Eyebrow: "NEW ARTIFACT / EVOLVED", Headline: "A new way to encounter the signal.", Body: "This scene was introduced because the previous generation felt too familiar.", Items: []BlueprintItem{{Label: "GENE 01", Title: "Living evidence", Body: "The interface responds as a spatial field."}, {Label: "GENE 02", Title: "Unexpected scale", Body: "Small signals become the dominant visual event."}}}
		blueprint.Sections = append(blueprint.Sections, SectionNode{})
		copy(blueprint.Sections[insertAt+1:], blueprint.Sections[insertAt:])
		blueprint.Sections[insertAt] = node
	case "clarity":
		for index := range blueprint.Sections {
			if blueprint.Sections[index].Kind == "hero" {
				blueprint.Sections[index].Layout = "split"
				blueprint.Sections[index].Visual = "dashboard"
				break
			}
		}
	case "trust":
		insertAt := len(blueprint.Sections) - 1
		node := SectionNode{ID: fmt.Sprintf("proof-g%d", generation), Kind: "terminal", Layout: "console", Visual: "telemetry", Eyebrow: "VERIFIED IN THIS GENERATION", Headline: "Proof you can inspect.", Body: "Every claim is connected to an observable system state.", Items: []BlueprintItem{{Label: "STATUS", Title: "Traceable", Value: "100%"}, {Label: "RUNTIME", Title: "Observed", Value: "LIVE"}}}
		blueprint.Sections = append(blueprint.Sections, SectionNode{})
		copy(blueprint.Sections[insertAt+1:], blueprint.Sections[insertAt:])
		blueprint.Sections[insertAt] = node
	case "conversion":
		for index := len(blueprint.Sections) - 1; index >= 0; index-- {
			if blueprint.Sections[index].Kind == "cta" {
				blueprint.Sections[index].Layout = "fullbleed"
				blueprint.Sections[index].Visual = "portal"
				blueprint.Sections[index].Headline = "Enter the system now."
				break
			}
		}
	}
	blueprint.Sections = sanitizeSections(blueprint.Sections, 0)
	return blueprint
}
