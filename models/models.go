package models

import "time"

type Project struct {
	Name      string `json:"-"`
	Path      string `json:"path"`
	Cmd       string `json:"cmd"`
	AutoStart bool   `json:"auto_start"`
	Status    string `json:"status"`
	PID       int    `json:"pid"`
	Created   string `json:"created"`
}

type StoreSource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type StoreApp struct {
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Icon   string `json:"icon"`
	URL    string `json:"url"`
	Cmd    string `json:"cmd"`
	Author string `json:"author"`
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
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	URL      string    `json:"url"`
	Size     int64     `json:"size"`
	Downloaded int64   `json:"downloaded"`
	Status   string    `json:"status"`
	Progress int       `json:"progress"`
	Cmd      string    `json:"cmd"`
	last     time.Time `json:"-"`
}
