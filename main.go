package main

import (
	"fmt"

	"lightpanel/config"
	"lightpanel/handlers"
)

func main() {
	if err := handlers.InitConfig(); err != nil {
		fmt.Printf("配置初始化失败: %v\n", err)
		return
	}

	handlers.InitTemplates()
	handlers.SetupRoutes()

	fmt.Printf("LightPanel %s\n", config.Version)
	fmt.Printf("端口: %s\n", config.Port)
	fmt.Printf("地址: http://127.0.0.1:%s\n", config.Port)
	fmt.Printf("数据: %s\n", config.DataDir)
	fmt.Printf("默认: admin / admin\n")
	fmt.Println("---")

	if err := handlers.ListenAndServe(); err != nil {
		fmt.Printf("启动失败: %v\n", err)
	}
}
