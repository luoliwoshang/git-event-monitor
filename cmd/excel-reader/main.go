package main

import (
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法: go run main.go <excel文件路径>")
		fmt.Println("示例: go run main.go \"议题三 待筛选 名单.xlsx\"")
		os.Exit(1)
	}

	filename := os.Args[1]

	// 打开Excel文件
	fmt.Printf("📖 正在读取Excel文件: %s\n", filename)
	file, err := excelize.OpenFile(filename)
	if err != nil {
		fmt.Printf("❌ 无法打开文件: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		// 关闭Excel文件
		if err := file.Close(); err != nil {
			fmt.Printf("⚠️ 关闭文件时出错: %v\n", err)
		}
	}()

	// 获取所有工作表名称
	sheets := file.GetSheetList()
	fmt.Printf("📋 发现 %d 个工作表: %v\n\n", len(sheets), sheets)

	// 遍历每个工作表
	for i, sheetName := range sheets {
		fmt.Printf("==== 工作表 %d: %s ====\n", i+1, sheetName)

		// 获取工作表中的所有行
		rows, err := file.GetRows(sheetName)
		if err != nil {
			fmt.Printf("❌ 读取工作表 '%s' 失败: %v\n", sheetName, err)
			continue
		}

		if len(rows) == 0 {
			fmt.Printf("⚠️ 工作表 '%s' 为空\n\n", sheetName)
			continue
		}

		// 显示表头（第一行）
		if len(rows) > 0 {
			fmt.Printf("🏷️ 表头: ")
			for j, cell := range rows[0] {
				if j > 0 {
					fmt.Print(" | ")
				}
				fmt.Printf("%s", cell)
			}
			fmt.Println()
		}

		// 显示数据行数和前几行示例
		dataRows := len(rows) - 1 // 减去表头
		fmt.Printf("📊 数据行数: %d\n", dataRows)

		// 显示前5行数据作为示例
		displayRows := 5
		if dataRows < displayRows {
			displayRows = dataRows
		}

		if displayRows > 0 {
			fmt.Printf("📝 前 %d 行数据:\n", displayRows)
			for i := 1; i <= displayRows; i++ {
				if i < len(rows) {
					fmt.Printf("  行 %d: ", i)
					for j, cell := range rows[i] {
						if j > 0 {
							fmt.Print(" | ")
						}
						// 限制每个单元格显示长度，避免输出过长
						if len(cell) > 30 {
							fmt.Printf("%.30s...", cell)
						} else {
							fmt.Printf("%s", cell)
						}
					}
					fmt.Println()
				}
			}
		}

		fmt.Println()
	}

	fmt.Printf("✅ Excel文件读取完成！\n")
}