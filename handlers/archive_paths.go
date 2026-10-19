package handlers

import (
	"fmt"
	"lightpanel/config"
	"os"
	"path/filepath"
	"strings"
)

// Reject prefix siblings, absolute paths and existing symlink ancestors.
func archiveTarget(root, name string) (string, error) {
	name = strings.ReplaceAll(name, "\\", "/")
	for _, component := range strings.Split(name, "/") {
		if component != "." && component != ".." && strings.TrimRight(component, " .") != component {
			return "", fmt.Errorf("路径不能包含尾随空格或句点")
		}
	}
	if strings.HasPrefix(name, "/") || strings.Contains(name, ":") {
		return "", fmt.Errorf("无效归档路径")
	}
	base, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	target := filepath.Join(base, filepath.FromSlash(name))
	rel, err := filepath.Rel(base, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("归档路径超出目标目录")
	}
	current := base
	if info, err := os.Lstat(current); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("目标目录不能为符号链接")
	}
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("归档路径包含符号链接")
		}
	}
	return target, nil
}

func managedDirectory(candidate string) (string, error) {
	base, err := filepath.Abs(config.AppsDir)
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(base, target)
	if err != nil || rel == "." {
		return "", fmt.Errorf("应用目录必须是独立子目录")
	}
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		if strings.HasPrefix(part, ".install-") {
			return "", fmt.Errorf("不能使用安装临时目录")
		}
	}
	return archiveTarget(base, rel)
}

func pathsOverlap(a, b string) bool {
	a, b = pathKey(a), pathKey(b)
	sep := string(os.PathSeparator)
	return a == b || strings.HasPrefix(a, b+sep) || strings.HasPrefix(b, a+sep)
}
