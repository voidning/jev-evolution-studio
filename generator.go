package main

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
)

const grammarCandidatesPerUniverse = 48

// GenerateBlueprints is a deterministic search over a bounded page grammar.
// Jev supplies semantic choices; code owns every candidate, constraint and mutation.
func GenerateBlueprints(prompt string, specs []PageSpec) []PageBlueprint {
	seeds := fallbackBlueprints(specs)
	result := make([]PageBlueprint, 0, len(seeds))
	for index, base := range seeds {
		candidates := make([]scoredBlueprint, 0, grammarCandidatesPerUniverse)
		for variant := 0; variant < grammarCandidatesPerUniverse; variant++ {
			candidate := varyBlueprint(base, specs[index], promptSeed(prompt)+uint64(index*997+variant*37), variant)
			candidates = append(candidates, scoredBlueprint{Blueprint: candidate, Score: grammarFitness(candidate, specs[index], variant)})
		}
		sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].Score > candidates[j].Score })
		result = append(result, candidates[0].Blueprint)
	}
	return normalizeBlueprints(result, specs)
}

type scoredBlueprint struct {
	Blueprint PageBlueprint
	Score     int
}

func promptSeed(prompt string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(strings.ToLower(strings.TrimSpace(prompt))))
	return h.Sum64()
}

func varyBlueprint(base PageBlueprint, spec PageSpec, seed uint64, variant int) PageBlueprint {
	blueprint := base
	blueprint.Sections = append([]SectionNode(nil), base.Sections...)
	layouts := map[string][]string{
		"hero": {"split", "centered", "asymmetric", "fullbleed", "console"}, "metrics": {"horizontal", "grid", "mosaic"},
		"manifesto": {"editorial", "centered", "fullbleed"}, "features": {"grid", "mosaic", "asymmetric"},
		"timeline": {"sticky", "horizontal", "split"}, "gallery": {"orbit", "mosaic", "fullbleed"},
		"terminal": {"console", "split", "sticky"}, "cluster": {"asymmetric", "grid", "mosaic"}, "cta": {"centered", "split", "fullbleed"},
	}
	visuals := map[string][]string{
		"hero": {"dashboard", "portal", "constellation", "code"}, "metrics": {"waveform", "telemetry", "dashboard"},
		"manifesto": {"typography", "particles", "constellation"}, "features": {"specimens", "dashboard", "none"},
		"timeline": {"waveform", "code", "telemetry"}, "gallery": {"particles", "constellation", "specimens"},
		"terminal": {"code", "telemetry", "dashboard"}, "cluster": {"dashboard", "constellation", "specimens"}, "cta": {"portal", "constellation", "typography"},
	}
	for i := range blueprint.Sections {
		node := &blueprint.Sections[i]
		gene := int(seed>>uint((i%8)*8)) + variant*11 + i*17
		if options := layouts[node.Kind]; len(options) > 0 {
			node.Layout = options[positiveMod(gene, len(options))]
		}
		if options := visuals[node.Kind]; len(options) > 0 {
			node.Visual = options[positiveMod(gene/3+variant, len(options))]
		}
	}
	if len(blueprint.Sections) > 4 {
		middle := append([]SectionNode(nil), blueprint.Sections[1:len(blueprint.Sections)-1]...)
		rotation := positiveMod(int(seed)+variant, len(middle))
		middle = append(middle[rotation:], middle[:rotation]...)
		copy(blueprint.Sections[1:len(blueprint.Sections)-1], middle)
	}
	if (variant+int(seed%7))%3 == 0 && len(blueprint.Sections) < 8 {
		artifact := grammarArtifact(spec, variant, seed)
		insertAt := 1 + positiveMod(variant+int(seed%5), len(blueprint.Sections)-1)
		blueprint.Sections = append(blueprint.Sections, SectionNode{})
		copy(blueprint.Sections[insertAt+1:], blueprint.Sections[insertAt:])
		blueprint.Sections[insertAt] = artifact
	}
	blueprint.CreativeDirection = fmt.Sprintf("%s · procedural genome %02d", base.CreativeDirection, variant+1)
	return blueprint
}

func grammarArtifact(spec PageSpec, variant int, seed uint64) SectionNode {
	kinds := []string{"gallery", "manifesto", "timeline", "cluster", "terminal"}
	kind := kinds[positiveMod(variant+int(seed%11), len(kinds))]
	items := []BlueprintItem{}
	for _, feature := range spec.FeaturesContent {
		items = append(items, BlueprintItem{Label: feature.Kicker, Title: feature.Title, Body: feature.Body})
	}
	return SectionNode{ID: fmt.Sprintf("artifact-%d", variant), Kind: kind, Layout: "mosaic", Visual: "particles", Eyebrow: "GENERATIVE ARTIFACT", Headline: spec.SectionTitle, Body: spec.Description, Items: items}
}

func grammarFitness(blueprint PageBlueprint, spec PageSpec, variant int) int {
	score := 100 + len(blueprint.Sections)*4
	seen := map[string]bool{}
	for _, node := range blueprint.Sections {
		signature := node.Kind + "/" + node.Layout + "/" + node.Visual
		if !seen[signature] {
			score += 7
		}
		seen[signature] = true
		if spec.Strategy == "impact" && (node.Layout == "fullbleed" || node.Layout == "orbit") {
			score += 5
		}
		if spec.Strategy == "trust" && (node.Visual == "telemetry" || node.Visual == "code") {
			score += 5
		}
		if spec.Strategy == "clarity" && (node.Layout == "split" || node.Layout == "grid") {
			score += 5
		}
	}
	return score - variant/12
}

func positiveMod(value, divisor int) int {
	value %= divisor
	if value < 0 {
		value += divisor
	}
	return value
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
