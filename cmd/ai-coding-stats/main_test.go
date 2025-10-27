package main

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessCSV(t *testing.T) {
	// 检查是否设置了 GitHub token
	githubToken := os.Getenv("GH_TOKEN")
	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_TOKEN")
	}

	if githubToken == "" {
		t.Fatal("GH_TOKEN or GITHUB_TOKEN not set - test requires valid token")
	}

	// 读取测试输入
	inputFile := filepath.Join("testdata", "test_input.csv")
	records, err := readCSV(inputFile)
	if err != nil {
		t.Fatalf("Failed to read test input: %v", err)
	}

	// 处理记录
	processedRecords, err := processRecords(records)
	if err != nil {
		t.Fatalf("Failed to process records: %v", err)
	}

	// 检查是否是生成模式
	genMode := os.Getenv("GEN_EXPECTED")
	if genMode == "true" || genMode == "1" {
		// 生成模式：写入 expected 文件
		expectedFile := filepath.Join("testdata", "test_expected.csv")
		err = writeCSVFile(expectedFile, processedRecords)
		if err != nil {
			t.Fatalf("Failed to write expected file: %v", err)
		}
		t.Logf("✅ Generated expected file: %s", expectedFile)
		return
	}

	// 验证模式：读取 expected 文件并对比
	expectedFile := filepath.Join("testdata", "test_expected.csv")
	expectedRecords, err := readCSV(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read expected file: %v", err)
	}

	// 对比结果
	if len(processedRecords) != len(expectedRecords) {
		t.Fatalf("Record count mismatch: got %d, expected %d", len(processedRecords), len(expectedRecords))
	}

	for i, processed := range processedRecords {
		expected := expectedRecords[i]
		if len(processed) != len(expected) {
			t.Errorf("Row %d: field count mismatch: got %d, expected %d", i, len(processed), len(expected))
			continue
		}

		for j := range processed {
			if processed[j] != expected[j] {
				t.Errorf("Row %d, Column %d: got %q, expected %q", i, j, processed[j], expected[j])
			}
		}
	}

	t.Logf("✅ Verified %d records", len(processedRecords)-1) // -1 for header
}

// writeCSVFile 写入 CSV 到文件
func writeCSVFile(filename string, records [][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return writer.Error()
}

// TestParseRepoURL 测试 URL 解析
func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://github.com/ShaoLongFei/XBot", "ShaoLongFei/XBot"},
		{"https://github.com/Wintercom/c-cube.git", "Wintercom/c-cube"},
		{"http://github.com/2788/qiniu-hackathon", "2788/qiniu-hackathon"},
		{"github.com/codefarmer009/codedance", "codefarmer009/codedance"},
		{"ShaoLongFei/XBot", "ShaoLongFei/XBot"},
		{"invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseRepoURL(tt.input)
			if result != tt.expected {
				t.Errorf("parseRepoURL(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestWriteCSV 测试 CSV 写入
func TestWriteCSV(t *testing.T) {
	records := [][]string{
		{"议题名称", "仓库地址", "访问状态", "合并PR总数", "AI生成PR数", "AI Coding浓度"},
		{"A.议题1：智能客服", "https://github.com/test/repo", "可访问", "10", "5", "50.00%"},
	}

	var buf bytes.Buffer
	err := writeCSV(&buf, records)
	if err != nil {
		t.Fatalf("writeCSV failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "议题名称") {
		t.Error("Output doesn't contain header")
	}
	if !strings.Contains(output, "test/repo") {
		t.Error("Output doesn't contain data")
	}
}
