package main

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
)

const (
	grammarCandidatesPerUniverse = 2048
	grammarDesignSpace           = "10^14+"
)

// GenerateBlueprints is a deterministic search over a bounded page grammar.
// Jev supplies semantic choices; code owns every candidate, constraint and mutation.
func GenerateBlueprints(prompt string, specs []PageSpec) []PageBlueprint {
	result := make([]PageBlueprint, 0, len(specs))
	for index, spec := range specs {
		candidates := make([]scoredBlueprint, 0, grammarCandidatesPerUniverse)
		for variant := 0; variant < grammarCandidatesPerUniverse; variant++ {
			seed := promptSeed(prompt) ^ uint64((index+1)*104729) ^ uint64((variant+1)*13007)
			candidate := composeBlueprint(spec, seed, variant)
			candidates = append(candidates, scoredBlueprint{Blueprint: candidate, Score: grammarFitness(candidate, spec, variant)})
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

type grammarRNG struct{ state uint64 }

func newGrammarRNG(seed uint64) *grammarRNG {
	if seed == 0 {
		seed = 0x9e3779b97f4a7c15
	}
	return &grammarRNG{state: seed}
}

func (r *grammarRNG) next() uint64 {
	r.state ^= r.state << 13
	r.state ^= r.state >> 7
	r.state ^= r.state << 17
	return r.state
}

func (r *grammarRNG) pick(values []string) string { return values[int(r.next()%uint64(len(values)))] }
func (r *grammarRNG) chance(percent uint64) bool  { return r.next()%100 < percent }

var grammarLayouts = map[string][]string{
	"hero": {"split", "centered", "asymmetric", "fullbleed", "console"}, "metrics": {"horizontal", "grid", "mosaic"},
	"manifesto": {"editorial", "centered", "fullbleed"}, "features": {"grid", "mosaic", "asymmetric", "split"},
	"timeline": {"sticky", "horizontal", "split"}, "gallery": {"orbit", "mosaic", "fullbleed", "asymmetric"},
	"terminal": {"console", "split", "sticky"}, "quote": {"centered", "editorial", "fullbleed"},
	"cluster": {"asymmetric", "grid", "mosaic", "split"}, "cta": {"centered", "split", "fullbleed", "asymmetric"},
}

var grammarVisuals = map[string][]string{
	"hero": {"dashboard", "portal", "constellation", "code", "particles"}, "metrics": {"waveform", "telemetry", "dashboard"},
	"manifesto": {"typography", "particles", "constellation", "none"}, "features": {"specimens", "dashboard", "constellation", "none"},
	"timeline": {"waveform", "code", "telemetry", "constellation"}, "gallery": {"particles", "constellation", "specimens", "portal"},
	"terminal": {"code", "telemetry", "dashboard"}, "quote": {"typography", "particles", "none"},
	"cluster": {"dashboard", "constellation", "specimens", "telemetry"}, "cta": {"portal", "constellation", "typography", "dashboard"},
}

func composeBlueprint(spec PageSpec, seed uint64, variant int) PageBlueprint {
	rng := newGrammarRNG(seed)
	sectionCount := 5 + int(rng.next()%4)
	middleCount := sectionCount - 2
	kindPool := []string{"metrics", "manifesto", "features", "timeline", "gallery", "terminal", "quote", "cluster"}
	if spec.ShowStats {
		kindPool = append(kindPool, "metrics", "terminal")
	}
	if spec.Strategy == "impact" {
		kindPool = append(kindPool, "gallery", "manifesto", "quote")
	} else if spec.Strategy == "trust" {
		kindPool = append(kindPool, "terminal", "metrics", "cluster")
	} else {
		kindPool = append(kindPool, "features", "timeline", "metrics")
	}
	sections := make([]SectionNode, 0, sectionCount)
	sections = append(sections, composeNode("hero", spec, rng, variant, 0))
	previous := "hero"
	for position := 0; position < middleCount; position++ {
		kind := rng.pick(kindPool)
		for attempts := 0; kind == previous && attempts < 5; attempts++ {
			kind = rng.pick(kindPool)
		}
		node := composeNode(kind, spec, rng, variant, position+1)
		if kind == "cluster" || (rng.chance(18) && position > 0) {
			childKinds := []string{"features", "metrics", "timeline", "gallery", "quote"}
			childCount := 1 + int(rng.next()%2)
			for child := 0; child < childCount; child++ {
				node.Children = append(node.Children, composeNode(rng.pick(childKinds), spec, rng, variant, 10+position*2+child))
			}
		}
		sections = append(sections, node)
		previous = kind
	}
	sections = append(sections, composeNode("cta", spec, rng, variant, sectionCount-1))
	directions := []string{"Adaptive editorial system", "Immersive signal journey", "Operational evidence field", "Kinetic product narrative", "Spatial intelligence atlas", "Living interface organism"}
	return PageBlueprint{Version: 2, ID: strings.Split(spec.ID, "-")[0], CreativeDirection: rng.pick(directions) + fmt.Sprintf(" · genome %04x", seed&0xffff), Sections: sections}
}

func composeNode(kind string, spec PageSpec, rng *grammarRNG, variant, position int) SectionNode {
	items := featureItems(spec)
	if kind == "metrics" || kind == "terminal" {
		items = metricItems(spec)
	}
	headlines := map[string][]string{
		"metrics":   {"The system is already moving.", "Evidence at the speed of the event.", "Every signal, visible."},
		"manifesto": {spec.SectionTitle, "The interface should feel like the subject itself.", "A new operating rhythm begins here."},
		"features":  {spec.SectionTitle, "Different instruments. One coherent intelligence.", "Built as a system, not a feature list."},
		"timeline":  {"From first signal to confident action.", "Follow the change as it happens.", "One continuous path through the unknown."},
		"gallery":   {"A field of evidence you can enter.", "The invisible becomes an environment.", "Every artifact tells part of the story."},
		"terminal":  {"Reality, instrumented.", "Proof you can inspect.", "The live system underneath the promise."},
		"quote":     {"Build an interface people remember after the screen goes dark.", "Clarity can still feel impossible.", "The next decision is already taking shape."},
		"cluster":   {"A system of systems.", "Signals connect before they become obvious.", spec.SectionTitle},
		"cta":       {"Enter the system now.", "See the whole thing live.", "Start with one real question."},
	}
	eyebrows := map[string][]string{
		"hero": {spec.Eyebrow, "LIVE SYSTEM / 01", "A NEW INTERFACE"}, "metrics": {"LIVE EVIDENCE", "SIGNAL PULSE", "MEASURED NOW"},
		"manifesto": {"THE CENTRAL IDEA", "WHY THIS EXISTS", "A DIFFERENT POSSIBILITY"}, "features": {"CAPABILITIES", "SYSTEM PRIMITIVES", "WHAT CHANGES"},
		"timeline": {"THE SEQUENCE", "HOW IT MOVES", "FROM SIGNAL TO ACTION"}, "gallery": {"ARTIFACT FIELD", "SIGNALS EMERGING", "VISUAL EVIDENCE"},
		"terminal": {"SYSTEM ONLINE", "VERIFIED RUNTIME", "UNDER THE SURFACE"}, "quote": {"POINT OF VIEW", "HUMAN SIGNAL", "THE BELIEF"},
		"cluster": {"COMPOSITE SYSTEM", "EVIDENCE GRAPH", "CONNECTED FIELD"}, "cta": {"BEGIN", "READY WHEN YOU ARE", "NEXT MOVE"},
	}
	node := SectionNode{ID: fmt.Sprintf("%s-%d-%d", kind, position, variant), Kind: kind, Layout: rng.pick(grammarLayouts[kind]), Visual: rng.pick(grammarVisuals[kind]), Eyebrow: rng.pick(eyebrows[kind]), Body: spec.Description, Items: items}
	if kind == "hero" {
		node.Headline = spec.Title
	} else {
		node.Headline = rng.pick(headlines[kind])
	}
	return node
}

func featureItems(spec PageSpec) []BlueprintItem {
	items := make([]BlueprintItem, 0, len(spec.FeaturesContent))
	for _, feature := range spec.FeaturesContent {
		items = append(items, BlueprintItem{Label: feature.Kicker, Title: feature.Title, Body: feature.Body})
	}
	return items
}

func metricItems(spec PageSpec) []BlueprintItem {
	items := make([]BlueprintItem, 0, len(spec.Metrics))
	for _, metric := range spec.Metrics {
		items = append(items, BlueprintItem{Label: metric.Label, Value: metric.Value + metric.Unit})
	}
	return items
}

func grammarFitness(blueprint PageBlueprint, spec PageSpec, variant int) int {
	targetLength := map[string]int{"airy": 5, "balanced": 6, "compact": 8}[spec.Density]
	if targetLength == 0 {
		targetLength = 6
	}
	lengthDistance := len(blueprint.Sections) - targetLength
	if lengthDistance < 0 {
		lengthDistance = -lengthDistance
	}
	score := 200 - lengthDistance*100
	seen, kinds := map[string]bool{}, map[string]bool{}
	for _, node := range blueprint.Sections {
		signature := node.Kind + "/" + node.Layout + "/" + node.Visual
		if !seen[signature] {
			score += 7
		}
		seen[signature] = true
		kinds[node.Kind] = true
		score += len(node.Children) * 4
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
	score += len(kinds) * 6
	if blueprint.Sections[0].Kind == "hero" && blueprint.Sections[len(blueprint.Sections)-1].Kind == "cta" {
		score += 20
	}
	return score - variant/256
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
	result.CandidateCount = grammarCandidatesPerUniverse * len(result.Specs)
	result.DesignSpace = grammarDesignSpace
	return result
}

func fallbackBlueprints(specs []PageSpec) []PageBlueprint {
	result := make([]PageBlueprint, 0, len(specs))
	for index, spec := range specs {
		seed := promptSeed(spec.World+"/"+spec.Strategy+"/fallback") ^ uint64((index+1)*7919)
		result = append(result, composeBlueprint(spec, seed, index))
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
