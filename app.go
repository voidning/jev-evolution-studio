package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	draft       *Draft
	ctx         context.Context
	mu          sync.Mutex
	op          sync.Mutex
	state       EditorState
	token       string
	server      *http.Server
	baseURL     string
	upstream    string
	meta        ProjectMeta
	process     *exec.Cmd
	snapshot    *Snapshot
	command     *Command
	response    chan BrowserReport
	lastBrowser time.Time
	headless    bool
	offline     bool
}

func NewApp() *App {
	return &App{ctx: context.Background(), token: randomToken(), offline: os.Getenv("JEV_OFFLINE") == "1"}
}
func randomToken() string {
	b := make([]byte, 24)
	if _, e := rand.Read(b); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b)
}
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if e := a.startServer(); e != nil {
		a.mu.Lock()
		a.state.Error = e.Error()
		a.mu.Unlock()
	}
}
func (a *App) shutdown(_ context.Context) {
	a.stopProcess()
	if a.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = a.server.Shutdown(ctx)
	}
}
func (a *App) stopProcess() {
	if a.process != nil && a.process.Process != nil {
		_ = a.process.Process.Kill()
		a.process = nil
	}
}
func (a *App) GetState() EditorState {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.state
	s.BrowserConnected = time.Since(a.lastBrowser) < 3*time.Second
	s.HasAPIKey = !a.offline && loadAPIKey() != ""
	return s
}
func (a *App) ChooseProject() (EditorState, error) {
	p, e := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "选择已接入 Jev 的 React / Vite 项目"})
	if e != nil {
		return a.GetState(), e
	}
	if p == "" {
		return a.GetState(), nil
	}
	return a.OpenProject(p)
}
func (a *App) OpenProject(path string) (EditorState, error) {
	if !a.op.TryLock() {
		return a.GetState(), fmt.Errorf("正在处理修改")
	}
	defer a.op.Unlock()
	if a.GetState().Pending {
		return a.GetState(), fmt.Errorf("请先 Accept 或 Undo 当前修改")
	}
	root, e := filepath.Abs(path)
	if e != nil {
		return a.GetState(), e
	}
	root, e = filepath.EvalSymlinks(root)
	if e != nil {
		return a.GetState(), e
	}
	data, e := os.ReadFile(filepath.Join(root, "package.json"))
	if e != nil {
		return a.GetState(), fmt.Errorf("请选择 React / Vite 项目")
	}
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return a.GetState(), fmt.Errorf("无效 package.json")
	}
	if pkg.Dependencies["react"] == "" && pkg.DevDependencies["react"] == "" {
		return a.GetState(), fmt.Errorf("仅支持 React")
	}
	if pkg.Dependencies["vite"] == "" && pkg.DevDependencies["vite"] == "" {
		return a.GetState(), fmt.Errorf("仅支持 Vite")
	}
	if pkg.Dependencies["tailwindcss"] == "" && pkg.DevDependencies["tailwindcss"] == "" {
		if _, e = patchPath(root); e != nil {
			return a.GetState(), e
		}
	}
	snap, e := readSnapshot(root)
	if e != nil {
		return a.GetState(), e
	}
	a.stopProcess()
	a.mu.Lock()
	a.upstream = ""
	a.meta = ProjectMeta{}
	a.lastBrowser = time.Time{}
	a.snapshot = snap
	a.state = EditorState{Project: root, Session: randomToken(), Mode: "offline", Selecting: true}
	a.mu.Unlock()
	if snap != nil {
		files := snapshotFiles(snap)
		matches := true
		for _, f := range files {
			b, e := fileBytes(root, f.Path)
			if e != nil || string(b) != f.After {
				matches = false
			}
		}
		a.mu.Lock()
		if matches {
			a.state.Pending = snap.Pending && len(snap.Files) == 0
			a.state.CanUndo = true
			a.state.Diff = sourceDiff(files)
			a.state.ChangeID = snap.ID
			for _, f := range files {
				a.state.Files = append(a.state.Files, f.Path)
			}
		} else {
			a.state.Error = "撤销记录与当前文件不一致；外部文件未覆盖"
		}
		a.mu.Unlock()
	}

	return a.GetState(), nil
}
func cleanEnvironment(env []string) []string {
	out := []string{}
	for _, v := range env {
		key := strings.ToUpper(strings.SplitN(v, "=", 2)[0])
		if strings.Contains(key, "TYPESAFE") || strings.Contains(key, "API_KEY") || strings.HasPrefix(key, "JEV_") {
			continue
		}
		out = append(out, v)
	}
	return out
}
func (a *App) StartPreview() (EditorState, error) {
	if !a.op.TryLock() {
		return a.GetState(), fmt.Errorf("正在处理修改")
	}
	defer a.op.Unlock()
	s := a.GetState()
	if s.Project == "" {
		return s, fmt.Errorf("请先选择项目")
	}
	if s.Connected {
		a.mu.Lock()
		upstream := a.upstream
		a.mu.Unlock()
		if meta, e := fetchMeta(upstream); e == nil && meta.Root == s.Project {
			return s, nil
		}
		a.stopProcess()
		a.mu.Lock()
		a.state.Connected = false
		a.mu.Unlock()
	}
	entry := filepath.Join(s.Project, "node_modules", "vite", "bin", "vite.js")
	if _, e := os.Stat(entry); e != nil {
		return s, fmt.Errorf("缺少 Vite 依赖，请先在项目中运行 npm install")
	}
	l, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		return s, e
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	cmd := exec.Command("node", entry, "--host", "127.0.0.1", "--port", fmt.Sprint(port), "--strictPort")
	cmd.Dir = s.Project
	cmd.Env = cleanEnvironment(os.Environ())
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if e = cmd.Start(); e != nil {
		return s, fmt.Errorf("无法启动 Vite：%w", e)
	}
	a.process = cmd
	go func() { _ = cmd.Wait() }()
	upstream := fmt.Sprintf("http://127.0.0.1:%d", port)
	var last error
	for deadline := time.Now().Add(12 * time.Second); time.Now().Before(deadline); {
		if last = a.connect(upstream); last == nil {
			return a.GetState(), nil
		}
		time.Sleep(120 * time.Millisecond)
	}
	a.stopProcess()
	return a.GetState(), fmt.Errorf("Vite 接入失败：请检查 integration/jev-vite.mjs 配置。%v", last)
}
func (a *App) ConnectPreview(raw string) (EditorState, error) {
	if !a.op.TryLock() {
		return a.GetState(), fmt.Errorf("正在处理修改")
	}
	defer a.op.Unlock()
	if a.GetState().Pending {
		return a.GetState(), fmt.Errorf("请先处理当前修改")
	}
	e := a.connect(raw)
	return a.GetState(), e
}
func (a *App) connect(raw string) error {
	u, e := url.Parse(raw)
	if e != nil || u.Scheme != "http" || u.Hostname() != "127.0.0.1" || u.Port() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return fmt.Errorf("仅接受 http://127.0.0.1:端口")
	}
	s := a.GetState()
	if s.Project == "" {
		return fmt.Errorf("请先选择项目")
	}
	meta, e := fetchMeta(strings.TrimSuffix(raw, "/"))
	if e != nil {
		return e
	}
	if meta.Root != s.Project || meta.Version != 1 || len(meta.Targets) == 0 {
		return fmt.Errorf("预览项目身份不匹配或没有可编辑元素")
	}
	ids := map[string]bool{}
	for _, t := range meta.Targets {
		if ids[t.ID] || !idRE.MatchString(t.ID) {
			return fmt.Errorf("目标 ID 非法或重复")
		}
		ids[t.ID] = true
		if e = validateSource(s.Project, t); e != nil {
			return e
		}
	}
	a.mu.Lock()
	if a.upstream != strings.TrimSuffix(raw, "/") {
		a.state.Session = randomToken()
		a.state.Selected = nil
		a.state.ClientID = ""
		a.lastBrowser = time.Time{}
	}
	a.upstream = strings.TrimSuffix(raw, "/")
	a.meta = meta
	a.state.Executor = "css"
	if meta.Tailwind {
		a.state.Executor = "tailwind"
	}
	a.state.Connected = true
	a.state.PreviewURL = a.baseURL + "/?token=" + a.token
	a.mu.Unlock()
	return nil
}
func fetchMeta(upstream string) (ProjectMeta, error) {
	var m ProjectMeta
	client := &http.Client{Timeout: time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	res, e := client.Get(upstream + "/__jev/meta")
	if e != nil {
		return m, fmt.Errorf("Vite 尚未就绪")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return m, fmt.Errorf("缺少 Jev Vite 插件")
	}
	e = json.NewDecoder(io.LimitReader(res.Body, 1024*1024)).Decode(&m)
	return m, e
}
func validateSource(root string, t Target) error {
	if filepath.IsAbs(t.Source) || !inside(root, filepath.Join(root, t.Source)) {
		return fmt.Errorf("非法源文件路径")
	}
	real, e := filepath.EvalSymlinks(filepath.Join(root, t.Source))
	if e != nil || !inside(root, real) {
		return fmt.Errorf("源文件不在项目内")
	}
	return nil
}
func (a *App) OpenStage() string {
	s := a.GetState()
	if s.PreviewURL != "" && !a.headless {
		runtime.BrowserOpenURL(a.ctx, s.PreviewURL)
	}
	return s.PreviewURL
}
func (a *App) SetSelecting(on bool) EditorState {
	a.mu.Lock()
	a.state.Selecting = on
	a.mu.Unlock()
	return a.GetState()
}
func (a *App) SetOffline(on bool) (EditorState, error) {
	if !a.op.TryLock() {
		return a.GetState(), fmt.Errorf("正在处理修改")
	}
	defer a.op.Unlock()
	a.mu.Lock()
	a.offline = on
	a.mu.Unlock()
	return a.GetState(), nil
}
func (a *App) stage(name string, start time.Time) {
	a.mu.Lock()
	a.state.Stages = append(a.state.Stages, Stage{name, time.Since(start).Milliseconds()})
	a.mu.Unlock()
}
func (a *App) browserCommand(clientID, kind string, ids []string, hash string, expectations ...Expectation) (BrowserReport, error) {
	cmd := &Command{ClientID: clientID, ID: randomToken(), Kind: kind, Targets: ids, Hash: hash, Expectations: expectations}
	ch := make(chan BrowserReport, 1)
	a.mu.Lock()
	a.command = cmd
	a.response = ch
	a.mu.Unlock()
	defer func() { a.mu.Lock(); a.command = nil; a.response = nil; a.mu.Unlock() }()
	select {
	case r := <-ch:
		if r.Error != "" {
			return r, errors.New(r.Error)
		}
		return r, nil
	case <-time.After(5 * time.Second):
		return BrowserReport{}, fmt.Errorf("浏览器未确认 %s；请保持预览连接", kind)
	}
}
func resolveTargets(root, id string, intents []EditIntent, meta ProjectMeta, live []Target) ([]Target, error) {
	indexed := map[string]Target{}
	for _, t := range meta.Targets {
		if _, ok := indexed[t.ID]; ok {
			return nil, fmt.Errorf("源代码 ID 重复")
		}
		if e := validateSource(root, t); e != nil {
			return nil, e
		}
		indexed[t.ID] = t
	}
	selected := Target{}
	for _, t := range live {
		if t.ID == id {
			selected = t
		}
	}
	if selected.ID == "" {
		return nil, fmt.Errorf("所选目标已消失")
	}
	scope := intents[0].Scope
	ids := []string{id}
	if scope == "container" {
		if selected.Parent == "" {
			return nil, fmt.Errorf("没有已标注的父容器")
		}
		ids = []string{selected.Parent}
	}
	if scope == "siblings" {
		if selected.Parent == "" {
			return nil, fmt.Errorf("无法确定同级元素")
		}
		ids = nil
		for _, t := range live {
			if t.Parent == selected.Parent {
				ids = append(ids, t.ID)
			}
		}
	}
	var targets []Target
	for _, id := range ids {
		found := false
		for _, t := range live {
			if t.ID != id {
				continue
			}
			source, ok := indexed[id]
			if !ok || source.Tag != t.Tag || t.Count != 1 {
				return nil, fmt.Errorf("目标 ID、标签或源代码不一致")
			}
			t.Source = source.Source
			t.Line = source.Line
			t.Revision = source.Revision
			t.Fingerprint = source.Fingerprint
			t.Start = source.Start
			t.End = source.End
			t.ClassKind = source.ClassKind
			t.ClassName = source.ClassName
			t.Repeated = source.Repeated
			t.Empty = source.Empty
			for _, in := range intents {
				if in.Scope != scope {
					return nil, unsupported
				}
				if e := validateIntent(in, t); e != nil {
					return nil, e
				}
			}
			targets = append(targets, t)
			found = true
			break
		}
		if !found {
			return nil, fmt.Errorf("目标不存在")
		}
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("没有可编辑目标")
	}
	return targets, nil
}
func (a *App) startServer() error {
	l, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		return e
	}
	a.baseURL = "http://" + l.Addr().String()
	a.server = &http.Server{Handler: http.HandlerFunc(a.serve), ReadHeaderTimeout: 3 * time.Second}
	go func() { _ = a.server.Serve(l) }()
	return nil
}
func (a *App) authorized(r *http.Request) bool {
	c, e := r.Cookie("jev_session")
	return e == nil && c.Value == a.token
}
func (a *App) serve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if r.Host != strings.TrimPrefix(a.baseURL, "http://") {
		http.Error(w, "invalid host", 403)
		return
	}
	if (r.URL.Path == "/" || r.URL.Path == "/console/") && r.URL.Query().Get("token") == a.token {
		http.SetCookie(w, &http.Cookie{Name: "jev_session", Value: a.token, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode})
		http.Redirect(w, r, r.URL.Path, 303)
		return
	}
	if !a.authorized(r) {
		http.Error(w, "unauthorized", 401)
		return
	}
	if origin := r.Header.Get("Origin"); origin != "" && origin != a.baseURL {
		http.Error(w, "invalid origin", 403)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/__jev/api/") {
		a.api(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/console/") {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/console")
		http.FileServer(http.FS(frontendFS())).ServeHTTP(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/assets/") {
		http.FileServer(http.FS(frontendFS())).ServeHTTP(w, r)
		return
	}
	a.mu.Lock()
	upstream := a.upstream
	a.mu.Unlock()
	if upstream == "" {
		http.Error(w, "请先启动项目预览", 503)
		return
	}
	u, _ := url.Parse(upstream)
	proxy := httputil.NewSingleHostReverseProxy(u)
	original := proxy.Director
	proxy.Director = func(req *http.Request) {
		original(req)
		req.Host = u.Host
		req.Header.Del("Cookie")
		req.Header.Del("Authorization")
		if req.Header.Get("Origin") != "" {
			req.Header.Set("Origin", upstream)
		}
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) { http.Error(w, "Vite disconnected", 502) }
	proxy.ServeHTTP(w, r)
}
func (a *App) api(w http.ResponseWriter, r *http.Request) {
	action := strings.TrimPrefix(r.URL.Path, "/__jev/api/")
	if r.Method == "GET" && (action == "state" || action == "bridge") {
		w.Header().Set("Content-Type", "application/json")
		if action == "state" {
			_ = json.NewEncoder(w).Encode(a.GetState())
			return
		}
		a.mu.Lock()
		payload := struct {
			Project   string   `json:"project"`
			Session   string   `json:"session"`
			Selecting bool     `json:"selecting"`
			Command   *Command `json:"command"`
		}{a.state.Project, a.state.Session, a.state.Selecting, a.command}
		a.mu.Unlock()
		_ = json.NewEncoder(w).Encode(payload)
		return
	}
	if r.Method != "POST" || r.Header.Get("Content-Type") != "application/json" || r.Header.Get("Origin") != a.baseURL {
		http.Error(w, "invalid request", 403)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 256*1024)
	var raw json.RawMessage
	if json.NewDecoder(r.Body).Decode(&raw) != nil {
		http.Error(w, "invalid JSON", 400)
		return
	}
	var result any
	var err error
	switch action {
	case "report":
		var report BrowserReport
		if err = json.Unmarshal(raw, &report); err == nil {
			a.mu.Lock()
			if report.Session != a.state.Session || report.Project != a.state.Project || !idRE.MatchString(report.ClientID) {
				err = fmt.Errorf("stale session")
			} else {
				if (report.Selected != nil && report.CommandID == "") || a.state.ClientID == "" {
					a.state.ClientID = report.ClientID
				}
				if a.state.ClientID == report.ClientID {
					a.lastBrowser = time.Now()
				}
				if report.Selected != nil && (report.CommandID == "" || (a.state.ClientID == report.ClientID && a.state.Selected != nil && a.state.Selected.ID == report.Selected.ID)) {
					t := *report.Selected
					for _, source := range a.meta.Targets {
						if source.ID == t.ID {
							t.Source = source.Source
							t.Line = source.Line
						}
					}
					a.state.Selected = &t
				}
				if report.CommandID == "escape" {
					a.state.Selecting = false
				}
				if a.command != nil && report.CommandID == a.command.ID && report.ClientID == a.command.ClientID && a.response != nil {
					if a.command.Kind != "refresh" || report.Hash == a.command.Hash || report.Error != "" {
						select {
						case a.response <- report:
						default:
						}
					}
				}
			}
			a.mu.Unlock()
		}
		result = map[string]bool{"ok": err == nil}
	case "open":
		var v struct {
			Path string `json:"path"`
		}
		err = json.Unmarshal(raw, &v)
		if err == nil {
			result, err = a.OpenProject(v.Path)
		}
	case "start":
		result, err = a.StartPreview()
	case "connect":
		var v struct {
			URL string `json:"url"`
		}
		err = json.Unmarshal(raw, &v)
		if err == nil {
			result, err = a.ConnectPreview(v.URL)
		}
	case "select":
		var v struct {
			On bool `json:"on"`
		}
		err = json.Unmarshal(raw, &v)
		if err == nil {
			result = a.SetSelecting(v.On)
		}
	case "offline":
		var v struct {
			On bool `json:"on"`
		}
		err = json.Unmarshal(raw, &v)
		if err == nil {
			result, err = a.SetOffline(v.On)
		}
	case "apply":
		var v ApplyRequest
		err = json.Unmarshal(raw, &v)
		if err == nil {
			result, err = a.ApplyEdit(v)
		}
	case "accept":
		result, err = a.Accept()
	case "undo":
		result, err = a.Undo()
	case "reject":
		result, err = a.Reject()
	default:
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(400)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(result)
}
