# SensMatch 敏感信息检测工具

SensMatch是一个用于检测文件中敏感信息的工具，支持多种文件格式，并提供图形界面展示检测结果。

## 功能特点

- 支持多种文件格式的敏感信息检测
- 使用MD5进行文件唯一性校验
- 提供详细的检测报告，包括：
  - 文件名和路径
  - 敏感信息数量统计
  - 触发的规则编号
  - 检测时间
- 图形界面展示检测结果
- 支持按文件名、规则编号、敏感信息数量进行搜索
- 点击文件名可查看完整文件路径

## 使用方法

1. 运行检测程序：
```bash
go run src/main.go
```

2. 启动图形界面：
```bash
python app.py
```

## 数据库结构

检测结果存储在SQLite数据库中，主要字段包括：
- file_name: 文件名
- file_path: 文件完整路径
- md5: 文件MD5值
- detect_time: 检测时间
- total_sensitive_count: 敏感信息总数
- rule_numbers: 触发的规则编号

## 注意事项

- 确保数据库文件(output.db)位于正确的位置
- 检测结果会同时保存为JSON格式(output.json)
- 图形界面需要PyQt5支持

## 项目结构

```
new_match/
├── read_path/           # 文件索引生成工具目录
│   └── read_path.py     # 用于生成文件索引的Python脚本
└── sens_match/          # 敏感信息检测工具目录
    ├── bin/            # 可执行文件目录
    │   └── sens_match.exe
    ├── src/            # 源代码目录
    ├── lib/            # 依赖库目录
    ├── test_data/      # 测试数据目录
    ├── file_index.json # 文件索引文件
    └── output.json     # 检测结果输出文件
```

## 使用流程

### 1. 生成文件索引

首先，使用 `read_path` 目录下的 `read_path.py` 脚本生成文件索引：

```bash
cd read_path
python read_path.py                                                                                                                                                                                      
```

这将生成 `file_index.json` 文件，包含需要检测的文件列表。

### 2. 运行敏感信息检测

进入 `sens_match/bin` 目录，运行检测程序：

```bash
cd sens_match/bin
.\sens_match.exe ..\file_index.json ..\output.json
```

程序将读取 `file_index.json` 中的文件列表，进行敏感信息检测，并将结果输出到 `output.json`。

## 编译说明

### 1. 环境要求

- Go 1.16 或更高版本
- Python 3.6 或更高版本
- HanLP 库（用于中文地址和人名识别）

### 2. 安装依赖

1. 安装 Go 依赖：
```bash
cd src
go mod init sens_match
go get golang.org/x/text/encoding/simplifiedchinese
go get golang.org/x/text/transform
```

2. 安装 Python 依赖：
```bash
pip install pyhanlp
```

### 3. 编译程序

在 `src` 目录下执行：
```bash
go build -o ../bin/sens_match.exe
```

## 使用方法

### 1. 准备输入文件

创建一个 JSON 格式的输入文件（例如：`file_index.json`），格式如下：
```json
{
    "file_dict": {
        "文件路径1": {
            "path": "文件路径1",
            "size": 文件大小,
            "modified_time": 修改时间戳
        },
        "文件路径2": {
            "path": "文件路径2",
            "size": 文件大小,
            "modified_time": 修改时间戳
        }
    },
    "inverted_index": {
        "文件名1": ["文件路径1", "文件路径2"],
        "文件名2": ["文件路径3"]
    }
}
```

### 2. 运行程序

```bash
.\sens_match.exe <输入JSON文件> <输出JSON文件>
```

例如：
```bash
.\sens_match.exe file_index.json output.json
```

### 3. 查看结果

程序会生成一个 JSON 格式的输出文件，包含每个文件的检测结果，格式如下：
```json
[
    {
        "file_name": "文件名",
        "md5": "文件MD5值",
        "detect_time": "检测时间",
        "match_counts": {
            "id_card": 1,
            "phone": 2,
            "email": 1
        },
        "matches": {
            "id_card": ["匹配到的身份证号"],
            "phone": ["匹配到的手机号"],
            "email": ["匹配到的邮箱"]
        }
    }
]
```

## 功能特点

- 支持多种敏感信息类型的检测
- 支持批量文件处理
- 自动计算文件MD5值
- 支持中文地址和人名识别
- 可配置跳过特定文件类型
- 输出详细的检测报告

## 支持的敏感信息类型

1. 身份证号（18位）
2. 手机号（11位）
3. 邮箱地址
4. IP地址（IPv4和IPv6）
5. MAC地址
6. 银行卡号
7. 护照号
8. 中文地址
9. 中文人名
10. 统一社会信用代码
11. 组织机构代码
12. 营业执照号
13. 税务登记号
14. 企业名称
15. 信用卡号

## 注意事项

1. 程序会自动跳过以下文件类型：
   - 图片文件（.jpg, .jpeg, .png, .gif, .bmp, .ico, .svg, .webp）
   - 音视频文件（.mp3, .mp4, .avi, .mov, .wmv）
   - 压缩文件（.zip, .rar, .7z, .tar, .gz）
   - 可执行文件（.exe, .dll, .so, .dylib）
   - 安装包（.msi）

2. 中文地址和人名识别需要 Python 环境和 HanLP 库支持：
   ```bash
   pip install pyhanlp
   ```

3. 确保 `lib` 目录下有 `address_name.py` 文件，用于中文地址和人名识别。

## 错误处理

如果遇到以下错误：

1. 文件编码问题：
   - 程序会自动尝试多种编码（UTF-8、GBK、GB2312等）
   - 如果仍然失败，会跳过该文件并继续处理其他文件

2. Python脚本执行失败：
   - 检查 Python 环境是否正确安装
   - 检查 HanLP 库是否正确安装
   - 检查 `address_name.py` 是否在正确位置

## 性能考虑

1. 大文件处理：
   - 程序会一次性读取整个文件内容
   - 对于特别大的文件，建议先分割处理

2. 内存使用：
   - 程序会缓存文件内容用于检测
   - 处理大量文件时注意内存使用情况

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个工具。

## 许可证

MIT License 