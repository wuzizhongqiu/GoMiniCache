package main

import (
	"GoMiniCache/config"
	"GoMiniCache/lib/logger"
	"GoMiniCache/tcp"
	EchoHandler "GoMiniCache/tcp"
	"fmt"
	"os"
)

const configFile string = "GoMiniCache.conf"

var defaultProperties = &config.ServerProperties{
	Bind: "0.0.0.0",
	Port: 9999,
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func main() {
	// 初始化日志配置
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "GoMiniCache",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	})

	// 初始化配置
	if fileExists(configFile) {
		config.SetupConfig(configFile)
	} else {
		config.Properties = defaultProperties
	}

	// 启动 GoMiniCache 的服务器
	err := tcp.ListenAndServeWithSignal(
		&tcp.Config{
			Address: fmt.Sprintf("%s:%d",
				config.Properties.Bind,
				config.Properties.Port),
		},
		EchoHandler.MakeHandler())
	if err != nil {
		logger.Error(err)
	}
}
