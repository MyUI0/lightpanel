package handlers

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
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
	fsResult := detectNewExecutable(before)

	var scriptPaths []string
	_ = filepath.WalkDir(sandboxPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".sh") || strings.Contains(strings.ToLower(d.Name()), "install") {
			scriptPaths = append(scriptPaths, path)
		}
		return nil
	})

	var analyzedPaths []string
	for _, sp := range scriptPaths {
		analyzedPaths = append(analyzedPaths, analyzeScript(sp)...)
	}

	if fsResult.Confidence == "high" {
		for _, ap := range analyzedPaths {
			if strings.HasPrefix(fsResult.WorkDir, ap) || strings.HasPrefix(ap, fsResult.WorkDir) {
				fsResult.Note += " (脚本确认: " + ap + ")"
				fsResult.Confidence = "confirmed"
				break
			}
		}
		return fsResult
	}

	if len(analyzedPaths) > 0 {
		best := analyzedPaths[0]
		for _, p := range analyzedPaths {
			if len(p) > len(best) {
				best = p
			}
		}
		if info, err := os.Stat(best); err == nil && info.IsDir() {
			var foundBin string
			_ = filepath.WalkDir(best, func(p string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				i, e := d.Info()
				if e != nil {
					return nil
				}
				if i.Mode()&0111 != 0 && i.Size() > 100 {
					foundBin = "./" + d.Name()
					return fs.SkipDir
				}
				return nil
			})
			note := "脚本分析推断目录: " + best
			if foundBin != "" {
				note += " (发现: " + foundBin + ")"
			}
			return detectResult{
				WorkDir:    best,
				Binary:     foundBin,
				Note:       note,
				Confidence: "script",
			}
		}
	}

	return fsResult
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
			found = path
			return fs.SkipDir
		}
		return nil
	})
	return found
}
