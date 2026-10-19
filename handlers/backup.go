package handlers

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"lightpanel/config"
	"lightpanel/models"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const maxRestoreFileSize = 100 * 1024 * 1024

func backupData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	// Build the archive first so errors never produce a successful partial download.
	f, err := os.CreateTemp("", "lightpanel-backup-*.tar.gz")
	if err != nil {
		http.Error(w, "无法创建备份", 500)
		return
	}
	defer os.Remove(f.Name())
	defer f.Close()
	appOpMu.Lock()
	if len(installReservations) > 0 {
		appOpMu.Unlock()
		http.Error(w, "请先完成或取消安装任务再备份", 409)
		return
	}
	err = writeBackup(f, config.DataDir)
	appOpMu.Unlock()
	if err != nil {
		http.Error(w, "备份失败："+err.Error(), 500)
		return
	}
	if _, err = f.Seek(0, 0); err != nil {
		http.Error(w, "读取备份失败", 500)
		return
	}
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", "attachment; filename=lightpanel_backup_"+time.Now().Format("20060102_150405")+".tar.gz")
	_, _ = io.Copy(w, f)
}

func writeBackup(out io.Writer, root string) error {
	gz := gzip.NewWriter(out)
	tw := tar.NewWriter(gz)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if rel == "config/sessions" {
			return filepath.SkipDir
		}
		if info.Mode()&os.ModeSymlink != 0 || (!info.IsDir() && !info.Mode().IsRegular()) {
			return nil
		}
		if !info.IsDir() && (info.Name() == "pid.pid" || rel == "config/initial-password.txt" || strings.HasPrefix(info.Name(), ".json-")) {
			return nil
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = rel
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(tw, file)
		closeErr := file.Close()
		if err != nil {
			return err
		}
		return closeErr
	})
	closeTar := tw.Close()
	closeGz := gz.Close()
	if err != nil {
		return err
	}
	if closeTar != nil {
		return closeTar
	}
	return closeGz
}

// Validate the complete archive in staging before touching the live data.
func stageRestore(reader io.Reader, stage string) error {
	tr := tar.NewReader(reader)
	var total int64
	count := 0
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		count++
		total += header.Size
		if count > 10000 || header.Size < 0 || header.Size > maxRestoreFileSize || total > int64(config.MaxDownBytes) {
			return fmt.Errorf("备份超过恢复限制")
		}
		name := strings.ReplaceAll(header.Name, "\\", "/")
		// Older backups included a leading data directory.
		if name == "data" || name == "data/" {
			continue
		}
		name = strings.TrimPrefix(name, "data/")
		name = path.Clean(name)
		policyName := strings.ToLower(name)
		if policyName == "config/initial-password.txt" || policyName == "config/sessions" || strings.HasPrefix(policyName, "config/sessions/") || filepath.Base(policyName) == "pid.pid" {
			continue
		}
		target, err := archiveTarget(stage, name)
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			mode := os.FileMode(header.Mode) & 0777
			if strings.HasPrefix(policyName, "config/") {
				mode = 0600
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
			if err != nil {
				return err
			}
			_, copyErr := io.CopyN(f, tr, header.Size)
			closeErr := f.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
			if strings.HasPrefix(policyName, "config/") && strings.HasSuffix(policyName, ".json") {
				b, err := os.ReadFile(target)
				if err != nil {
					return err
				}
				if !json.Valid(b) {
					return fmt.Errorf("配置 JSON 无效：%s", name)
				}
				if policyName == "config/apps.json" {
					var apps map[string]models.Project
					if json.Unmarshal(b, &apps) != nil || apps == nil {
						return fmt.Errorf("应用配置必须是有效对象")
					}
				}
				if policyName == "config/user.json" {
					var user models.UserConfig
					if json.Unmarshal(b, &user) != nil || user.Username == "" || user.Password == "" {
						return fmt.Errorf("账号配置缺少必要字段")
					}
				}
			}
		default:
			return fmt.Errorf("备份包含不支持的文件类型")
		}
	}
}

func restoreData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	file, header, err := r.FormFile("backup_file")
	if err != nil {
		http.Error(w, "请选择备份文件", 400)
		return
	}
	defer file.Close()
	stage, err := os.MkdirTemp(filepath.Dir(config.DataDir), ".restore-*")
	if err != nil {
		http.Error(w, "无法创建恢复目录", 500)
		return
	}
	defer os.RemoveAll(stage)
	var reader io.Reader = file
	if strings.HasSuffix(header.Filename, ".tar.gz") {
		gz, err := gzip.NewReader(file)
		if err != nil {
			http.Error(w, "备份格式错误", 400)
			return
		}
		defer gz.Close()
		reader = gz
	} else if !strings.HasSuffix(header.Filename, ".tar") {
		http.Error(w, "仅支持 tar 或 tar.gz", 400)
		return
	}
	if err := stageRestore(reader, stage); err != nil {
		http.Error(w, "恢复校验失败："+err.Error(), 400)
		return
	}
	appOpMu.Lock()
	defer appOpMu.Unlock()
	if len(installReservations) > 0 {
		http.Error(w, "请先取消或完成所有安装任务", 409)
		return
	}
	var apps map[string]models.Project
	if err := LoadJSON(config.ConfigApps, &apps); err != nil {
		http.Error(w, "无法读取当前应用配置", 500)
		return
	}
	for _, app := range apps {
		checkAppStatus(&app)
		if app.PID > 0 {
			http.Error(w, "请先停止全部应用再恢复备份", 409)
			return
		}
	}
	// Check every destination before copying; preserve executable permissions.
	paths := []string{}
	err = filepath.Walk(stage, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(stage, path)
		if _, err := archiveTarget(config.DataDir, rel); err != nil {
			return err
		}
		paths = append(paths, rel)
		return nil
	})
	if err != nil {
		http.Error(w, "恢复路径校验失败", 400)
		return
	}
	for _, rel := range paths {
		source := filepath.Join(stage, rel)
		target, err := archiveTarget(config.DataDir, rel)
		if err != nil {
			http.Error(w, "恢复路径校验失败", 400)
			return
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			http.Error(w, "创建恢复目录失败", 500)
			return
		}
		data, err := os.ReadFile(source)
		if err != nil {
			http.Error(w, "读取备份失败", 500)
			return
		}
		info, _ := os.Stat(source)
		temp, err := os.CreateTemp(filepath.Dir(target), ".restore-file-*")
		if err != nil {
			http.Error(w, "恢复写入失败", 500)
			return
		}
		tempName := temp.Name()
		_, err = temp.Write(data)
		if err == nil {
			err = temp.Chmod(info.Mode().Perm())
		}
		closeErr := temp.Close()
		if err == nil {
			err = closeErr
		}
		if err == nil {
			err = os.Rename(tempName, target)
		}
		_ = os.Remove(tempName)
		if err != nil {
			http.Error(w, "恢复写入失败，请保留备份重试", 500)
			return
		}
	}
	for name := range generations {
		generations[name]++
	}
	http.Redirect(w, r, "/setting?ok=1", http.StatusSeeOther)
}
