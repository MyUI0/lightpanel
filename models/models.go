package models

import (
	"sync"
	"time"
)

type Project struct {
	Name      string `json:"-"`
	Path      string `json:"path"`
	Cmd       string `json:"cmd"`
	SetupCmd  string `json:"setup_cmd,omitempty"`
	WorkDir   string `json:"work_dir,omitempty"`
	SourceURL string `json:"source_url,omitempty"`
	URL       string `json:"url,omitempty"`
	Port      int    `json:"port,omitempty"`
	AutoStart bool   `json:"auto_start"`
	Status    string `json:"status"`
	PID       int    `json:"pid"`
	Created   string `json:"created"`
	Version   string `json:"version,omitempty"`
}

type StoreSource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type StoreApp struct {
	Name        string `json:"name"`
	Desc        string `json:"desc"`
	Icon        string `json:"icon"`
	URL         string `json:"url"`
	Cmd         string `json:"cmd"`
	SetupCmd    string `json:"setup_cmd,omitempty"`
	WorkDir     string `json:"work_dir,omitempty"`
	Author      string `json:"author"`
	Version     string `json:"version,omitempty"`
	AutoExtract bool   `json:"auto_extract,omitempty"`
	MakeExec    bool   `json:"make_exec,omitempty"`
	AskParams   bool   `json:"ask_params,omitempty"`
	ParamsHint  string `json:"params_hint,omitempty"`
	Port        int    `json:"port,omitempty"`
}

type StoreManifest struct {
	SchemaVersion int            `json:"schema_version"`
	SourceName    string         `json:"source_name"`
	Maintainer    string         `json:"maintainer,omitempty"`
	Apps          []StoreApp     `json:"apps"`
}

type UserConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BgConfig struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type ProcInfo struct {
	PID  int32
	Name string
	Cpu  float64
	Mem  float64
}

type DownloadTask struct {
	mu        sync.Mutex
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Size      int64     `json:"size"`
	Downloaded int64    `json:"downloaded"`
	Status    string    `json:"status"`
	Progress  int       `json:"progress"`
	Cmd       string    `json:"cmd"`
	Version   string    `json:"version"`
	SetupCmd  string    `json:"setup_cmd"`
	WorkDir   string    `json:"work_dir"`
	Last        time.Time `json:"-"`
	AutoExtract bool      `json:"auto_extract"`
	MakeExec    bool      `json:"make_exec"`
	Port        int       `json:"port"`
}

func (t *DownloadTask) SetStatus(s string) {
	t.mu.Lock()
	t.Status = s
	t.mu.Unlock()
}

func (t *DownloadTask) GetStatus() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Status
}

func (t *DownloadTask) IsDownloading() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Status == "downloading"
}

func (t *DownloadTask) Snapshot() DownloadTask {
	t.mu.Lock()
	defer t.mu.Unlock()
	return DownloadTask{
		ID:         t.ID,
		Name:       t.Name,
		URL:        t.URL,
		Size:       t.Size,
		Downloaded: t.Downloaded,
		Status:     t.Status,
		Progress:   t.Progress,
		Cmd:        t.Cmd,
		Last:       t.Last,
	}
}

func (t *DownloadTask) UpdateProgress(downloaded int64, size int64, progress int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Downloaded = downloaded
	if size > 0 {
		t.Size = size
	}
	t.Progress = progress
	t.Last = time.Now()
}

func (t *DownloadTask) UpdateSize(size int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Size = size
}

func (t *DownloadTask) GetSize() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Size
}

func (t *DownloadTask) GetDownloaded() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Downloaded
}

func (t *DownloadTask) GetProgress() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Progress
}

type DownloadHistory struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Cmd       string    `json:"cmd,omitempty"`
	Version   string    `json:"version,omitempty"`
	Status    string    `json:"status"`
	Size      int64     `json:"size"`
	Downloaded int64    `json:"downloaded"`
	Progress  int       `json:"progress"`
	Installed bool      `json:"installed"`
	Timestamp time.Time `json:"timestamp"`
}
