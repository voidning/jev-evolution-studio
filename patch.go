package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func digest(data []byte) string { h := sha256.Sum256(data); return hex.EncodeToString(h[:]) }
func inside(root, path string) bool {
	rel, e := filepath.Rel(root, path)
	return e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
func patchPath(root string) (string, error) {
	real, e := filepath.EvalSymlinks(root)
	if e != nil || real != root {
		return "", fmt.Errorf("项目路径已变化")
	}
	path := filepath.Join(root, "src", "jev-edits.css")
	real, e = filepath.EvalSymlinks(path)
	if e != nil {
		return "", fmt.Errorf("项目需要已有 src/jev-edits.css：%w", e)
	}
	if real != path || !inside(root, real) {
		return "", fmt.Errorf("拒绝符号链接或项目外路径")
	}
	info, e := os.Stat(real)
	if e != nil || !info.Mode().IsRegular() || info.Size() > 1024*1024 {
		return "", fmt.Errorf("CSS Patch 必须是小于 1MB 的普通文件")
	}
	return real, nil
}
func atomicWrite(path string, data []byte, mode os.FileMode) error {
	f, e := os.CreateTemp(filepath.Dir(path), ".jev-tmp-*")
	if e != nil {
		return e
	}
	name := f.Name()
	defer os.Remove(name)
	if e = f.Chmod(mode); e == nil {
		_, e = f.Write(data)
	}
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
	return os.Rename(name, path)
}
func replacePatch(root string, expected, next []byte) error {
	if _, e := patchPath(root); e != nil {
		return e
	}
	confined, e := os.OpenRoot(root)
	if e != nil {
		return e
	}
	defer confined.Close()
	const relative = "src/jev-edits.css"
	current, e := confined.ReadFile(relative)
	if e != nil {
		return e
	}
	if !bytes.Equal(current, expected) {
		return fmt.Errorf("文件已被外部修改；未覆盖，请先处理冲突")
	}
	info, e := confined.Stat(relative)
	if e != nil {
		return e
	}
	temp := "src/.jev-tmp-" + randomToken()
	f, e := confined.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if e != nil {
		return e
	}
	defer confined.Remove(temp)
	_, e = f.Write(next)
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
	// Recheck after fsync: an editor may have saved while the temporary file was written.
	current, e = confined.ReadFile(relative)
	if e != nil {
		return e
	}
	if !bytes.Equal(current, expected) {
		return fmt.Errorf("文件已被外部修改；未覆盖，请先处理冲突")
	}
	return confined.Rename(temp, relative)
}
func snapshotPath(root string) (string, error) {
	base, e := os.UserConfigDir()
	if e != nil {
		return "", e
	}
	dir := filepath.Join(base, "JevEditor", "history")
	if e = os.MkdirAll(dir, 0700); e != nil {
		return "", e
	}
	return filepath.Join(dir, digest([]byte(root))+".json"), nil
}
func saveSnapshot(s *Snapshot) error {
	p, e := snapshotPath(s.Project)
	if e != nil {
		return e
	}
	b, e := json.Marshal(s)
	if e != nil {
		return e
	}
	return atomicWrite(p, b, 0600)
}
func readSnapshot(root string) (*Snapshot, error) {
	p, e := snapshotPath(root)
	if e != nil {
		return nil, e
	}
	b, e := os.ReadFile(p)
	if os.IsNotExist(e) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	var s Snapshot
	if e = json.Unmarshal(b, &s); e != nil {
		return nil, e
	}
	if s.Project != root {
		return nil, fmt.Errorf("撤销记录项目不匹配")
	}
	return &s, nil
}
func clearSnapshot(root string) error {
	p, e := snapshotPath(root)
	if e != nil {
		return e
	}
	e = os.Remove(p)
	if os.IsNotExist(e) {
		return nil
	}
	return e
}

// A full-file unified diff deliberately preserves every byte represented by the transaction.
func actualDiff(before, after []byte) string {
	if bytes.Equal(before, after) {
		return ""
	}
	lines := func(b []byte) []string {
		if len(b) == 0 {
			return nil
		}
		return strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
	}
	a, b := lines(before), lines(after)
	var out strings.Builder
	fmt.Fprintf(&out, "--- a/src/jev-edits.css\n+++ b/src/jev-edits.css\n@@ -%d,%d +%d,%d @@\n", min(1, len(a)), len(a), min(1, len(b)), len(b))
	for _, part := range []struct {
		ls     []string
		prefix string
		raw    []byte
	}{{a, "-", before}, {b, "+", after}} {
		for _, l := range part.ls {
			out.WriteString(part.prefix + l + "\n")
		}
		if len(part.raw) > 0 && part.raw[len(part.raw)-1] != '\n' {
			out.WriteString("\\ No newline at end of file\n")
		}
	}
	return out.String()
}
