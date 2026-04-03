package main

import (
	"log"

	"lightpanel/config"
	"lightpanel/handlers"
)

func main() {
	if err := handlers.InitConfig(); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	handlers.CleanupSessions()
	handlers.InitTemplates()
	handlers.SetupRoutes()

	go handlers.RunHook("panel_start")

	log.Printf("LightPanel %s", config.Version)
	log.Printf("端口: %s", config.Port)
	log.Printf("数据目录: %s", config.DataDir)
	log.Printf("配置文件: %s", config.ConfigDir)
	log.Printf("沙盒目录: %s", config.AppsDir)
	log.Printf("默认账号: admin / admin")

	if err := handlers.ListenAndServe(); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
