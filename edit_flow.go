package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"
)

type Draft struct {
	Patch                 SourcePatch
	Targets               []Target
	Intents               []EditIntent
	Session, ClientID, ID string
}

func (a *App) finish(err *error, result *EditorState) {
	a.mu.Lock()
	a.state.Busy = false
	if *err != nil {
		a.state.Error = (*err).Error()
	}
	a.mu.Unlock()
	*result = a.GetState()
}
func (a *App) ApplyEdit(req ApplyRequest) (result EditorState, err error) {
	if !a.op.TryLock() {
		return a.GetState(), fmt.Errorf("正在处理修改")
	}
	defer a.op.Unlock()
	s := a.GetState()
	if s.Pending {
		return s, fmt.Errorf("请先 Accept 或 Reject 当前 Diff")
	}
	if len(req.Prompt) > 2048 {
		return s, unsupported
	}
	if !s.Connected || !s.BrowserConnected || s.Selected == nil || s.Session != req.Session || s.Selected.ID != req.TargetID {
		return s, fmt.Errorf("请在当前项目预览中选择元素")
	}
	a.mu.Lock()
	a.state.Busy = true
	a.state.Error = ""
	a.state.Stages = nil
	a.state.Intents = nil
	upstream := a.upstream
	a.mu.Unlock()
	defer a.finish(&err, &result)
	start := time.Now()
	a.stage("Understanding edit", start)
	meta, e := fetchMeta(upstream)
	if e != nil || meta.Root != s.Project {
		return s, fmt.Errorf("项目预览已失效")
	}
	intents, mode, e := understandEdit(a.ctx, req.Prompt, *s.Selected, meta, a.offline)
	a.mu.Lock()
	a.state.Mode = mode
	a.state.Intents = intents
	a.mu.Unlock()
	if e != nil {
		return s, e
	}
	a.stage("Validating target", start)
	live, e := a.browserCommand(s.ClientID, "inspect", nil, "")
	if e != nil {
		return s, e
	}
	targets, e := resolveTargets(s.Project, req.TargetID, intents, meta, live.Targets)
	if e != nil {
		return s, e
	}
	for _, target := range targets {
		if target.Repeated {
			return s, fmt.Errorf("不支持：map() 目标会影响共享实例，请选择静态元素")
		}
	}
	for _, fresh := range meta.Targets {
		if fresh.ID == req.TargetID && s.Selected.Revision != "" && (fresh.Revision != s.Selected.Revision || fresh.Fingerprint != s.Selected.Fingerprint) {
			return s, fmt.Errorf("源码 revision 已变化，请重新选择元素后再试")
		}
	}
	var executor StyleExecutor = CssExecutor{}
	if meta.Tailwind {
		executor = TailwindExecutor{}
		if intents[0].Operation == "insert_child" {
			executor = JSXExecutor{}
		}
	} else {
		for _, i := range intents {
			if i.State != "" || !contains([]string{"all", "mobile", "desktop"}, i.Breakpoint) {
				return s, fmt.Errorf("普通 CSS 模式不支持此状态或断点")
			}
			if i.Operation == "insert_child" || i.Operation == "set_color" || i.Operation == "set_utility" || i.Operation == "margin" {
				return s, fmt.Errorf("该操作需要 Tailwind 项目；普通 CSS 兼容操作仍可使用")
			}
		}
	}
	ctx, cancel := context.WithTimeout(a.ctx, 12*time.Second)
	defer cancel()
	patch, e := executor.CreatePatch(ctx, s.Project, intents, targets)
	if e != nil {
		return s, e
	}
	if e = executor.Validate(patch); e != nil {
		return s, e
	}
	id := randomToken()
	a.mu.Lock()
	a.draft = &Draft{patch, targets, intents, s.Session, s.ClientID, id}
	a.state.Pending = true
	a.state.Diff = sourceDiff(patch.Files)
	a.state.ChangeID = id
	a.state.Files = nil
	for _, f := range patch.Files {
		a.state.Files = append(a.state.Files, f.Path)
	}
	a.state.Executor = patch.Executor
	a.mu.Unlock()
	a.stage("Patch ready — awaiting Accept", start)
	return a.GetState(), nil
}
func (a *App) Reject() (EditorState, error) {
	if !a.op.TryLock() {
		return a.GetState(), fmt.Errorf("正在处理修改")
	}
	defer a.op.Unlock()
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.draft == nil {
		return a.state, fmt.Errorf("没有待拒绝的 Diff")
	}
	a.draft = nil
	a.state.Pending = false
	a.state.Diff = ""
	a.state.Files = nil
	a.state.Error = ""
	a.state.Stages = nil
	return a.state, nil
}
func (a *App) sourceRefresh(client string, files []SourceFile, checks []SourceCheck, reverse bool) (BrowserReport, error) {
	revisions := map[string]string{}
	for _, f := range files {
		v := f.After
		if reverse {
			v = f.Before
		}
		revisions[f.Path] = digest([]byte(v))
	}
	cmd := &Command{ID: randomToken(), ClientID: client, Kind: "source-refresh", Revisions: revisions, Checks: checks}
	if target := a.GetState().Selected; target != nil {
		cmd.Targets = []string{target.ID}
	}
	ch := make(chan BrowserReport, 1)
	a.mu.Lock()
	a.command = cmd
	a.response = ch
	a.mu.Unlock()
	defer func() { a.mu.Lock(); a.command = nil; a.response = nil; a.mu.Unlock() }()
	select {
	case r := <-ch:
		if r.Error != "" {
			return r, fmt.Errorf("%s", r.Error)
		}
		return r, nil
	case <-time.After(8 * time.Second):
		return BrowserReport{}, fmt.Errorf("浏览器未确认源码 HMR；修改已取消")
	}
}
func (a *App) Accept() (result EditorState, err error) {
	if !a.op.TryLock() {
		return a.GetState(), fmt.Errorf("正在处理修改")
	}
	defer a.op.Unlock()
	s := a.GetState()
	d := a.draft
	if d == nil { // Compatibility for a pre-upgrade CSS snapshot already written by the previous release.
		if a.snapshot != nil && a.snapshot.Pending && len(a.snapshot.Files) == 0 {
			root, e := os.OpenRoot(s.Project)
			if e != nil {
				return s, e
			}
			defer root.Close()
			current, e := root.ReadFile("src/jev-edits.css")
			if e != nil || !bytes.Equal(current, a.snapshot.After) {
				return s, fmt.Errorf("文件已被外部修改，不能接受旧快照")
			}
			next := *a.snapshot
			next.Pending = false
			if e := saveSnapshot(&next); e != nil {
				return s, e
			}
			a.mu.Lock()
			a.snapshot = &next
			a.state.Pending = false
			a.mu.Unlock()
			return a.GetState(), nil
		}
		return s, fmt.Errorf("没有待接受的 Diff")
	}
	if d.Session != s.Session {
		return s, fmt.Errorf("项目会话已变化")
	}
	a.mu.Lock()
	a.state.Busy = true
	a.state.Error = ""
	upstream := a.upstream
	a.mu.Unlock()
	defer a.finish(&err, &result)
	start := time.Now()
	meta, e := fetchMeta(upstream)
	if e != nil {
		return s, e
	}
	live, e := a.browserCommand(d.ClientID, "inspect", nil, "")
	if e != nil {
		return s, e
	}
	for _, original := range d.Targets {
		found := false
		for _, fresh := range meta.Targets {
			if fresh.ID == original.ID {
				if fresh.Revision != original.Revision || fresh.Fingerprint != original.Fingerprint {
					return s, fmt.Errorf("源码 revision 或节点指纹已变化，请 Reject 后重新生成")
				}
				for _, t := range live.Targets {
					if t.ID == fresh.ID && t.Count == 1 {
						found = true
					}
				}
			}
		}
		if !found {
			return s, fmt.Errorf("目标消失或匹配不唯一，未写入")
		}
	}
	snap := &Snapshot{Project: s.Project, Files: d.Patch.Files, Executor: d.Patch.Executor, ID: d.ID}
	old := a.snapshot
	if e = saveSnapshot(snap); e != nil {
		return s, e
	}
	a.stage("Writing patch", start)
	restoreHistory := func() {
		if old != nil {
			_ = saveSnapshot(old)
		} else {
			_ = clearSnapshot(s.Project)
		}
	}
	if e = writeFiles(s.Project, d.Patch.Files, false); e != nil {
		restoreHistory()
		return s, e
	}
	rollback := func(cause error) error {
		if re := writeFiles(s.Project, d.Patch.Files, true); re != nil {
			a.mu.Lock()
			a.snapshot = snap
			a.state.CanUndo = true
			a.mu.Unlock()
			return fmt.Errorf("%v；恢复被外部修改阻止：%v，快照已保留", cause, re)
		}
		restoreHistory()
		return fmt.Errorf("%v；文件已自动恢复", cause)
	}
	if d.Patch.Executor != "css" {
		a.stage("Checking source", start)
		if e = validateBuild(a.ctx, s.Project); e != nil {
			return s, rollback(e)
		}
		if _, e = a.sourceRefresh(d.ClientID, d.Patch.Files, d.Patch.Checks, false); e != nil {
			return s, rollback(e)
		}
	} else {
		var expected []Expectation
		for _, in := range d.Intents {
			for _, t := range d.Targets {
				props, _ := registry[in.Operation].Generate(in, t)
				expected = append(expected, Expectation{t.ID, in.Breakpoint, props})
			}
		}
		if _, e = a.browserCommand(d.ClientID, "refresh", nil, digest([]byte(d.Patch.Files[0].After)), expected...); e != nil {
			return s, rollback(e)
		}
	}
	a.mu.Lock()
	a.snapshot = snap
	a.draft = nil
	a.state.Pending = false
	a.state.CanUndo = true
	a.mu.Unlock()
	a.stage("Vite refreshed", start)
	return a.GetState(), nil
}
func (a *App) Undo() (result EditorState, err error) {
	if !a.op.TryLock() {
		return a.GetState(), fmt.Errorf("正在处理修改")
	}
	defer a.op.Unlock()
	s := a.GetState()
	if a.draft != nil {
		return s, fmt.Errorf("请先 Reject 待接受的 Diff")
	}
	snap := a.snapshot
	if snap == nil || !s.CanUndo {
		return s, fmt.Errorf("没有可撤销的修改")
	}
	a.mu.Lock()
	a.state.Busy = true
	a.state.Error = ""
	a.state.Stages = nil
	a.mu.Unlock()
	defer a.finish(&err, &result)
	start := time.Now()
	files := snapshotFiles(snap)
	a.stage("Writing patch", start)
	if e := writeFiles(s.Project, files, true); e != nil {
		return s, e
	}
	rollback := func(e error) error {
		if re := writeFiles(s.Project, files, false); re != nil {
			return fmt.Errorf("%v；恢复失败：%v", e, re)
		}
		return e
	}
	if s.BrowserConnected {
		var e error
		if snap.Executor == "css" || snap.Executor == "" {
			_, e = a.browserCommand(s.ClientID, "refresh", nil, digest([]byte(files[0].Before)))
		} else {
			e = validateBuild(a.ctx, s.Project)
			if e == nil {
				_, e = a.sourceRefresh(s.ClientID, files, nil, true)
			}
		}
		if e != nil {
			return s, rollback(e)
		}
		a.stage("Vite refreshed", start)
	}
	if e := clearSnapshot(s.Project); e != nil {
		return s, rollback(e)
	}
	for i := range files {
		files[i].Before, files[i].After = files[i].After, files[i].Before
	}
	a.mu.Lock()
	a.snapshot = nil
	a.state.Pending = false
	a.state.CanUndo = false
	a.state.Diff = sourceDiff(files)
	a.mu.Unlock()
	return a.GetState(), nil
}
