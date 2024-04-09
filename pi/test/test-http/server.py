from http.server import HTTPServer, ThreadingHTTPServer, BaseHTTPRequestHandler
import time

class MyHandler(BaseHTTPRequestHandler):
    
    def do_GET(self):
        content = "<h1>Hello World</h1><p>Default page</p>"
        try:
                with open("./templates/index.txt") as template:
                        content = template.read()
                        template.close()
        except FileNotFoundError as fnf:
                print("Template not found, sending default response")
        
        self.send_response(200)
        self.send_header("Content-type", "text/html")
        self.send_header("Content-length", str(len(content)))
        self.end_headers()
        self.wfile.write(bytes(content, encoding='utf8'))
        
        #x = 0
        #while True:
        #        print(x)
        #        x += 1
        #        time.sleep(3)

server = HTTPServer(("", 8000), MyHandler)
#server = ThreadingHTTPServer(("", 8000), MyHandler)
print(server.server_address)
server.serve_forever()
