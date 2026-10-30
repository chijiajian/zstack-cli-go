#!/usr/bin/env python3
"""Opt-in CLI write regression confined to resources owned by a durable journal."""
import argparse
import json
import os
from pathlib import Path
import re
import secrets
import shutil
import subprocess
import tempfile
import time
import uuid

from readonly_smoke import CLI as ReadCLI, interruption_handlers, RunInterrupted, validate_auth

QUERIES = {"volume": "volumes", "instance": "instances", "snapshot": "snapshots",
           "account": "accounts", "user": "users", "role": "roles", "policy": "policies",
           "organization": "organizations", "project": "projects"}


class Failure(Exception):
    def __init__(self, message, rejected=False):
        super().__init__(message)
        self.rejected = rejected


class Journal:
    def __init__(self, path, existing=False, endpoint=""):
        if not endpoint:
            raise Failure("journal requires the selected endpoint")
        self.path = Path(path)
        if existing:
            self.data = json.loads(self.path.read_text())
            self.validate()
            if self.data.get("endpoint") != endpoint:
                raise Failure("journal belongs to a different endpoint")
        else:
            self.data = {"version": 1, "run": uuid.uuid4().hex, "endpoint": endpoint, "resources": []}
            # Never overwrite another run's cleanup evidence.
            with self.path.open("x") as stream:
                os.chmod(self.path, 0o600)
                json.dump(self.data, stream)

    def validate(self):
        if self.data.get("version") != 1 or not re.fullmatch(r"[a-f0-9]{32}", self.data.get("run", "")):
            raise Failure("invalid ownership journal")
        if not isinstance(self.data.get("resources"), list):
            raise Failure("invalid journal resources")
        for item in self.data["resources"]:
            if not isinstance(item, dict) or item.get("state") not in {"planned", "submitted", "created", "rejected", "cleaned"}:
                raise Failure("invalid resource journal state")
            if item.get("root_volume") is not None and not re.fullmatch(r"[a-f0-9]{32}", item["root_volume"]):
                raise Failure("invalid root volume identity")
            if item.get("kind") not in QUERIES or not re.fullmatch(r"[a-f0-9]{32}", item.get("uuid", "")) or item.get("name") != self.name(item["kind"], item["uuid"]):
                raise Failure("journal contains an unowned resource")

    def name(self, kind, identifier):
        return f"codex-e2e-{self.data['run'][:12]}-{kind}-{identifier[:8]}"

    def save(self):
        descriptor, path = tempfile.mkstemp(prefix=".write-e2e-", dir=self.path.parent)
        try:
            with os.fdopen(descriptor, "w") as stream:
                json.dump(self.data, stream, indent=2)
                stream.flush()
                os.fsync(stream.fileno())
            os.replace(path, self.path)
        finally:
            if os.path.exists(path):
                os.unlink(path)

    def plan(self, kind):
        if kind not in QUERIES:
            raise Failure("unsupported resource kind")
        identifier = uuid.uuid4().hex
        item = {"kind": kind, "uuid": identifier, "name": self.name(kind, identifier), "state": "planned"}
        self.data["resources"].append(item)
        self.save()  # Intent and requested UUID are durable before the mutation.
        return item


class CLI:
    def __init__(self, binary, config, timeout=180):
        self.binary, self.config, self.timeout = binary, config, timeout
        self.deadline = None

    def run(self, args):
        env = {k: v for k, v in os.environ.items() if not k.startswith("ZSTACK_")}
        env["ZSTACK_CONFIG"] = str(self.config)
        remaining = self.timeout if self.deadline is None else min(self.timeout, self.deadline - time.monotonic())
        if remaining <= 0:
            raise Failure("suite deadline exceeded before request", rejected=True)
        try:
            result = subprocess.run([self.binary, *args, "--error-format", "json"], env=env, capture_output=True, text=True, timeout=remaining)
        except (OSError, UnicodeError, subprocess.TimeoutExpired):
            raise Failure("CLI unavailable or timed out; inspect the journal before retrying") from None
        if result.returncode:
            try:
                diagnostic = json.loads(result.stderr)["error"]
                if not isinstance(diagnostic, dict):
                    raise ValueError("invalid diagnostic")
                code = diagnostic.get("apiCode", "")
                if not isinstance(code, str) or not re.fullmatch(r"[A-Z][A-Z0-9_]{0,20}\.[0-9]{1,8}", code):
                    code = ""
            except (ValueError, KeyError, TypeError):
                code = ""
            suffix = ": " + code if code else ""
            raise Failure(f"{args[0]} {args[1]} exited with status {result.returncode}{suffix}", rejected=code.startswith("LICENSE."))
        return result.stdout

    def document(self, args):
        try:
            return json.loads(self.run([*args, "-o", "json"]))
        except ValueError:
            raise Failure("CLI returned invalid JSON; operation may have completed") from None


class Runner:
    def __init__(self, cli, journal):
        self.cli, self.journal = cli, journal

    def lookup(self, item):
        self.journal.validate()
        records = self.cli.document(["get", QUERIES[item["kind"]], "-q", "uuid=" + item["uuid"], "--limit", "2"])
        if not isinstance(records, list) or len(records) > 1 or any(not isinstance(row, dict) for row in records):
            raise Failure("ambiguous resource lookup")
        if records and (records[0].get("uuid") != item["uuid"] or records[0].get("name") != item["name"]):
            raise Failure("resource ownership differs from journal; refusing mutation")
        return records

    def require(self, item):
        records = self.lookup(item)
        if not records:
            raise Failure("owned resource disappeared")
        return records[0]

    def create(self, kind, flags=()):
        item = self.journal.plan(kind)
        if self.lookup(item):
            raise Failure("planned UUID already exists; refusing creation")
        item["state"] = "submitted"
        self.journal.save()
        try:
            result = self.cli.document(["create", kind, item["name"], "--resource-uuid", item["uuid"], *flags])
        except Failure as error:
            if error.rejected:
                item["state"] = "rejected"
                self.journal.save()
            raise
        if not isinstance(result, dict) or result.get("uuid") != item["uuid"] or not self.lookup(item):
            raise Failure("creation could not be confirmed")
        item["state"] = "created"
        self.journal.save()
        return item

    def update(self, item):
        if not self.lookup(item):
            raise Failure("created resource disappeared")
        value = "CLI E2E " + self.journal.data["run"]
        self.cli.run(["update", item["kind"], item["uuid"], "--description", value])
        records = self.lookup(item)
        if not records or records[0].get("description") != value:
            raise Failure("updated description was not persisted")

    def recover(self, item):
        before = self.require(item).copy()
        field = "status" if item["kind"] == "volume" else "state"
        expected = before.get(field)
        if not isinstance(expected, str) or not expected or expected in ("Destroyed", "Deleted"):
            raise Failure("resource has no recoverable initial state")
        self.cli.run(["delete", item["kind"], item["uuid"], "--yes"])
        deleted = self.lookup(item)
        if not deleted:
            raise Failure("resource was expunged by server policy; recovery cannot be validated")
        if deleted[0].get(field) not in ("Destroyed", "Deleted"):
            raise Failure("deletion state was not confirmed before recovery")
        self.cli.run([item["kind"], "recover", item["uuid"], "--yes"] if item["kind"] == "project" else [item["kind"], "recover", item["uuid"]])
        record = self.require(item)
        if record.get(field) != expected:
            raise Failure("recovery was not confirmed")
        item["recovery_verified"] = True
        self.journal.save()

    def cleanup(self):
        errors = []
        self.cli.deadline = None
        # Independent of test failure/deadline: each cleanup attempt has its own bound.
        for item in reversed(self.journal.data["resources"]):
            if item["state"] == "cleaned":
                continue
            if item["state"] == "planned":
                item["state"] = "cleaned"
                self.journal.save()
                continue
            try:
                records = self.lookup(item)
                if not records:
                    if item["state"] == "submitted":
                        raise Failure("creation outcome unknown; rerun cleanup after the API job settles")
                else:
                    if item["kind"] == "instance":
                        root = records[0].get("rootVolumeUuid")
                        if not isinstance(root, str) or not re.fullmatch(r"[a-f0-9]{32}", root):
                            raise Failure("instance root volume identity unavailable; refusing incomplete cleanup")
                        if item.get("root_volume") and item["root_volume"] != root:
                            raise Failure("instance root volume differs from journal")
                        item["root_volume"] = root
                        self.journal.save()
                    attached = records[0].get("vmInstanceUuid") if item["kind"] == "volume" else None
                    if attached:
                        owners = [r for r in self.journal.data["resources"] if r["kind"] == "instance" and r["uuid"] == attached]
                        if len(owners) != 1 or not self.lookup(owners[0]):
                            raise Failure("volume is attached outside this run; refusing detach")
                        self.cli.run(["volume", "detach", item["uuid"]])
                    if records[0].get("state") not in ("Deleted", "Destroyed") and records[0].get("status") != "Deleted":
                        self.cli.run(["delete", item["kind"], item["uuid"], "--yes"])
                    if item["kind"] in ("volume", "instance", "project") and self.lookup(item):
                        self.cli.run(["expunge", item["kind"], item["uuid"], "--yes"])
                    if self.lookup(item):
                        raise Failure("cleanup did not remove the resource")
                if item.get("root_volume"):
                    remaining = self.cli.document(["get", "volumes", "-q", "uuid=" + item["root_volume"], "--limit", "1"])
                    if not isinstance(remaining, list) or remaining:
                        raise Failure("instance root volume remains after cleanup")
                item["state"] = "cleaned"
                self.journal.save()
            except (Failure, OSError) as error:
                reason = str(error) if isinstance(error, Failure) else "cleanup journal could not be saved"
                errors.append({"kind": item["kind"], "uuid": item["uuid"], "reason": reason})
        return errors


def run_metadata(runner, directory, primary_storage=None, organization_type="Department", resources=None):
    password = Path(directory) / "password"
    password.touch(mode=0o600)
    password.write_text(secrets.token_urlsafe(36))
    statements = Path(directory) / "statements.json"
    statements.write_text(json.dumps([{"name": "query-account", "effect": "Allow", "actions": ["org.zstack.header.identity.APIQueryAccountMsg"]}]))
    plans = [
        ("volume", ["--size", "1G"] + (["--primary-storage", primary_storage] if primary_storage else [])),
        ("account", ["--password-file", str(password)]),
        ("user", ["--password-file", str(password)]),
        ("policy", ["--statements-file", str(statements)]),
        ("role", ["--statements-file", str(statements)]),
        ("organization", ["--type", organization_type]),
        ("project", []),
    ]
    for kind, flags in plans:
        if resources is not None and kind not in resources:
            continue
        item = runner.create(kind, flags)
        if kind != "policy":  # SDK does not expose policy update.
            runner.update(item)
        if kind == "project":
            runner.recover(item)


def run_compute(runner, image, network, host=None, primary_storage=None):
    flags = ["--image", image, "--l3-network", network, "--cpu", "1", "--memory", "1G", "--strategy", "CreateStopped"]
    if host:
        flags += ["--host", host]
    if primary_storage:
        flags += ["--primary-storage", primary_storage]
    vm = runner.create("instance", flags)
    record = runner.require(vm)
    if record.get("state") != "Stopped":
        raise Failure("test instance was not created stopped")
    root = record.get("rootVolumeUuid")
    if not isinstance(root, str) or not re.fullmatch(r"[a-f0-9]{32}", root):
        raise Failure("created instance has no valid root volume identity")
    vm["root_volume"] = root
    runner.journal.save()
    runner.update(vm)
    volume = runner.create("volume", ["--size", "1G"])
    runner.update(volume)
    # All mutating targets below must still match this run's UUID/name ownership.
    runner.lookup(vm)
    runner.lookup(volume)
    runner.cli.run(["volume", "attach", volume["uuid"], "--instance", vm["uuid"]])
    if runner.require(volume).get("vmInstanceUuid") != vm["uuid"]:
        raise Failure("volume attachment was not persisted")
    snapshot = runner.journal.plan("snapshot")
    if runner.lookup(snapshot):
        raise Failure("snapshot UUID collision")
    snapshot["state"] = "submitted"
    runner.journal.save()
    result = runner.cli.document(["snapshot", "create", volume["uuid"], "--name", snapshot["name"], "--resource-uuid", snapshot["uuid"]])
    if result.get("uuid") != snapshot["uuid"] or runner.require(snapshot).get("volumeUuid") != volume["uuid"]:
        raise Failure("snapshot creation was not confirmed")
    snapshot["state"] = "created"
    runner.journal.save()
    runner.cli.run(["snapshot", "revert", snapshot["uuid"], "--yes"])
    runner.cli.run(["delete", "snapshot", snapshot["uuid"], "--yes"])
    if runner.lookup(snapshot):
        raise Failure("snapshot deletion was not confirmed")
    snapshot["state"] = "cleaned"
    runner.journal.save()
    runner.lookup(volume)
    runner.cli.run(["volume", "resize", volume["uuid"], "--size", "2G"])
    detail = runner.cli.document(["describe", "volume", volume["uuid"]])
    if detail.get("resource", {}).get("size") != 2 * 1024 ** 3:
        raise Failure("volume resize was not confirmed")
    runner.cli.run(["volume", "detach", volume["uuid"]])
    if runner.require(volume).get("vmInstanceUuid"):
        raise Failure("volume detachment was not confirmed")
    for item in [volume, vm]:
        runner.recover(item)
    runner.journal.data["compute_checks"] = ["instance-create-stopped", "metadata-update", "volume-attach", "snapshot-create-revert-delete", "volume-resize", "volume-detach", "volume-recover", "instance-recover"]
    runner.journal.save()


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--cli", default="./zstack-cli")
    parser.add_argument("--config", required=True, help="session-only or AK config, copied into isolation")
    parser.add_argument("--journal", required=True, help="new durable journal path, retained for cleanup evidence")
    parser.add_argument("--execute", action="store_true", help="explicitly enable real writes")
    parser.add_argument("--cleanup-only", action="store_true", help="clean only resources belonging to an existing journal")
    parser.add_argument("--timeout", type=int, default=180)
    parser.add_argument("--suite-timeout", type=int, default=1200)
    parser.add_argument("--profile", choices=("metadata", "compute", "all"), default="metadata")
    parser.add_argument("--image", help="existing root image UUID for a stopped test instance")
    parser.add_argument("--l3-network", help="existing L3 UUID for a stopped test instance")
    parser.add_argument("--host", help="optional existing host UUID")
    parser.add_argument("--primary-storage")
    parser.add_argument("--organization-type", default="Department")
    parser.add_argument("--resources", default="volume,account,user,policy,role,organization,project", help="Explicit metadata resource selection; reports retain the executed scope")
    args = parser.parse_args(argv)
    resources=args.resources.split(",")
    if not resources or len(set(resources))!=len(resources) or any(r not in {"volume","account","user","policy","role","organization","project"} for r in resources):
        parser.error("invalid metadata resources")
    if args.profile in ("compute", "all"):
        if not args.image or not args.l3_network:
            parser.error("compute requires --image and --l3-network")
        if any(v is not None and not re.fullmatch(r"[a-fA-F0-9]{32}", v) for v in (args.image, args.l3_network, args.host, args.primary_storage)):
            parser.error("compute fixture references must be UUIDs")
    if not 1 <= args.suite_timeout <= 3600:
        parser.error("suite-timeout must be 1..3600 seconds")
    if not args.execute:
        parser.error("real writes require --execute")
    if not 1 <= args.timeout <= 600:
        parser.error("timeout must be 1..600 seconds")
    binary = shutil.which(args.cli)
    if not binary or not Path(args.config).is_file():
        parser.error("CLI executable or config unavailable")
    failure = None
    cleanup_errors = []
    try:
        with interruption_handlers(), tempfile.TemporaryDirectory(prefix="zstack-write-e2e-") as directory:
            config = Path(directory) / "config.yaml"
            config.touch(mode=0o600)
            config.write_bytes(Path(args.config).read_bytes())
            validate_auth(ReadCLI(binary, config))
            view = json.loads(ReadCLI(binary, config).run(["config", "view", "-o", "json"]))
            endpoint = view["contexts"][view["selected-context"]]["endpoint"]
            journal = Journal(args.journal, args.cleanup_only, endpoint)
            runner = Runner(CLI(binary, config, args.timeout), journal)
            runner.cli.deadline = time.monotonic() + args.suite_timeout
            try:
                if not args.cleanup_only:
                    if args.profile in ("metadata", "all"):
                        run_metadata(runner, directory, args.primary_storage, args.organization_type, resources)
                    if args.profile in ("compute", "all"):
                        run_compute(runner, args.image, args.l3_network, args.host, args.primary_storage)
            except (Failure, RunInterrupted) as error:
                failure = str(error) if isinstance(error, Failure) else "interrupted"
            finally:
                cleanup_errors = runner.cleanup()
    except Exception as error:
        failure = "runner or preflight failed: " + type(error).__name__
    report = {"passed": failure is None and not cleanup_errors, "failure": failure, "cleanup_errors": cleanup_errors, "journal": str(Path(args.journal).absolute()), "resources": resources if args.profile != "compute" else [], "profile": args.profile, "cleanup_only": args.cleanup_only}
    print(json.dumps(report, indent=2))
    return 0 if report["passed"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
