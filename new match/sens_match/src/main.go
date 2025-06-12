package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// 全局变量
var fileDict FileDict
var results []SensitiveInfo

// 不需要检测的文件扩展名
var skipExtensions = map[string]bool{
	".jpg":   true,
	".jpeg":  true,
	".png":   true,
	".gif":   true,
	".bmp":   true,
	".ico":   true,
	".svg":   true,
	".webp":  true,
	".mp3":   true,
	".mp4":   true,
	".avi":   true,
	".mov":   true,
	".wmv":   true,
	".zip":   true,
	".rar":   true,
	".7z":    true,
	".tar":   true,
	".gz":    true,
	".msi":   true,
	".exe":   true,
	".dll":   true,
	".so":    true,
	".dylib": true,
}

// copyFileIndex 从read_path文件夹复制file_index.json到sens_match文件夹
func copyFileIndex() error {
	// 获取当前可执行文件所在目录（工作目录）
	exeDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %v", err)
	}

	// 计算正确的相对路径：从工作目录向上两级到new_match，再进入read_path
	baseDir := filepath.Dir(filepath.Dir(exeDir))
	sourcePath := filepath.Join(baseDir, "read_path", "file_index.json")
	targetPath := filepath.Join(filepath.Dir(exeDir), "file_index.json")

	// 检查源文件是否存在
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("源文件不存在: %s", sourcePath)
	}

	// 打开源文件
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %v", err)
	}
	defer sourceFile.Close()

	// 创建目标文件
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer targetFile.Close()

	// 复制文件内容
	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %v", err)
	}

	fmt.Printf("成功复制 file_index.json 从 %s 到 %s\n", sourcePath, targetPath)
	return nil
}

// FileProcessor 处理文件扫描和敏感信息检测
type FileProcessor struct {
	sensMatch *SensMatch
}

// NewFileProcessor 创建新的 FileProcessor 实例
func NewFileProcessor() *FileProcessor {
	return &FileProcessor{
		sensMatch: NewSensMatch(),
	}
}

// shouldSkipFile 检查是否应该跳过该文件
func shouldSkipFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return skipExtensions[ext]
}

// ProcessFile 处理单个文件
func (p *FileProcessor) ProcessFile(filePath string) (*SensitiveInfo, error) {
	// 检查是否需要跳过该文件
	if shouldSkipFile(filePath) {
		return nil, fmt.Errorf("跳过不支持的文件类型: %s", filePath)
	}

	// 获取文件读取器
	reader, err := GetFileReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件读取器失败: %v", err)
	}

	// 计算MD5
	md5Value, err := calculateMD5(filePath)
	if err != nil {
		return nil, fmt.Errorf("计算MD5失败: %v", err)
	}

	// 流式读取文件内容
	bufferSize := 4096 // 每次读取4KB
	buffer := make([]byte, bufferSize)
	var allMatches map[string][]string
	allMatches = make(map[string][]string)

	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("读取文件失败: %v", err)
		}
		if n == 0 {
			break
		}

		// 对当前块进行敏感信息检测
		chunkMatches := p.sensMatch.RunAllChecks(string(buffer[:n]))

		// 合并结果
		for k, v := range chunkMatches {
			allMatches[k] = append(allMatches[k], v...)
		}
	}

	// 统计匹配数量
	matchCounts := make(map[string]int)
	totalCount := 0
	ruleNumbersMap := make(map[int]bool) // 使用map来去重
	for k, v := range allMatches {
		matchCounts[k] = len(v)
		totalCount += len(v)
		// 根据规则名称获取规则编号
		switch k {
		case "phone":
			ruleNumbersMap[1] = true
		case "ip":
			ruleNumbersMap[2] = true
		case "mac":
			ruleNumbersMap[3] = true
		case "ipv6":
			ruleNumbersMap[4] = true
		case "bank_card":
			ruleNumbersMap[5] = true
		case "email":
			ruleNumbersMap[6] = true
		case "passport":
			ruleNumbersMap[7] = true
		case "id_number":
			ruleNumbersMap[8] = true
		case "gender":
			ruleNumbersMap[9] = true
		case "national":
			ruleNumbersMap[10] = true
		case "carnum":
			ruleNumbersMap[11] = true
		case "telephone":
			ruleNumbersMap[12] = true
		case "officer":
			ruleNumbersMap[13] = true
		case "HM_pass":
			ruleNumbersMap[14] = true
		case "jdbc":
			ruleNumbersMap[15] = true
		case "organization":
			ruleNumbersMap[16] = true
		case "business":
			ruleNumbersMap[17] = true
		case "credit":
			ruleNumbersMap[18] = true
		case "address_name":
			ruleNumbersMap[19] = true
		}
	}

	// 将规则编号转换为有序字符串
	var ruleNumbers []int
	for num := range ruleNumbersMap {
		ruleNumbers = append(ruleNumbers, num)
	}
	// 排序
	sort.Ints(ruleNumbers)

	// 将规则编号转换为字符串，用顿号分隔
	ruleNumbersStr := ""
	if len(ruleNumbers) > 0 {
		strNums := make([]string, len(ruleNumbers))
		for i, num := range ruleNumbers {
			strNums[i] = fmt.Sprintf("%d", num)
		}
		ruleNumbersStr = strings.Join(strNums, "、")
	}

	// 获取文件的绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件绝对路径失败: %v", err)
	}

	return &SensitiveInfo{
		FileName:            filepath.Base(filePath), // 只使用文件名
		FilePath:            absPath,                 // 使用绝对路径
		MD5:                 md5Value,
		DetectTime:          time.Now().Format("2006-01-02 15:04:05"),
		MatchCounts:         matchCounts,
		Matches:             allMatches,
		TotalSensitiveCount: totalCount,
		RuleNumbers:         ruleNumbersStr,
	}, nil
}

// ProcessFileList 处理文件列表
func (p *FileProcessor) ProcessFileList(files []FileInfo) []SensitiveInfo {
	results = nil // 清空之前的结果
	for _, file := range files {
		// 检查是否需要跳过该文件
		if shouldSkipFile(file.Path) {
			fmt.Printf("跳过不支持的文件类型: %s\n", file.Path)
			continue
		}

		info, err := p.ProcessFile(file.Path)
		if err != nil {
			fmt.Printf("处理文件 %s 失败: %v\n", file.Path, err)
			continue
		}
		results = append(results, *info)
	}
	return results
}

// SaveResults 保存结果到JSON文件和SQLite数据库
func (p *FileProcessor) SaveResults(results []SensitiveInfo, outputFile string) error {
	// 保存到JSON文件
	data, err := json.MarshalIndent(results, "", "    ")
	if err != nil {
		return fmt.Errorf("序列化结果失败: %v", err)
	}

	if err := ioutil.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("写入结果文件失败: %v", err)
	}

	// 尝试导出到SQLite数据库
	sqlitePath := strings.TrimSuffix(outputFile, filepath.Ext(outputFile)) + ".db"
	fmt.Printf("尝试导出SQLite到: %s\n", sqlitePath)

	if err := ExportToSQLite(sqlitePath, results); err != nil {
		fmt.Printf("导出到SQLite失败: %v\n", err)
		if strings.Contains(err.Error(), "cgo") {
			fmt.Println("提示: 请使用 'set CGO_ENABLED=1' 重新编译程序")
		}
	} else {
		fmt.Printf("成功导出SQLite数据库: %s\n", sqlitePath)
	}

	return nil
}

func main() {
	// 检查命令行参数
	if len(os.Args) != 3 {
		fmt.Println("用法: sens_match.exe <输入JSON文件> <输出JSON文件>")
		fmt.Println("默认命令: .\\sens_match.exe ..\\file_index.json ..\\output.json")
		os.Exit(1)
	}

	// 复制 file_index.json
	if err := copyFileIndex(); err != nil {
		fmt.Printf("警告: 复制 file_index.json 失败: %v\n", err)
		fmt.Println("程序将继续运行，但可能无法正确读取文件索引")
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	// 读取输入文件列表
	fileList, err := ReadFileList(inputFile)
	if err != nil {
		fmt.Printf("读取文件列表失败: %v\n", err)
		os.Exit(1)
	}

	// 处理文件
	processor := NewFileProcessor()
	results := processor.ProcessFileList(fileList)

	// 保存结果
	if err := processor.SaveResults(results, outputFile); err != nil {
		fmt.Printf("保存结果失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("处理完成，结果已保存到: %s\n", outputFile)
}
