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
    "vm-a": {"uuid": "vm-a", "name": "开发 VM", "state": "Stopped", "cpuNum": 2, "memorySize": 1610612736,
             "description": "fixture detail", "hostUuid": "host-a", "imageUuid": "missing-image",
             "vmNics": [{"l3NetworkUuid": "l3-a"}]},
    "vm-b": {"uuid": "vm-b", "name": "web-prod", "state": "Running", "cpuNum": 4, "memorySize": 4294967296},
    "vm-c": {"uuid": "vm-c", "name": "Paused", "state": "Paused"},
}
RESOURCE_ROWS = {
    "volumes": [
        {"uuid": "disk-a", "name": "system", "vmInstanceUuid": "vm-a", "type": "Root", "size": 9007199254740993, "actualSize": 0, "primaryStorageUuid": "primary-a"},
        {"uuid": "disk-b", "name": "data", "vmInstanceUuid": "vm-b", "type": "Data", "size": 1610612737},
        {"uuid": "disk-c", "name": "spare", "lastVmInstanceUuid": "vm-a", "type": "Data"},
    ],
    "hosts": [{"uuid": "host-a", "name": "esx-like", "clusterUuid": "cluster-a", "zoneUuid": "zone-a", "status": "Connected", "totalCpuCapacity": 3000000000, "cpuNum": 8, "totalMemoryCapacity": 8589934592, "availableMemoryCapacity": 0}],
    "clusters": [{"uuid": "cluster-a", "name": "cluster", "zoneUuid": "zone-a", "hypervisorType": "KVM"}],
    "zones": [{"uuid": "zone-a", "name": "zone", "state": "Enabled"}],
    "vm-instances/nics": [
        {"uuid": "nic-a", "vmInstanceUuid": "vm-a", "l3NetworkUuid": "l3-a", "mac": "00:00:5e:00:53:01", "deviceId": 0},
        {"uuid": "nic-b", "vmInstanceUuid": "vm-b", "l3NetworkUuid": "l3-a", "mac": "00:00:5e:00:53:02", "deviceId": 0},
    ],
    "primary-storage": [{"uuid": "primary-a", "name": "primary", "zoneUuid": "zone-a", "totalCapacity": 4294967296, "availableCapacity": 0, "totalPhysicalCapacity": 2147483648, "systemUsedCapacity": -4294967297}],
    "backup-storage": [{"uuid": "storage-a", "name": "backup", "totalCapacity": 1073741824}],
    "images": [{"uuid": "image-a", "name": "installer", "mediaType": "ISO", "size": 1073741824, "actualSize": 161061273}],
    "instance-offerings": [{"uuid": "offering-a", "name": "small", "memorySize": 2147483648, "cpuNum": 2}],
    "disk-offerings": [{"uuid": "offering-disk", "name": "one-gib", "diskSize": 1073741824}],
    "l2-networks": [{"uuid": "l2-a", "name": "transport", "zoneUuid": "zone-a", "type": "L2VlanNetwork", "virtualNetworkId": 0}],
    "l3-networks": [{"uuid": "l3-a", "name": "app-net", "l2NetworkUuid": "l2-a", "zoneUuid": "zone-a", "ipVersion": 4}],
    "l3-networks/ip-ranges": [{"uuid": "range-a", "name": "app-range", "l3NetworkUuid": "l3-a", "startIp": "192.0.2.20", "endIp": "192.0.2.30", "networkCidr": "192.0.2.0/24"}],
}
RESOURCE_REQUESTS = {}
RESOURCE_CONDITIONS = {}
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
                        "startRequests": START_REQUESTS, "startJobPolls": START_JOB_POLLS,
                        "resourceRequests": RESOURCE_REQUESTS, "resourceConditions": RESOURCE_CONDITIONS})
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
        resource = url.path.removeprefix("/zstack/v1/")
        query = parse_qs(url.query)
        condition = query.get("q", [None])[0]
        if resource == "vm-instances":
            READS += 1
            LAST_CONDITION = condition
            rows = list(VMS.values())
        elif resource in RESOURCE_ROWS:
            RESOURCE_REQUESTS[resource] = RESOURCE_REQUESTS.get(resource, 0) + 1
            RESOURCE_CONDITIONS[resource] = condition
            rows = RESOURCE_ROWS[resource]
        else:
            collection, _, identifier = resource.rpartition("/")
            records = VMS.values() if collection == "vm-instances" else RESOURCE_ROWS.get(collection, [])
            rows = [row for row in records if row["uuid"] == identifier]
            if resource == "images/missing-image":
                self.reply({"error": {"code": "TEST.4", "details": "synthetic-related-response-marker"}}, 404)
                return
            if not rows:
                self.reply({"error": {"code": "TEST.MISSING"}}, 404)
                return
        for clause in query.get("q", []):
            operator = "?=" if "?=" in clause else "="
            field, _, values = clause.partition(operator)
            allowed = values.split(",") if operator == "?=" else [values]
            rows = [row for row in rows if row.get(field) in allowed]
        if resource == "vm-instances" and condition == "uuid=wrong-identity":
            rows = [VMS["vm-b"]]
        if resource == "volumes" and condition == "uuid=wrong-inventory":
            rows = RESOURCE_ROWS["volumes"][:1]

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
