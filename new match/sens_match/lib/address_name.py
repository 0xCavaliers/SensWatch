import re, sys, json
from pyhanlp import *
import chardet
import codecs

class address_name:
    """地址和人名识别"""
    def __init__(self):
        self.address = r'(ns|nsf)'  # 地址
        self.person_name = r'nr'  # 人名
        self.CRFnewSegment = HanLP.newSegment("crf")  # 通过crf算法识别实体

    def check_chinese_address(self, value: str) -> dict:
        try:
            # 确保输入是字符串类型
            if not isinstance(value, str):
                value = str(value)

            address_list = self.CRFnewSegment.seg(value)
            dict = {}
            for i in address_list:
                dict[str(i.word)] = [str(i.nature)]

            addresses = []
            names = []
            for key, value in dict.items():
                value = str(value)
                if re.search(self.address, value):
                    addresses.append(key)
                if re.search(self.person_name, value):
                    names.append(key)
            
            return {
                "addresses": addresses,
                "names": names
            }
        except Exception as e:
            print(json.dumps({
                "error": f"处理文本时出错: {str(e)}",
                "addresses": [],
                "names": []
            }))
            sys.exit(1)

def detect_encoding(file_path):
    """检测文件编码"""
    try:
        with open(file_path, 'rb') as f:
            raw_data = f.read()
            result = chardet.detect(raw_data)
            return result['encoding'] or 'utf-8'
    except Exception as e:
        print(json.dumps({
            "error": f"检测文件编码失败: {str(e)}",
            "addresses": [],
            "names": []
        }))
        sys.exit(1)

def read_file_content(file_path):
    """读取文件内容，尝试多种编码"""
    encodings = ['utf-8', 'gbk', 'gb2312', 'gb18030', 'big5']
    
    # 首先尝试检测编码
    detected_encoding = detect_encoding(file_path)
    if detected_encoding:
        encodings.insert(0, detected_encoding)
    
    # 尝试不同的编码
    for encoding in encodings:
        try:
            with codecs.open(file_path, 'r', encoding=encoding) as f:
                return f.read()
        except UnicodeDecodeError:
            continue
        except Exception as e:
            print(json.dumps({
                "error": f"读取文件失败: {str(e)}",
                "addresses": [],
                "names": []
            }))
            sys.exit(1)
    
    # 如果所有编码都失败，使用二进制模式读取
    try:
        with open(file_path, 'rb') as f:
            return f.read().decode('utf-8', errors='ignore')
    except Exception as e:
        print(json.dumps({
            "error": f"无法读取文件: {str(e)}",
            "addresses": [],
            "names": []
        }))
        sys.exit(1)

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print(json.dumps({
            "error": "Usage: python address_name.py <input_file>",
            "addresses": [],
            "names": []
        }))
        sys.exit(1)
        
    try:
        file_path = sys.argv[1]
        value = read_file_content(file_path)
        
        address_checker = address_name()
        result = address_checker.check_chinese_address(value)
        print(json.dumps(result))
    except Exception as e:
        print(json.dumps({
            "error": f"处理失败: {str(e)}",
            "addresses": [],
            "names": []
        }))
        sys.exit(1)
        sys.exit(1)