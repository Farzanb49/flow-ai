import os
import time
from flask import Flask, jsonify

app = Flask(__name__)
port = int(os.environ.get('PORT', 8080))

@app.route('/')
def hello():
    return jsonify({
        'message': 'Hello from Flow Test Python App!',
        'timestamp': time.strftime('%Y-%m-%d %H:%M:%S UTC', time.gmtime()),
        'environment': os.environ.get('FLASK_ENV', 'development'),
        'port': port
    })

@app.route('/health')
def health():
    return jsonify({
        'status': 'healthy',
        'uptime': 'N/A (psutil not available)',
        'memory': 'N/A (psutil not available)'
    })

@app.route('/env')
def env():
    return jsonify({
        'environment': dict(os.environ),
        'python_version': os.sys.version,
        'platform': os.sys.platform
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=port, debug=False)
