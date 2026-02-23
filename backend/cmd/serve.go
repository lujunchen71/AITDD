package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/server"
	"github.com/aitdd/backend/internal/services"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动AITDD服务",
	Long: `启动AITDD HTTP服务器和前端服务。
默认监听 localhost:34567`,
	Run: runServe,
}

var (
	port int
)

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&port, "port", "p", 34567, "服务端口")
}

func runServe(cmd *cobra.Command, args []string) {
	// 加载配置
	configService := services.NewConfigService(".aitdd/config.json")
	config, err := configService.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		fmt.Println("提示: 请先运行 'aitdd init' 初始化项目")
		os.Exit(1)
	}

	// 使用配置中的端口（如果未指定）
	if port == 34567 && config.ServerPort != 0 {
		port = config.ServerPort
	}

	// 初始化数据库
	dbPath := config.Database.Path
	if dbPath == "" {
		dbPath = ".aitdd/data/aitdd.db"
	}
	if err := database.Initialize(&database.Config{Path: dbPath}); err != nil {
		fmt.Fprintf(os.Stderr, "初始化数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// 创建HTTP服务器
	httpServer := server.NewHTTPServer(port)

	// 启动服务器
	fmt.Printf("🚀 AITDD服务启动中...\n")
	fmt.Printf("📡 HTTP服务: http://localhost:%d\n", port)
	fmt.Printf("🌐 前端界面: http://localhost:%d\n", port)
	fmt.Printf("📊 API端点: http://localhost:%d/api/v1\n", port)
	fmt.Println("\n按 Ctrl+C 停止服务")

	// 启动服务（非阻塞）
	go func() {
		if err := httpServer.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "启动服务失败: %v\n", err)
			os.Exit(1)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n正在停止服务...")
	if err := httpServer.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "停止服务失败: %v\n", err)
	}
	fmt.Println("服务已停止")
}
