package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"

	"lightpanel/config"
	"lightpanel/handlers"
)

func main() {
	if os.Getenv("LIGHTPANEL_DAEMON") == "1" && os.Getenv("RUNNING") != "1" {
		os.Setenv("RUNNING", "1")
		cmd := exec.Command(os.Args[0])
		cmd.Stdin = nil
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		configureDaemon(cmd)
		if err := cmd.Start(); err != nil {
			log.Fatalf("启动失败: %v", err)
		}
		log.Printf("服务已在后台启动，PID: %d", cmd.Process.Pid)
		os.Exit(0)
	}

	if err := handlers.InitConfig(); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	handlers.CleanupSessions()
	handlers.InitTemplates()
	handlers.SetupRoutes()

	go handlers.RunHook("panel_start")

	log.Printf("朱雀面板 %s", config.Version)
	log.Printf("端口: %s", config.Port)
	log.Printf("数据目录: %s", config.DataDir)
	log.Printf("配置文件: %s", config.ConfigDir)
	log.Printf("沙盒目录: %s", config.AppsDir)
	log.Printf("监听地址: %s:%s", config.Host, config.Port)

	if err := handlers.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("启动失败: %v", err)
	}
}
