package main

import (
	"fmt"
	"os"

	"github.com/aitdd/backend/internal/api"
	"github.com/aitdd/backend/internal/database"
	"github.com/spf13/cobra"
)

var servePort int

var rootCmd = &cobra.Command{
	Use:   "aitdd",
	Short: "AITDD - AI辅助可视化任务治理系统",
	Long: `AITDD是一个可视化任务治理系统，通过MCP接口供第三方AI编程软件调用，
提供本地SQLite数据库存储和React前端可视化界面。`,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动API服务器",
	Long:  `启动AITDD API服务器，监听指定端口并提供RESTful API服务。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化数据库
		if err := database.Initialize(nil); err != nil {
			fmt.Fprintf(os.Stderr, "数据库初始化失败: %v\n", err)
			os.Exit(1)
		}
		defer database.Close()

		// 设置路由
		router := api.SetupRouter()

		// 启动服务器
		addr := fmt.Sprintf(":%d", servePort)
		fmt.Printf("AITDD API服务器启动在 http://localhost%s\n", addr)
		if err := router.Run(addr); err != nil {
			fmt.Fprintf(os.Stderr, "服务器启动失败: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 34567, "服务器监听端口")
	rootCmd.AddCommand(serveCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
