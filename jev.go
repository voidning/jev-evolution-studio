package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
)

func AskJev(prompt string, current []PageSpec) (Answers, error) {
	questions := map[string]any{
		"world":               choice("Choose the semantic product world that should drive the site's brand, copy, evidence, and imagery", map[string]string{"ai_data": "AI, analytics, software, or live data", "quantum": "Quantum computing or advanced physics", "ocean": "Ocean, marine science, diving, or deep-sea exploration", "space": "Space, astronomy, satellites, or aerospace", "biotech": "Biology, healthcare, or life science", "creative": "Creative tools, media, culture, or design"}),
		"archetype":           choice("Choose the product archetype that best matches the requested website", map[string]string{"developer": "Developer infrastructure or technical tool", "enterprise": "Trust-oriented B2B or enterprise system", "data_product": "Data, intelligence, analytics, or live operations product"}),
		"goal":                choice("Choose the primary job this website must perform", map[string]string{"explore": "Help visitors understand and explore", "convert": "Drive a signup, purchase, or application", "trust": "Establish authority and reduce perceived risk"}),
		"tone":                choice("Choose the visual voice that best expresses the request", map[string]string{"technical": "Precise restrained technical minimalism", "bold": "Futuristic high-impact visual confidence", "editorial": "Warm spacious editorial sophistication"}),
		"theme":               choice("Choose the most fitting accent color family", map[string]string{"violet": "Electric violet for intelligence and depth", "mint": "Signal green for speed and live systems", "ember": "Warm orange for a human editorial feel"}),
		"hero":                choice("Choose the best hero composition", map[string]string{"split": "Copy paired with a live product artifact", "centered": "A cinematic centered statement", "terminal": "A developer terminal is the main proof"}),
		"visual":              choice("Choose the primary product visual", map[string]string{"dashboard": "Live operational dashboard", "terminal": "Executable code terminal", "abstract": "Generative signal field or system map"}),
		"features":            choice("Choose the best feature narrative", map[string]string{"bento": "Asymmetric capability system", "columns": "Calm editorial sequence", "timeline": "Continuous event flow"}),
		"density":             choice("Choose the correct information density", map[string]string{"airy": "Cinematic focus and large pauses", "balanced": "Balanced proof and breathing room", "compact": "Dense fast operational detail"}),
		"navigation":          choice("Choose the navigation behavior", map[string]string{"minimal": "Very few destinations and strong focus", "exploratory": "Multiple pathways for product exploration", "action": "Navigation organized around one conversion action"}),
		"motion":              choice("Choose the motion system", map[string]string{"still": "Mostly static and calm", "fluid": "Continuous subtle morphing", "kinetic": "Energetic staged transformations"}),
		"story":               choice("Choose the page storytelling order", map[string]string{"proof_first": "Lead with evidence and product reality", "vision_first": "Lead with a bold future-facing idea", "problem_first": "Lead with the user's pain then resolve it"}),
		"cta":                 choice("Choose the primary call to action", map[string]string{"demo": "See or book a product demonstration", "trial": "Start using the product now", "waitlist": "Apply for early access"}),
		"contrast":            score("Rate how dramatic the visual contrast should be", []string{"Very quiet", "Restrained", "Noticeable", "Dramatic", "Extreme"}),
		"novelty":             score("Rate how far the design should depart from conventional SaaS patterns", []string{"Conventional", "Familiar", "Distinctive", "Experimental", "Unprecedented"}),
		"spatial_tension":     score("Rate the desired spatial tension between ordered and deliberately off-axis composition", qualityLevels("ordered", "high-tension")),
		"visual_abstraction":  score("Rate whether the visual language should be literal or abstract", qualityLevels("literal", "abstract")),
		"information_density": score("Rate the desired amount of information visible at once", qualityLevels("sparse", "dense")),
		"narrative_depth":     score("Rate how strongly the page should unfold as a journey", qualityLevels("direct", "deeply narrative")),
		"trust_priority":      score("Rate how strongly evidence and credibility should shape the composition", qualityLevels("expressive", "evidence-led")),
		"motion_energy":       score("Rate the desired energy of the motion system", qualityLevels("still", "kinetic")),
		"symmetry":            score("Rate whether the composition should be asymmetric or symmetrical", qualityLevels("asymmetric", "symmetrical")),
		"pricing":             noul("The requested page should contain a pricing section"),
		"logos":               noul("The requested page should show a customer or partner trust strip"),
		"stats":               noul("The requested page should show quantitative product evidence"),
		"testimony":           noul("The requested page should contain customer testimony or human proof"),
	}
	state := map[string]any{"request": prompt, "current_page": current, "instruction": "Interpret this as a UI design or edit request. Preserve prior choices unless the request asks to change them."}
	return askSystemOne(state, questions)
}

func AskJevCritic(prompt string, specs []PageSpec) (Answers, error) {
	questions := map[string]any{}
	for _, spec := range specs {
		prefix := spec.ID + "_"
		questions[prefix+"originality"] = score("Rate how original and memorable this concept is for the user's request", qualityLevels("generic", "category-defining"))
		questions[prefix+"clarity"] = score("Rate how immediately understandable this concept is for the target audience", qualityLevels("confusing", "instantly clear"))
		questions[prefix+"trust"] = score("Rate how credible and trustworthy this concept feels for the product", qualityLevels("unconvincing", "highly credible"))
		questions[prefix+"conversion"] = score("Rate how effectively this concept guides a visitor to the requested primary action", qualityLevels("directionless", "highly persuasive"))
	}
	state := map[string]any{"request": prompt, "concepts": specs, "instruction": "Judge each concept independently. Use only the provided design genome and user request."}
	return askSystemOne(state, questions)
}

func askSystemOne(state any, questions map[string]any) (Answers, error) {
	key := loadAPIKey()
	if key == "" {
		return nil, errors.New("TYPESAFE_API_KEY is not configured")
	}
	payload := map[string]any{"model": "jev-1.13.0", "state": state, "questions": questions}
	body, _ := json.Marshal(payload)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.typesafe.ai/v1/systemone", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("TypeSafe returned %s: %s", resp.Status, strings.TrimSpace(string(limited)))
	}
	var result struct {
		Answers Answers `json:"answers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Answers) == 0 {
		return nil, errors.New("TypeSafe returned no answers")
	}
	return result.Answers, nil
}

func choice(instructions string, criteria map[string]string) map[string]any {
	return map[string]any{"type": "choice", "instructions": instructions, "criteria": criteria}
}
func noul(instructions string) map[string]any {
	return map[string]any{"type": "noul", "instructions": instructions}
}
func score(instructions string, criteria []string) map[string]any {
	return map[string]any{"type": "score", "instructions": instructions, "criteria": criteria}
}
func qualityLevels(low, high string) []string {
	return []string{"Very " + low, low, "Adequate", "Strong", high}
}

func LocalAnswers(prompt string) Answers {
	p := strings.ToLower(prompt)
	has := func(words ...string) bool {
		for _, word := range words {
			if strings.Contains(p, word) {
				return true
			}
		}
		return false
	}
	warm, bold := has("暖", "warm", "橙", "人文"), has("大胆", "bold", "冲击", "未来", "震撼")
	compact, airy := has("紧凑", "dense", "信息密度", "数据密集"), has("留白", "克制", "极简", "airy", "calm")
	pricing := .28
	if has("删除定价", "不要定价", "remove pricing", "without pricing") {
		pricing = .05
	} else if has("定价", "pricing", "价格", "套餐") {
		pricing = .94
	}
	return Answers{
		"world":      {Choice: localWorld(has), Confidence: .92, Probabilities: dist("ai_data", "quantum", "ocean")},
		"archetype":  {Choice: ternary(has("企业", "b2b", "机构"), "enterprise", "developer"), Confidence: .9, Probabilities: dist("developer", "enterprise", "data_product")},
		"goal":       {Choice: ternary(has("购买", "转化", "定价", "pricing"), "convert", "explore"), Confidence: .86, Probabilities: dist("explore", "convert", "trust")},
		"tone":       {Choice: ternary(warm, "editorial", ternary(bold, "bold", "technical")), Confidence: .89, Probabilities: dist("technical", "bold", "editorial")},
		"theme":      {Choice: ternary(warm, "ember", ternary(bold, "mint", "violet")), Confidence: .92, Probabilities: dist("violet", "mint", "ember")},
		"hero":       {Choice: ternary(bold, "terminal", ternary(airy, "centered", "split")), Confidence: .87, Probabilities: dist("split", "centered", "terminal")},
		"visual":     {Choice: ternary(bold, "abstract", ternary(compact, "dashboard", "terminal")), Confidence: .83, Probabilities: dist("dashboard", "terminal", "abstract")},
		"features":   {Choice: ternary(compact, "timeline", ternary(airy, "columns", "bento")), Confidence: .82, Probabilities: dist("bento", "columns", "timeline")},
		"density":    {Choice: ternary(compact, "compact", ternary(airy, "airy", "balanced")), Confidence: .79, Probabilities: dist("airy", "balanced", "compact")},
		"navigation": {Choice: ternary(has("转化", "购买"), "action", "minimal"), Confidence: .78, Probabilities: dist("minimal", "exploratory", "action")},
		"motion":     {Choice: ternary(bold, "kinetic", "fluid"), Confidence: .84, Probabilities: dist("still", "fluid", "kinetic")},
		"story":      {Choice: ternary(bold, "vision_first", "proof_first"), Confidence: .81, Probabilities: dist("proof_first", "vision_first", "problem_first")},
		"cta":        {Choice: ternary(has("内测", "waitlist"), "waitlist", ternary(has("试用", "购买"), "trial", "demo")), Confidence: .86, Probabilities: dist("demo", "trial", "waitlist")},
		"contrast":   {Score: ternaryFloat(bold, 3.7, 2.3), Confidence: .78}, "novelty": {Score: ternaryFloat(bold, 3.6, 2.4), Confidence: .8},
		"spatial_tension":     {Score: ternaryFloat(bold, 3.8, ternaryFloat(airy, 1.7, 2.6)), Confidence: .82},
		"visual_abstraction":  {Score: ternaryFloat(bold, 3.7, 2.1), Confidence: .8},
		"information_density": {Score: ternaryFloat(compact, 3.8, ternaryFloat(airy, 1.2, 2.5)), Confidence: .84},
		"narrative_depth":     {Score: ternaryFloat(has("故事", "旅程", "沉浸", "cinematic"), 3.8, 2.5), Confidence: .78},
		"trust_priority":      {Score: ternaryFloat(has("可信", "科研", "证据", "企业", "trust"), 3.9, 2.4), Confidence: .86},
		"motion_energy":       {Score: ternaryFloat(bold, 3.8, ternaryFloat(airy, 1.3, 2.4)), Confidence: .81},
		"symmetry":            {Score: ternaryFloat(airy, 3.3, ternaryFloat(bold, 1.2, 2.4)), Confidence: .76},
		"pricing":             {Noul: pricing}, "logos": {Noul: ternaryFloat(has("logo", "客户", "信任", "企业"), .88, .68)}, "stats": {Noul: ternaryFloat(has("数据", "实时", "速度", "分析", "指标"), .93, .61)}, "testimony": {Noul: ternaryFloat(has("客户", "案例", "证言"), .86, .34)},
	}
}

func CompileConcepts(answers Answers, mode string, latency int64) DesignResult {
	baseTheme := pick(answers["theme"].Choice, "violet", "violet", "ember", "mint")
	alt := map[string][]string{"violet": {"mint", "ember"}, "mint": {"violet", "ember"}, "ember": {"violet", "mint"}}[baseTheme]
	clarity := makeSpec("signal", "Signal", "Precision universe", "clarity", answers, map[string]any{"theme": baseTheme, "hero": "split", "visual": "dashboard", "features": "bento", "density": "balanced"})
	impact := makeSpec("pulse", "Pulse", "Impact universe", "impact", answers, map[string]any{"theme": alt[0], "hero": "centered", "visual": "abstract", "features": "timeline", "density": "compact", "motion": "kinetic", "pricing": false, "stats": false})
	trust := makeSpec("atlas", "Atlas", "Trust universe", "trust", answers, map[string]any{"theme": alt[1], "hero": "terminal", "visual": "terminal", "features": "columns", "density": "airy", "logos": true, "stats": true})
	return DesignResult{Mode: mode, LatencyMS: latency, Prompt: defaultPrompt, Generation: 1, Winner: 0, SwarmSize: len(answers), Specs: []PageSpec{clarity, impact, trust}}
}

func makeSpec(id, name, descriptor, strategy string, answers Answers, overrides map[string]any) PageSpec {
	theme := pick(answers["theme"].Choice, "violet", "violet", "ember", "mint")
	hero := pick(answers["hero"].Choice, "split", "split", "centered", "terminal")
	visual := pick(answers["visual"].Choice, "dashboard", "dashboard", "terminal", "abstract")
	features := pick(answers["features"].Choice, "bento", "bento", "columns", "timeline")
	density := pick(answers["density"].Choice, "balanced", "airy", "balanced", "compact")
	tone := pick(answers["tone"].Choice, "technical", "technical", "bold", "editorial")
	navigation := pick(answers["navigation"].Choice, "minimal", "minimal", "exploratory", "action")
	motion := pick(answers["motion"].Choice, "fluid", "still", "fluid", "kinetic")
	story := pick(answers["story"].Choice, "proof_first", "proof_first", "vision_first", "problem_first")
	cta := pick(answers["cta"].Choice, "demo", "demo", "trial", "waitlist")
	world := pick(answers["world"].Choice, "ai_data", "ai_data", "quantum", "ocean", "space", "biotech", "creative")
	for key, target := range map[string]*string{"theme": &theme, "hero": &hero, "visual": &visual, "features": &features, "density": &density, "navigation": &navigation, "motion": &motion, "story": &story, "cta": &cta} {
		if value, ok := overrides[key].(string); ok {
			*target = value
		}
	}
	content := contentFor(world, strategy)
	titles := map[string]string{"clarity": content.Title, "impact": content.Title, "trust": content.Title}
	descriptions := map[string]string{"clarity": content.Description, "impact": content.Description, "trust": content.Description}
	if world == "ai_data" && tone == "editorial" && strategy == "clarity" {
		titles[strategy] = "Intelligence that\nkeeps up."
	}
	logos, pricing, stats := answers["logos"].Noul > .5, answers["pricing"].Noul > .55, answers["stats"].Noul > .5
	if v, ok := overrides["logos"].(bool); ok {
		logos = v
	}
	if v, ok := overrides["pricing"].(bool); ok {
		pricing = v
	}
	if v, ok := overrides["stats"].(bool); ok {
		stats = v
	}
	decision := func(labelText, value, group, key string, fallback float64) Decision {
		a := answers[key]
		return Decision{Label: labelText, Value: value, Group: group, Confidence: conf(a, fallback), Distribution: a.Probabilities}
	}
	decisions := []Decision{
		decision("Product world", humanLabel(world), "Intent", "world", .9), decision("Archetype", humanLabel(answers["archetype"].Choice), "Intent", "archetype", .9), decision("Primary goal", humanLabel(answers["goal"].Choice), "Intent", "goal", .84), decision("Story order", humanLabel(story), "Intent", "story", .8), decision("CTA", humanLabel(cta), "Intent", "cta", .82),
		decision("Hero", humanLabel(hero), "Experience", "hero", .86), decision("Product visual", humanLabel(visual), "Experience", "visual", .82), decision("Feature system", humanLabel(features), "Experience", "features", .82), decision("Navigation", humanLabel(navigation), "Experience", "navigation", .78),
		decision("Visual tone", humanLabel(tone), "Visual", "tone", .9), decision("Accent", humanLabel(theme), "Visual", "theme", .88), decision("Density", humanLabel(density), "Visual", "density", .78), decision("Motion", humanLabel(motion), "Visual", "motion", .82),
		{Label: "Contrast", Value: scoreLabel(answers["contrast"].Score), Group: "Signals", Confidence: conf(answers["contrast"], .78)}, {Label: "Novelty", Value: scoreLabel(answers["novelty"].Score), Group: "Signals", Confidence: conf(answers["novelty"], .8)}, {Label: "Pricing", Value: yesNo(pricing), Group: "Signals", Confidence: math.Abs(answers["pricing"].Noul-.5) * 2}, {Label: "Proof", Value: yesNo(stats || logos), Group: "Signals", Confidence: .84},
	}
	controls := DesignControls{
		SpatialTension: controlValue(answers["spatial_tension"], .55), VisualAbstraction: controlValue(answers["visual_abstraction"], .55),
		InformationDensity: controlValue(answers["information_density"], .55), NarrativeDepth: controlValue(answers["narrative_depth"], .55),
		TrustPriority: controlValue(answers["trust_priority"], .55), MotionEnergy: controlValue(answers["motion_energy"], .55), Symmetry: controlValue(answers["symmetry"], .5),
	}
	// The three universes are continuous lenses, not fixed page skeletons.
	if strategy == "clarity" {
		controls.SpatialTension *= .72
		controls.Symmetry = clamp01(controls.Symmetry + .16)
	}
	if strategy == "impact" {
		controls.SpatialTension = clamp01(controls.SpatialTension + .22)
		controls.VisualAbstraction = clamp01(controls.VisualAbstraction + .2)
		controls.MotionEnergy = clamp01(controls.MotionEnergy + .2)
	}
	if strategy == "trust" {
		controls.TrustPriority = clamp01(controls.TrustPriority + .24)
		controls.MotionEnergy *= .72
	}
	return PageSpec{ID: id, Name: name, Descriptor: descriptor, Strategy: strategy, Generation: 1, Theme: theme, Hero: hero, Visual: visual, Features: features, Density: density, Navigation: navigation, Motion: motion, Story: story, CTA: cta, World: world, Brand: content.Brand, Eyebrow: content.Eyebrow, SectionLabel: content.SectionLabel, SectionTitle: content.SectionTitle, FeaturesContent: content.Features, Metrics: content.Metrics, ShowLogos: logos, ShowPricing: pricing, ShowStats: stats, Title: titles[strategy], Description: descriptions[strategy], Decisions: decisions, Controls: controls}
}

func controlValue(answer Answer, fallback float64) float64 {
	if answer.Score == 0 {
		return fallback
	}
	return clamp01(answer.Score / 4)
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

type pageContent struct {
	Brand        string
	Eyebrow      string
	Title        string
	Description  string
	SectionLabel string
	SectionTitle string
	Features     []FeatureContent
	Metrics      []MetricContent
}

func contentFor(world, strategy string) pageContent {
	titles := map[string]map[string]string{
		"ai_data":  {"clarity": "See the signal.\nWhile it still matters.", "impact": "Make the invisible\nimpossible to miss.", "trust": "The live system\nyou can trust."},
		"quantum":  {"clarity": "Quantum advantage,\nmade observable.", "impact": "Compute beyond\nthe possible.", "trust": "Quantum systems.\nEnterprise ready."},
		"ocean":    {"clarity": "Enter the deep.\nSee what lives there.", "impact": "The ocean has\na second sky.", "trust": "Explore the abyss.\nTrust every signal."},
		"space":    {"clarity": "Every orbit.\nOne living picture.", "impact": "Build beyond\nthe horizon.", "trust": "Mission intelligence\nwithout blind spots."},
		"biotech":  {"clarity": "See biology\nas it changes.", "impact": "Life is writing\nthe next answer.", "trust": "Evidence for every\nclinical decision."},
		"creative": {"clarity": "From raw idea\nto clear direction.", "impact": "Make culture\nmove differently.", "trust": "Creative systems\nbuilt to ship."},
	}
	descriptions := map[string]map[string]string{
		"ai_data":  {"clarity": "Turn live product data into decisions your team can act on — without waiting for another dashboard refresh.", "impact": "A living intelligence layer that detects the shift, reveals the pattern, and moves before the moment disappears.", "trust": "Decision-grade intelligence with observable pipelines, measurable latency, and proof at every layer."},
		"quantum":  {"clarity": "Translate quantum behavior into results engineers and executives can inspect, compare, and act on.", "impact": "A computational frontier where impossible state spaces become visible, programmable, and immediate.", "trust": "Auditable orchestration, reproducible benchmarks, and secure hybrid workflows for serious quantum programs."},
		"ocean":    {"clarity": "Join a live expedition into the lightless ocean, where every dive reveals another living system.", "impact": "Descend past the last sunlight and watch an unseen universe ignite around you.", "trust": "Research-grade navigation, verified observations, and a transparent record of every discovery."},
		"space":    {"clarity": "Turn fragmented mission telemetry into one precise, continuously updated operational view.", "impact": "Design, launch, and command the systems that will carry us beyond familiar space.", "trust": "Flight-proven monitoring and explainable mission decisions from ground station to orbit."},
		"biotech":  {"clarity": "Connect experiments, samples, and evidence in one live model of biological change.", "impact": "Explore the patterns inside living systems and accelerate the moment discovery becomes medicine.", "trust": "Traceable data, validated workflows, and decision-ready evidence for regulated teams."},
		"creative": {"clarity": "Organize inspiration, direction, and production into one clear creative operating system.", "impact": "A kinetic space where ideas collide, mutate, and become work nobody has seen before.", "trust": "A repeatable creative workflow with clear ownership, review, and delivery at every stage."},
	}
	brands := map[string]string{"ai_data": "Velocity", "quantum": "Axiom Q", "ocean": "Abyssal", "space": "Apogee", "biotech": "Helix", "creative": "Motive"}
	eyebrows := map[string]string{"ai_data": "BUILT FOR LIVE PRODUCTS", "quantum": "THE QUANTUM OPERATING LAYER", "ocean": "LIVE FROM THE HADAL ZONE", "space": "MISSION SYSTEMS / ALWAYS ON", "biotech": "BIOLOGICAL INTELLIGENCE / LIVE", "creative": "A NEW CREATIVE FREQUENCY"}
	sections := map[string][2]string{"ai_data": {"ONE CONTINUOUS SYSTEM", "From raw event to\nclear decision."}, "quantum": {"FROM QUBIT TO OUTCOME", "Control the system.\nProve the advantage."}, "ocean": {"THE LIVING ABYSS", "Descend through\na world of light."}, "space": {"ONE MISSION PICTURE", "From first signal\nto final orbit."}, "biotech": {"EVIDENCE IN MOTION", "From living signal\nto clear discovery."}, "creative": {"THE IDEA ENGINE", "From first spark\nto finished work."}}
	featureSets := map[string][]FeatureContent{
		"ai_data":  {{"REAL-TIME ENGINE", "Every signal, while it matters.", "Stream product events into one live model of what users are doing right now."}, {"SEMANTIC LAYER", "Ask in plain language.", "Describe the change you need to understand without brittle queries."}, {"DEVELOPER FIRST", "Ship the decision.", "Move from detected pattern to product action in a few typed lines."}},
		"quantum":  {{"HYBRID RUNTIME", "Orchestrate every backend.", "Route workloads across simulators, QPUs, and classical infrastructure."}, {"BENCHMARK LAYER", "Prove useful advantage.", "Compare fidelity, cost, and runtime with reproducible experiments."}, {"SECURE CONTROL", "Deploy with confidence.", "Bring quantum workflows into governed enterprise environments."}},
		"ocean":    {{"LIVE DESCENT", "Follow the dive in real time.", "Track depth, temperature, and navigation as the vehicle enters the abyss."}, {"SPECIES SIGNAL", "See life reveal itself.", "Surface bioluminescent encounters and annotate them with marine scientists."}, {"EXPEDITION LOG", "Every discovery preserved.", "Build an open, verifiable record from raw sensor stream to field note."}},
		"space":    {{"ORBITAL TWIN", "Every vehicle, one system.", "Fuse spacecraft state and ground telemetry into a live operational model."}, {"ANOMALY SIGNAL", "Detect the shift early.", "Find mission-critical deviations before they become irreversible."}, {"COMMAND LAYER", "Act with full context.", "Move from observation to validated command through governed workflows."}},
		"biotech":  {{"LIVE ASSAY", "Watch biology respond.", "Unify experimental signals across instruments, samples, and timepoints."}, {"EVIDENCE GRAPH", "Trace every conclusion.", "Connect results back to protocols, batches, and source observations."}, {"TRANSLATION LAYER", "Move discovery forward.", "Turn validated findings into the next experiment or clinical decision."}},
		"creative": {{"DIRECTION SPACE", "Align around the idea.", "Turn references and language into a shared visual direction."}, {"MUTATION ENGINE", "Explore without losing intent.", "Generate structured variations while preserving what the team locks."}, {"SHIP SYSTEM", "Finish with momentum.", "Move concepts through review, production, and delivery in one flow."}},
	}
	metricSets := map[string][]MetricContent{
		"ai_data":  {{"84", "ms", "median signal latency"}, {"12.4", "B", "events processed"}, {"99.99", "%", "pipeline uptime"}},
		"quantum":  {{"127", "q", "managed qubits"}, {"42", "x", "workflow speedup"}, {"99.9", "%", "run traceability"}},
		"ocean":    {{"6,200", "m", "live descent depth"}, {"98", "", "species signals"}, {"4K", "", "expedition stream"}},
		"space":    {{"18", "", "active vehicles"}, {"240", "ms", "command latency"}, {"99.99", "%", "mission uptime"}},
		"biotech":  {{"4.2", "M", "cells observed"}, {"31", "%", "faster iteration"}, {"100", "%", "evidence traceable"}},
		"creative": {{"8.4", "K", "directions explored"}, {"3.1", "x", "faster review"}, {"92", "%", "intent retained"}},
	}
	section := sections[world]
	return pageContent{Brand: brands[world], Eyebrow: eyebrows[world], Title: titles[world][strategy], Description: descriptions[world][strategy], SectionLabel: section[0], SectionTitle: section[1], Features: featureSets[world], Metrics: metricSets[world]}
}

func localWorld(has func(...string) bool) string {
	if has("量子", "quantum", "qubit") {
		return "quantum"
	}
	if has("深海", "海洋", "潜水", "ocean", "marine", "abyss") {
		return "ocean"
	}
	if has("太空", "航天", "卫星", "space", "orbit") {
		return "space"
	}
	if has("生物", "医疗", "药物", "biotech", "clinical") {
		return "biotech"
	}
	if has("创意", "设计", "媒体", "creative", "culture") {
		return "creative"
	}
	return "ai_data"
}

func ApplyCritic(result DesignResult, answers Answers) DesignResult {
	for i := range result.Specs {
		s := &result.Specs[i]
		fallback := localScorecard(*s)
		s.Scores = Scorecard{Originality: scorePercent(answers[s.ID+"_originality"], fallback.Originality), Clarity: scorePercent(answers[s.ID+"_clarity"], fallback.Clarity), Trust: scorePercent(answers[s.ID+"_trust"], fallback.Trust), Conversion: scorePercent(answers[s.ID+"_conversion"], fallback.Conversion)}
		s.Scores.Composite = composite(*s)
	}
	result.Winner = winner(result.Specs)
	return result
}

func EvolveConcepts(result DesignResult) DesignResult {
	result.Generation++
	result.MutationLog = nil
	for i := range result.Specs {
		s := &result.Specs[i]
		weakness := weakest(s.Scores)
		s.Generation = result.Generation
		s.ID = strings.Split(s.ID, "-")[0] + fmt.Sprintf("-g%d", result.Generation)
		s.Descriptor = "Gen " + fmt.Sprint(result.Generation) + " · " + strings.Title(weakness) + " mutation"
		s.Mutation = map[string]string{"originality": "Distinctive artifact injected", "clarity": "Narrative simplified", "trust": "Proof layer reinforced", "conversion": "Action path intensified"}[weakness]
		mutationDecision := Decision{Label: "Mutation", Value: strings.Title(weakness), Confidence: .94, Group: "Signals"}
		replaced := false
		for index := range s.Decisions {
			if s.Decisions[index].Label == "Mutation" {
				s.Decisions[index] = mutationDecision
				replaced = true
				break
			}
		}
		if !replaced {
			s.Decisions = append(s.Decisions, mutationDecision)
		}
		s.Title = strings.ReplaceAll(s.Title, "Before everyone else.", "While it still matters.")
		s.ShowStats = s.ShowStats || weakness == "trust"
		s.ShowLogos = s.ShowLogos || weakness == "trust"
		s.ShowPricing = s.ShowPricing || weakness == "conversion"
		if weakness == "clarity" {
			s.Features, s.Density = "columns", "airy"
		}
		if weakness == "originality" {
			s.Motion = "kinetic"
			if s.Strategy == "clarity" {
				s.Features = "timeline"
			}
		}
		if weakness == "conversion" {
			s.Navigation, s.CTA = "action", "trial"
		}
		s.Blueprint = evolveBlueprint(s.Blueprint, weakness, result.Generation, *s)
		result.MutationLog = append(result.MutationLog, s.Name+": "+s.Mutation)
	}
	return result
}

func LocalCritic(specs []PageSpec) Answers {
	answers := Answers{}
	for _, spec := range specs {
		s := localScorecard(spec)
		for key, value := range map[string]int{"originality": s.Originality, "clarity": s.Clarity, "trust": s.Trust, "conversion": s.Conversion} {
			answers[spec.ID+"_"+key] = Answer{Type: "score", Score: float64(value) * 4 / 100, Confidence: .78}
		}
	}
	return answers
}

func localScorecard(s PageSpec) Scorecard {
	base := map[string]Scorecard{"clarity": {72, 91, 77, 84, 0}, "impact": {94, 66, 69, 79, 0}, "trust": {70, 82, 95, 76, 0}}[s.Strategy]
	if s.Generation > 1 {
		base.Originality += 4
		base.Clarity += 4
		base.Trust += 4
		base.Conversion += 4
	}
	return base
}
func composite(s PageSpec) int {
	w := map[string][4]float64{"clarity": {.15, .4, .25, .2}, "impact": {.4, .18, .17, .25}, "trust": {.12, .23, .45, .2}}[s.Strategy]
	return int(math.Round(float64(s.Scores.Originality)*w[0] + float64(s.Scores.Clarity)*w[1] + float64(s.Scores.Trust)*w[2] + float64(s.Scores.Conversion)*w[3]))
}
func weakest(s Scorecard) string {
	values := map[string]int{"originality": s.Originality, "clarity": s.Clarity, "trust": s.Trust, "conversion": s.Conversion}
	keys := []string{"originality", "clarity", "trust", "conversion"}
	sort.SliceStable(keys, func(i, j int) bool { return values[keys[i]] < values[keys[j]] })
	return keys[0]
}
func winner(specs []PageSpec) int {
	best := 0
	for i := range specs {
		if specs[i].Scores.Composite > specs[best].Scores.Composite {
			best = i
		}
	}
	return best
}
func scorePercent(a Answer, fallback int) int {
	if a.Type == "" && a.Score == 0 {
		return fallback
	}
	value := int(math.Round(a.Score * 25))
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
func dist(keys ...string) map[string]float64 {
	out := map[string]float64{}
	values := []float64{.68, .21, .11}
	for i, key := range keys {
		out[key] = values[i%3]
	}
	return out
}
func humanLabel(value string) string {
	if value == "" {
		return "Adaptive"
	}
	return strings.Title(strings.ReplaceAll(value, "_", " "))
}
func scoreLabel(value float64) string {
	labels := []string{"Quiet", "Restrained", "Distinct", "Dramatic", "Extreme"}
	index := int(math.Round(value))
	if index < 0 {
		index = 0
	}
	if index > 4 {
		index = 4
	}
	return labels[index]
}
func yesNo(value bool) string {
	if value {
		return "Included"
	}
	return "Omitted"
}
func pick(value, fallback string, allowed ...string) string {
	for _, v := range allowed {
		if value == v {
			return value
		}
	}
	return fallback
}
func conf(a Answer, fallback float64) float64 {
	if a.Confidence <= 0 {
		return fallback
	}
	if a.Confidence < .35 {
		return .35
	}
	if a.Confidence > .99 {
		return .99
	}
	return a.Confidence
}
func ternary[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}
func ternaryFloat(condition bool, a, b float64) float64 {
	if condition {
		return a
	}
	return b
}
