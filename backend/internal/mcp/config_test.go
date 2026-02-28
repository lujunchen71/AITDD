package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigManager_Load(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "aitdd-config-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建 .aitdd 目录
	aitddDir := filepath.Join(tempDir, ".aitdd")
	if err := os.MkdirAll(aitddDir, 0755); err != nil {
		t.Fatalf("创建 .aitdd 目录失败: %v", err)
	}

	t.Run("加载存在的配置文件", func(t *testing.T) {
		configPath := filepath.Join(aitddDir, "project.json")
		expectedConfig := &Config{
			PathName:   "test-project",
			ApiBaseUrl: "http://localhost:8080/api/v1",
		}
		configData, _ := json.Marshal(expectedConfig)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("写入配置文件失败: %v", err)
		}

		cm := NewConfigManager(configPath)
		if cm.GetProjectPathName() != expectedConfig.PathName {
			t.Errorf("期望 pathName = %s, 实际 = %s", expectedConfig.PathName, cm.GetProjectPathName())
		}
		if cm.GetApiBaseUrl() != expectedConfig.ApiBaseUrl {
			t.Errorf("期望 apiBaseUrl = %s, 实际 = %s", expectedConfig.ApiBaseUrl, cm.GetApiBaseUrl())
		}
	})

	t.Run("加载不存在的配置文件时创建默认配置", func(t *testing.T) {
		configPath := filepath.Join(aitddDir, "nonexistent.json")
		cm := NewConfigManager(configPath)

		if cm.GetApiBaseUrl() != "http://localhost:34567/api/v1" {
			t.Errorf("期望默认 apiBaseUrl = http://localhost:34567/api/v1, 实际 = %s", cm.GetApiBaseUrl())
		}
	})
}

func TestConfigManager_Save(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "aitdd-config-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "project.json")
	cm := NewConfigManager(configPath)

	// 使用 SetProject 更新配置
	if err := cm.SetProject("", "new-project", "new-project"); err != nil {
		t.Fatalf("保存配置失败: %v", err)
	}

	// 重新加载验证
	cm2 := NewConfigManager(configPath)
	if cm2.GetProjectPathName() != "new-project" {
		t.Errorf("期望 pathName = new-project, 实际 = %s", cm2.GetProjectPathName())
	}
}

func TestConfigManager_SetApiBaseUrl(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "aitdd-config-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "project.json")
	cm := NewConfigManager(configPath)

	newUrl := "http://custom-host:1234/api/v1"
	cm.SetApiBaseUrl(newUrl)

	if cm.GetApiBaseUrl() != newUrl {
		t.Errorf("期望 apiBaseUrl = %s, 实际 = %s", newUrl, cm.GetApiBaseUrl())
	}
}

func TestConfigManager_SetProject(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "aitdd-config-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "project.json")
	cm := NewConfigManager(configPath)

	newPathName := "my-awesome-project"
	if err := cm.SetProject("project-id", "My Awesome Project", newPathName); err != nil {
		t.Fatalf("SetProject 失败: %v", err)
	}

	if cm.GetProjectPathName() != newPathName {
		t.Errorf("期望 pathName = %s, 实际 = %s", newPathName, cm.GetProjectPathName())
	}
}

func TestConfigManager_GetConfig(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "aitdd-config-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "project.json")
	cm := NewConfigManager(configPath)

	config := cm.GetConfig()
	if config == nil {
		t.Fatal("配置不应为 nil")
	}
}

func TestConfigManager_ConcurrentAccess(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "aitdd-config-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "project.json")
	cm := NewConfigManager(configPath)

	// 并发读写测试
	done := make(chan bool)

	// 并发写入
	for i := 0; i < 10; i++ {
		go func(idx int) {
			projectName := string(rune('0' + idx))
			_ = cm.SetProject("", projectName, projectName)
			done <- true
		}(i)
	}

	// 并发读取
	for i := 0; i < 10; i++ {
		go func() {
			_ = cm.GetProjectPathName()
			_ = cm.GetApiBaseUrl()
			done <- true
		}()
	}

	// 等待所有操作完成
	for i := 0; i < 20; i++ {
		<-done
	}
}
