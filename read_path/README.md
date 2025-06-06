# 文件索引与搜索系统

这是一个用于文件索引和搜索的系统，包含两个主要组件：一个Python实现的图形界面应用程序和一个Go语言实现的文件读取工具。

## 功能特点

### Python应用程序 (read_path.py)
- 图形用户界面，支持直观的文件操作
- 实时文件系统监控
- 多线程文件索引
- 模糊搜索功能
- 自动保存索引数据
- 支持多个目录的监控和索引
- 文件变更实时更新
- 详细的日志记录

### Go程序 (read_files.go)
- 高效的JSON索引文件读取
- 支持大文件处理
- 文件路径导出功能
- 基于缓冲的文件读取

## 系统要求

### Python程序
- Python 3.x
- 依赖包：
  - tkinter
  - watchdog
  - pathlib
  - logging

### Go程序
- Go 1.x 或更高版本

## 使用方法

### Python程序
1. 运行程序：
   ```bash
   python read_path.py
   ```
2. 通过图形界面：
   - 点击"浏览"选择要索引的目录
   - 使用搜索框进行文件搜索
   - 点击搜索结果可以打开文件位置
   - 使用"清除索引"按钮重置所有索引

### Go程序
1. 编译程序：
   ```bash
   go build read_files.go
   ```
2. 运行程序：
   ```bash
   ./read_files
   ```
   程序会读取 `file_index.json` 文件并生成 `output.txt` 文件。

## 文件说明

- `read_path.py`: 主程序文件，包含图形界面和文件索引功能
- `read_files.go`: Go语言实现的文件读取工具
- `file_index.json`: 索引数据文件
- `directory_log.json`: 目录监控日志文件
- `simple_everything.log`: 程序运行日志文件
- `output.txt`: Go程序生成的输出文件

## 注意事项

1. 确保有足够的磁盘空间用于存储索引文件
2. 程序会自动保存索引数据，无需手动保存
3. 首次运行可能需要较长时间来建立索引
4. 建议定期清理不需要的索引数据以优化性能 