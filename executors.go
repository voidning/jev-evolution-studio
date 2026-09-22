package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/pmezard/go-difflib/difflib"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//go:embed integration/dist/source-executor.cjs
var sourceExecutor []byte

type StyleExecutor interface {
	Inspect(context.Context, string) (ProjectMeta, error)
	CreatePatch(context.Context, string, []EditIntent, []Target) (SourcePatch, error)
	Validate(SourcePatch) error
}
type TailwindExecutor struct{}
type CssExecutor struct{}
type JSXExecutor struct{ TailwindExecutor }

func (JSXExecutor) CreatePatch(ctx context.Context, root string, intents []EditIntent, targets []Target) (SourcePatch, error) {
	for _, i := range intents {
		if i.Operation != "insert_child" {
			return SourcePatch{}, fmt.Errorf("结构执行器只接受注册结构操作")
		}
	}
	return TailwindExecutor{}.CreatePatch(ctx, root, intents, targets)
}

func runSource(ctx context.Context, request any) ([]byte, error) {
	f, e := os.CreateTemp("", "jev-source-*.cjs")
	if e != nil {
		return nil, e
	}
	name := f.Name()
	defer os.Remove(name)
	if _, e = f.Write(sourceExecutor); e != nil {
		f.Close()
		return nil, e
	}
	f.Close()
	body, _ := json.Marshal(request)
	cmd := exec.CommandContext(ctx, "node", name)
	cmd.Env = cleanEnvironment(os.Environ())
	cmd.Stdin = bytes.NewReader(body)
	out, e := cmd.Output()
	if e != nil {
		var result struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(out, &result) == nil && result.Error != "" {
			return nil, fmt.Errorf("%s", result.Error)
		}
		return nil, fmt.Errorf("源码解析器运行失败，请检查 Node 版本")
	}
	return out, nil
}
func (TailwindExecutor) Inspect(ctx context.Context, root string) (ProjectMeta, error) {
	var m ProjectMeta
	b, e := runSource(ctx, map[string]any{"action": "inspect", "root": root})
	if e == nil {
		e = json.Unmarshal(b, &m)
	}
	return m, e
}
func (TailwindExecutor) CreatePatch(ctx context.Context, root string, intents []EditIntent, targets []Target) (SourcePatch, error) {
	var p SourcePatch
	b, e := runSource(ctx, map[string]any{"action": "patch", "root": root, "intents": intents, "targets": targets})
	if e == nil {
		e = json.Unmarshal(b, &p)
	}
	return p, e
}
func (TailwindExecutor) Validate(p SourcePatch) error {
	if len(p.Files) == 0 {
		return fmt.Errorf("没有源码修改")
	}
	for _, f := range p.Files {
		if f.Before == f.After || (!strings.HasSuffix(f.Path, ".jsx") && !strings.HasSuffix(f.Path, ".tsx")) {
			return fmt.Errorf("非法源码 Patch")
		}
	}
	return nil
}
func (CssExecutor) Inspect(ctx context.Context, root string) (ProjectMeta, error) {
	return TailwindExecutor{}.Inspect(ctx, root)
}
func (CssExecutor) CreatePatch(_ context.Context, root string, intents []EditIntent, targets []Target) (SourcePatch, error) {
	p, e := patchPath(root)
	if e != nil {
		return SourcePatch{}, e
	}
	before, e := os.ReadFile(p)
	if e != nil {
		return SourcePatch{}, e
	}
	after, e := makePatch(before, randomToken(), intents, targets)
	if e != nil {
		return SourcePatch{}, e
	}
	return SourcePatch{Executor: "css", Files: []SourceFile{{"src/jev-edits.css", string(before), string(after)}}}, nil
}
func (CssExecutor) Validate(p SourcePatch) error {
	if len(p.Files) != 1 || p.Files[0].Path != "src/jev-edits.css" || p.Files[0].Before == p.Files[0].After {
		return fmt.Errorf("没有需要修改的样式")
	}
	return nil
}
func sourceDiff(files []SourceFile) string {
	var out strings.Builder
	for _, f := range files {
		if f.Path == "src/jev-edits.css" {
			out.WriteString(actualDiff([]byte(f.Before), []byte(f.After)))
			continue
		}
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{A: difflib.SplitLines(f.Before), B: difflib.SplitLines(f.After), FromFile: "a/" + f.Path, ToFile: "b/" + f.Path, Context: 3})
		out.WriteString(diff)
	}
	return out.String()
}
func snapshotFiles(s *Snapshot) []SourceFile {
	if len(s.Files) > 0 {
		return s.Files
	}
	return []SourceFile{{"src/jev-edits.css", string(s.Before), string(s.After)}}
}
func fileBytes(root, relative string) ([]byte, error) {
	if filepath.IsAbs(relative) || !inside(root, filepath.Join(root, relative)) {
		return nil, fmt.Errorf("源路径越界")
	}
	real, e := filepath.EvalSymlinks(filepath.Join(root, relative))
	if e != nil || real != filepath.Join(root, relative) {
		return nil, fmt.Errorf("源路径无效或包含符号链接")
	}
	r, e := os.OpenRoot(root)
	if e != nil {
		return nil, e
	}
	defer r.Close()
	return r.ReadFile(relative)
}
func replaceSource(root, relative, before, after string) error {
	current, e := fileBytes(root, relative)
	if e != nil {
		return e
	}
	if string(current) != before {
		return fmt.Errorf("文件 revision 已变化；未覆盖，请重新生成 Diff")
	}
	r, e := os.OpenRoot(root)
	if e != nil {
		return e
	}
	defer r.Close()
	info, e := r.Stat(relative)
	if e != nil {
		return e
	}
	temp := filepath.Join(filepath.Dir(relative), ".jev-tmp-"+randomToken())
	f, e := r.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
	if e != nil {
		return e
	}
	defer r.Remove(temp)
	_, e = f.Write([]byte(after))
	if e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e == nil {
		e = ce
	}
	if e != nil {
		return e
	}
	current, e = r.ReadFile(relative)
	if e != nil || string(current) != before {
		return fmt.Errorf("文件已被外部修改；未覆盖")
	}
	return r.Rename(temp, relative)
}
func writeFiles(root string, files []SourceFile, reverse bool) error {
	for _, f := range files {
		expected := f.Before
		if reverse {
			expected = f.After
		}
		b, e := fileBytes(root, f.Path)
		if e != nil || string(b) != expected {
			return fmt.Errorf("%s revision 已变化；未覆盖", f.Path)
		}
	}
	for i, f := range files {
		before, after := f.Before, f.After
		if reverse {
			before, after = after, before
		}
		if e := replaceSource(root, f.Path, before, after); e != nil {
			for j := i - 1; j >= 0; j-- {
				v := files[j]
				from, to := v.After, v.Before
				if reverse {
					from, to = to, from
				}
				if rollback := replaceSource(root, v.Path, from, to); rollback != nil {
					return fmt.Errorf("%v；恢复失败：%v", e, rollback)
				}
			}
			return e
		}
	}
	return nil
}
func validateBuild(ctx context.Context, root string) error {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	temp, e := os.MkdirTemp(root, ".jev-check-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(temp)
	run := func(entry string, args ...string) error {
		cmd := exec.CommandContext(ctx, "node", append([]string{entry}, args...)...)
		cmd.Dir = root
		cmd.Env = cleanEnvironment(os.Environ())
		if e := cmd.Run(); e != nil {
			return fmt.Errorf("源码检查失败（%s）；已取消修改，请先修复项目类型或构建错误", filepath.Base(entry))
		}
		return nil
	}
	if _, e = os.Stat(filepath.Join(root, "tsconfig.json")); e == nil {
		tsc := filepath.Join(root, "node_modules/typescript/bin/tsc")
		if _, e = os.Stat(tsc); e != nil {
			return fmt.Errorf("存在 tsconfig.json 但缺少 TypeScript，请先安装项目依赖")
		}
		if e = run(tsc, "--noEmit", "--incremental", "--tsBuildInfoFile", filepath.Join(temp, "types.tsbuildinfo")); e != nil {
			return e
		}
	}
	return run(filepath.Join(root, "node_modules/vite/bin/vite.js"), "build", "--outDir", filepath.Join(temp, "dist"), "--emptyOutDir")
}
