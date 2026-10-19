package handlers

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var commonBinPaths = []string{
	"/usr/local/bin",
	"/usr/bin",
	"/bin",
	"/opt",
	"/usr/local",
	"/usr/lib",
	"/root",
	"/home",
}

var excludeExts = map[string]bool{
	".bash": true, ".py": true, ".conf": true,
	".json": true, ".yaml": true, ".yml": true, ".txt": true,
	".log": true, ".md": true, ".pid": true, ".tmp": true,
	".ini": true, ".cfg": true, ".xml": true, ".html": true,
	".css": true, ".js": true, ".png": true, ".jpg": true,
	".ico": true, ".svg": true, ".gz": true, ".zip": true,
	".tar": true, ".xz": true, ".bz2": true, ".7z": true,
	".deb": true, ".rpm": true,
}

var excludeNames = map[string]bool{
	"find": true, "xargs": true, "grep": true, "ls": true,
	"cat": true, "chmod": true, "chown": true, "mkdir": true,
	"rm": true, "cp": true, "mv": true, "echo": true,
	"bash": true, "sh": true, "env": true, "test": true,
}

type detectResult struct {
	WorkDir    string
	Binary     string
	Note       string
	Confidence string
}

func analyzeScript(scriptPath string) []string {
	f, err := os.Open(scriptPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var paths []string
	seen := make(map[string]bool)

	mkdirRe := regexp.MustCompile(`(?:mkdir\s+-p?\s+|install\s+-d\s+)(/[\w/.-]+)`)
	cpRe := regexp.MustCompile(`(?:cp\s+\S+\s+|install\s+\S+\s+)(/[\w/.-]+)`)
	mvRe := regexp.MustCompile(`mv\s+\S+\s+(/[\w/.-]+)`)
	curlRe := regexp.MustCompile(`(?:curl|wget).*?-o\s+(/[\w/.-]+)`)
	destRe := regexp.MustCompile(`(?:DEST_DIR|INSTALL_DIR|PREFIX|TARGET|WORKDIR)=["']?(/[\w/.-]+)`)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		for _, re := range []*regexp.Regexp{mkdirRe, cpRe, mvRe, curlRe, destRe} {
			matches := re.FindAllStringSubmatch(line, -1)
			for _, m := range matches {
				if len(m) > 1 && len(m[1]) > 3 && !seen[m[1]] {
					seen[m[1]] = true
					paths = append(paths, m[1])
				}
			}
		}
	}
	return paths
}

func detectNewExecutable(before time.Time) detectResult {
	var candidates []struct {
		path string
		mod  time.Time
		size int64
	}

	for _, dir := range commonBinPaths {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				rel, _ := filepath.Rel(dir, path)
				depth := 0
				if rel != "." {
					depth = len(strings.Split(rel, string(os.PathSeparator)))
				}
				if depth > 5 {
					return fs.SkipDir
				}
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			modTime := info.ModTime()
			if modTime.Before(before) {
				return nil
			}
			if info.Mode()&0111 == 0 {
				return nil
			}
			if info.Size() < 100 {
				return nil
			}
			base := filepath.Base(path)
			if excludeNames[base] {
				return nil
			}
			ext := filepath.Ext(base)
			if excludeExts[ext] {
				return nil
			}
			candidates = append(candidates, struct {
				path string
				mod  time.Time
				size int64
			}{path, modTime, info.Size()})
			return nil
		})
	}

	if len(candidates) == 0 {
		return detectResult{Note: "自动检测未发现新安装的可执行文件，请手动设置工作目录和启动命令", Confidence: "none"}
	}

	latest := candidates[0]
	for _, c := range candidates {
		if c.mod.After(latest.mod) {
			latest = c
		}
	}

	workDir := filepath.Dir(latest.path)
	binary := "./" + filepath.Base(latest.path)
	sizeStr := formatDetectedSize(latest.size)
	return detectResult{
		WorkDir:    workDir,
		Binary:     binary,
		Note:       "自动检测到: " + latest.path + " (" + sizeStr + ")",
		Confidence: "high",
	}
}

func combinedDetect(sandboxPath string, before time.Time) detectResult {
	binary := findBinaryInDir(sandboxPath)
	if binary == "" {
		return detectResult{WorkDir: sandboxPath, Note: "应用目录内未检测到可执行文件，请填写启动命令", Confidence: "none"}
	}
	return detectResult{WorkDir: sandboxPath, Binary: binary, Note: "检测到应用目录内的可执行文件：" + binary, Confidence: "local"}
}

func formatDetectedSize(size int64) string {
	if size <= 0 {
		return "0 B"
	}
	if size > 1024*1024 {
		v := float64(size) / (1024 * 1024)
		if v >= 10 {
			return fmt.Sprintf("%.0f MB", v)
		}
		return fmt.Sprintf("%.1f MB", v)
	}
	if size > 1024 {
		v := float64(size) / 1024
		if v >= 10 {
			return fmt.Sprintf("%.0f KB", v)
		}
		return fmt.Sprintf("%.1f KB", v)
	}
	return fmt.Sprintf("%d B", size)
}

func findBinaryInDir(dir string) string {
	var found string
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		depth := len(strings.Split(rel, string(os.PathSeparator)))
		if depth > 3 {
			return nil
		}
		i, e := d.Info()
		if e != nil {
			return nil
		}
		if i.Mode()&0111 != 0 && i.Size() > 100 {
			base := filepath.Base(path)
			if excludeNames[base] || excludeExts[filepath.Ext(base)] {
				return nil
			}
			found = rel
			return fs.SkipDir
		}
		return nil
	})
	if found == "" {
		return ""
	}
	value := "./" + filepath.ToSlash(found)
	if runtime.GOOS == "windows" {
		return "\"" + strings.ReplaceAll(value, "\"", "") + "\""
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
