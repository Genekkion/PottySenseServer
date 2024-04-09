from flask import Flask, render_template
import time

app = Flask(__name__)

@app.route('/')
def index():
    return render_template("index.html")

@app.route('/ping')
def ping():
    return "PING"
    
if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0')
    while True:
        print("Hello")
        time.sleep(3)
