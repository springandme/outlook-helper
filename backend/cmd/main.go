package main

import (
	"log"
	"outlook-helper/backend/internal/api"
	"outlook-helper/backend/internal/config"
	"outlook-helper/backend/internal/database"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库连接
	conn, err := database.Initialize(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer conn.Close()

	// 运行数据库迁移
	if err := database.Migrate(conn); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// 创建数据库管理器
	db := database.NewDB(conn)

	// 初始化种子数据
	if err := database.SeedData(conn); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	}

	// 检查数据库完整性
	if err := database.CheckDatabaseIntegrity(conn); err != nil {
		log.Printf("Warning: Database integrity check failed: %v", err)
	}

	// 启动API服务器
	server := api.NewServer(cfg, db)
	log.Printf("Starting server on port %s", cfg.Port)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
