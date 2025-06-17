package main

// FileInfo 表示文件信息
type FileInfo struct {
	Path         string  `json:"path"`
	Size         int64   `json:"size"`
	ModifiedTime float64 `json:"modified_time"`
}

// FileDict 表示文件索引的JSON结构
type FileDict struct {
	FileDict      map[string]FileInfo `json:"file_dict"`
	InvertedIndex map[string][]string `json:"inverted_index"`
}

// SensitiveInfo 表示敏感信息检测结果
type SensitiveInfo struct {
	FileName            string              `json:"file_name"`
	FilePath            string              `json:"file_path"`
	MD5                 string              `json:"md5"`
	DetectTime          string              `json:"detect_time"`
	MatchCounts         map[string]int      `json:"match_counts"`
	Matches             map[string][]string `json:"matches"`
	TotalSensitiveCount int                 `json:"total_sensitive_count"`
	RuleNumbers         string              `json:"rule_numbers"`
}
