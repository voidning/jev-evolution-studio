package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx         context.Context
	mu          sync.RWMutex
	server      *http.Server
	listener    net.Listener
	previewURL  string
	token       string
	current     DesignResult
	subscribers map[chan []byte]struct{}
}

func NewApp() *App { return &App{subscribers: make(map[chan []byte]struct{})} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.token = randomToken()
	initial := CompileConcepts(LocalAnswers(defaultPrompt), "local", 0)
	initial = applyBlueprints(initial, GenerateBlueprints(defaultPrompt, initial.Specs), "procedural-search", "jev-grammar")
	a.current = ApplyCritic(initial, LocalCritic(initial.Specs))
	if err := a.startPreviewServer(); err != nil {
		runtime.LogErrorf(ctx, "preview server: %v", err)
	}
}

func (a *App) shutdown(ctx context.Context) {
	if a.server != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = a.server.Shutdown(shutdownCtx)
	}
}

func (a *App) GetState() AppState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return AppState{PreviewURL: a.previewURL, Result: a.current, HasAPIKey: loadAPIKey() != ""}
}

func (a *App) OpenStage() string {
	a.mu.RLock()
	url := a.previewURL
	a.mu.RUnlock()
	if url != "" {
		runtime.BrowserOpenURL(a.ctx, url)
	}
	return url
}

func (a *App) Generate(prompt string) DesignResult {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		prompt = defaultPrompt
	}
	started := time.Now()
	a.mu.RLock()
	currentSpecs := append([]PageSpec(nil), a.current.Specs...)
	a.mu.RUnlock()
	type jevResult struct {
		answers Answers
		err     error
	}
	jevChannel := make(chan jevResult, 1)
	go func() {
		answers, err := AskJev(prompt, currentSpecs)
		jevChannel <- jevResult{answers: answers, err: err}
	}()
	jevOutput := <-jevChannel
	answers := jevOutput.answers
	mode := "jev"
	if jevOutput.err != nil {
		answers = LocalAnswers(prompt)
		mode = "local"
		runtime.LogDebugf(a.ctx, "Jev fallback: %v", jevOutput.err)
	}
	result := CompileConcepts(answers, mode, time.Since(started).Milliseconds())
	result.Prompt = prompt
	result = applyBlueprints(result, GenerateBlueprints(prompt, result.Specs), "procedural-search", "jev-grammar")
	result.Mode += "+procedural-ast"
	a.broadcastEvent("tournament", result)
	critique, critiqueErr := AskJevCritic(prompt, result.Specs)
	if critiqueErr != nil {
		critique = LocalCritic(result.Specs)
	}
	result = ApplyCritic(result, critique)
	a.broadcastEvent("critique", result)
	result = EvolveConcepts(result)
	a.broadcastEvent("mutation", result)
	critique, critiqueErr = AskJevCritic(prompt, result.Specs)
	if critiqueErr != nil {
		critique = LocalCritic(result.Specs)
	}
	result = ApplyCritic(result, critique)
	result.LatencyMS = time.Since(started).Milliseconds()
	a.mu.Lock()
	a.current = result
	a.mu.Unlock()
	a.broadcastEvent("finale", result)
	a.broadcast(result.Specs[result.Winner])
	return result
}

func (a *App) Evolve() DesignResult {
	started := time.Now()
	a.mu.RLock()
	result := a.current
	a.mu.RUnlock()
	generationSeed := fmt.Sprintf("%s / generation %d", result.Prompt, result.Generation+1)
	result = applyBlueprints(result, GenerateBlueprints(generationSeed, result.Specs), "procedural-search", "jev-grammar")
	result = EvolveConcepts(result)
	a.broadcastEvent("mutation", result)
	critique, err := AskJevCritic(result.Prompt, result.Specs)
	if err != nil {
		critique = LocalCritic(result.Specs)
	}
	result = ApplyCritic(result, critique)
	result.LatencyMS = time.Since(started).Milliseconds()
	a.mu.Lock()
	a.current = result
	a.mu.Unlock()
	a.broadcastEvent("finale", result)
	a.broadcast(result.Specs[result.Winner])
	return result
}

func (a *App) SelectConcept(index int) PageSpec {
	a.mu.RLock()
	if index < 0 || index >= len(a.current.Specs) {
		index = 0
	}
	spec := a.current.Specs[index]
	a.mu.RUnlock()
	a.broadcast(spec)
	return spec
}

func (a *App) broadcast(spec PageSpec) {
	a.broadcastEvent("pagespec", spec)
}

func (a *App) broadcastEvent(name string, value any) {
	payload, _ := json.Marshal(value)
	frame := []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", name, payload))
	a.mu.RLock()
	channels := make([]chan []byte, 0, len(a.subscribers))
	for ch := range a.subscribers {
		channels = append(channels, ch)
	}
	a.mu.RUnlock()
	for _, ch := range channels {
		select {
		case ch <- frame:
		default:
		}
	}
}

func (a *App) startPreviewServer() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	a.listener = listener
	port := listener.Addr().(*net.TCPAddr).Port
	a.previewURL = fmt.Sprintf("http://127.0.0.1:%d/?token=%s", port, a.token)
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handlePreview)
	mux.HandleFunc("/events", a.handleEvents)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })
	a.server = &http.Server{Handler: mux, ReadHeaderTimeout: 3 * time.Second}
	go func() {
		if err := a.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			runtime.LogErrorf(a.ctx, "preview serve: %v", err)
		}
	}()
	return nil
}

func (a *App) validToken(r *http.Request) bool { return r.URL.Query().Get("token") == a.token }

func (a *App) handlePreview(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" || !a.validToken(r) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(previewDocument()))
}

func (a *App) handleEvents(w http.ResponseWriter, r *http.Request) {
	if !a.validToken(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ch := make(chan []byte, 4)
	a.mu.Lock()
	a.subscribers[ch] = struct{}{}
	initial, _ := json.Marshal(a.current.Specs[0])
	a.mu.Unlock()
	defer func() { a.mu.Lock(); delete(a.subscribers, ch); close(ch); a.mu.Unlock() }()
	fmt.Fprintf(w, "event: pagespec\ndata: %s\n\n", initial)
	flusher.Flush()
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case data := <-ch:
			_, _ = w.Write(data)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

func randomToken() string { b := make([]byte, 18); _, _ = rand.Read(b); return hex.EncodeToString(b) }

func loadAPIKey() string {
	if value := strings.TrimSpace(os.Getenv("TYPESAFE_API_KEY")); value != "" {
		return value
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
			if len(parts) == 2 && parts[0] == "TYPESAFE_API_KEY" {
				return strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			}
		}
	}
	return ""
}

func apiKeyPaths(executable string) []string {
	developmentPaths := []string{".env.local", "web-prototype/.env.local"}
	if executable == "" {
		return developmentPaths
	}
	executableDir := filepath.Dir(executable)
	contentsDir := filepath.Dir(executableDir)
	appBundle := filepath.Dir(contentsDir)
	if filepath.Base(executableDir) != "MacOS" || filepath.Base(contentsDir) != "Contents" || !strings.HasSuffix(filepath.Base(appBundle), ".app") {
		return append([]string{filepath.Join(executableDir, ".env.local")}, developmentPaths...)
	}
	projectRoot := filepath.Clean(filepath.Join(executableDir, "..", "..", "..", "..", ".."))
	return []string{
		filepath.Join(projectRoot, ".env.local"),
		filepath.Join(projectRoot, "web-prototype", ".env.local"),
		filepath.Join(executableDir, ".env.local"),
		developmentPaths[0],
		developmentPaths[1],
	}
}
