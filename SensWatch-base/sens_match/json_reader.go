package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// ReadFileList 从JSON文件读取文件列表
func ReadFileList(filePath string) ([]FileInfo, error) {
	// 读取JSON文件
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	// 解析JSON
	if err := json.Unmarshal(data, &fileDict); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	// 将map转换为slice
	var fileList []FileInfo
	for _, info := range fileDict.FileDict {
		fileList = append(fileList, info)
	}

	return fileList, nil
}
