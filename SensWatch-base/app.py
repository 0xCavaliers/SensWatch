import sys
import sqlite3
from PyQt5.QtWidgets import (
    QApplication, QMainWindow, QWidget, QVBoxLayout, QHBoxLayout,
    QPushButton, QLabel, QLineEdit, QTableWidget, QTableWidgetItem,
    QSizePolicy, QHeaderView, QMessageBox
)
from PyQt5.QtCore import Qt, QTimer
from PyQt5.QtGui import QFont, QColor, QPalette
import os

class SensitiveFileWindow(QMainWindow):
    def __init__(self):
        super().__init__()
        self.setWindowTitle("SensWatch")
        self.setGeometry(100, 100, 900, 600)
        
        # 连接数据库
        script_dir = os.path.dirname(os.path.abspath(__file__))
        self.db_path = os.path.join(script_dir, "output.db")
        self.conn = sqlite3.connect(self.db_path)
        self.cursor = self.conn.cursor()
        
        # 存储文件路径的字典
        self.file_paths = {}
        
        # 创建新列（如果不存在）
        try:
            self.cursor.execute("""
                ALTER TABLE detection_results 
                ADD COLUMN rule_numbers TEXT;
            """)
            self.cursor.execute("""
                ALTER TABLE detection_results 
                ADD COLUMN total_sensitive_count INTEGER;
            """)
            self.conn.commit()
        except sqlite3.Error:
            # 如果列已存在，忽略错误
            pass
        
        # 创建定时器用于定期刷新数据
        self.refresh_timer = QTimer()
        self.refresh_timer.timeout.connect(self.refresh_table_data)
        self.refresh_timer.start(30000)  # 每5秒刷新一次
        
        # ---------- 主容器 ----------
        main_widget = QWidget()
        main_layout = QVBoxLayout(main_widget)
        main_layout.setContentsMargins(0, 0, 0, 0)
        main_layout.setSpacing(0)

        # ---------- 顶部导航 ----------
        nav = QWidget()
        nav.setStyleSheet("background:#6ea6e6;")
        nav_layout = QHBoxLayout(nav)
        nav_layout.setContentsMargins(0, 0, 0, 0)
        nav_layout.setSpacing(0)

        title = QLabel("SensWatch")
        # title.setFont(QFont("微软雅黑", 28, QFont.Bold))
        title.setFont(QFont("Microsoft YaHei", 28, QFont.Bold))
        title.setStyleSheet("color:white;")
        nav_layout.addStretch(1)
        nav_layout.addWidget(title, alignment=Qt.AlignCenter)
        nav_layout.addStretch(1)
        main_layout.addWidget(nav)

        # ---------- 搜索框 ----------
        search_container = QWidget()
        s_layout = QHBoxLayout(search_container)
        s_layout.setContentsMargins(200, 60, 200, 40)
        s_layout.addStretch(1)

        search_box = QWidget()
        search_box.setFixedWidth(860)
        sb_layout = QHBoxLayout(search_box)
        sb_layout.setContentsMargins(0, 0, 0, 0)
        sb_layout.setSpacing(8)

        self.search_edit = QLineEdit()
        self.search_edit.setPlaceholderText("输入文件名、规则编号或敏感信息数进行搜索")
        self.search_edit.setFixedHeight(40)
        self.search_edit.setStyleSheet("""
            QLineEdit {
                font-size:20px;
                border:1.5px solid #d0e3f5;
                border-radius:8px;
                padding-left:12px;
                background:white;
                color: black;
            }
        """)
        self.search_edit.returnPressed.connect(self.search_files)  # 添加回车搜索功能

        search_btn = QPushButton("🔍")
        search_btn.setFixedSize(40, 40)
        search_btn.setStyleSheet("""
            QPushButton {
                background:white;
                font-size:20px;
                border:1.5px solid #d0e3f5;
                border-radius:8px;
            }
            QPushButton:hover {
                background:#f0f8ff;
            }
        """)
        search_btn.clicked.connect(self.search_files)  # 添加点击搜索功能

        sb_layout.addWidget(self.search_edit)
        sb_layout.addWidget(search_btn)
        s_layout.addWidget(search_box)
        s_layout.addStretch(1)
        main_layout.addWidget(search_container)

        # ---------- 表格 ----------
        table_area = QWidget()
        t_area_layout = QHBoxLayout(table_area)
        t_area_layout.setContentsMargins(200, 0, 200, 0)
        t_area_layout.setSpacing(0)

        self.table = QTableWidget()
        self.table.setColumnCount(5)  # 设置为5列
        self.table.setHorizontalHeaderLabels([
            "文件名", "敏感信息数", "MD5码", "规则编号", "发现时间"
        ])
        
        # 从数据库加载数据
        self.load_table_data()
        
        self.table.setSizePolicy(QSizePolicy.Expanding, QSizePolicy.Expanding)
        self.table.horizontalHeader().setSectionResizeMode(QHeaderView.Stretch)
        self.table.verticalHeader().setVisible(False)
        self.table.setEditTriggers(QTableWidget.NoEditTriggers)
        self.table.setSelectionMode(QTableWidget.NoSelection)
        self.table.setShowGrid(False)

        # 行高 & 对齐
        for r in range(self.table.rowCount()):
            self.table.setRowHeight(r, 44)
            for c in range(self.table.columnCount()):
                item = self.table.item(r, c)
                if item:
                    item.setTextAlignment(Qt.AlignCenter)

        # 样式
        self.table.setStyleSheet("""
            QTableWidget {
                font-size:20px;
                border:none;
                border-radius:8px;
                background:#f0f8ff;
            }
            QTableWidget::item {
                border-bottom:1px solid #6ea6e6;
            }
            QHeaderView::section {
                background:#dbe7f8;
                font-weight:bold;
                font-size:20px;
                height:48px;
                border:none;
            }
            QTableWidget, QTableWidget::item {
                color: #333333;
            }
        """)

        t_area_layout.addWidget(self.table)
        main_layout.addWidget(table_area)

        self.setCentralWidget(main_widget)
        
    def search_files(self):
        search_text = self.search_edit.text().strip()
        if not search_text:
            self.load_table_data()  # 如果搜索框为空，显示所有数据
            return

        try:
            # 构建搜索条件
            conditions = []
            params = []
            
            # 添加文件名搜索条件
            conditions.append("file_name LIKE ?")
            params.append(f"%{search_text}%")
            
            # 添加规则编号搜索条件
            conditions.append("rule_numbers LIKE ?")
            params.append(f"%{search_text}%")
            
            # 尝试将搜索文本转换为数字（用于敏感信息数搜索）
            try:
                count = int(search_text)
                conditions.append("total_sensitive_count = ?")
                params.append(count)
            except ValueError:
                pass

            # 构建SQL查询
            query = f"""
                SELECT file_name, total_sensitive_count, md5, rule_numbers, detect_time, file_path 
                FROM detection_results 
                WHERE {' OR '.join(conditions)}
            """
            
            self.cursor.execute(query, params)
            rows = self.cursor.fetchall()
            
            # 更新表格
            self.table.setRowCount(len(rows))
            for i, row in enumerate(rows):
                # 存储文件路径
                self.file_paths[row[0]] = row[5]
                
                for j, value in enumerate(row[:5]):  # 只显示前5列
                    if j == 3:  # 规则编号列
                        item = QTableWidgetItem("无" if row[1] == 0 else str(value) if value is not None else "")
                    else:
                        item = QTableWidgetItem(str(value) if value is not None else "")
                    self.table.setItem(i, j, item)
                    
            if len(rows) == 0:
                QMessageBox.information(self, "搜索结果", "未找到匹配的记录")
                
        except sqlite3.Error as e:
            QMessageBox.warning(self, "搜索错误", f"搜索时发生错误：{str(e)}")
            
    def load_table_data(self):
        try:
            # 查询数据库，只选择需要的列
            self.cursor.execute("""
                SELECT file_name, total_sensitive_count, md5, rule_numbers, detect_time, file_path 
                FROM detection_results
            """)
            rows = self.cursor.fetchall()
            
            # 设置表格行数
            self.table.setRowCount(len(rows))
            
            # 填充数据
            for i, row in enumerate(rows):
                # 存储文件路径
                self.file_paths[row[0]] = row[5]
                
                for j, value in enumerate(row[:5]):  # 只显示前5列
                    if j == 3:  # 规则编号列
                        item = QTableWidgetItem("无" if row[1] == 0 else str(value) if value is not None else "")
                    else:
                        item = QTableWidgetItem(str(value) if value is not None else "")
                    self.table.setItem(i, j, item)
                    
            # 先断开已有的绑定，再绑定一次
            try:
                self.table.cellClicked.disconnect()
            except Exception:
                pass
            self.table.cellClicked.connect(self.show_file_path)
            
        except sqlite3.Error as e:
            QMessageBox.warning(self, "数据加载错误", f"加载数据时发生错误：{str(e)}")
            
    def show_file_path(self, row, column):
        # 文件名、MD5码、规则编号、发现时间都可弹窗显示完整信息
        col_titles = ["文件名", "敏感信息数", "MD5码", "规则编号", "发现时间"]
        if column in [0, 2, 3, 4]:
            value = self.table.item(row, column).text()
            title = col_titles[column]
            # 文件名列显示文件路径，其它列显示内容本身
            if column == 0 and value in self.file_paths:
                content = self.file_paths[value]
            else:
                content = value
            msg_box = QMessageBox(self)
            msg_box.setWindowTitle(title)
            msg_box.setText(content)
            msg_box.setStyleSheet("QLabel{color: black;}")
            msg_box.exec_()
            
    def refresh_table_data(self):
        """刷新表格数据"""
        try:
            # 保存当前搜索文本
            current_search = self.search_edit.text().strip()
            
            if current_search:
                # 如果有搜索条件，使用搜索方法刷新
                self.search_files()
            else:
                # 否则直接重新加载所有数据
                self.load_table_data()
                
        except sqlite3.Error as e:
            QMessageBox.warning(self, "数据刷新错误", f"刷新数据时发生错误：{str(e)}")

    def closeEvent(self, event):
        # 停止定时器
        self.refresh_timer.stop()
        # 关闭数据库连接
        self.conn.close()
        event.accept()


if __name__ == "__main__":
    app = QApplication(sys.argv)
    app.setStyle("Fusion")
    palette = QPalette()
    palette.setColor(QPalette.Window, QColor(240, 248, 255))   # 整体浅蓝底
    app.setPalette(palette)

    window = SensitiveFileWindow()
    window.show()
    sys.exit(app.exec_())
