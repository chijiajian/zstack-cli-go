"""Loopback-only test service for the packaged CLI and PowerShell contract."""
import base64
import hashlib
import hmac
import json
import sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import parse_qs, urlsplit

SECRET = ' 测试密码 "quotes"\n '
VMS = {
    "vm-a": {"uuid": "vm-a", "name": "开发 VM", "state": "Stopped", "cpuNum": 2, "memorySize": 1610612736},
    "vm-b": {"uuid": "vm-b", "name": "web-prod", "state": "Running", "cpuNum": 4, "memorySize": 4294967296},
    "vm-c": {"uuid": "vm-c", "name": "Paused", "state": "Paused"},
}
READS = 0
LAST_CONDITION = None
START_REQUESTS = 0
START_JOB_POLLS = 0
START_JOB = "/zstack/v1/api-jobs/test-start"
KEY_ID = "synthetic-access-key"
KEY_SECRET = ' synthetic signing secret "quotes" '
if len(sys.argv) > 2:
    VMS = {
        f"vm-{i:05d}": {"uuid": f"vm-{i:05d}", "name": f"VM {i}", "state": "Running", "cpuNum": 2, "memorySize": 2**30}
        for i in range(int(sys.argv[2]))
    }


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *_args):
        pass

    def reply(self, body, code=200):
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(json.dumps(body).encode())

    def do_PUT(self):
        global START_REQUESTS
        path = urlsplit(self.path).path
        if path == "/zstack/v1/accounts/login":
            request = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
            login = request.get("logInByAccount", {})
            if login.get("accountName") != "mock" or login.get("password") != hashlib.sha512(SECRET.encode()).hexdigest():
                self.reply({"error": {"code": "TEST.1", "description": "test-secret-echo"}}, 401)
                return
            self.reply({"inventory": {"uuid": "test-session"}})
            return
        if path == "/zstack/v1/vm-instances/vm-a/actions":
            START_REQUESTS += 1
            self.reply({"location": START_JOB}, 202)
            return
        self.reply({"error": {"code": "TEST.2"}}, 409)

    def do_GET(self):
        global READS, LAST_CONDITION, START_JOB_POLLS
        url = urlsplit(self.path)
        if url.path == "/test/reads":
            self.reply({"inventoryReads": READS, "lastInventoryCondition": LAST_CONDITION,
                        "startRequests": START_REQUESTS, "startJobPolls": START_JOB_POLLS})
            return
        if url.path == START_JOB and START_REQUESTS and self.headers.get("Authorization") == "OAuth test-session":
            START_JOB_POLLS += 1
            if START_JOB_POLLS == 1:
                self.reply({}, 202)
            else:
                VMS["vm-a"]["state"] = "Running"
                self.reply({"inventory": {**VMS["vm-a"], "state": "Starting"}})
            return
        if url.path == "/zstack/v1/accounts/sessions/test-session/valid":
            self.reply({"valid": True})
            return
        auth = self.headers.get("Authorization", "")
        if auth.startswith("ZStack "):
            date = self.headers.get("Date", "")
            payload = f"GET\n{date}\n{url.path.removeprefix('/zstack')}".encode()
            signature = base64.b64encode(hmac.digest(KEY_SECRET.encode(), payload, "sha1")).decode()
            if not date.endswith(" CST") or not hmac.compare_digest(auth, f"ZStack {KEY_ID}:{signature}"):
                self.reply({"error": {"code": "TEST.KEY"}}, 401)
                return
        elif auth != "OAuth test-session":
            self.reply({"error": {"code": "TEST.3"}}, 404)
            return
        if url.path != "/zstack/v1/vm-instances":
            self.reply({"error": {"code": "TEST.3"}}, 404)
            return
        query = parse_qs(url.query)
        READS += 1
        LAST_CONDITION = query.get("q", [None])[0]
        rows = [vm for uuid, vm in VMS.items() if not query.get("q") or f"uuid={uuid}" in query["q"]]
        if LAST_CONDITION == "uuid=wrong-identity":
            rows = [VMS["vm-b"]]
        start = int(query.get("start", ["0"])[0])
        limit = int(query.get("limit", ["100"])[0])
        self.reply({"inventories": rows[start:start + limit], "total": len(rows)})

    def do_DELETE(self):
        if self.path == "/zstack/v1/accounts/sessions/test-session":
            self.reply({})
            return
        self.reply({}, 404)


with ThreadingHTTPServer(("127.0.0.1", 0), Handler) as server:
    Path(sys.argv[1]).write_text(str(server.server_address[1]), encoding="ascii")
    server.serve_forever()
