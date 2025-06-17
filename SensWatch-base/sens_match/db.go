package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// ExportToSQLite 导出数据到SQLite数据库文件
func ExportToSQLite(sqlitePath string, results []SensitiveInfo) error {
	// 创建SQLite数据库连接
	sqliteDB, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		return fmt.Errorf("创建SQLite数据库失败: %v", err)
	}
	defer sqliteDB.Close()

	// 先删除旧表
	dropTableSQL := `DROP TABLE IF EXISTS detection_results;`
	_, err = sqliteDB.Exec(dropTableSQL)
	if err != nil {
		return fmt.Errorf("删除旧SQLite表失败: %v", err)
	}

	// 创建表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS detection_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_path TEXT NOT NULL UNIQUE,
		file_name TEXT NOT NULL,
		md5 TEXT NOT NULL,
		detect_time DATETIME NOT NULL,
		match_counts TEXT,
		matches TEXT,
		total_sensitive_count INTEGER DEFAULT 0,
		rule_numbers TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = sqliteDB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("创建SQLite表失败: %v", err)
	}

	// 准备SQLite插入语句
	insertSQL := `
	INSERT OR REPLACE INTO detection_results 
	(file_path, file_name, md5, detect_time, match_counts, matches, total_sensitive_count, rule_numbers)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := sqliteDB.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("准备SQLite插入语句失败: %v", err)
	}
	defer stmt.Close()

	// 从全局变量中获取所有结果
	for _, result := range results {
		// 将match_counts和matches转换为JSON字符串
		matchCountsJSON, err := json.Marshal(result.MatchCounts)
		if err != nil {
			return fmt.Errorf("转换match_counts为JSON失败: %v", err)
		}

		matchesJSON, err := json.Marshal(result.Matches)
		if err != nil {
			return fmt.Errorf("转换matches为JSON失败: %v", err)
		}

		// 执行插入
		_, err = stmt.Exec(
			result.FilePath, // 使用完整路径
			result.FileName, // 使用文件名
			result.MD5,
			result.DetectTime,
			string(matchCountsJSON),
			string(matchesJSON),
			result.TotalSensitiveCount,
			result.RuleNumbers)
		if err != nil {
			return fmt.Errorf("插入数据到SQLite失败: %v", err)
		}
	}

	return nil
}
