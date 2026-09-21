package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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
		if len(blueprint.Sections) < 5 {
			t.Fatalf("blueprint %s is too shallow: %d sections", blueprint.ID, len(blueprint.Sections))
		}
		parts := make([]string, 0, len(blueprint.Sections))
		for _, section := range blueprint.Sections {
			parts = append(parts, section.Kind+"/"+section.Layout+"/"+section.Visual)
			foundNested = foundNested || len(section.Children) > 0
		}
		signature := strings.Join(parts, "|")
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
	before := len(result.Specs[0].Blueprint.Sections)
	result = EvolveConcepts(result)
	if len(result.Specs[0].Blueprint.Sections) <= before {
		t.Fatal("originality mutation did not grow the page AST")
	}
}

func TestGeneratorUsesStructuredPageAST(t *testing.T) {
	seed := CompileConcepts(LocalAnswers("ocean exploration"), "local", 0)
	want := fallbackBlueprints(seed.Specs)
	encoded, err := json.Marshal(map[string]any{"blueprints": want})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected authorization header")
		}
		body, _ := io.ReadAll(r.Body)
		var request map[string]any
		if err := json.Unmarshal(body, &request); err != nil {
			t.Errorf("invalid request: %v", err)
		}
		text, _ := request["text"].(map[string]any)
		format, _ := text["format"].(map[string]any)
		if format["type"] != "json_schema" || format["strict"] != true {
			t.Errorf("generator did not request strict structured output: %v", format)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"output": []any{map[string]any{"content": []any{map[string]any{"type": "output_text", "text": string(encoded)}}}}})
	}))
	defer server.Close()
	t.Setenv("GENERATOR_API_KEY", "test-key")
	t.Setenv("GENERATOR_BASE_URL", server.URL)
	t.Setenv("GENERATOR_MODEL", "test-fast-model")
	blueprints, model, err := GenerateBlueprints("ocean exploration", seed.Specs)
	if err != nil {
		t.Fatal(err)
	}
	if model != "test-fast-model" || len(blueprints) != 3 || blueprints[0].Version != 2 {
		t.Fatalf("unexpected generator output: model=%s blueprints=%+v", model, blueprints)
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
	if !strings.Contains(string(page), "renderGenerated") || !strings.Contains(string(page), "GENERATED PAGE AST") {
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
