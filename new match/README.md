# New Match 文件处理工具集

New Match是一个集成了多个文件处理功能的工具集，包括文件索引、敏感信息检测等功能。

## 项目结构

```
new match/
├── read_path/           # 文件索引工具
│   ├── read_path.py     # 文件索引主程序
│   └── md5_calculator.c # MD5计算工具
├── sens_match/          # 敏感信息检测工具
│   ├── src/            # Go源代码
│   ├── app.py          # 图形界面
│   └── output.db       # 检测结果数据库
└── test/               # 测试数据目录
```

## 主要功能

### 1. 文件索引 (read_path)
- 快速建立文件索引
- 支持文件变动监控
- 提供文件搜索功能
- 计算文件MD5值

### 2. 敏感信息检测 (sens_match)
- 检测文件中的敏感信息
- 支持多种文件格式
- 提供图形界面展示结果
- 支持按多种条件搜索

## 使用方法

### 文件索引工具
```bash
cd read_path
python read_path.py
```

### 敏感信息检测
```bash
cd sens_match
go run src/main.go  # 运行检测
python app.py       # 启动图形界面
```

## 依赖要求

- Python 3.x
- Go 1.x
- PyQt5
- SQLite3

## 注意事项

- 确保所有依赖已正确安装
- 运行前检查配置文件
- 建议在测试目录中先进行测试 