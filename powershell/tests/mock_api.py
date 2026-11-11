"""Loopback-only test service for the packaged CLI and PowerShell contract."""
import base64
import hashlib
import hmac
import json
import re
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
    "volume-snapshots": [],
    "vm-instances/cdroms": [
        {"uuid": "cdrom-a", "name": "primary drive", "vmInstanceUuid": "vm-a", "isoUuid": "", "deviceId": 0},
        {"uuid": "cdrom-b", "name": "secondary drive", "vmInstanceUuid": "vm-a", "isoUuid": "image-other", "deviceId": 1},
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
    "images": [{"uuid": "image-a", "name": "installer", "mediaType": "ISO", "state": "Enabled", "status": "Ready", "size": 1073741824, "actualSize": 161061273},
               {"uuid": "image-root", "name": "root template", "mediaType": "RootVolumeTemplate", "size": 1073741824}],
    "instance-offerings": [{"uuid": "offering-a", "name": "small", "memorySize": 2147483648, "cpuNum": 2}],
    "disk-offerings": [{"uuid": "offering-disk", "name": "one-gib", "diskSize": 1073741824}],
    "l2-networks": [{"uuid": "l2-a", "name": "transport", "zoneUuid": "zone-a", "type": "L2VlanNetwork", "virtualNetworkId": 0}],
    "l3-networks": [{"uuid": "l3-a", "name": "app-net", "l2NetworkUuid": "l2-a", "zoneUuid": "zone-a", "ipVersion": 4}],
    "l3-networks/ip-ranges": [{"uuid": "range-a", "name": "app-range", "l3NetworkUuid": "l3-a", "startIp": "192.0.2.20", "endIp": "192.0.2.30", "networkCidr": "192.0.2.0/24"}],
}
RESOURCE_ROWS.update({
    "longjobs": [
        {"uuid": f"job-{state.lower()}", "name": f"fixture {state}", "jobName": "Fixture", "state": state,
         "targetResourceUuid": "vm-a", "jobResult": "synthetic-job-result",
         "description": "fixture job metadata", "executeTime": 1234, "createDate": "2026-11-11T02:03:04Z"}
        for state in ("Succeeded", "Failed", "Canceled", "Suspended", "Running")
    ] + [{"uuid": "job-unnamed", "state": "Succeeded"},
         {"uuid": "job-resume-stuck", "name": "stuck", "state": "Suspended"}],
    "accounts": [
        {"uuid": f"acct-{i:03}", "name": f"account-{i:03}", "type": "Normal", "password": "synthetic-field"}
        for i in range(105)
    ],
    "accounts/users": [{"uuid": "user-a", "name": "alice", "accountUuid": "acct-000", "password": "synthetic-field"}],
    "identities/roles": [{"uuid": "role-a", "name": "operator", "identity": "fixture-role"}],
    "accounts/policies": [{"uuid": "policy-a", "name": "audit", "accountUuid": "acct-000"}],
    "iam2/projects": [{"uuid": "project-a", "name": "project", "linkedAccountUuid": "acct-000"}],
    "security-groups": [{"uuid": "sg-a", "name": "apps"}, {"uuid": "sg-b", "name": "other"}],
    "security-groups/rules": [
        {"uuid": "sgr-a", "securityGroupUuid": "sg-a", "type": "Ingress", "protocol": "TCP", "startPort": 443, "endPort": 443, "allowedCidr": "192.0.2.0/24"},
        {"uuid": "sgr-b", "securityGroupUuid": "sg-b", "type": "Ingress", "protocol": "UDP", "startPort": 53, "endPort": 53},
    ],
    "vips": [{"uuid": "vip-a", "name": "front door", "l3NetworkUuid": "l3-a", "ip": "192.0.2.20"}],
    "load-balancers": [{"uuid": "lb-a", "name": "edge", "vipUuid": "vip-a"}, {"uuid": "lb-b", "name": "other"}],
    "load-balancers/listeners": [
        {"uuid": "listener-a", "name": "https", "loadBalancerUuid": "lb-a", "loadBalancerPort": 443, "serverGroupUuid": "pool-a"},
        {"uuid": "listener-b", "name": "other", "loadBalancerUuid": "lb-b"},
    ],
    "load-balancers/servergroups": [
        {"uuid": "pool-a", "name": "apps", "loadBalancerUuid": "lb-a"},
        {"uuid": "pool-b", "name": "other", "loadBalancerUuid": "lb-b"},
    ],
})

LIFECYCLE_REQUESTS = {}
LIFECYCLE_PAYLOADS = {}
REBOOT_JOBS = {}
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

    def record_write(self, action, payload):
        LIFECYCLE_REQUESTS[action] = LIFECYCLE_REQUESTS.get(action, 0) + 1
        LIFECYCLE_PAYLOADS[action] = payload

    def device_write(self, method):
        url = urlsplit(self.path)
        path = url.path.removeprefix("/zstack/v1/")
        routes = (
            ("POST", r"volumes/data", "create-volume"),
            ("DELETE", r"volumes/([^/]+)", "delete-volume"),
            ("PUT", r"volumes/([^/]+)/actions", "recover-volume"),
            ("POST", r"volumes/([^/]+)/vm-instances/([^/]+)", "attach-volume"),
            ("DELETE", r"volumes/([^/]+)/vm-instances", "detach-volume"),
            ("PUT", r"volumes/(?:data/)?resize/([^/]+)/actions", "resize-volume"),
            ("POST", r"volumes/([^/]+)/volume-snapshots", "create-snapshot"),
            ("PUT", r"volume-snapshots/([^/]+)/actions", "revert-snapshot"),
            ("DELETE", r"volume-snapshots/([^/]+)", "delete-snapshot"),
            ("POST", r"vm-instances/([^/]+)/l3-networks/([^/]+)", "attach-nic"),
            ("DELETE", r"vm-instances/nics/([^/]+)", "detach-nic"),
            ("POST", r"vm-instances/([^/]+)/iso/([^/]+)", "attach-iso"),
            ("DELETE", r"vm-instances/([^/]+)/iso", "detach-iso"),
        )
        for verb, pattern, action in routes:
            match = re.fullmatch(pattern, path)
            if verb == method and match:
                break
        else:
            return False
        if self.headers.get("Authorization") != "OAuth test-session":
            self.reply({}, 403)
            return True
        size = int(self.headers.get("Content-Length", 0))
        body = json.loads(self.rfile.read(size)) if size else {}
        query = parse_qs(url.query)
        self.record_write(action, {"path": path, "query": query, "body": body})
        ids = match.groups()
        identifier = ids[0] if ids else ""
        disks = RESOURCE_ROWS["volumes"]
        snaps = RESOURCE_ROWS["volume-snapshots"]
        nics = RESOURCE_ROWS["vm-instances/nics"]
        drives = RESOURCE_ROWS["vm-instances/cdroms"]
        volume = next((row for row in disks if row["uuid"] == identifier), None)
        if action == "create-volume":
            detail = body["params"]
            if detail["name"] == "reject-volume":
                self.reply({"error": {"code": "TEST.5", "details": "synthetic-device-secret"}}, 403)
                return True
            requested_size = detail.get("diskSize", 1073741824)
            disks.append({"uuid": detail["resourceUuid"], "name": detail["name"], "type": "Data",
                          "status": "Ready", "size": requested_size + (detail["name"] == "misreported-volume"),
                          "vmInstanceUuid": "", "diskOfferingUuid": detail.get("diskOfferingUuid", ""),
                          "primaryStorageUuid": detail.get("primaryStorageUuid", "")})
        elif action == "delete-volume":
            volume["status"] = "Deleted"
        elif action == "recover-volume":
            assert body == {"recoverDataVolume": {}}
            volume["status"] = "Ready"
        elif action == "attach-volume":
            volume["vmInstanceUuid"] = ids[1]
        elif action == "detach-volume":
            assert query == {"vmUuid": [volume["vmInstanceUuid"]]}
            volume["vmInstanceUuid"] = ""
        elif action == "resize-volume":
            volume["size"] = next(iter(body.values()))["size"]
        elif action == "create-snapshot":
            detail = body["params"]
            snaps.append({"uuid": detail["resourceUuid"], "name": detail["name"], "volumeUuid": identifier,
                          "status": "Ready", "state": "Enabled", "latest": True})
        elif action == "revert-snapshot":
            assert body == {"revertVolumeFromSnapshot": {}}
        elif action == "delete-snapshot":
            snaps.remove(next(row for row in snaps if row["uuid"] == identifier))
        elif action == "attach-nic":
            params = body["params"]
            nic = {"uuid": "nic-new", "vmInstanceUuid": identifier, "l3NetworkUuid": ids[1], "deviceId": 1,
                   "mac": params.get("customMac", "").lower(), "ip": params.get("staticIp", "")}
            if nic["ip"] == "192.0.2.202":
                nic["ip"] = "192.0.2.203"
            nics.append(nic)
            self.reply({"inventory": {"uuid": identifier, "vmNics": [row for row in nics if row["vmInstanceUuid"] == identifier]}})
            return True
        elif action == "detach-nic":
            nics.remove(next(row for row in nics if row["uuid"] == identifier))
        elif action == "attach-iso":
            drive_id = body["systemTags"][0].removeprefix("cdromUuid::")
            next(row for row in drives if row["uuid"] == drive_id)["isoUuid"] = ids[1]
        elif action == "detach-iso":
            iso = query["isoUuid"][0]
            next(row for row in drives if row["vmInstanceUuid"] == identifier and row["isoUuid"] == iso)["isoUuid"] = ""
        self.reply({})
        return True

    def do_POST(self):
        if self.device_write("POST"):
            return
        if self.path != "/zstack/v1/vm-instances" or self.headers.get("Authorization") != "OAuth test-session":
            self.reply({}, 404)
            return
        params = json.loads(self.rfile.read(int(self.headers["Content-Length"])))["params"]
        self.record_write("create", params)
        if params["name"] == "reject-vm":
            self.reply({"error": {"code": "TEST.5", "details": "synthetic secret"}}, 403)
            return
        uuid = params["resourceUuid"]
        assert uuid not in VMS
        state = {"CreateStopped": "Stopped", "InstantStart": "Running"}[params["strategy"]]
        vm = {"uuid": uuid, "name": params["name"], "state": state, "bootId": 0,
              "description": params.get("description", ""), "cpuNum": params.get("cpuNum", 2),
              "memorySize": params.get("memorySize", 2147483648)}
        VMS[uuid] = vm
        RESOURCE_ROWS["volumes"].append({"uuid": f"{uuid}-root", "name": "fixture root", "type": "Root",
                                          "vmInstanceUuid": uuid, "status": "Ready", "size": 1073741824})
        response = {**vm, "uuid": "wrong-response"} if params["name"] == "unknown-vm" else vm
        self.reply({"inventory": response})

    def do_PUT(self):
        global START_REQUESTS
        if self.device_write("PUT"):
            return
        path = urlsplit(self.path).path
        if path == "/zstack/v1/accounts/login":
            request = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
            login = request.get("logInByAccount", {})
            if login.get("accountName") != "mock" or login.get("password") != hashlib.sha512(SECRET.encode()).hexdigest():
                self.reply({"error": {"code": "TEST.1", "description": "test-secret-echo"}}, 401)
                return
            self.reply({"inventory": {"uuid": "test-session"}})
            return
        if path.startswith("/zstack/v1/longjobs/") and path.endswith("/actions") and self.headers.get("Authorization") == "OAuth test-session":
            job_id = path.split("/")[-2]
            body = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
            assert body == {"resumeLongJob": {}}
            job = next(row for row in RESOURCE_ROWS["longjobs"] if row["uuid"] == job_id)
            self.record_write("resume-long-job", {"uuid": job_id})
            if job_id != "job-resume-stuck":
                job["state"] = "Succeeded"
            self.reply({"inventory": {"uuid": job_id, "state": "Running"}})
            return
        if path == "/zstack/v1/vm-instances/vm-a/actions":
            START_REQUESTS += 1
            self.reply({"location": START_JOB}, 202)
            return
        if path.startswith("/zstack/v1/vm-instances/") and path.endswith("/actions") and self.headers.get("Authorization") == "OAuth test-session":
            uuid = path.split("/")[-2]
            vm = VMS[uuid]
            body = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
            action, payload = next(iter(body.items()))
            self.record_write(action, payload)
            if action == "rebootVmInstance":
                location = f"/zstack/v1/api-jobs/reboot-{uuid}"
                REBOOT_JOBS[location] = uuid
                self.reply({"location": location}, 202)
                return
            if action == "updateVmInstance":
                vm.update(payload)
            else:
                vm["state"] = {"pauseVmInstance": "Paused", "resumeVmInstance": "Running",
                               "stopVmInstance": "Stopped", "recoverVmInstance": "Stopped"}[action]
            if action == "recoverVmInstance":
                for disk in RESOURCE_ROWS["volumes"]:
                    if disk.get("vmInstanceUuid") == uuid and disk["type"] == "Root":
                        disk["status"] = "Ready"
            self.reply({"inventory": vm})
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
        if url.path == "/test/lifecycle":
            self.reply({"requests": LIFECYCLE_REQUESTS, "payloads": LIFECYCLE_PAYLOADS,
                        "bootIds": {uuid: vm["bootId"] for uuid, vm in VMS.items() if "bootId" in vm}})
            return
        if url.path in REBOOT_JOBS and self.headers.get("Authorization") == "OAuth test-session":
            vm = VMS[REBOOT_JOBS.pop(url.path)]
            vm["bootId"] += 1
            self.reply({"inventory": vm})
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
        if resource == "zql":
            RESOURCE_REQUESTS[resource] = RESOURCE_REQUESTS.get(resource, 0) + 1
            statement = query.get("zql", [""])[0]
            if statement == "query denied":
                self.reply({"error": {"code": "TEST.ZQL", "details": "synthetic-zql-secret"}}, 403)
            elif statement == "query invalid":
                self.reply({"results": [{"inventories": []}, "invalid"]})
            else:
                assert statement == "query vmInstance.uuid where name='中文;quoted'; count vmInstance"
                self.reply({"results": [
                    {"inventories": [{"uuid": "vm-a", "dateText": "2026-11-11T02:03:04Z"}], "returnWith": {"cpu": [{"value": 0}]}},
                    {"total": 0, "inventoryAggregateFunctions": {"max": 7}},
                ]})
            return
        if resource == "hosts/capacities/cpu-memory":
            self.reply({"totalCpu": 0, "availableCpu": 0, "totalMemory": 9007199254740993, "availableMemory": 1})
            return
        if resource == "primary-storage/capacities":
            self.reply({"totalCapacity": 10, "availableCapacity": 5, "totalPhysicalCapacity": 20, "availablePhysicalCapacity": 10})
            return
        if resource == "zwatch/metrics":
            self.reply({"data": []})
            return
        if resource == "accounts/policies" and condition == "uuid=denied":
            self.reply({"error": {"code": "ID.1002", "details": "synthetic-policy-secret"}}, 503)
            return
        if resource == "vm-instances":
            READS += 1
            LAST_CONDITION = condition
            rows = list(VMS.values())
        elif resource == "global-configurations":
            rows = [{"category": category, "name": "deletionPolicy", "value": "Delay"} for category in ("vm", "volume")]
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
        if self.device_write("DELETE"):
            return
        url = urlsplit(self.path)
        if url.path.startswith("/zstack/v1/vm-instances/") and self.headers.get("Authorization") == "OAuth test-session":
            uuid = url.path.rsplit("/", 1)[1]
            assert parse_qs(url.query) == {"deleteMode": ["Permissive"]}
            self.record_write("delete", {"mode": "Permissive"})
            VMS[uuid]["state"] = "Destroyed"
            for disk in RESOURCE_ROWS["volumes"]:
                if disk.get("vmInstanceUuid") == uuid and disk["type"] == "Root":
                    disk["status"] = "Deleted"
            self.reply({})
            return
        if self.path == "/zstack/v1/accounts/sessions/test-session":
            self.reply({})
            return
        self.reply({}, 404)


with ThreadingHTTPServer(("127.0.0.1", 0), Handler) as server:
    Path(sys.argv[1]).write_text(str(server.server_address[1]), encoding="ascii")
    server.serve_forever()
