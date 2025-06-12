package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// calculateMD5 计算文件的 MD5 值
func calculateMD5(filePath string) (string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 创建 MD5 哈希对象
	hash := md5.New()

	// 读取文件内容并计算哈希值
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("读取文件内容失败: %v", err)
	}

	// 获取 MD5 值并转换为十六进制字符串
	md5Value := hex.EncodeToString(hash.Sum(nil))
	return md5Value, nil
}
