#导入包
import os
import json
import time
from collections import defaultdict
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor, as_completed
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler
import threading
import tkinter as tk
from tkinter import ttk, filedialog, messagebox, scrolledtext
import logging

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    handlers=[
        logging.FileHandler("simple_everything.log"),  # 输出到文件
        logging.StreamHandler()  # 输出到控制台
    ]
)
logger = logging.getLogger(__name__)

# 主类
class SimpleEverything:
    def __init__(self, index_file: str = "file_index.json", save_interval: int = 300, log_file: str = "directory_log.json"):
        self.index_file = index_file
        self.save_interval = save_interval  # 定期保存的时间间隔（秒）
        self.log_file = log_file  # 记录目录和时间戳的文件
        self.file_dict = {}  # 存储文件路径和元数据
        self.inverted_index = defaultdict(list)  # 倒排索引
        self.lock = threading.Lock()  # 线程锁，用于保护共享资源
        self.observer = Observer()  # 初始化监控器
        self.event_handlers = []  # 存储所有的 FileChangeHandler 实例
        self.timer = None  # 定时器，用于定期保存索引
        self.directory_log = {}  # 存储目录及其最新时间戳
        self.load_index()  # 加载已有索引
        self.load_directory_log()  # 加载目录记录
        self.start_auto_save()  # 启动定期保存索引

    def load_index(self):
        """从磁盘加载索引"""
        if os.path.exists(self.index_file):
            try:
                with open(self.index_file, "r", encoding="utf-8") as f:
                    data = json.load(f)
                    self.file_dict = data.get("file_dict", {})
                    self.inverted_index = defaultdict(list, data.get("inverted_index", {}))
                logger.info("索引加载完成。")
            except Exception as e:
                logger.error(f"加载索引时出错: {e}")
        else:
            logger.warning("索引文件不存在，跳过加载。")

    def load_directory_log(self):
        """从磁盘加载目录记录"""
        if os.path.exists(self.log_file):
            try:
                with open(self.log_file, "r", encoding="utf-8") as f:
                    self.directory_log = json.load(f)
                logger.info("目录记录加载完成。")

                # 为每个目录建立文件变动监控
                for directory in self.directory_log.keys():
                    logger.info(f"为目录 {directory} 建立文件变动监控...")
                    event_handler = FileChangeHandler(self, directory)
                    self.observer.schedule(event_handler, directory, recursive=True)
                    self.event_handlers.append(event_handler)

                    if not self.observer.is_alive():
                        self.observer.start()

                # 打印已建立的目录索引及其最新时间戳
                for directory, timestamp in self.directory_log.items():
                    logger.info(f"已建立目录索引: {directory}, 最新时间戳: {time.ctime(timestamp)}")
            except Exception as e:
                logger.error(f"加载目录记录时出错: {e}")
        else:
            logger.warning("目录记录文件不存在，跳过加载。")


    def save_index(self):
        """保存索引到磁盘"""
        with self.lock:
            try:
                with open(self.index_file, "w", encoding="utf-8") as f:
                    json.dump({
                        "file_dict": self.file_dict,
                        "inverted_index": {k: v for k, v in self.inverted_index.items()}
                    }, f, ensure_ascii=False)
                logger.info("索引保存完成。")
            except Exception as e:
                logger.error(f"保存索引时出错: {e}")

    def save_directory_log(self):
        """保存目录记录到磁盘"""
        with self.lock:
            try:
                with open(self.log_file, "w", encoding="utf-8") as f:
                    json.dump(self.directory_log, f, ensure_ascii=False)
                logger.info("目录记录保存完成。")
            except Exception as e:
                logger.error(f"保存目录记录时出错: {e}")

    def start_auto_save(self):
        """启动定期保存索引"""
        def auto_save_task():
            self.save_index()
            self.save_directory_log()
            self.timer = threading.Timer(self.save_interval, auto_save_task)
            self.timer.start()

        self.timer = threading.Timer(self.save_interval, auto_save_task)
        self.timer.start()
        #print(f"已启动定期保存索引，每隔 {self.save_interval} 秒保存一次。")
        logger.info(f"已启动定期保存索引，每隔 {self.save_interval} 秒保存一次。")

    def stop_auto_save(self):
        """停止定期保存索引"""
        if self.timer:
            self.timer.cancel()
        print("已停止定期保存索引。")
        logger.info("已停止定期保存索引。")

    def create_index(self, target_dir: str):
        """构建或更新索引"""
        if not os.path.isdir(target_dir):
            logger.error(f"输入的目录无效: {target_dir}")
            return "输入的目录无效，请重试。"

        start_time = time.time()
        self._index_directory(target_dir)
        logger.info(f"{target_dir}索引构建完成，耗时 {time.time() - start_time:.2f} 秒，共索引 {len(self.file_dict)} 个文件。")
        self.save_index()
        self.directory_log[target_dir] = time.time()
        self.save_directory_log()

        # 为新增目录创建监控
        #print("****************")
        event_handler = FileChangeHandler(self, target_dir)
        self.observer.schedule(event_handler, target_dir, recursive=True)
        self.event_handlers.append(event_handler)
        #print("++++++++++++++++++")
        if not self.observer.is_alive():
            self.observer.start()
        logger.info(f"已启动对目录 {target_dir} 的监控。")


    def _index_directory(self, root_dir: str):
        """多线程遍历并构建索引"""
        with ThreadPoolExecutor() as executor:
            futures = []
            for root, dirs, files in os.walk(root_dir):
                for file_name in files:
                    futures.append(executor.submit(self._index_file, root, file_name))

            # 等待所有任务完成
            for future in as_completed(futures):
                future.result()

            # 关闭线程池
            executor.shutdown(wait=True)
    
    def _index_file(self, root: str, file_name: str):
        try:
            """索引单个文件"""
            file_path = Path(root) / file_name
            if not file_path.is_file():
                # logger.warning(f"跳过非文件项: {file_path}")
                return
            
            file_info = {
                "path": str(file_path),
                "size": file_path.stat().st_size,
                "modified_time": file_path.stat().st_mtime,
            }
            with self.lock:
                if str(file_path) not in self.file_dict:  # 仅添加新文件或修改文件
                    self.file_dict[str(file_path)] = file_info
                    for word in file_name.lower().split():
                        self.inverted_index[word].append(str(file_path))
        except Exception as e:
            logger.error(f"索引文件 {file_name} 时出错: {e}")
            



    def fuzzy_search(self, keyword: str):
        """文件名模糊匹配查询"""
        keyword = keyword.lower()
        matched_files = []
        for file_path, file_info in self.file_dict.items():
            if keyword in Path(file_path).name.lower():  # 检查文件名是否包含关键词
                matched_files.append(file_path)
        return matched_files

    def search(self, keyword: str):
        """根据关键词搜索文件"""
        logger.info(f"正在搜索关键词: {keyword}")
        matched_files = self.fuzzy_search(keyword)
        valid_files = []
        for file_path in matched_files:
            if not os.path.exists(file_path):
                with self.lock:
                    if file_path in self.file_dict:
                        del self.file_dict[file_path]
                logger.debug(f"文件已删除，从索引中移除: {file_path}")
            else:
                current_mtime = os.path.getmtime(file_path)
                if current_mtime != self.file_dict[file_path]["modified_time"]:
                    self._index_file(str(Path(file_path).parent), Path(file_path).name)
                valid_files.append(file_path)
        logger.info(f"找到 {len(valid_files)} 个匹配文件。")
        return valid_files

    def get_file_info(self, file_path: str):
        """获取文件的详细信息"""
        return self.file_dict.get(file_path, None)

    def stop(self):
        """停止监控器和定时器"""
        self.stop_auto_save()
        if self.observer.is_alive():
            self.observer.stop()
            self.observer.join()
        print("所有监控和定时器已停止。")
        
    def clear_all_indices(self):
        """清理所有索引数据"""
        with self.lock:
            # 停止定时器线程
            self.stop_auto_save()

            # 停止并清除现有监控线程
            if hasattr(self, "observer") and self.observer.is_alive():
                self.observer.stop()
                self.observer.join()

            # 创建一个新的 Observer 实例
            self.observer = Observer()
            self.event_handlers = []  # 清空事件处理器列表

            # 清理内存中的索引数据
            self.file_dict.clear()
            self.inverted_index.clear()
            self.directory_log.clear()

            # 删除索引文件和目录记录文件
            try:
                if os.path.exists(self.index_file):
                    os.remove(self.index_file)
                    logger.info(f"索引文件 {self.index_file} 已删除。")
                if os.path.exists(self.log_file):
                    os.remove(self.log_file)
                    logger.info(f"目录记录文件 {self.log_file} 已删除。")
            except Exception as e:
                logger.error(f"删除文件时出错: {e}")

            logger.info("所有索引数据和线程已清理。")

# 监控句柄类
class FileChangeHandler(FileSystemEventHandler):
    def __init__(self, searcher: SimpleEverything, target_dir: str):
        self.searcher = searcher
        self.target_dir = target_dir

    def on_created(self, event):
        """处理文件创建事件"""
        if not event.is_directory:
            file_path = os.path.normpath(event.src_path)
            self.searcher._index_file(str(Path(file_path).parent), Path(file_path).name)
            logger.info(f"文件创建事件: {file_path}")
            self.update_directory_timestamp()

    def on_deleted(self, event):
        """处理文件删除事件"""
        if not event.is_directory:
            file_path = event.src_path
            with self.searcher.lock:
                if file_path in self.searcher.file_dict:
                    del self.searcher.file_dict[file_path]
            logger.info(f"文件删除事件: {file_path}")
            self.update_directory_timestamp()
    
    
    def on_modified(self, event):
        """处理文件修改事件"""
        if not event.is_directory:
            file_path = os.path.normpath(event.src_path)
            self.searcher._index_file(str(Path(file_path).parent), Path(file_path).name)
            logger.info(f"文件修改事件: {file_path}")
            #self.update_directory_timestamp()
    
    def on_moved(self, event):
        """处理文件移动/重命名事件"""
        if not event.is_directory:
            src_path = os.path.normpath(event.src_path)
            dest_path = os.path.normpath(event.dest_path)
            with self.searcher.lock:
                if src_path in self.searcher.file_dict:
                    del self.searcher.file_dict[src_path]
            self.searcher._index_file(str(Path(dest_path).parent), Path(dest_path).name)
            logger.info(f"文件移动/重命名事件: {src_path} -> {dest_path}")
            self.update_directory_timestamp()

    def update_directory_timestamp(self):
        """更新目录的时间戳"""
        self.searcher.directory_log[self.target_dir] = time.time()
        #self.searcher.save_directory_log()

# 界面类
class SimpleEverythingApp:
    def __init__(self, root):
        self.root = root
        self.root.title("Simple Everything")

        # 初始化配置
        self.root.geometry("800x600")  # 设置初始窗口大小
        self.root.minsize(600, 400)  # 设置最小窗口大小

        self.searcher = SimpleEverything()
        self.setup_ui()

    def setup_ui(self):
        # 配置行和列的权重，使界面可以缩放
        self.root.grid_rowconfigure(0, weight=1)
        self.root.grid_columnconfigure(0, weight=1)
        self.root.grid_columnconfigure(1, weight=1)

        # 左侧面板
        left_frame = ttk.Frame(self.root, padding="10")
        left_frame.grid(row=0, column=0, sticky=(tk.W, tk.E, tk.N, tk.S))
        left_frame.grid_rowconfigure(5, weight=1)  # 搜索结果部分可缩放
        left_frame.grid_columnconfigure(0, weight=1)

        # 选择目录部分
        ttk.Label(left_frame, text="选择目录:").grid(row=0, column=0, sticky=tk.W)
        self.dir_var = tk.StringVar()
        self.dir_entry = ttk.Entry(left_frame, textvariable=self.dir_var, width=50)
        self.dir_entry.grid(row=1, column=0, sticky=(tk.W, tk.E))
        ttk.Button(left_frame, text="浏览...", command=self.browse_directory).grid(row=1, column=1, sticky=tk.W)
        ttk.Button(left_frame, text="确认", command=self.confirm_directory).grid(row=2, column=0, columnspan=2, pady=5)

        # 搜索文件部分
        ttk.Label(left_frame, text="搜索文件:").grid(row=3, column=0, sticky=tk.W)
        self.search_var = tk.StringVar()
        self.search_entry = ttk.Entry(left_frame, textvariable=self.search_var, width=50)
        self.search_entry.grid(row=4, column=0, sticky=(tk.W, tk.E))
        ttk.Button(left_frame, text="检索", command=self.search_files).grid(row=4, column=1, sticky=tk.W)

        # 清理索引按钮
        ttk.Button(left_frame, text="清理索引", command=self.clear_index).grid(row=6, column=0, columnspan=2, pady=5)

        # 搜索结果部分
        self.results_text = scrolledtext.ScrolledText(left_frame, width=60, height=10, wrap=tk.WORD)
        self.results_text.grid(row=5, column=0, columnspan=2, pady=5, sticky=(tk.W, tk.E, tk.N, tk.S))
        self.results_text.bind("<Double-Button-1>", self.open_file_location)

        # 右侧面板：显示已索引目录及其时间戳
        right_frame = ttk.Frame(self.root, padding="10")
        right_frame.grid(row=0, column=1, sticky=(tk.W, tk.E, tk.N, tk.S))
        right_frame.grid_rowconfigure(1, weight=1)

        ttk.Label(right_frame, text="已索引的目录:").grid(row=0, column=0, sticky=tk.W)
        self.directories_text = scrolledtext.ScrolledText(right_frame, width=40, height=20, wrap=tk.WORD)
        self.directories_text.grid(row=1, column=0, sticky=(tk.W, tk.E, tk.N, tk.S))

        self.update_directories_display()

    def browse_directory(self):
        selected_dir = filedialog.askdirectory()
        if selected_dir:
            self.dir_var.set(selected_dir)

    def confirm_directory(self):
        target_dir = self.dir_var.get()
        if not target_dir:
            messagebox.showwarning("警告", "请先选择一个目录！")
            return
        if not os.path.isdir(target_dir):
            messagebox.showerror("错误", "选择的目录无效，请重试！")
            return
        self.searcher.create_index(target_dir)
        self.update_directories_display()

    def clear_index(self):
        """清理所有已建立的索引"""
        confirm = messagebox.askyesno("确认", "确定要清理所有索引吗？此操作不可恢复！")
        if not confirm:
            return

        # 调用清理方法
        self.searcher.clear_all_indices()

        # 清空搜索结果和已索引目录的显示
        self.results_text.delete(1.0, tk.END)
        self.update_directories_display()

        messagebox.showinfo("成功", "所有索引已清理！")


    def update_directories_display(self):
        """更新右侧的已索引目录显示"""
        self.directories_text.delete(1.0, tk.END)
        for directory, timestamp in self.searcher.directory_log.items():
            self.directories_text.insert(tk.END, f"目录: {directory}\n")
            self.directories_text.insert(tk.END, f"最新时间戳: {time.ctime(timestamp)}\n")
            self.directories_text.insert(tk.END, "=" * 50 + "\n")

    def search_files(self):
        keyword = self.search_var.get()
        if not keyword:
            messagebox.showwarning("警告", "请输入搜索关键词！")
            return
        results = self.searcher.search(keyword)
        self.results_text.delete(1.0, tk.END)
        if results:
            self.results_text.insert(tk.END, f"找到 {len(results)} 个匹配文件:\n")
            for result in results:
                file_info = self.searcher.get_file_info(result)
                self.results_text.insert(tk.END, f"文件路径: {file_info['path']}\n")
                self.results_text.insert(tk.END, f"文件大小: {file_info['size']} 字节\n")
                self.results_text.insert(tk.END, f"修改时间: {time.ctime(file_info['modified_time'])}\n")
                self.results_text.insert(tk.END, "=" * 50 + "\n")
        else:
            self.results_text.insert(tk.END, "未找到匹配文件。\n")

    def open_file_location(self, event):
        index = self.results_text.index("@%s,%s" % (event.x, event.y))
        line_start = f"{index.split('.')[0]}.0"
        line_end = f"{index.split('.')[0]}.end"
        line_text = self.results_text.get(line_start, line_end)
        if line_text.startswith("文件路径: "):
            file_path = line_text[len("文件路径: "):].strip()
            if os.path.exists(file_path):
                if os.name == "nt":  # Windows 系统
                    os.startfile(os.path.dirname(file_path))
                    os.system(f'explorer /select,"{file_path}"')
                else:
                    messagebox.showinfo("信息", "此功能仅在 Windows 系统中可用。")
            else:
                messagebox.showerror("错误", "文件不存在！")
        else:
            messagebox.showinfo("信息", "请双击文件路径以跳转。")

    def on_closing(self):
        self.searcher.stop()
        self.searcher.save_index()
        self.searcher.save_directory_log()
        # 清理内存中的索引数据
        self.searcher.file_dict.clear()
        self.searcher.inverted_index.clear()
        self.searcher.directory_log.clear()

        self.root.destroy()

# 主函数
# 主程序入口
if __name__ == "__main__":
    root = tk.Tk()
    app = SimpleEverythingApp(root)
    root.protocol("WM_DELETE_WINDOW", app.on_closing)
    root.mainloop()