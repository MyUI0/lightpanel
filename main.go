package main

import (
	"log"
	"os"

	"github.com/sevlyar/go-daemon"
	"lightpanel/config"
	"lightpanel/handlers"
)

func main() {
	cntxt := &daemon.Context{
		PidFileName: "lightpanel.pid",
		PidFilePerm: 0644,
		WorkDir:     ".",
		Umask:       027,
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatalf("守护进程化失败: %v", err)
	}
	if d != nil {
		log.Printf("服务已在后台启动，PID: %d", d.Pid)
		return
	}
	defer cntxt.Release()

	log.SetFlags(0)
	log.SetOutput(os.Stdout)

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
	log.Printf("默认账号: admin / admin")

	if err := handlers.ListenAndServe(); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
