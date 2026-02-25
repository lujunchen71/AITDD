package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aitdd/backend/internal/services"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化AITDD项目",
	Long: `初始化AITDD项目，创建必要的目录结构和配置文件。
支持选择不同的AI插件类型（KiloCode、OpenCode、ClaudeCode）。`,
	Run: runInit,
}

var (
	pluginType string
	projectName string
)

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&pluginType, "plugin", "p", "", "AI插件类型 (kilocode, opencode, claudecode)")
	initCmd.Flags().StringVarP(&projectName, "name", "n", "", "项目名称")
}

func runInit(cmd *cobra.Command, args []string) {
	fmt.Println("🚀 初始化AITDD项目...")

	// 获取当前目录作为项目目录
	projectDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取当前目录失败: %v\n", err)
		os.Exit(1)
	}

	// 如果没有指定插件类型，使用交互式选择
	if pluginType == "" {
		pluginType = selectPluginType()
	}

	// 如果没有指定项目名称，使用目录名
	if projectName == "" {
		projectName = filepath.Base(projectDir)
	}

	// 验证插件类型
	if !isValidPluginType(pluginType) {
		fmt.Fprintf(os.Stderr, "无效的插件类型: %s\n", pluginType)
		fmt.Println("支持的插件类型: kilocode, opencode, claudecode")
		os.Exit(1)
	}

	// 创建初始化服务
	initService := services.NewInitService(projectDir, projectName, pluginType)

	// 执行初始化
	if err := initService.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "初始化失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 项目初始化完成!")
	fmt.Printf("📁 项目目录: %s\n", projectDir)
	fmt.Printf("🔌 插件类型: %s\n", pluginType)
	fmt.Println("\n下一步:")
	fmt.Println("  1. 运行 'aitdd serve' 启动服务")
	fmt.Println("  2. 在浏览器访问 http://localhost:34567")
}

func selectPluginType() string {
	fmt.Println("\n请选择AI插件类型:")
	fmt.Println("  1. KiloCode")
	fmt.Println("  2. OpenCode")
	fmt.Println("  3. ClaudeCode")
	fmt.Print("\n请输入选项 (1-3): ")

	var choice int
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		return "kilocode"
	case 2:
		return "opencode"
	case 3:
		return "claudecode"
	default:
		fmt.Println("无效选项，使用默认值: kilocode")
		return "kilocode"
	}
}

func isValidPluginType(pluginType string) bool {
	validTypes := []string{"kilocode", "opencode", "claudecode"}
	for _, t := range validTypes {
		if t == pluginType {
			return true
		}
	}
	return false
}
