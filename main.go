package main

import (
	"log"
	"nezha_sec/internal/views"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// 初始化应用
	log.Println("初始化应用...")
	m, err := views.NewChatModel()
	if err != nil {
		log.Fatalf("初始化模型失败: %v", err)
	}
	
	// 创建程序实例
	p := tea.NewProgram(m)
	
	// 运行程序
	log.Println("启动 TUI 应用...")
	if _, err := p.Run(); err != nil {
		log.Printf("应用错误: %v", err)
	}
	
	log.Println("应用已退出")
}
