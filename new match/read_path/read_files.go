package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// IndexData 结构体用于解析整个 JSON 文件
type IndexData struct {
	FileDict map[string]struct {
		Path string `json:"path"`
	} `json:"file_dict"`
}

// 读取文件内容
func readFileContent(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("无法打开文件 %s: %v", filePath, err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	buffer := make([]byte, 4096) // 4KB 的缓冲区

	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("读取文件时出错: %v", err)
		}
		if n > 0 {
			// 这里可以添加内容匹配规则
			// 暂时不输出内容
		}
		if err == io.EOF {
			break
		}
	}
	return nil
}

func main() {
	// 读取索引文件
	indexFile, err := os.Open("file_index.json")
	if err != nil {
		fmt.Printf("无法打开索引文件: %v\n", err)
		return
	}
	defer indexFile.Close()

	// 解析 JSON 数据
	var indexData IndexData
	if err := json.NewDecoder(indexFile).Decode(&indexData); err != nil {
		fmt.Printf("解析 JSON 数据时出错: %v\n", err)
		return
	}

	// 创建输出文件
	outputFile, err := os.Create("output.txt")
	if err != nil {
		fmt.Printf("创建输出文件时出错: %v\n", err)
		return
	}
	defer outputFile.Close()

	// 创建带缓冲的写入器
	writer := bufio.NewWriter(outputFile)
	defer writer.Flush()

	// 先保存所有文件路径
	for path := range indexData.FileDict {
		writer.WriteString(path + "\n")
	}
	writer.Flush() // 确保所有路径都已写入文件

	// 再读取所有文件内容
	for path := range indexData.FileDict {
		if err := readFileContent(path); err != nil {
			fmt.Printf("读取文件内容时出错: %v\n", err)
			continue
		}
	}

	fmt.Println("文件路径已保存到 output.txt")
}
