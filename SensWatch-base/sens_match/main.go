package main

import (
	"bufio"
	"database/sql"
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
		// 只添加包含敏感信息的文件
		if info.TotalSensitiveCount > 0 {
			results = append(results, *info)
		} else {
			fmt.Printf("跳过不包含敏感信息的文件: %s\n", file.Path)
		}
	}
	return results
}

// SaveResults 保存结果到JSON文件和SQLite数据库
func (p *FileProcessor) SaveResults(results []SensitiveInfo, outputFile string) error {
	// 如果没有包含敏感信息的文件，直接返回
	if len(results) == 0 {
		fmt.Println("没有发现包含敏感信息的文件，不生成输出文件")
		return nil
	}

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

// LogTracker 用于跟踪日志文件的变化
type LogTracker struct {
	lastPosition int64
	logFile      string
}

// FileChange 表示文件变化信息
type FileChange struct {
	Type     string // "modify", "move", "delete"
	OldPath  string // 移动/重命名前的路径
	NewPath  string // 移动/重命名后的路径
	FilePath string // 修改/删除的文件路径
}

// NewLogTracker 创建新的日志跟踪器
func NewLogTracker(logFile string) *LogTracker {
	return &LogTracker{
		lastPosition: 0,
		logFile:      logFile,
	}
}

// getModifiedFiles 获取自上次检查以来修改过的文件
func (lt *LogTracker) getModifiedFiles() ([]FileChange, error) {
	file, err := os.Open(lt.logFile)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %v", err)
	}
	defer file.Close()

	// 获取文件当前大小
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 如果文件大小小于上次位置，说明文件被清空或重置
	if fileInfo.Size() < lt.lastPosition {
		lt.lastPosition = 0
	}

	// 移动到上次处理的位置
	if _, err := file.Seek(lt.lastPosition, 0); err != nil {
		return nil, fmt.Errorf("移动文件指针失败: %v", err)
	}

	// 读取新的日志内容
	scanner := bufio.NewScanner(file)
	var changes []FileChange
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "文件修改事件:") {
			parts := strings.Split(line, "文件修改事件: ")
			if len(parts) > 1 {
				changes = append(changes, FileChange{
					Type:     "modify",
					FilePath: parts[1],
				})
			}
		} else if strings.Contains(line, "文件移动/重命名事件:") {
			parts := strings.Split(line, "文件移动/重命名事件: ")
			if len(parts) > 1 {
				pathParts := strings.Split(parts[1], " -> ")
				if len(pathParts) == 2 {
					changes = append(changes, FileChange{
						Type:    "move",
						OldPath: pathParts[0],
						NewPath: pathParts[1],
					})
				}
			}
		} else if strings.Contains(line, "文件删除事件:") {
			parts := strings.Split(line, "文件删除事件: ")
			if len(parts) > 1 {
				changes = append(changes, FileChange{
					Type:     "delete",
					FilePath: parts[1],
				})
			}
		}
	}

	// 更新最后处理的位置
	lt.lastPosition, err = file.Seek(0, 1)
	if err != nil {
		return nil, fmt.Errorf("获取当前位置失败: %v", err)
	}

	return changes, nil
}

// updateDatabase 更新数据库
func (p *FileProcessor) updateDatabase(filesToScan map[string]bool, outputFile string) error {
	// 打开数据库连接
	sqlitePath := strings.TrimSuffix(outputFile, filepath.Ext(outputFile)) + ".db"
	db, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}
	defer db.Close()

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 准备SQLite插入语句
	insertSQL := `
	INSERT OR REPLACE INTO detection_results 
	(file_path, file_name, md5, detect_time, match_counts, matches, total_sensitive_count, rule_numbers)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("准备SQLite插入语句失败: %v", err)
	}
	defer stmt.Close()

	// 处理需要扫描的文件
	for filePath, shouldScan := range filesToScan {
		// 如果文件被标记为false，说明是移动/重命名/删除的旧路径
		if !shouldScan {
			// 检查数据库中是否存在该文件
			var count int
			err = tx.QueryRow("SELECT COUNT(*) FROM detection_results WHERE file_path = ?", filePath).Scan(&count)
			if err != nil {
				return fmt.Errorf("查询数据库失败: %v", err)
			}

			// 如果数据库中存在该文件，需要删除记录
			if count > 0 {
				_, err = tx.Exec("DELETE FROM detection_results WHERE file_path = ?", filePath)
				if err != nil {
					return fmt.Errorf("删除文件记录失败: %v", err)
				}
				fmt.Printf("文件 %s 已从数据库中删除\n", filePath)
			}
			continue
		}

		info, err := p.ProcessFile(filePath)
		if err != nil {
			fmt.Printf("处理文件 %s 失败: %v\n", filePath, err)
			continue
		}

		if info.TotalSensitiveCount > 0 {
			// 如果文件包含敏感信息，更新数据库
			matchCountsJSON, err := json.Marshal(info.MatchCounts)
			if err != nil {
				return fmt.Errorf("转换match_counts为JSON失败: %v", err)
			}

			matchesJSON, err := json.Marshal(info.Matches)
			if err != nil {
				return fmt.Errorf("转换matches为JSON失败: %v", err)
			}

			// 执行插入或更新
			_, err = stmt.Exec(
				info.FilePath,
				info.FileName,
				info.MD5,
				info.DetectTime,
				string(matchCountsJSON),
				string(matchesJSON),
				info.TotalSensitiveCount,
				info.RuleNumbers)
			if err != nil {
				return fmt.Errorf("插入数据到SQLite失败: %v", err)
			}
		} else {
			// 如果文件不再包含敏感信息，检查数据库中是否存在该文件
			var count int
			err = tx.QueryRow("SELECT COUNT(*) FROM detection_results WHERE file_path = ?", filePath).Scan(&count)
			if err != nil {
				return fmt.Errorf("查询数据库失败: %v", err)
			}

			// 如果数据库中存在该文件，说明它从敏感文件变成了非敏感文件，需要删除
			if count > 0 {
				_, err = tx.Exec("DELETE FROM detection_results WHERE file_path = ?", filePath)
				if err != nil {
					return fmt.Errorf("删除非敏感文件记录失败: %v", err)
				}
				fmt.Printf("文件 %s 从敏感文件变为非敏感文件，已从数据库中删除\n", filePath)
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// processFileChanges 处理文件变化并返回需要扫描的文件集合
func (p *FileProcessor) processFileChanges(changes []FileChange, db *sql.DB) (map[string]bool, error) {
	// 创建需要扫描的文件集合
	filesToScan := make(map[string]bool)

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback() // 如果提交成功，这个回滚会被忽略

	// 处理每个变化
	for _, change := range changes {
		switch change.Type {
		case "modify":
			// 将修改的文件添加到扫描集合
			filesToScan[change.FilePath] = true

		case "move":
			// 检查旧路径是否在数据库中
			var count int
			err := tx.QueryRow("SELECT COUNT(*) FROM detection_results WHERE file_path = ?", change.OldPath).Scan(&count)
			if err != nil {
				return nil, fmt.Errorf("查询数据库失败: %v", err)
			}

			if count > 0 {
				// 如果旧路径在数据库中，更新为新路径
				_, err = tx.Exec("UPDATE detection_results SET file_path = ? WHERE file_path = ?",
					change.NewPath, change.OldPath)
				if err != nil {
					return nil, fmt.Errorf("更新文件路径失败: %v", err)
				}
			}

			// 将新路径添加到扫描集合，旧路径标记为false
			filesToScan[change.NewPath] = true
			filesToScan[change.OldPath] = false

		case "delete":
			// 从数据库中删除记录
			_, err = tx.Exec("DELETE FROM detection_results WHERE file_path = ?", change.FilePath)
			if err != nil {
				return nil, fmt.Errorf("删除文件记录失败: %v", err)
			}
			filesToScan[change.FilePath] = false
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}

	return filesToScan, nil
}

func realTimeMatch() {
	// 创建日志跟踪器
	exeDir, err := os.Getwd()
	baseDir := filepath.Dir(exeDir)
	log_path := filepath.Join(baseDir, "simple_everything.log")
	tracker := NewLogTracker(log_path)
	processor := NewFileProcessor()

	// 打开数据库连接
	sqlitePath := filepath.Join(baseDir, "output.db")
	db, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		fmt.Printf("打开数据库失败: %v\n", err)
		return
	}
	defer db.Close()

	for {
		// 等待3秒后再次检查
		time.Sleep(3 * time.Second)

		// 获取文件变化
		changes, err := tracker.getModifiedFiles()
		if err != nil {
			fmt.Printf("获取文件变化失败: %v\n", err)
			continue
		}

		if len(changes) > 0 {
			fmt.Printf("检测到 %d 个文件变化\n", len(changes))

			// 处理文件变化，获取需要扫描的文件集合
			filesToScan, err := processor.processFileChanges(changes, db)
			if err != nil {
				fmt.Printf("处理文件变化失败: %v\n", err)
				continue
			}

			// 更新数据库
			rootDir := filepath.Dir(filepath.Dir(exeDir))
			outputFile := filepath.Join(rootDir, "output.json")
			if err := processor.updateDatabase(filesToScan, outputFile); err != nil {
				fmt.Printf("更新数据库失败: %v\n", err)
			} else {
				fmt.Printf("成功更新数据库，处理了 %d 个文件\n", len(filesToScan))
			}
		}
	}
}

func main() {
	exeDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前工作目录失败: %v\n", err)
		os.Exit(1)
	}
	rootDir := filepath.Dir(exeDir)
	inputFile := filepath.Join(rootDir, "file_index.json")
	outputFile := filepath.Join(rootDir, "output.json")

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

	realTimeMatch()
}
