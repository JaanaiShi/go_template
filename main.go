package main

import (
	"context"
	"fmt"
	"github.com/JaanaiShi/flint/common"
	"github.com/JaanaiShi/flint/initialize/config"
	"github.com/JaanaiShi/flint/initialize/db"
	"github.com/JaanaiShi/flint/initialize/logger"
	"github.com/JaanaiShi/flint/initialize/redis"
)

func main() {
	// 初始化日志
	ctx := context.Background()
	config.Init()
	fmt.Println("config:", common.Conf)
	l := logger.NewLogger("")
	logger.InitLog(l.Logger)

	logger.Info(ctx, "初始化日志完成")

	logger.Info(ctx, "初始化mysql数据库 start")
	db.Init()
	logger.Info(ctx, "初始化mysql数据库 end")

	logger.Info(ctx, "初始化redis数据库 start")
	redis.Init()
	logger.Info(ctx, "初始化redis数据库 end")

}
