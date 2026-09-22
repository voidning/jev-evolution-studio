package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFallbackBlueprintsAreRecursiveAndStructurallyDistinct(t *testing.T) {
	result := CompileConcepts(LocalAnswers("为深海探险生成一套完全不同的互动网页"), "local", 0)
	blueprints := fallbackBlueprints(result.Specs)
	if len(blueprints) != 3 {
		t.Fatalf("expected three blueprints, got %d", len(blueprints))
	}
	sequences := map[string]bool{}
	foundNested := false
	for _, blueprint := range blueprints {
		if topFrameCount(blueprint) < 5 {
			t.Fatalf("blueprint %s is too shallow: %d frames", blueprint.ID, topFrameCount(blueprint))
		}
		signature := blueprintSignature(blueprint)
		foundNested = foundNested || inspectBlueprint(blueprint).maxDepth >= 3
		if sequences[signature] {
			t.Fatalf("duplicate AST sequence: %s", signature)
		}
		sequences[signature] = true
	}
	if !foundNested {
		t.Fatal("expected at least one recursive section composition")
	}
	result = applyBlueprints(result, blueprints, "local-grammar", "local-grammar")
	result.Specs[0].Scores = Scorecard{Originality: 10, Clarity: 90, Trust: 90, Conversion: 90}
	before := topFrameCount(result.Specs[0].Blueprint)
	result = EvolveConcepts(result)
	if topFrameCount(result.Specs[0].Blueprint) <= before {
		t.Fatal("originality mutation did not grow the page AST")
	}
}

func TestProceduralSearchIsDeterministicAndPromptSensitive(t *testing.T) {
	seed := CompileConcepts(LocalAnswers("ocean exploration"), "local", 0)
	first := GenerateBlueprints("ocean exploration", seed.Specs)
	repeat := GenerateBlueprints("ocean exploration", seed.Specs)
	second := GenerateBlueprints("quiet ocean research", seed.Specs)
	if len(first) != 3 || len(repeat) != 3 || len(second) != 3 {
		t.Fatalf("expected three procedural winners")
	}
	signature := func(blueprints []PageBlueprint) string {
		parts := []string{}
		for _, blueprint := range blueprints {
			parts = append(parts, blueprintSignature(blueprint))
		}
		return strings.Join(parts, "|")
	}
	if signature(first) != signature(repeat) {
		t.Fatal("same prompt did not reproduce the same AST")
	}
	if signature(first) == signature(second) {
		t.Fatal("different prompts collapsed to the same AST")
	}
	for _, blueprint := range first {
		if topFrameCount(blueprint) < 5 || !strings.Contains(blueprint.CreativeDirection, "genome") {
			t.Fatalf("procedural winner is incomplete: %+v", blueprint)
		}
	}
}

func TestCompositionalGrammarDoesNotCollapseToThreeSkeletons(t *testing.T) {
	prompts := []string{
		"deep ocean expedition", "quiet biotech laboratory", "kinetic music platform", "trusted banking infrastructure",
		"playful robotics school", "cinematic space observatory", "minimal developer database", "warm creative studio",
	}
	signatures := map[string]bool{}
	for _, prompt := range prompts {
		seed := CompileConcepts(LocalAnswers(prompt), "local", 0)
		blueprints := GenerateBlueprints(prompt, seed.Specs)
		for _, blueprint := range blueprints {
			signatures[blueprintSignature(blueprint)] = true
		}
	}
	if len(signatures) < 20 {
		t.Fatalf("grammar collapsed to too few structures: got %d unique signatures from 24 pages", len(signatures))
	}
}

func TestTournamentWinnersSpanPageLengths(t *testing.T) {
	seed := CompileConcepts(LocalAnswers("cinematic live research system"), "local", 0)
	blueprints := GenerateBlueprints("cinematic live research system", seed.Specs)
	lengths := map[int]bool{}
	for _, blueprint := range blueprints {
		lengths[topFrameCount(blueprint)] = true
	}
	if len(lengths) < 3 || !lengths[5] || !lengths[6] || !lengths[7] {
		t.Fatalf("expected continuous lenses to select 5, 6 and 7 frame pages, got %v", lengths)
	}
}

func topFrameCount(blueprint PageBlueprint) int {
	count := 0
	for _, node := range blueprint.Root.Children {
		if node.Primitive == "frame" {
			count++
		}
	}
	return count
}

func blueprintSignature(blueprint PageBlueprint) string {
	parts := []string{}
	var walk func(DesignNode, int)
	walk = func(node DesignNode, depth int) {
		parts = append(parts, fmt.Sprintf("%d:%s/%s/%d/%s/%s", depth, node.Primitive, node.Role, node.Layout.Columns, node.Visual.Kind, node.Layout.Align))
		for _, child := range node.Children {
			walk(child, depth+1)
		}
	}
	walk(blueprint.Root, 0)
	return strings.Join(parts, "|")
}

func TestPrimitiveRegistryConstrainsEveryGeneratedNode(t *testing.T) {
	if len(primitiveRegistry) < 10 {
		t.Fatalf("primitive registry is unexpectedly small: %d", len(primitiveRegistry))
	}
	result := CompileConcepts(LocalAnswers("cinematic ocean research"), "local", 0)
	for _, blueprint := range GenerateBlueprints("cinematic ocean research", result.Specs) {
		stats := inspectBlueprint(blueprint)
		if stats.invalid != 0 {
			t.Fatalf("blueprint %s violates primitive registry: %+v", blueprint.ID, stats)
		}
		if stats.maxDepth < 3 {
			t.Fatalf("blueprint %s is not meaningfully recursive", blueprint.ID)
		}
	}
}

func BenchmarkGenerateBlueprints(b *testing.B) {
	seed := CompileConcepts(LocalAnswers("cinematic ocean research"), "local", 0)
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		GenerateBlueprints("cinematic ocean research", seed.Specs)
	}
}

func TestPackagedAppFindsProjectAPIKey(t *testing.T) {
	executable := "/Users/demo/Desktop/jev-ui-studio-wails/build/bin/ForgeDesktop.app/Contents/MacOS/jev-ui-studio-wails"
	wantRoot := filepath.Clean("/Users/demo/Desktop/jev-ui-studio-wails/.env.local")
	wantSnapshot := filepath.Clean("/Users/demo/Desktop/jev-ui-studio-wails/web-prototype/.env.local")
	paths := apiKeyPaths(executable)
	if len(paths) < 4 {
		t.Fatalf("not enough packaged app candidates: %v", paths)
	}
	if paths[0] != wantRoot || paths[1] != wantSnapshot {
		t.Fatalf("project keys must precede working-directory keys: %v", paths)
	}
	for index, path := range paths {
		if (path == ".env.local" || path == "web-prototype/.env.local") && index < 2 {
			t.Fatalf("working-directory key has unsafe priority: %v", paths)
		}
	}
}

func TestSemanticWorldAndUniverseStructuresSurviveEvolution(t *testing.T) {
	result := CompileConcepts(LocalAnswers("为深海生物发光探险做主页，强调实时下潜与科研可信度"), "local", 0)
	if result.SwarmSize < 19 || result.Specs[0].World != "ocean" || result.Specs[0].Brand != "Abyssal" {
		t.Fatalf("ocean world was not compiled: %+v", result.Specs[0])
	}
	assertDistinct := func(specs []PageSpec) {
		t.Helper()
		seen := map[string]bool{}
		for _, spec := range specs {
			signature := spec.Hero + "/" + spec.Visual + "/" + spec.Features
			if seen[signature] {
				t.Fatalf("universe structure collapsed to %q", signature)
			}
			seen[signature] = true
		}
	}
	assertDistinct(result.Specs)
	result = ApplyCritic(result, LocalCritic(result.Specs))
	result = EvolveConcepts(result)
	result = EvolveConcepts(result)
	assertDistinct(result.Specs)
	for _, spec := range result.Specs {
		mutations := 0
		for _, decision := range spec.Decisions {
			if decision.Label == "Mutation" {
				mutations++
			}
		}
		if mutations != 1 {
			t.Fatalf("expected one mutation node, got %d", mutations)
		}
	}
}

func TestPreviewServerStreamsConceptChanges(t *testing.T) {
	app := NewApp()
	app.token = randomToken()
	app.current = CompileConcepts(LocalAnswers(defaultPrompt), "local", 0)
	if err := app.startPreviewServer(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = app.server.Shutdown(ctx)
	})

	response, err := http.Get(app.previewURL)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("preview returned %s", response.Status)
	}
	page, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), "renderGenerated") || !strings.Contains(string(page), "GENERATED DESIGN AST") {
		t.Fatal("preview did not include the recursive AST renderer")
	}

	badResponse, err := http.Get(strings.Split(app.previewURL, "?")[0])
	if err != nil {
		t.Fatal(err)
	}
	badResponse.Body.Close()
	if badResponse.StatusCode != http.StatusNotFound {
		t.Fatalf("preview without token returned %s", badResponse.Status)
	}

	eventsURL := strings.Replace(app.previewURL, "/?", "/events?", 1)
	request, err := http.NewRequest(http.MethodGet, eventsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	stream, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Body.Close()
	reader := bufio.NewReader(stream.Body)

	readEvent := func() string {
		t.Helper()
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(line, "data: ") {
				return strings.TrimSpace(strings.TrimPrefix(line, "data: "))
			}
		}
	}

	if initial := readEvent(); !strings.Contains(initial, `"id":"signal"`) {
		t.Fatalf("unexpected initial event: %s", initial)
	}
	app.SelectConcept(1)
	if update := readEvent(); !strings.Contains(update, `"id":"pulse"`) {
		t.Fatalf("unexpected concept update: %s", update)
	}
}

func TestJevLive(t *testing.T) {
	if os.Getenv("RUN_LIVE_JEV") != "1" {
		t.Skip("set RUN_LIVE_JEV=1 to exercise the configured TypeSafe account")
	}
	if loadAPIKey() == "" {
		t.Fatal("TYPESAFE_API_KEY is not configured")
	}
	answers, err := AskJev("做一个温暖、留白很多、不显示定价的 AI 编辑器主页", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"archetype", "goal", "tone", "theme", "hero", "features", "density", "pricing"} {
		if _, ok := answers[key]; !ok {
			t.Errorf("missing Jev decision %q", key)
		}
	}
	if len(answers) < 18 {
		t.Fatalf("expected decision swarm, got only %d answers", len(answers))
	}
	result := CompileConcepts(answers, "jev", 0)
	critique, err := AskJevCritic("做一个温暖、留白很多、不显示定价的 AI 编辑器主页", result.Specs)
	if err != nil {
		t.Fatal(err)
	}
	result = ApplyCritic(result, critique)
	result = EvolveConcepts(result)
	secondCritique, err := AskJevCritic(result.Prompt, result.Specs)
	if err != nil {
		t.Fatal(err)
	}
	result = ApplyCritic(result, secondCritique)
	if result.Generation != 2 || len(result.MutationLog) != 3 || result.Specs[result.Winner].Scores.Composite == 0 {
		t.Fatalf("unexpected evolution result: generation=%d mutations=%d winner=%+v", result.Generation, len(result.MutationLog), result.Specs[result.Winner])
	}
}
