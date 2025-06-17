import subprocess
import os
import sys
import platform
import json
import atexit
import signal
import hashlib

script_dir = os.path.dirname(os.path.abspath(__file__))

def main_workflow(target_dir):
    # 1. 建立/更新索引
    from sens_match.read_path import SimpleEverything
    searcher = SimpleEverything()
    print("正在清理旧索引...")
    searcher.clear_all_indices()
    print("正在建立新索引...")
    searcher.create_index(target_dir)
    print("索引建立完成。")

    # 2. 调用 Go 敏感检测（假设有 main.go）
    go_detect_dir = os.path.join('sens_match')
    print("正在进行敏感信息检测...")
    go_proc = subprocess.Popen(['go', 'run', '.'], cwd=go_detect_dir)
    print("敏感信息检测已在后台运行。")

    def cleanup():
        print("正在关闭敏感信息检测进程...")
        if go_proc.poll() is None:
            try:
                go_proc.terminate()
                go_proc.wait(timeout=10)
            except Exception:
                go_proc.kill()
        print("敏感信息检测进程已关闭。")

    atexit.register(cleanup)

    # 3. 启动结果展示界面
    print("启动结果展示界面...")
    gui_proc = subprocess.Popen([sys.executable, 'app.py'])

    # 保持主进程运行，直到用户中断
    try:
        gui_proc.wait()  # 阻塞直到GUI关闭
        print("结果展示界面已关闭，主程序即将退出。")
    except KeyboardInterrupt:
        print("主程序收到中断信号，准备退出。")

if __name__ == '__main__':
    target_dir = input("请输入要索引的目录路径：").strip()
    main_workflow(target_dir) 