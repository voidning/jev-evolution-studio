package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func sampleTarget(id, tag string) Target {
	return Target{ID: id, Tag: tag, Source: "src/App.jsx", Count: 1, Styles: map[string]string{"font-size": "32px", "font-weight": "400", "padding-top": "24px", "padding-right": "24px", "padding-bottom": "24px", "padding-left": "24px", "gap": "24px", "color": "rgb(255, 255, 255)", "background-color": "rgb(255, 255, 255)"}}
}
func fixture(t *testing.T) string {
	t.Helper()
	root, e := filepath.EvalSymlinks(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	_ = os.Mkdir(filepath.Join(root, "src"), 0700)
	_ = os.WriteFile(filepath.Join(root, "src/jev-edits.css"), []byte("/* initial */\n"), 0600)
	_ = os.WriteFile(filepath.Join(root, "src/App.jsx"), []byte(`<h1 data-jev-id="title">Hello</h1>`), 0600)
	_ = os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"dependencies":{"react":"19"},"devDependencies":{"vite":"7"}}`), 0600)
	t.Cleanup(func() { _ = clearSnapshot(root) })
	return root
}
func TestOfflineOperations(t *testing.T) {
	cases := []struct{ prompt, op, value, tag string }{
		{"这个标题再大一点", "fontSize", "", "h1"}, {"文字变小一点", "fontSize", "", "p"}, {"这里更紧凑", "density", "", "div"}, {"更宽松", "density", "", "div"}, {"padding增加", "padding", "", "div"}, {"内边距减少", "padding", "", "div"}, {"gap减少", "gap", "", "div"}, {"gap增加", "gap", "", "div"}, {"字重增强", "fontWeight", "", "h1"}, {"字重减弱", "fontWeight", "", "p"}, {"左对齐", "textAlign", "left", "p"}, {"文字居中", "textAlign", "center", "p"}, {"右对齐", "textAlign", "right", "p"}, {"把这块内容居中", "center", "center", "section"}, {"四列", "columns", "4", "div"}, {"横向排列", "axis", "row", "div"}, {"纵向排列", "axis", "column", "div"}, {"宽度narrow", "width", "narrow", "div"}, {"宽度normal", "width", "normal", "div"}, {"宽度wide", "width", "wide", "div"}, {"全宽", "width", "full", "div"}, {"圆角小一点", "radius", "sharp", "button"}, {"圆角适中", "radius", "medium", "button"}, {"圆角大一点", "radius", "round", "button"}, {"背景更强", "background", "stronger", "div"}, {"背景更柔和", "background", "softer", "div"}, {"背景透明", "background", "transparent", "div"}, {"显示边框", "border", "show", "div"}, {"隐藏边框", "border", "hide", "div"}, {"让这个按钮更突出", "emphasis", "primary", "button"}, {"强调normal", "emphasis", "normal", "button"}, {"弱化按钮", "emphasis", "muted", "button"}, {"隐藏元素", "visibility", "hidden", "div"}, {"恢复元素", "visibility", "visible", "div"},
	}
	for _, c := range cases {
		t.Run(c.prompt, func(t *testing.T) {
			is, e := LocalIntent(c.prompt)
			if e != nil || len(is) != 1 {
				t.Fatalf("%v %v", is, e)
			}
			if is[0].Operation != c.op || is[0].Value != c.value {
				t.Fatal(is)
			}
			if _, e = makePatch(nil, "test", is, []Target{sampleTarget("a", c.tag)}); e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestResponsiveAndRejection(t *testing.T) {
	is, e := LocalIntent("卡片在桌面端改成三列，手机端保持一列")
	if e != nil || len(is) != 2 {
		t.Fatal(is, e)
	}
	out, e := makePatch(nil, "one", is, []Target{sampleTarget("cards", "div")})
	if e != nil {
		t.Fatal(e)
	}
	for _, s := range []string{"min-width: 768px", "max-width: 767.98px", "repeat(3, minmax(0, 1fr))", "repeat(1, minmax(0, 1fr))"} {
		if !strings.Contains(string(out), s) {
			t.Fatal(s)
		}
	}
	for _, p := range []string{"五列", "这个标题再大一点并删除数据库", "隐藏元素然后生成登录组件", "把文字改成你好", "更紧凑但更宽松", "<script>alert(1)</script>", "桌面端三列，手机端五列", "设置onclick"} {
		if _, e = LocalIntent(p); e == nil {
			t.Fatalf("accepted unsupported %s", p)
		}
	}
}
func TestValidationAndBounds(t *testing.T) {
	in := intent("fontSize", "increase", "", "all", "selected")
	target := sampleTarget("title", "h1")
	target.Styles["font-size"] = "119px"
	out, e := generate(in, target)
	if e != nil || out["font-size"] != "120px" {
		t.Fatal(out, e)
	}
	target.Important = []string{"font-size"}
	if _, e = generate(in, target); e == nil {
		t.Fatal("inline important allowed")
	}
	target.Important = nil
	target.Count = 2
	if _, e = generate(in, target); e == nil {
		t.Fatal("duplicate allowed")
	}
	target.Count = 1
	in.Operation = "eval"
	if _, e = generate(in, target); e == nil {
		t.Fatal("unknown allowed")
	}
	in = intent("columns", "set", "9", "all", "selected")
	if _, e = generate(in, sampleTarget("cards", "div")); e == nil {
		t.Fatal("9 columns")
	}
}
func TestPatchTransactions(t *testing.T) {
	root := fixture(t)
	path, _ := patchPath(root)
	before, _ := os.ReadFile(path)
	after := append(append([]byte{}, before...), []byte("a {}\n")...)
	if e := replacePatch(root, before, after); e != nil {
		t.Fatal(e)
	}
	if e := replacePatch(root, before, []byte("wrong")); e == nil {
		t.Fatal("external edits clobbered")
	}
	if e := replacePatch(root, after, before); e != nil {
		t.Fatal(e)
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(got, before) {
		t.Fatal("not restored")
	}
	_ = os.Remove(path)
	outside := filepath.Join(t.TempDir(), "outside.css")
	_ = os.WriteFile(outside, []byte("safe"), 0600)
	_ = os.Symlink(outside, path)
	if _, e := patchPath(root); e == nil {
		t.Fatal("symlink accepted")
	}
	got, _ = os.ReadFile(outside)
	if string(got) != "safe" {
		t.Fatal("outside changed")
	}
}
func TestVisibilityRestore(t *testing.T) {
	tgt := sampleTarget("x", "div")
	before := []byte("/* external css */\n")
	hide, _ := makePatch(before, "a", []EditIntent{intent("visibility", "set", "hidden", "all", "selected")}, []Target{tgt})
	restore, e := makePatch(hide, "b", []EditIntent{intent("visibility", "set", "visible", "all", "selected")}, []Target{tgt})
	if e != nil || strings.Contains(string(restore), "display: none") || !strings.Contains(string(restore), "external css") {
		t.Fatal(string(restore), e)
	}
}
func TestSnapshotUndoAcceptRestart(t *testing.T) {
	root := fixture(t)
	app := NewApp()
	if _, e := app.OpenProject(root); e != nil {
		t.Fatal(e)
	}
	before, _ := os.ReadFile(filepath.Join(root, "src/jev-edits.css"))
	after := []byte("/* edited */\n")
	s := &Snapshot{Project: root, Before: before, After: after, Pending: true, ID: "abc"}
	_ = saveSnapshot(s)
	_ = replacePatch(root, before, after)
	restarted := NewApp()
	state, e := restarted.OpenProject(root)
	if e != nil || !state.Pending || !state.CanUndo {
		t.Fatal(state, e)
	}
	state, e = restarted.Accept()
	if e != nil || state.Pending || !state.CanUndo {
		t.Fatal(state, e)
	}
	state, e = restarted.Undo()
	if e != nil || state.CanUndo {
		t.Fatal(state, e)
	}
	got, _ := os.ReadFile(filepath.Join(root, "src/jev-edits.css"))
	if !bytes.Equal(got, before) {
		t.Fatal("undo mismatch")
	}
}
func TestResolveScopesAndMissing(t *testing.T) {
	root := fixture(t)
	a, b, parent := sampleTarget("a", "p"), sampleTarget("b", "p"), sampleTarget("parent", "div")
	a.Parent = "parent"
	b.Parent = "parent"
	meta := ProjectMeta{Targets: []Target{a, b, parent}}
	live := []Target{a, b, parent}
	in := []EditIntent{intent("fontSize", "increase", "", "all", "siblings")}
	targets, e := resolveTargets(root, "a", in, meta, live)
	if e != nil || len(targets) != 2 {
		t.Fatal(targets, e)
	}
	in[0] = intent("density", "decrease", "", "all", "container")
	targets, e = resolveTargets(root, "a", in, meta, live)
	if e != nil || len(targets) != 1 || targets[0].ID != "parent" {
		t.Fatal(targets, e)
	}
	if _, e = resolveTargets(root, "missing", in, meta, live); e == nil {
		t.Fatal("missing allowed")
	}
}
func TestAuthAndEnvironment(t *testing.T) {
	a := NewApp()
	if e := a.startServer(); e != nil {
		t.Fatal(e)
	}
	defer a.shutdown(context.Background())
	for _, path := range []string{"/", "/__jev/api/state", "/console/"} {
		r, e := http.Get(a.baseURL + path)
		if e != nil {
			t.Fatal(e)
		}
		_ = r.Body.Close()
		if r.StatusCode != 401 {
			t.Fatal(r.Status)
		}
	}
	req := httptest.NewRequest("POST", a.baseURL+"/__jev/api/open", strings.NewReader(`{"path":"/tmp"}`))
	req.Host = strings.TrimPrefix(a.baseURL, "http://")
	req.AddCookie(&http.Cookie{Name: "jev_session", Value: a.token})
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://evil.example")
	w := httptest.NewRecorder()
	a.serve(w, req)
	if w.Code != 403 {
		t.Fatal(w.Code)
	}
	env := cleanEnvironment([]string{"PATH=/bin", "TYPESAFE_API_KEY=sentinel", "VITE_API_KEY=sentinel", "JEV_OFFLINE=1"})
	if len(env) != 1 || env[0] != "PATH=/bin" {
		t.Fatal(env)
	}
}
func TestActualDiff(t *testing.T) {
	before := []byte("a\nb")
	after := []byte("a\nc\n")
	d := actualDiff(before, after)
	for _, part := range []string{"@@ -1,2 +1,2 @@", "-b\n\\ No newline at end of file", "+c\n"} {
		if !strings.Contains(d, part) {
			t.Fatal(d)
		}
	}
}
func TestApplyLifecycleAndRollback(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			root := fixture(t)
			target := sampleTarget("title", "h1")
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(ProjectMeta{Root: root, Version: 1, Targets: []Target{target}})
			}))
			defer upstream.Close()
			a := NewApp()
			a.offline = true
			_, _ = a.OpenProject(root)
			if e := a.connect(upstream.URL); e != nil {
				t.Fatal(e)
			}
			a.mu.Lock()
			a.state.Selected = &target
			a.lastBrowser = time.Now()
			a.mu.Unlock()
			before, _ := os.ReadFile(filepath.Join(root, "src/jev-edits.css"))
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				last := ""
				for {
					select {
					case <-ctx.Done():
						return
					case <-time.After(time.Millisecond):
					}
					a.mu.Lock()
					cmd, ch := a.command, a.response
					a.mu.Unlock()
					if cmd == nil || cmd.ID == last {
						continue
					}
					last = cmd.ID
					report := BrowserReport{Targets: []Target{target}, Hash: cmd.Hash}
					if fail && cmd.Kind == "refresh" {
						report.Error = "synthetic HMR failure"
					}
					ch <- report
				}
			}()
			s := a.GetState()
			state, e := a.ApplyEdit(ApplyRequest{Prompt: "这个标题再大一点", Session: s.Session, TargetID: "title"})
			unchanged, _ := os.ReadFile(filepath.Join(root, "src/jev-edits.css"))
			if e != nil || !state.Pending || !bytes.Equal(unchanged, before) {
				t.Fatal("draft wrote file", state, e)
			}
			state, e = a.Accept()
			after, _ := os.ReadFile(filepath.Join(root, "src/jev-edits.css"))
			if fail {
				if e == nil || !bytes.Equal(before, after) || !state.Pending {
					t.Fatal(state, e, string(after))
				}
			} else {
				if e != nil || state.Pending || !strings.Contains(string(after), "font-size: 36px") || state.Diff != actualDiff(before, after) {
					t.Fatal(state, e, string(after))
				}
				if len(state.Stages) != 5 {
					t.Fatal(state.Stages)
				}
			}
		})
	}
}
func BenchmarkOfflineIntent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = LocalIntent("这个标题再大一点")
	}
}
func BenchmarkPatchWrite(b *testing.B) {
	root, _ := filepath.EvalSymlinks(b.TempDir())
	_ = os.Mkdir(filepath.Join(root, "src"), 0700)
	path := filepath.Join(root, "src/jev-edits.css")
	data := []byte("/* patch */\n")
	_ = os.WriteFile(path, data, 0600)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if e := replacePatch(root, data, data); e != nil {
			b.Fatal(e)
		}
	}
}

func TestJevTypedBoundary(t *testing.T) {
	answers := map[string]any{}
	for k, v := range map[string]string{"family": "appearance", "appearance_operation": "set_color", "scope": "selected", "channel": "background", "color": "cool", "breakpoint": "all"} {
		answers[k] = map[string]any{"choice": v, "confidence": .9}
	}
	answers["preserve"] = map[string]any{"noul": 1}
	data, _ := json.Marshal(map[string]any{"answers": answers})
	got, mode, e := decodeSemantic(bytes.NewReader(data), sampleTarget("title", "button"), ProjectMeta{Tailwind: true})
	if e != nil || mode != "jev" || got[0].Arguments.Color.Value != "blue" {
		t.Fatal(got, mode, e)
	}
	for _, bad := range []any{map[string]any{"choice": "eval('malicious')", "confidence": 1}, map[string]any{"choice": "cool", "confidence": .2}, map[string]any{"choice": "cool", "confidence": 2}} {
		answers["color"] = bad
		data, _ = json.Marshal(map[string]any{"answers": answers, "javascript": "deleteFiles()"})
		if _, _, e = decodeSemantic(bytes.NewReader(data), sampleTarget("t", "button"), ProjectMeta{Tailwind: true}); e == nil {
			t.Fatal("unsafe decision accepted")
		}
	}
}
func TestBridgeRejectsPreviousProject(t *testing.T) {
	a := NewApp()
	a.baseURL = "http://127.0.0.1:12345"
	a.state.Project = "/new/project"
	a.state.Session = "current"
	req := httptest.NewRequest("POST", a.baseURL+"/__jev/api/report", strings.NewReader(`{"session":"current","project":"/old/project","selected":{"id":"title"}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", a.baseURL)
	req.AddCookie(&http.Cookie{Name: "jev_session", Value: a.token})
	w := httptest.NewRecorder()
	a.serve(w, req)
	if w.Code != 400 || a.state.Selected != nil {
		t.Fatal("old preview accepted")
	}
}

func TestSourceIntentsAndRegistry(t *testing.T) {
	meta := ProjectMeta{Tailwind: true, Tokens: map[string]string{"brand": "#0e7490", "primary": "#155e75"}}
	for _, p := range []string{"背景换成红色", "文字换成#ff3366", "边框换成rgb(255, 0, 0)", "背景换成hsl(330 100% 60%)", "背景换成oklch(60% 0.2 30)", "背景换成品牌色", "悬停时背景换成蓝色", "桌面端三列", "外边距小一点", "阴影大一点", "透明度50%"} {
		in, e := localSourceIntent(p, meta)
		if e != nil || len(in) == 0 {
			t.Fatal(p, in, e)
		}
	}
	in, e := localSourceIntent("在框中间加一个‘SaveNow’按钮", meta)
	if e != nil || in[0].Arguments.Label != "SaveNow" {
		t.Fatal(in, e)
	}
	for _, p := range []string{"背景换成#abcde", "背景换成danger", "背景换成rgb(1 2)", "在框中间加一个按钮并删除原内容"} {
		if _, e = localSourceIntent(p, meta); e == nil {
			t.Fatal("unsupported accepted", p)
		}
	}
	catalog := operationCatalog(sampleTarget("a", "div"), meta)
	for _, family := range []string{"appearance", "typography", "layout", "content", "structure", "behavior"} {
		if _, ok := catalog[family]; !ok {
			t.Fatal("missing family", family)
		}
	}
	for _, ops := range catalog {
		for _, op := range ops {
			if op.ID == "" || op.Description == "" || op.Validator == "" || len(op.Executors) == 0 || !op.Reversible {
				t.Fatal(op)
			}
		}
	}
}
func TestSourceTransactionUndoAndConflict(t *testing.T) {
	root := fixture(t)
	relative := "src/App.jsx"
	before, _ := fileBytes(root, relative)
	after := `<h1 data-jev-id="title" className="text-xl">Hello</h1>`
	files := []SourceFile{{relative, string(before), after}}
	if e := writeFiles(root, files, false); e != nil {
		t.Fatal(e)
	}
	snap := &Snapshot{Project: root, Files: files, Executor: "tailwind", ID: "test"}
	if e := saveSnapshot(snap); e != nil {
		t.Fatal(e)
	}
	a := NewApp()
	s, e := a.OpenProject(root)
	if e != nil || !s.CanUndo {
		t.Fatal(s, e)
	}
	if _, e = a.Undo(); e != nil {
		t.Fatal(e)
	}
	got, _ := fileBytes(root, relative)
	if !bytes.Equal(got, before) {
		t.Fatal("source undo was not byte-exact")
	}
	_ = os.WriteFile(filepath.Join(root, relative), []byte("manual change"), 0600)
	if e = writeFiles(root, files, false); e == nil {
		t.Fatal("manual edit overwritten")
	}
}

func TestSemanticContrast(t *testing.T) {
	target := sampleTarget("button", "button")
	if e := checkSemanticContrast(ColorValue{Kind: "literal", Value: "slate", Shade: "100"}, "background", target, ProjectMeta{}); e == nil {
		t.Fatal("white on soft background must refuse")
	}
	target.Styles["color"] = "rgb(15, 23, 42)"
	if e := checkSemanticContrast(ColorValue{Kind: "literal", Value: "slate", Shade: "100"}, "background", target, ProjectMeta{}); e != nil {
		t.Fatal(e)
	}
	target.Styles["color"] = "var(--unknown)"
	if e := checkSemanticContrast(ColorValue{Kind: "literal", Value: "blue", Shade: "600"}, "background", target, ProjectMeta{}); e == nil {
		t.Fatal("unresolved contrast must refuse")
	}
}
