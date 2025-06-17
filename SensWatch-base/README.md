# SensWatch 跨平台文件敏感信息检测客户端

SensWatch 是一个集文件索引、敏感信息检测、实时监控与可视化展示于一体的跨平台客户端，支持 Windows、macOS、Linux。适用于批量文件合规检测、数据安全审计等场景。

---

## 主要特性

- **一键自动化**：`python main.py` 一步完成索引、检测、结果展示，自动管理所有进程。
- **多语言引擎**：Python 负责索引与调度，Go 实现高性能敏感信息检测，PyQt5 提供现代化前端。
- **实时监控**：自动追踪目录变动，检测结果实时同步。
- **多格式支持**：支持 docx、pdf、xlsx、txt、pptx 等主流办公文档。
- **可视化前端**：敏感文件、规则编号、MD5、发现时间等一览无余，支持搜索与弹窗详情。
- **跨平台兼容**：支持 Windows、macOS、Linux，自动适配本地环境。

---

## 项目结构

```
new match/
├── main.py                # 一键自动化主程序（入口）
├── app.py                 # PyQt5 图形界面
├── sens_match/            # 敏感信息检测（Go）
│   ├── main.go
│   ├── sens_match.go
│   ├── types.go
│   ├── file_reader.go
│   ├── db.go
│   ├──read_path.py
│   └── ...
├── file_index.json        # 文件索引（自动生成）
├── output.json            # 检测结果（自动生成）
├── output.db              # 检测结果数据库（自动生成）
├── simple_everything.log  # 日志文件（自动生成）
├── directory_log.json     # 目录监控记录（自动生成）
└── test 1/                # 测试数据目录
```

---

## 快速开始

### 1. 一键自动检测

在项目根目录下运行：

```bash
python main.py
```

按提示输入要检测的目录路径，系统将自动完成索引、敏感检测、结果展示等所有流程。

### 2. 单独查看检测结果

如需单独查看检测结果，可直接运行：

```bash
python app.py
```

---

## 依赖环境

- Python 3.9+
- Go 1.16+
- PyQt5
- watchdog
- HanLP（如需中文地址/人名识别）

### 安装依赖

```bash
pip install pyqt5 watchdog pyhanlp
cd sens_match
go mod tidy
```

---

## 数据与结果说明

### 文件索引（file_index.json）

```json
{
  "file_dict": {
    "绝对路径": {
      "path": "绝对路径",
      "size": 文件大小,
      "modified_time": 修改时间戳
    }
  },
  "inverted_index": {
    "文件名": ["绝对路径1", "绝对路径2"]
  }
}
```

### 检测结果（output.json）

```json
[
  {
    "file_name": "xxx.pdf",
    "file_path": "/abs/path/xxx.pdf",
    "md5": "xxxx",
    "detect_time": "2025-06-16 22:15:34",
    "match_counts": {"email": 2, "phone": 1},
    "matches": {"email": ["a@b.com"], "phone": ["138..."]},
    "total_sensitive_count": 3,
    "rule_numbers": "1、6"
  }
]
```

### 数据库结构（output.db）

表 detection_results 字段：

- file_name
- file_path
- md5
- detect_time
- match_counts
- matches
- total_sensitive_count
- rule_numbers

---

## 日志与监控

- `simple_everything.log`：记录索引、监控、异常等全流程日志。
- `directory_log.json`：记录已监控目录及最新索引时间戳，实现断点续扫。

---

## 支持的敏感信息类型（部分）

- 身份证号、手机号、邮箱、IP、MAC、银行卡、护照、中文地址、人名、企业信息等（详见 output.json 字段）。

---

## 常见问题

- **依赖未安装**：请先安装 Python/Go 依赖。
- **Go 检测进程未响应**：请检查 go 环境和 sens_match 目录下源码完整性。
- **前端无数据**：请确认 output.db/output.json 已生成且有检测结果。

---

## 贡献与反馈

如有建议、Bug 或需求，欢迎 Issue 或 PR！

---

如需进一步定制说明或英文版，请告知！ 