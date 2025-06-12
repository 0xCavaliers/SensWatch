package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// readDocx 读取docx文件内容
func readDocx(path string) (io.Reader, error) {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", path)
	}

	// 检查文件大小
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}
	if fileInfo.Size() == 0 {
		return nil, fmt.Errorf("文件为空: %s", path)
	}

	// 打开docx文件（实际上是一个zip文件）
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("打开docx文件失败: %v", err)
	}
	defer reader.Close()

	// 检查是否是有效的docx文件
	foundDocumentXML := false
	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			foundDocumentXML = true
			break
		}
	}
	if !foundDocumentXML {
		return nil, fmt.Errorf("不是有效的docx文件: %s", path)
	}

	// 查找document.xml文件
	var content strings.Builder
	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("打开document.xml失败: %v", err)
			}
			defer rc.Close()

			// 读取XML内容
			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("读取document.xml失败: %v", err)
			}

			// 使用更底层的XML解析
			decoder := xml.NewDecoder(bytes.NewReader(data))
			for {
				token, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					return nil, fmt.Errorf("解析XML失败: %v", err)
				}

				switch t := token.(type) {
				case xml.StartElement:
					if t.Name.Local == "t" {
						var text string
						if err := decoder.DecodeElement(&text, &t); err != nil {
							return nil, fmt.Errorf("解析文本节点失败: %v", err)
						}
						content.WriteString(text)
					}
				}
			}
			break
		}
	}

	return strings.NewReader(content.String()), nil
}

// readPdf 读取pdf文件内容
func readPdf(path string) (io.Reader, error) {
	// 创建临时文件用于存储Python脚本的输出
	tempFile := path + ".txt"
	defer os.Remove(tempFile)

	// 构建Python命令
	cmd := exec.Command("python", "-c", `
import PyPDF2
import sys

def extract_text_from_pdf(pdf_path):
    text = ""
    with open(pdf_path, 'rb') as file:
        reader = PyPDF2.PdfReader(file)
        for i, page in enumerate(reader.pages):
            page_text = page.extract_text()
            if page_text:
                text += f"\n--- Page {i+1} ---\n{page_text}\n"
    return text

pdf_file_path = sys.argv[1]
extracted_text = extract_text_from_pdf(pdf_file_path)
with open(sys.argv[2], "w", encoding="utf-8") as f:
    f.write(extracted_text)
`, path, tempFile)

	// 执行Python脚本
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("执行Python脚本失败: %v", err)
	}

	// 读取生成的文本文件
	content, err := os.ReadFile(tempFile)
	if err != nil {
		return nil, fmt.Errorf("读取PDF提取结果失败: %v", err)
	}

	return bytes.NewReader(content), nil
}

// readXlsx 读取xlsx文件内容
func readXlsx(path string) (io.Reader, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("打开xlsx文件失败: %v", err)
	}
	defer f.Close()

	var content strings.Builder
	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil {
			return nil, fmt.Errorf("读取xlsx工作表失败: %v", err)
		}
		for _, row := range rows {
			for _, cell := range row {
				content.WriteString(cell)
				content.WriteString("\t")
			}
			content.WriteString("\n")
		}
	}
	return strings.NewReader(content.String()), nil
}

// readPptx 读取pptx文件内容
func readPptx(path string) (io.Reader, error) {
	// 打开pptx文件（实际上是一个zip文件）
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("打开pptx文件失败: %v", err)
	}
	defer reader.Close()

	var content strings.Builder
	// 遍历所有幻灯片
	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "ppt/slides/slide") && strings.HasSuffix(file.Name, ".xml") {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("打开幻灯片文件失败: %v", err)
			}
			defer rc.Close()

			// 读取XML内容
			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("读取幻灯片内容失败: %v", err)
			}

			// 解析XML内容
			decoder := xml.NewDecoder(bytes.NewReader(data))
			for {
				token, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					return nil, fmt.Errorf("解析XML失败: %v", err)
				}

				switch t := token.(type) {
				case xml.StartElement:
					if t.Name.Local == "t" {
						var text string
						if err := decoder.DecodeElement(&text, &t); err != nil {
							return nil, fmt.Errorf("解析文本节点失败: %v", err)
						}
						content.WriteString(text)
						content.WriteString(" ")
					}
				}
			}
			content.WriteString("\n")
		}
	}

	return strings.NewReader(content.String()), nil
}

// GetFileReader 根据文件扩展名获取文件读取器
func GetFileReader(path string) (io.Reader, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".docx":
		return readDocx(path)
	case ".pdf":
		return readPdf(path)
	case ".xlsx":
		return readXlsx(path)
	case ".pptx":
		return readPptx(path)
	default:
		// 对于其他文件类型，直接返回文件句柄
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("打开文件失败: %v", err)
		}
		return file, nil
	}
}
