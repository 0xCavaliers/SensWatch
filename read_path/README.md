# 文件索引与搜索系统

这是一个用于文件索引和搜索的系统，包含三个主要组件：一个Python实现的图形界面应用程序、一个Go语言实现的文件读取工具和一个C语言实现的MD5计算工具。

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

### C程序 (md5_calculator.c)
- 使用OpenSSL计算文件MD5值
- 多线程处理大文件
- 实时显示处理进度
- 支持UTF-8编码
- 结果保存到文本文件

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

### C程序
- Visual Studio 2022（或其他支持C的编译器）
- OpenSSL库
- json-c库
- vcpkg包管理器

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

### C程序
1. 安装依赖：
   ```bash
   # 在vcpkg目录下运行
   .\vcpkg install openssl:x64-windows
   .\vcpkg install json-c:x64-windows
   ```

2. 编译程序：
   ```bash
   # 在x64 Native Tools Command Prompt for VS中运行
   cl /EHsc /MD /I"D:\DevTools\vcpkg\installed\x64-windows\include" md5_calculator.c /link /LIBPATH:"D:\DevTools\vcpkg\installed\x64-windows\lib" libssl.lib libcrypto.lib json-c.lib
   ```

3. 复制必要的DLL文件：
   ```bash
   copy "D:\DevTools\vcpkg\installed\x64-windows\bin\libcrypto-3-x64.dll" .
   copy "D:\DevTools\vcpkg\installed\x64-windows\bin\libssl-3-x64.dll" .
   copy "D:\DevTools\vcpkg\installed\x64-windows\bin\json-c.dll" .
   ```

4. 运行程序：
   ```bash
   md5_calculator.exe
   ```
   程序会读取 `file_index.json` 文件并生成 `output_MD5.txt` 文件。

## 文件说明

- `read_path.py`: 主程序文件，包含图形界面和文件索引功能
- `read_files.go`: Go语言实现的文件读取工具
- `md5_calculator.c`: C语言实现的MD5计算工具
- `file_index.json`: 索引数据文件
- `directory_log.json`: 目录监控日志文件
- `simple_everything.log`: 程序运行日志文件
- `output.txt`: Go程序生成的输出文件
- `output_MD5.txt`: C程序生成的MD5值文件

## 注意事项

1. 确保有足够的磁盘空间用于存储索引文件
2. 程序会自动保存索引数据，无需手动保存
3. 首次运行可能需要较长时间来建立索引
4. 建议定期清理不需要的索引数据以优化性能
5. 运行C程序前确保所有必要的DLL文件都在程序目录下

# 文件路径读取和 MD5 计算工具

这个项目包含三个程序，用于读取文件路径并计算文件的 MD5 值。

## 程序说明

### 1. read_files.go
Go 语言版本的文件路径读取程序。

使用方法：
```bash
go run read_files.go
```

### 2. read_path.py
Python 版本的文件路径读取程序。

使用方法：
```bash
python read_path.py
```

### 3. md5_calculator.c
C 语言版本的文件 MD5 计算程序。该程序使用多线程技术，可以快速计算大量文件的 MD5 值。

#### 依赖要求
- OpenSSL 3.0 或更高版本
- GCC 编译器

#### 安装 OpenSSL
推荐使用 vcpkg 安装 OpenSSL：
```bash
# 安装 vcpkg
git clone https://github.com/Microsoft/vcpkg.git
cd vcpkg
.\bootstrap-vcpkg.bat

# 安装 OpenSSL
.\vcpkg install openssl:x64-mingw-dynamic
.\vcpkg integrate install
```

#### 编译方法
```bash
# 使用 vcpkg 安装的 OpenSSL
gcc md5_calculator.c -o md5_calculator.exe -I"vcpkg安装路径/installed/x64-mingw-dynamic/include" -L"vcpkg安装路径/installed/x64-mingw-dynamic/lib" -lssl -lcrypto
```

#### 使用方法
1. 确保 `output.txt` 文件存在，该文件应包含要计算 MD5 的文件路径列表
2. 运行程序：
```bash
md5_calculator.exe
```
3. 程序会在当前目录生成 `md5_output.txt` 文件，格式为：
```
文件路径  MD5值
```

#### 输出说明
- 成功计算 MD5 时，输出格式为：`文件路径  MD5值`
- 计算失败时，输出格式为：`文件路径  ERROR`
- 程序运行过程中会显示处理进度

## 注意事项
1. 所有程序都会在当前目录生成 `output.txt` 文件
2. 文件路径使用 UTF-8 编码
3. MD5 计算程序支持多线程处理，默认使用 8 个线程
4. 如果文件数量超过 100000 个，需要修改 `MAX_FILES` 常量
5. 程序使用 OpenSSL 3.0 的 EVP 接口，确保使用兼容的 OpenSSL 版本 
