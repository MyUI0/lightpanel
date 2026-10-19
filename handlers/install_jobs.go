package handlers

import (
	"context"
	"fmt"
	"io"
	"lightpanel/config"
	"lightpanel/models"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Reservations are protected by appOpMu. A task owns only its unique staging directory.
var installReservations = map[string]string{}
var installJobs sync.Map

type installPlan struct {
	App    models.StoreApp
	Source string
}
type installJob struct {
	control sync.Mutex
	mu      sync.Mutex
	plan    installPlan
	target  string
	task    *models.DownloadTask
	create  *CreateTask
	cancel  context.CancelFunc
	done    chan struct{}
	intent  string
}

func normalizedAppName(name string) (string, error) {
	name = strings.ReplaceAll(strings.TrimSpace(name), " ", "-")
	if name == "" || len(name) > 120 || strings.ContainsAny(name, "./\\; &|'?*:#\"<>\x00\r\n\t") {
		return "", fmt.Errorf("应用名称包含非法字符或过长")
	}
	return name, nil
}
func loadAppsStrict() (map[string]models.Project, error) {
	var apps map[string]models.Project
	if err := LoadJSON(config.ConfigApps, &apps); err != nil {
		return nil, fmt.Errorf("无法读取应用配置：%w", err)
	}
	if apps == nil {
		return nil, fmt.Errorf("应用配置必须是 JSON 对象")
	}
	return apps, nil
}
func pathKey(p string) string { p, _ = filepath.Abs(p); return strings.ToLower(filepath.Clean(p)) }
func checkInstallAvailable(apps map[string]models.Project, name, target string) error {
	for reserved := range installReservations {
		if pathsOverlap(reserved, target) {
			return fmt.Errorf("该应用已有进行中的安装任务")
		}
	}
	for n, a := range apps {
		normalized, _ := normalizedAppName(n)
		if normalized == name || pathsOverlap(a.Path, target) {
			return fmt.Errorf("应用名称或目录已被占用")
		}
	}
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		return fmt.Errorf("目标目录已存在或无法检查")
	}
	return nil
}
func submitInstall(plan installPlan, create *CreateTask) (*installJob, error) {
	name, err := normalizedAppName(plan.App.Name)
	if err != nil {
		return nil, err
	}
	plan.App.Name = name
	if plan.App.URL != "" {
		if err = validateRemoteURL(plan.App.URL); err != nil {
			return nil, err
		}
	}
	if !validateCommand(plan.App.Cmd) || !validateCommand(plan.App.SetupCmd) {
		return nil, fmt.Errorf("命令包含非法字符")
	}
	if plan.App.Port < 0 || plan.App.Port > 65535 {
		return nil, fmt.Errorf("端口无效")
	}
	appOpMu.Lock()
	defer appOpMu.Unlock()
	apps, err := loadAppsStrict()
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(config.AppsDir, 0755); err != nil {
		return nil, err
	}
	target, err := archiveTarget(config.AppsDir, name)
	if err != nil {
		return nil, err
	}
	if err = checkInstallAvailable(apps, name, target); err != nil {
		return nil, err
	}
	id := "dl_" + generateSessionToken()
	a := plan.App
	task := &models.DownloadTask{ID: id, Name: name, URL: a.URL, Cmd: a.Cmd, SetupCmd: a.SetupCmd, WorkDir: a.WorkDir, Version: a.Version, Icon: a.Icon, Port: a.Port, AutoExtract: a.AutoExtract, MakeExec: a.MakeExec, Status: "downloading", Last: time.Now()}
	job := &installJob{plan: plan, target: target, task: task, create: create}
	installReservations[pathKey(target)] = id
	installJobs.Store(id, job)
	downloadTasks.Store(id, task)
	job.start()
	return job, nil
}
func (j *installJob) start() {
	j.mu.Lock()
	defer j.mu.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	j.cancel = cancel
	j.done = make(chan struct{})
	j.intent = ""
	j.task.SetStatus("downloading")
	j.task.UpdateProgress(0, 0, 0)
	if j.create != nil {
		j.create.SetStatus("creating", "正在下载...")
	}
	go j.run(ctx, j.done)
}
func (j *installJob) release() {
	appOpMu.Lock()
	defer appOpMu.Unlock()
	if installReservations[pathKey(j.target)] == j.task.ID {
		delete(installReservations, pathKey(j.target))
	}
}
func (j *installJob) run(ctx context.Context, done chan struct{}) {
	var stage string
	var runErr error
	defer func() {
		if recovered := recover(); recovered != nil {
			runErr = fmt.Errorf("安装异常：%v", recovered)
		}
		if stage != "" {
			_ = os.RemoveAll(stage)
		}
		j.mu.Lock()
		status := j.task.GetStatus()
		if status != "completed" {
			status = "error"
			if j.intent != "" {
				status = j.intent
			}
			j.task.SetStatus(status)
			msg := "安装已取消"
			if runErr != nil {
				msg = runErr.Error()
			}
			j.task.SetMessage(msg)
			if j.create != nil {
				j.create.SetStatus(status, msg)
			}
		}
		if status != "paused" {
			j.release()
		}
		j.cancel = nil
		j.mu.Unlock()
		close(done)
		if status != "paused" {
			time.AfterFunc(30*time.Minute, func() {
				installJobs.Delete(j.task.ID)
				downloadTasks.Delete(j.task.ID)
				if j.create != nil {
					createTasks.Delete(j.create.ID)
				}
			})
		}
	}()
	stage, runErr = os.MkdirTemp(config.AppsDir, ".install-")
	if runErr != nil {
		return
	}
	a := j.plan.App
	if a.URL != "" {
		file := filepath.Join(stage, downloadFilename(a.URL))
		runErr = j.download(ctx, file)
		if runErr != nil {
			return
		}
		if ctx.Err() != nil {
			runErr = ctx.Err()
			return
		}
		j.task.SetStatus("deploying")
		if a.AutoExtract {
			if !extractArchive(file, stage) {
				runErr = fmt.Errorf("解压失败：文件不是受支持的有效归档")
				return
			}
			if runErr = os.Remove(file); runErr != nil {
				return
			}
		} else if a.MakeExec {
			if runErr = os.Chmod(file, 0755); runErr != nil {
				return
			}
		}
	}
	if ctx.Err() != nil {
		runErr = ctx.Err()
		return
	}
	// Runtime markers belong to this panel, never to a downloaded package.
	for _, marker := range []string{"pid.pid", ".setup_done"} {
		if err := os.Remove(filepath.Join(stage, marker)); err != nil && !os.IsNotExist(err) {
			runErr = err
			return
		}
	}
	cmd := a.Cmd
	workDir := a.WorkDir
	if workDir != "" && !filepath.IsAbs(workDir) {
		workDir, runErr = archiveTarget(j.target, workDir)
		if runErr != nil {
			return
		}
	}
	if cmd == "" {
		detectionDir := stage
		if workDir != "" {
			rel, err := filepath.Rel(j.target, workDir)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
				runErr = fmt.Errorf("使用外部工作目录时请填写启动命令")
				return
			}
			detectionDir = filepath.Join(stage, rel)
		}
		cmd = findBinaryInDir(detectionDir)
	}
	project := models.Project{Path: j.target, Cmd: cmd, SetupCmd: a.SetupCmd, WorkDir: workDir, SourceURL: j.plan.Source, PackageURL: a.URL, Version: a.Version, Icon: a.Icon, Port: a.Port, Created: time.Now().Format("2006-01-02 15:04")}
	if a.Port > 0 {
		project.URL = fmt.Sprintf("http://%s:%d", getLocalIP(), a.Port)
	}
	// Serialize cancellation against the final commit. A successful commit is never undone by deleting its task record.
	j.mu.Lock()
	runErr = func() error {
		defer j.mu.Unlock()
		appOpMu.Lock()
		defer appOpMu.Unlock()
		if ctx.Err() != nil {
			return ctx.Err()
		}
		apps, err := loadAppsStrict()
		if err != nil {
			return err
		}
		for n, p := range apps {
			normalized, _ := normalizedAppName(n)
			if normalized == a.Name || pathsOverlap(p.Path, j.target) {
				return fmt.Errorf("提交失败：应用已存在")
			}
		}
		if _, err := os.Lstat(j.target); !os.IsNotExist(err) {
			return fmt.Errorf("提交失败：目标目录已存在")
		}
		if _, err := archiveTarget(config.AppsDir, a.Name); err != nil {
			return err
		}
		if err := os.Rename(stage, j.target); err != nil {
			return err
		}
		apps[a.Name] = project
		if err := WriteJSON(config.ConfigApps, apps); err != nil {
			if rollback := os.Rename(j.target, stage); rollback != nil {
				return fmt.Errorf("保存失败且回退失败，文件保留在 %s：%v", j.target, err)
			}
			return err
		}
		stage = ""
		j.task.UpdateProgress(j.task.GetDownloaded(), j.task.GetSize(), 100)
		j.task.SetStatus("completed")
		j.task.SetMessage("应用已创建")
		if j.create != nil {
			j.create.SetProgress(100, "")
			j.create.SetStatus("completed", "创建完成")
		}
		return nil
	}()
}
func (j *installJob) download(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", j.plan.App.URL, nil)
	if err != nil {
		return err
	}
	resp, err := remoteClient(120 * time.Second).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败：HTTP %d", resp.StatusCode)
	}
	limit := int64(config.MaxDownBytes)
	if resp.ContentLength > limit {
		return fmt.Errorf("下载文件超过大小限制")
	}
	j.task.UpdateSize(resp.ContentLength)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	reader := io.LimitReader(resp.Body, limit+1)
	buf := make([]byte, 32768)
	var total int64
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		n, e := reader.Read(buf)
		if total+int64(n) > limit {
			return fmt.Errorf("下载文件超过大小限制")
		}
		if n > 0 {
			if _, err = f.Write(buf[:n]); err != nil {
				return err
			}
			total += int64(n)
			p := 0
			if resp.ContentLength > 0 {
				p = int(total * 90 / resp.ContentLength)
			}
			j.task.UpdateProgress(total, resp.ContentLength, p)
			if j.create != nil {
				j.create.SetProgress(p, "正在下载...")
			}
		}
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}
	}
	if resp.ContentLength >= 0 && total != resp.ContentLength {
		return fmt.Errorf("下载不完整")
	}
	if total == 0 {
		return fmt.Errorf("下载文件为空")
	}
	if err = f.Sync(); err != nil {
		return err
	}
	return f.Close()
}
func (j *installJob) action(action string) error {
	j.control.Lock()
	defer j.control.Unlock()
	j.mu.Lock()
	status := j.task.GetStatus()
	if action == "resume" {
		done := j.done
		j.mu.Unlock()
		if status != "paused" {
			return fmt.Errorf("任务未暂停")
		}
		<-done
		j.start()
		return nil
	}
	if action != "pause" && action != "delete" {
		j.mu.Unlock()
		return fmt.Errorf("无效操作")
	}
	if action == "pause" && status != "downloading" {
		j.mu.Unlock()
		return fmt.Errorf("仅下载中的任务可暂停")
	}
	if status == "downloading" || status == "deploying" {
		j.intent = "paused"
		if action == "delete" {
			j.intent = "cancelled"
		}
		j.cancel()
	}
	done := j.done
	j.mu.Unlock()
	<-done
	if action == "delete" {
		j.release()
		installJobs.Delete(j.task.ID)
		downloadTasks.Delete(j.task.ID)
		if j.create != nil {
			j.create.SetStatus("cancelled", "下载任务已取消")
		}
	}
	return nil
}
