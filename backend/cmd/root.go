package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aitdd",
	Short: "AITDD - AI辅助可视化任务治理系统",
	Long: `AITDD是一个可视化任务治理系统，通过MCP接口供第三方AI编程软件调用。
提供本地SQLite数据库存储和React前端可视化界面。`,
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
