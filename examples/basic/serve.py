#!/usr/bin/env python3
"""Dev server with Cache-Control: no-cache for WASM development."""
import http.server
import os
import sys

PORT = int(sys.argv[1]) if len(sys.argv) > 1 else 8080
DIR = sys.argv[2] if len(sys.argv) > 2 else os.getcwd()


class NoCacheHandler(http.server.SimpleHTTPRequestHandler):
    def end_headers(self):
        self.send_header("Cache-Control", "no-cache, no-store, must-revalidate")
        self.send_header("Pragma", "no-cache")
        self.send_header("Expires", "0")
        super().end_headers()


os.chdir(DIR)
print(f"Serving {DIR} at http://localhost:{PORT}  (Ctrl+C to stop)")
http.server.HTTPServer(("", PORT), NoCacheHandler).serve_forever()
