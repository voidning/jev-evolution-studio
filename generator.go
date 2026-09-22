package main

import (
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"strings"
)

const (
	grammarCandidatesPerUniverse = 2048
	grammarDesignSpace           = "10^18+"
)

type PrimitiveDefinition struct {
	Name            string
	MinChildren     int
	MaxChildren     int
	AllowedChildren map[string]bool
	SemanticRoles   map[string]bool
	ResponsiveRule  string
}

var primitiveRegistry = map[string]PrimitiveDefinition{
	"page":       primitive("page", 2, 10, []string{"cluster", "frame"}, []string{"document"}, "single-flow"),
	"frame":      primitive("frame", 1, 4, []string{"stack", "grid", "overlay", "rail", "cluster", "rule"}, []string{"intro", "context", "mechanism", "evidence", "discovery", "proof", "invitation"}, "fluid-inset"),
	"stack":      primitive("stack", 1, 8, []string{"text", "collection", "artifact", "action", "cluster", "grid", "rail", "rule"}, nil, "vertical-collapse"),
	"grid":       primitive("grid", 2, 8, []string{"text", "collection", "artifact", "action", "cluster", "stack", "rule"}, nil, "columns-collapse"),
	"overlay":    primitive("overlay", 2, 5, []string{"text", "artifact", "action", "cluster", "rule"}, nil, "layer-to-flow"),
	"rail":       primitive("rail", 2, 8, []string{"text", "collection", "artifact", "cluster", "action"}, nil, "horizontal-to-scroll"),
	"cluster":    primitive("cluster", 1, 8, []string{"text", "collection", "artifact", "action", "rule"}, []string{"navigation", "copy", "evidence-group", "action-group"}, "wrap"),
	"text":       primitive("text", 0, 0, nil, []string{"mark", "eyebrow", "claim", "explanation", "quote", "label"}, "fluid-type"),
	"collection": primitive("collection", 0, 0, nil, []string{"features", "metrics", "steps", "specimens", "proof-list"}, "grid-to-stack"),
	"artifact":   primitive("artifact", 0, 0, nil, []string{"signal", "system", "evidence", "atmosphere"}, "aspect-ratio"),
	"action":     primitive("action", 0, 0, nil, []string{"primary", "secondary"}, "full-width-small"),
	"rule":       primitive("rule", 0, 0, nil, []string{"divider", "index"}, "hide-when-tight"),
}

func primitive(name string, min, max int, children, roles []string, responsive string) PrimitiveDefinition {
	return PrimitiveDefinition{Name: name, MinChildren: min, MaxChildren: max, AllowedChildren: stringSet(children...), SemanticRoles: stringSet(roles...), ResponsiveRule: responsive}
}

func stringSet(values ...string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

// GenerateBlueprints searches independently, then chooses winners jointly so
// later universes are rewarded for a different silhouette and tree topology.
func GenerateBlueprints(prompt string, specs []PageSpec) []PageBlueprint {
	pools := make([][]scoredBlueprint, len(specs))
	for index, spec := range specs {
		pool := make([]scoredBlueprint, 0, grammarCandidatesPerUniverse)
		for variant := 0; variant < grammarCandidatesPerUniverse; variant++ {
			seed := promptSeed(prompt) ^ uint64((index+1)*104729) ^ uint64((variant+1)*13007)
			candidate := composeBlueprint(spec, seed, variant)
			pool = append(pool, scoredBlueprint{Blueprint: candidate, Score: grammarFitness(candidate, spec, variant)})
		}
		sort.SliceStable(pool, func(i, j int) bool { return pool[i].Score > pool[j].Score })
		pools[index] = pool
	}

	selected := make([]PageBlueprint, 0, len(specs))
	for index, pool := range pools {
		limit := 128
		if len(pool) < limit {
			limit = len(pool)
		}
		bestIndex, bestScore := 0, math.MinInt
		for candidateIndex := 0; candidateIndex < limit; candidateIndex++ {
			candidate := pool[candidateIndex]
			score := candidate.Score
			for _, previous := range selected {
				score += blueprintNovelty(candidate.Blueprint, previous) * 12
			}
			score -= candidateIndex / (index + 1)
			if score > bestScore {
				bestIndex, bestScore = candidateIndex, score
			}
		}
		selected = append(selected, pool[bestIndex].Blueprint)
	}
	return normalizeBlueprints(selected, specs)
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

func composeBlueprint(spec PageSpec, seed uint64, variant int) PageBlueprint {
	rng := newGrammarRNG(seed)
	target := targetFrameCount(spec)
	roles := storyRoles(spec, rng, target)
	children := []DesignNode{navigationNode(spec, variant)}
	for position, role := range roles {
		children = append(children, composeFrame(spec, rng, variant, position, role))
	}
	root := DesignNode{ID: "page-root", Primitive: "page", Role: "document", Layout: LayoutSpec{Axis: "vertical"}, Children: children}
	directions := []string{"Spatial signal journal", "Living evidence field", "Layered operational atlas", "Kinetic editorial instrument", "Responsive discovery system", "Atmospheric data narrative"}
	return PageBlueprint{Version: 3, ID: strings.Split(spec.ID, "-")[0], CreativeDirection: rng.pick(directions) + fmt.Sprintf(" · genome %04x", seed&0xffff), Root: root}
}

func targetFrameCount(spec PageSpec) int {
	if spec.Strategy == "impact" {
		return 7
	}
	if spec.Strategy == "trust" {
		return 5
	}
	return 6
}

func storyRoles(spec PageSpec, rng *grammarRNG, count int) []string {
	middle := []string{"context", "mechanism", "evidence", "discovery", "proof"}
	if spec.Controls.TrustPriority > .65 {
		middle = append(middle, "proof", "evidence")
	}
	if spec.Controls.VisualAbstraction > .65 {
		middle = append(middle, "discovery", "context")
	}
	roles, previous := []string{"intro"}, "intro"
	for len(roles) < count-1 {
		role := rng.pick(middle)
		for attempts := 0; role == previous && attempts < 4; attempts++ {
			role = rng.pick(middle)
		}
		roles, previous = append(roles, role), role
	}
	return append(roles, "invitation")
}

func navigationNode(spec PageSpec, variant int) DesignNode {
	return DesignNode{ID: fmt.Sprintf("nav-%d", variant), Primitive: "cluster", Role: "navigation", Layout: LayoutSpec{Axis: "horizontal", Align: "between", Gap: "small"}, Style: StyleSpec{Surface: "transparent", Scale: "small"}, Children: []DesignNode{
		textNode("brand", "mark", NodeContent{Headline: spec.Brand}, "small", "strong"),
		textNode("direction", "label", NodeContent{Body: strings.ToUpper(spec.Strategy) + " / LIVE"}, "small", "quiet"),
		actionNode("nav-action", spec, "secondary"),
	}}
}

func composeFrame(spec PageSpec, rng *grammarRNG, variant, position int, role string) DesignNode {
	frame := DesignNode{ID: fmt.Sprintf("frame-%s-%d-%d", role, position, variant), Primitive: "frame", Role: role, Layout: LayoutSpec{Inset: rng.pick([]string{"compact", "balanced", "expansive"}), MinHeight: frameHeight(role, rng)}, Style: StyleSpec{Surface: frameSurface(role, rng), Emphasis: emphasisFor(role)}, Interaction: InteractionSpec{Trigger: "viewport", Motion: motionFor(spec, rng), Strength: spec.Controls.MotionEnergy}}
	copyNode, artifact, collection := copyCluster(spec, rng, role, position), artifactNode(spec, rng, role, position), collectionNode(spec, rng, role, position)
	action := actionNode(fmt.Sprintf("action-%d", position), spec, "primary")
	composition := rng.pick([]string{"split", "stack", "overlay", "mosaic", "rail"})
	if role == "intro" && spec.Controls.SpatialTension > .68 {
		composition = rng.pick([]string{"overlay", "mosaic"})
	}
	if role == "invitation" {
		composition = rng.pick([]string{"stack", "overlay"})
	}
	if role == "proof" && spec.Controls.TrustPriority > .6 {
		composition = rng.pick([]string{"split", "rail"})
	}

	var layout DesignNode
	switch composition {
	case "split":
		children := []DesignNode{copyNode, artifact}
		if rng.chance(28) && role != "intro" {
			children[1] = collection
		}
		layout = DesignNode{ID: frame.ID + "-grid", Primitive: "grid", Layout: LayoutSpec{Columns: 2, Gap: "large", Align: "center", Reverse: rng.chance(34)}, Children: children}
	case "overlay":
		layout = DesignNode{ID: frame.ID + "-overlay", Primitive: "overlay", Layout: LayoutSpec{Align: rng.pick([]string{"start", "center", "end"})}, Children: []DesignNode{artifact, copyNode}}
	case "mosaic":
		secondary := DesignNode{ID: frame.ID + "-side", Primitive: "stack", Layout: LayoutSpec{Gap: "small"}, Children: []DesignNode{collection}}
		if role == "intro" || role == "invitation" {
			secondary.Children = append(secondary.Children, action)
		}
		layout = DesignNode{ID: frame.ID + "-mosaic", Primitive: "grid", Layout: LayoutSpec{Columns: 12, Gap: "medium", Align: "stretch"}, Children: []DesignNode{withSpan(copyNode, 7), withSpan(artifact, 5), withSpan(secondary, 5)}}
	case "rail":
		layout = DesignNode{ID: frame.ID + "-rail", Primitive: "rail", Layout: LayoutSpec{Axis: "horizontal", Gap: "medium", Align: "stretch"}, Children: []DesignNode{copyNode, collection, artifact}}
	default:
		children := []DesignNode{copyNode}
		if role == "mechanism" || role == "evidence" || role == "proof" {
			children = append(children, collection)
		} else {
			children = append(children, artifact)
		}
		layout = DesignNode{ID: frame.ID + "-stack", Primitive: "stack", Layout: LayoutSpec{Gap: "large", Align: alignmentFromSymmetry(spec.Controls.Symmetry, rng)}, Children: children}
	}
	if role == "intro" || role == "invitation" {
		appendAction(&layout, action)
	}
	frame.Children = []DesignNode{layout}
	return frame
}

func copyCluster(spec PageSpec, rng *grammarRNG, role string, position int) DesignNode {
	eyebrow, headline, body := contentForRole(spec, role, position)
	scale := "large"
	if role == "intro" {
		scale = rng.pick([]string{"display", "monumental"})
	} else if role == "invitation" {
		scale = "display"
	}
	return DesignNode{ID: fmt.Sprintf("copy-%s-%d", role, position), Primitive: "cluster", Role: "copy", Layout: LayoutSpec{Axis: "vertical", Gap: "small", Align: alignmentFromSymmetry(spec.Controls.Symmetry, rng)}, Children: []DesignNode{
		textNode(fmt.Sprintf("eyebrow-%d", position), "eyebrow", NodeContent{Eyebrow: eyebrow}, "small", "accent"),
		textNode(fmt.Sprintf("claim-%d", position), "claim", NodeContent{Headline: headline}, scale, emphasisFor(role)),
		textNode(fmt.Sprintf("body-%d", position), "explanation", NodeContent{Body: body}, "body", "quiet"),
	}}
}

func contentForRole(spec PageSpec, role string, position int) (string, string, string) {
	features := spec.FeaturesContent
	feature := FeatureContent{Kicker: spec.SectionLabel, Title: spec.SectionTitle, Body: spec.Description}
	if len(features) > 0 {
		feature = features[position%len(features)]
	}
	switch role {
	case "intro":
		return spec.Eyebrow, spec.Title, spec.Description
	case "context":
		return "THE FIELD", spec.SectionTitle, "The subject becomes the interface: context, motion and evidence share one continuous spatial system."
	case "mechanism":
		return feature.Kicker, feature.Title, feature.Body
	case "evidence":
		return "MEASURED NOW", "Evidence should shape the composition.", "Live measurements are not decoration. They determine rhythm, hierarchy and the next available action."
	case "discovery":
		return "ANOTHER LAYER", "Move through the system, not around it.", "Each transition reveals a different relationship while preserving the same semantic direction."
	case "proof":
		return "VERIFIABLE BY DESIGN", "Every claim leaves a visible trace.", "Operational state, provenance and performance stay attached to the story instead of being buried below it."
	default:
		return "NEXT MOVE", "Enter the system while it is moving.", spec.Description
	}
}

func collectionNode(spec PageSpec, rng *grammarRNG, role string, position int) DesignNode {
	items, collectionRole := featureItems(spec), "features"
	if role == "evidence" || role == "proof" {
		items, collectionRole = metricItems(spec), "metrics"
	} else if role == "discovery" {
		collectionRole = "specimens"
	}
	return DesignNode{ID: fmt.Sprintf("collection-%d", position), Primitive: "collection", Role: collectionRole, Layout: LayoutSpec{Columns: 2 + int(rng.next()%3), Gap: rng.pick([]string{"small", "medium"})}, Style: StyleSpec{Surface: rng.pick([]string{"line", "soft", "none"}), Shape: rng.pick([]string{"sharp", "soft"})}, Content: NodeContent{Items: items}}
}

func artifactNode(spec PageSpec, rng *grammarRNG, role string, position int) DesignNode {
	visuals := []string{"signal-field", "telemetry", "orbital-map", "specimen-field", "type-sculpture", "depth-map", "runtime"}
	if spec.Controls.VisualAbstraction < .42 {
		visuals = []string{"telemetry", "runtime", "depth-map"}
	}
	if role == "proof" || role == "evidence" {
		visuals = append(visuals, "runtime", "telemetry")
	}
	return DesignNode{ID: fmt.Sprintf("artifact-%s-%d", role, position), Primitive: "artifact", Role: visualRole(role), Layout: LayoutSpec{Span: 5}, Style: StyleSpec{Surface: rng.pick([]string{"void", "line", "soft"}), Shape: rng.pick([]string{"sharp", "soft", "round"})}, Visual: VisualSpec{Kind: rng.pick(visuals), Position: rng.pick([]string{"center", "edge", "bleed"}), Intensity: spec.Controls.VisualAbstraction}, Interaction: InteractionSpec{Trigger: "viewport", Motion: motionFor(spec, rng), Strength: spec.Controls.MotionEnergy}}
}

func actionNode(id string, spec PageSpec, role string) DesignNode {
	labels := map[string]string{"demo": "Book a live demo", "trial": "Start building", "waitlist": "Request access"}
	return DesignNode{ID: id, Primitive: "action", Role: role, Style: StyleSpec{Surface: "accent", Scale: "small", Emphasis: "strong"}, Content: NodeContent{Headline: labels[spec.CTA]}}
}
func textNode(id, role string, content NodeContent, scale, emphasis string) DesignNode {
	return DesignNode{ID: id, Primitive: "text", Role: role, Style: StyleSpec{Scale: scale, Emphasis: emphasis}, Content: content}
}
func withSpan(node DesignNode, span int) DesignNode { node.Layout.Span = span; return node }
func appendAction(node *DesignNode, action DesignNode) {
	if node.Primitive == "overlay" && len(node.Children) > 1 && node.Children[1].Primitive == "cluster" {
		node.Children[1].Children = append(node.Children[1].Children, action)
		return
	}
	node.Children = append(node.Children, action)
}
func frameHeight(role string, rng *grammarRNG) string {
	if role == "intro" {
		return rng.pick([]string{"screen", "tall"})
	}
	if role == "invitation" {
		return "tall"
	}
	return rng.pick([]string{"auto", "medium", "tall"})
}
func frameSurface(role string, rng *grammarRNG) string {
	if role == "invitation" {
		return "accent-wash"
	}
	if role == "intro" {
		return rng.pick([]string{"void", "atmosphere"})
	}
	return rng.pick([]string{"void", "line", "soft", "contrast"})
}
func emphasisFor(role string) string {
	if role == "intro" || role == "invitation" {
		return "strong"
	}
	if role == "discovery" {
		return "expressive"
	}
	return "balanced"
}
func visualRole(role string) string {
	if role == "proof" || role == "evidence" {
		return "evidence"
	}
	if role == "intro" || role == "discovery" {
		return "atmosphere"
	}
	return "system"
}
func motionFor(spec PageSpec, rng *grammarRNG) string {
	if spec.Controls.MotionEnergy < .28 {
		return "still"
	}
	if spec.Controls.MotionEnergy > .72 {
		return rng.pick([]string{"drift", "orbit", "pulse"})
	}
	return rng.pick([]string{"reveal", "drift", "still"})
}
func alignmentFromSymmetry(symmetry float64, rng *grammarRNG) string {
	if symmetry > .68 {
		return "center"
	}
	if symmetry < .32 {
		return rng.pick([]string{"start", "end"})
	}
	return "start"
}

func featureItems(spec PageSpec) []BlueprintItem {
	result := make([]BlueprintItem, 0, len(spec.FeaturesContent))
	for _, feature := range spec.FeaturesContent {
		result = append(result, BlueprintItem{Label: feature.Kicker, Title: feature.Title, Body: feature.Body})
	}
	return result
}
func metricItems(spec PageSpec) []BlueprintItem {
	result := make([]BlueprintItem, 0, len(spec.Metrics))
	for _, metric := range spec.Metrics {
		result = append(result, BlueprintItem{Label: metric.Label, Value: metric.Value + metric.Unit})
	}
	return result
}

func grammarFitness(blueprint PageBlueprint, spec PageSpec, variant int) int {
	stats := inspectBlueprint(blueprint)
	score := 420 - abs(stats.frames-targetFrameCount(spec))*90
	score += len(stats.primitives)*18 + len(stats.roles)*9 + stats.maxDepth*12 + stats.artifacts*8 + stats.layoutChanges*11
	if stats.actions == 0 {
		score -= 120
	}
	if stats.invalid > 0 {
		score -= stats.invalid * 500
	}
	if spec.Controls.VisualAbstraction > .65 {
		score += stats.artifacts * 6
	}
	if spec.Controls.TrustPriority > .65 {
		score += stats.metrics * 12
	}
	return score - variant/256
}

type blueprintStats struct {
	frames, maxDepth, artifacts, actions, metrics, invalid, layoutChanges int
	primitives, roles, paths                                              map[string]bool
}

func inspectBlueprint(blueprint PageBlueprint) blueprintStats {
	stats := blueprintStats{primitives: map[string]bool{}, roles: map[string]bool{}, paths: map[string]bool{}}
	var walk func(DesignNode, int, string)
	walk = func(node DesignNode, depth int, parent string) {
		stats.primitives[node.Primitive] = true
		if node.Role != "" {
			stats.roles[node.Role] = true
		}
		stats.paths[parent+">"+node.Primitive+":"+node.Role+fmt.Sprintf(":%d", node.Layout.Columns)] = true
		if depth > stats.maxDepth {
			stats.maxDepth = depth
		}
		if node.Primitive == "frame" {
			stats.frames++
		}
		if node.Primitive == "artifact" {
			stats.artifacts++
		}
		if node.Primitive == "action" {
			stats.actions++
		}
		if node.Primitive == "collection" && node.Role == "metrics" {
			stats.metrics++
		}
		definition, ok := primitiveRegistry[node.Primitive]
		if !ok || len(node.Children) < definition.MinChildren || len(node.Children) > definition.MaxChildren {
			stats.invalid++
		}
		for _, child := range node.Children {
			if !definition.AllowedChildren[child.Primitive] {
				stats.invalid++
			}
			if child.Primitive != node.Primitive {
				stats.layoutChanges++
			}
			walk(child, depth+1, node.Primitive)
		}
	}
	walk(blueprint.Root, 0, "root")
	return stats
}

func blueprintNovelty(a, b PageBlueprint) int {
	aStats, bStats := inspectBlueprint(a), inspectBlueprint(b)
	intersection, union := 0, len(aStats.paths)
	for path := range bStats.paths {
		if aStats.paths[path] {
			intersection++
		} else {
			union++
		}
	}
	distance := 100
	if union > 0 {
		distance = 100 - intersection*100/union
	}
	return distance + abs(aStats.frames-bStats.frames)*8 + abs(aStats.maxDepth-bStats.maxDepth)*5
}

func normalizeBlueprints(input []PageBlueprint, specs []PageSpec) []PageBlueprint {
	fallbacks, byID := fallbackBlueprints(specs), map[string]PageBlueprint{}
	for _, blueprint := range input {
		if blueprint.ID != "signal" && blueprint.ID != "pulse" && blueprint.ID != "atlas" {
			continue
		}
		blueprint.Version, blueprint.Root = 3, sanitizeNode(blueprint.Root, 0)
		if inspectBlueprint(blueprint).invalid == 0 {
			byID[blueprint.ID] = blueprint
		}
	}
	result := make([]PageBlueprint, 0, len(fallbacks))
	for _, fallback := range fallbacks {
		if generated, ok := byID[fallback.ID]; ok {
			result = append(result, generated)
		} else {
			result = append(result, fallback)
		}
	}
	return result
}

func sanitizeNode(node DesignNode, depth int) DesignNode {
	definition, ok := primitiveRegistry[node.Primitive]
	if !ok {
		return DesignNode{ID: node.ID, Primitive: "rule", Role: "divider"}
	}
	if depth >= 6 {
		node.Children = nil
	}
	if len(node.Children) > definition.MaxChildren {
		node.Children = node.Children[:definition.MaxChildren]
	}
	children := make([]DesignNode, 0, len(node.Children))
	for _, child := range node.Children {
		if definition.AllowedChildren[child.Primitive] {
			children = append(children, sanitizeNode(child, depth+1))
		}
	}
	node.Children = children
	return node
}

func applyBlueprints(result DesignResult, blueprints []PageBlueprint, source, generator string) DesignResult {
	blueprints = normalizeBlueprints(blueprints, result.Specs)
	for index := range result.Specs {
		result.Specs[index].Blueprint = blueprints[index]
		result.Specs[index].Descriptor = blueprints[index].CreativeDirection
	}
	result.ASTSource, result.Generator = source, generator
	result.CandidateCount, result.DesignSpace = grammarCandidatesPerUniverse*len(result.Specs), grammarDesignSpace
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
	if blueprint.Root.Primitive == "" {
		blueprint = fallbackBlueprints([]PageSpec{spec})[0]
	}
	blueprint.Version, blueprint.ID = 3, strings.Split(spec.ID, "-")[0]
	blueprint.CreativeDirection = fmt.Sprintf("%s · generation %02d %s mutation", blueprint.CreativeDirection, generation, weakness)
	rng := newGrammarRNG(promptSeed(blueprint.ID+weakness) ^ uint64(generation*65537))
	switch weakness {
	case "originality", "trust":
		role := "discovery"
		if weakness == "trust" {
			role = "proof"
		}
		frame, insertAt := composeFrame(spec, rng, generation, generation+8, role), len(blueprint.Root.Children)-1
		blueprint.Root.Children = append(blueprint.Root.Children, DesignNode{})
		copy(blueprint.Root.Children[insertAt+1:], blueprint.Root.Children[insertAt:])
		blueprint.Root.Children[insertAt] = frame
	case "clarity":
		mutateFirstRole(&blueprint.Root, "claim", func(node *DesignNode) { node.Style.Scale, node.Style.Emphasis = "display", "strong" })
	case "conversion":
		mutateFirstRole(&blueprint.Root, "primary", func(node *DesignNode) { node.Style.Scale, node.Style.Surface = "large", "accent" })
	}
	blueprint.Root = sanitizeNode(blueprint.Root, 0)
	return blueprint
}

func mutateFirstRole(node *DesignNode, role string, mutate func(*DesignNode)) bool {
	if node.Role == role {
		mutate(node)
		return true
	}
	for index := range node.Children {
		if mutateFirstRole(&node.Children[index], role, mutate) {
			return true
		}
	}
	return false
}
func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
