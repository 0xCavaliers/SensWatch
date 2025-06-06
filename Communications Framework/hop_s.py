from flask import Flask, request, jsonify
from datetime import datetime
import threading
import logging
import json

# 配置日志
logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger(__name__)

app = Flask(__name__)
# 存储多个客户端状态，使用字典存储，key为客户端IP
clients = {}

# 简单的API密钥验证
API_KEY = "secure-api-key-123"

def check_auth():
    """验证API密钥"""
    auth_header = request.headers.get('Authorization')
    if not auth_header or auth_header != f"Bearer {API_KEY}":
        return False
    return True

def display_clients_status():
    """在控制台显示所有客户端状态"""
    print("\n" + "="*50)
    print("当前客户端状态:")
    print("="*50)
    for client_ip, status in clients.items():
        print(f"客户端IP: {client_ip}")
        print(f"在线状态: {'在线' if status['online'] else '离线'}")
        print(f"最后心跳: {status['last_heartbeat']}")
        print("-"*50)
    print("\n")

@app.route('/api/register', methods=['POST'])
def register():
    """客户端注册接口"""
    try:
        logger.debug("Received register request from %s", request.remote_addr)
        logger.debug("Headers: %s", request.headers)
        
        if not check_auth():
            logger.warning("Unauthorized register attempt")
            return jsonify({"status": "error", "message": "Unauthorized"}), 401
        
        data = request.get_json()
        client_ip = data.get('ip_address', request.remote_addr)
        logger.debug("Request data: %s", data)
        
        # 为每个客户端创建状态记录
        clients[client_ip] = {
            "online": True,
            "last_heartbeat": datetime.now().isoformat(),
            "ip_address": client_ip
        }
        
        response = jsonify({
            "status": "success",
            "message": "Client registered successfully",
            "client_status": clients[client_ip]
        })
        
        logger.debug("Sending response: %s", response.get_data())
        return response
        
    except Exception as e:
        logger.exception("Error in register endpoint")
        return jsonify({"status": "error", "message": str(e)}), 500

@app.route('/api/heartbeat', methods=['POST'])
def heartbeat():
    """心跳检测接口"""
    if not check_auth():
        return jsonify({"status": "error", "message": "Unauthorized"}), 401
    
    data = request.get_json()
    client_ip = data.get('ip_address', request.remote_addr)
    
    if client_ip not in clients:
        return jsonify({"status": "error", "message": "Client not registered"}), 404
    
    clients[client_ip]["last_heartbeat"] = datetime.now().isoformat()
    
    return jsonify({
        "status": "success",
        "message": "Heartbeat received",
        "timestamp": clients[client_ip]["last_heartbeat"]
    })

@app.route('/api/unregister', methods=['POST'])
def unregister():
    """客户端注销接口"""
    if not check_auth():
        return jsonify({"status": "error", "message": "Unauthorized"}), 401
    
    data = request.get_json()
    client_ip = data.get('ip_address', request.remote_addr)
    
    if client_ip in clients:
        clients[client_ip]["online"] = False
        clients[client_ip]["last_heartbeat"] = datetime.now().isoformat()
    
    return jsonify({
        "status": "success",
        "message": "Client unregistered successfully",
        "timestamp": datetime.now().isoformat()
    })

@app.route('/api/status', methods=['GET'])
def status():
    """获取客户端状态（只返回请求客户端的状态）"""
    if not check_auth():
        return jsonify({"status": "error", "message": "Unauthorized"}), 401
    
    # 从请求头中获取客户端IP
    client_ip = request.headers.get('X-Client-IP')
    if not client_ip:
        return jsonify({"status": "error", "message": "Client IP not provided"}), 400
    
    if client_ip not in clients:
        return jsonify({"status": "error", "message": "Client not registered"}), 404
    
    return jsonify({
        "status": "success",
        "client_status": clients[client_ip]
    })

def start_heartbeat_checker():
    """启动心跳检查线程"""
    def checker():
        while True:
            current_time = datetime.now()
            for client_ip, status in clients.items():
                if status["online"] and status["last_heartbeat"]:
                    last_heartbeat = datetime.fromisoformat(status["last_heartbeat"])
                    if (current_time - last_heartbeat).seconds > 30:  # 30秒超时
                        status["online"] = False
                        logger.info(f"Client {client_ip} marked as offline due to heartbeat timeout")
            display_clients_status()  # 显示所有客户端状态
            threading.Event().wait(10)  # 每10秒检查一次
    
    thread = threading.Thread(target=checker)
    thread.daemon = True
    thread.start()

if __name__ == '__main__':
    start_heartbeat_checker()
    app.run(host='0.0.0.0', port=5000, debug=True)