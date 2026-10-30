#!/usr/bin/env python3
"""Bounded read-only CLI regression; requires a pre-authenticated config."""
import argparse
from contextlib import contextmanager
import json
import os
from pathlib import Path
import re
import shutil
import signal
import time
import subprocess
import tempfile

# Canonical query commands; aliases intentionally share the same coverage.
RESOURCES = dict.fromkeys("""
zones clusters hosts instances volumes images primary-storages long-jobs
cdroms disk-offerings instance-offerings image-storages image-groups
l2-networks l3-networks ip-ranges management-nodes nics snapshots tags
vips eips virtual-router-offerings virtual-routers vm-scripts vm-script-records
accounts users roles policies projects organizations alarms
security-groups security-group-rules port-forwarding-rules load-balancers
load-balancer-listeners load-balancer-server-groups
""".split(), "uuid")
RESOURCES.update({"events": "id", "audits": "id", "global-configs": "id"})

DETAILS = {"instances": "instance", "volumes": "volume", "long-jobs": "job"}


class CheckFailure(Exception):
    pass


class CLI:
    def __init__(self, binary, config, timeout=30, suite_timeout=600):
        self.binary, self.config, self.timeout = binary, config, timeout
        self.deadline = time.monotonic() + suite_timeout

    def run(self, args):
        # No caller can use this runner for login or resource mutations.
        allowed = (len(args) >= 2 and (
            args[0] == "get" and args[1] in RESOURCES or
            args[0] == "describe" and args[1] in DETAILS.values() or
            args == ["config", "view", "-o", "json"]))
        if not allowed:
            raise CheckFailure("command outside read-only allowlist")
        remaining = self.deadline - time.monotonic()
        if remaining <= 0:
            raise CheckFailure("suite deadline exceeded")
        env = {k: v for k, v in os.environ.items() if not k.startswith("ZSTACK_")}
        env["ZSTACK_CONFIG"] = str(self.config)
        request_flags = [] if args[0] == "config" else ["--request-timeout", f"{min(self.timeout, remaining):.6f}s"]
        try:
            result = subprocess.run(
                [self.binary, *args, *request_flags],
                env=env, capture_output=True, text=True, timeout=min(self.timeout + 5, remaining))
        except (OSError, UnicodeError, subprocess.TimeoutExpired):
            raise CheckFailure("CLI unavailable or timed out") from None
        if result.returncode:
            # Do not copy command arguments, server bodies or credentials to reports.
            raise CheckFailure(f"CLI exited with status {result.returncode}")
        if args[0] != "describe" and result.stderr.strip():
            raise CheckFailure("unexpected diagnostic output from CLI")
        return result.stdout


def document(cli, args):
    try:
        return json.loads(cli.run([*args, "-o", "json"]))
    except (ValueError, TypeError):
        raise CheckFailure("invalid JSON output") from None


def require(condition, reason):
    if not condition:
        raise CheckFailure(reason)


def validate_auth(cli):
    try:
        config = document(cli, ["config", "view"])
        selected = config["contexts"][config["selected-context"]]
        auth = selected["auth_type"]
        if auth == "account":
            require(selected["password"] == "", "use a session config created with --save-password=false")
            require(bool(selected["session_uuid"]), "selected context has no saved session")
        elif auth == "access-key":
            require(bool(selected["access_key_id"]) and bool(selected["access_key_secret"]), "selected context has incomplete access keys")
        else:
            raise CheckFailure("unsupported authentication type")
    except (KeyError, TypeError):
        raise CheckFailure("invalid selected configuration") from None


def valid_identity(key, value, resource=None):
    if resource in ("alarms", "tags") and key == "uuid":
        # Built-in alarms/tags may use opaque alphanumeric IDs, not hexadecimal UUIDs.
        return isinstance(value, str) and bool(re.fullmatch(r"[A-Za-z0-9]{1,128}", value))
    if key == "uuid":
        return isinstance(value, str) and bool(re.fullmatch(r"[0-9a-fA-F]{32}", value))
    if key == "id":
        return type(value) is int and value >= 0
    return isinstance(value, str) and bool(value.strip())


def identity_rows(value, key, resource=None):
    return isinstance(value, list) and all(
        isinstance(row, dict) and set(row) == {key} and valid_identity(key, row[key], resource)
        for row in value)


def validate_projection(output, fmt, key, value, resource=None):
    expected = [{key: value}]
    if fmt == "json":
        projected = json.loads(output)
        require(identity_rows(projected, key, resource) and projected == expected, "projection differs from selected identity")
    elif fmt == "text":
        require(output.strip() == f"[map[{key}:{value}]]", "invalid text projection")
    elif fmt == "yaml":
        match = re.fullmatch(r"\s*- " + re.escape(key) + r": ([^\r\n]+)\s*", output)
        require(match is not None, "invalid YAML projection shape")
        scalar = match[1].strip()
        if scalar.startswith('"'):
            scalar = json.loads(scalar)
        elif scalar.startswith("'") and scalar.endswith("'"):
            scalar = scalar[1:-1].replace("''", "'")
        require(str(scalar) == str(value), "YAML projection differs from selected identity")
    else:
        cells = []
        for line in output.splitlines():
            if "│" in line or "|" in line:
                cells.append([cell.strip() for cell in re.split(r"[│|]", line.strip(" │|"))])
            else:
                require(all(c.isspace() or c in "+-=" or "\u2500" <= c <= "\u257f" for c in line), "unexpected table content")
        require(len(cells) == 2 and cells[0] == [key.upper()] and cells[1] == [str(value)], "invalid table projection")


def run_suite(cli, sample_size=3):
    checks = []

    def check(label, action):
        try:
            action()
        except CheckFailure as error:
            checks.append({"check": label, "status": "failed", "reason": str(error)})
        except (KeyError, TypeError, ValueError, IndexError):
            checks.append({"check": label, "status": "failed", "reason": "unexpected response shape"})
        else:
            checks.append({"check": label, "status": "passed"})

    for resource, key in RESOURCES.items():
        rows = []

        def sample():
            nonlocal rows
            value = document(cli, ["get", resource, "--limit", str(sample_size)])
            require(isinstance(value, list) and len(value) <= sample_size, "invalid sample size or shape")
            require(all(isinstance(row, dict) and valid_identity(key, row.get(key), resource) for row in value), "missing or invalid sample identity")
            if key in ("uuid", "id"):
                require(len({row[key] for row in value}) == len(value), "duplicate sample identities")
            rows = value

        check(f"{resource}: sample", sample)
        if checks[-1]["status"] == "failed":
            continue
        if not rows:
            checks.append({"check": f"{resource}: dependent checks", "status": "skipped", "reason": "no sample records"})
            continue
        for fmt in ("json", "yaml", "text", "table"):
            def fields(fmt=fmt):
                selected = rows[0][key]
                output = cli.run(["get", resource, "-q", f"{key}={selected}", "--limit", "1", "--fields", key, "-o", fmt])
                validate_projection(output, fmt, key, selected, resource)
            check(f"{resource}: fields/{fmt}", fields)
        # Only stable numeric/UUID identifiers can bound subsequent pagination.
        values = [str(row[key]) for row in rows]
        safe = key == "id" and all(re.fullmatch(r"[0-9]+", value) for value in values)
        safe = safe or key == "uuid" and all(re.fullmatch(r"[0-9a-fA-F]{32}", value) for value in values)
        safe = safe or resource in ("alarms", "tags") and all(valid_identity(key, value, resource) for value in values)
        if not safe:
            checks.append({"check": f"{resource}: pagination", "status": "skipped", "reason": "no supported stable identity"})
            continue
        conditions = ["-q", key + "?=" + ",".join(values)]

        def pagination():
            for direction in ("+", "-"):
                args = ["get", resource, *conditions, "--sort", direction + key, "--fields", key]
                baseline = document(cli, [*args, "--limit", str(sample_size)])
                pages = document(cli, [*args, "--all-pages", "--limit", "1"])
                expected = [{key: value} for value in sorted([row[key] for row in rows], reverse=direction == "-")]
                require(identity_rows(baseline, key, resource) and identity_rows(pages, key, resource) and baseline == expected and pages == expected, "pagination lost, duplicated or reordered sampled identities")
                offset = document(cli, [*args, "--all-pages", "--limit", "1", "--start", "1"])
                require(identity_rows(offset, key, resource) and offset == pages[1:], "offset differs from paginated identities")
                count = document(cli, [*args, "--count"])
                require(type(count) is int and count == len(pages), "count differs from sample identities")
        check(f"{resource}: pagination", pagination)
        if resource in DETAILS:
            relation_warnings = False
            def detail():
                nonlocal relation_warnings
                for row in rows:
                    result = document(cli, ["describe", DETAILS[resource], row[key]])
                    record = result if resource == "long-jobs" else result["resource"]
                    require(record["uuid"] == row[key], "detail identity mismatch")
                    warnings = result.get("warnings", [])
                    require(isinstance(warnings, list) and all(isinstance(w, str) for w in warnings), "invalid detail warnings")
                    require(all(re.search(r"HTTP (403|404):|expected one resource .* got 0$", warning) for warning in warnings), "unexpected related-resource failure")
                    relation_warnings = relation_warnings or bool(warnings)
            check(f"{resource}: detail identity", detail)
            if relation_warnings:
                checks.append({"check": f"{resource}: related resources", "status": "skipped", "reason": "optional relations unavailable; main identity checked separately"})
    return checks


class RunInterrupted(BaseException):
    def __init__(self, signum):
        self.signum = signum


@contextmanager
def interruption_handlers():
    def interrupt(signum, frame):
        raise RunInterrupted(signum)
    previous = {sig: signal.signal(sig, interrupt) for sig in (signal.SIGINT, signal.SIGTERM)}
    try:
        yield
    finally:
        for sig, handler in previous.items():
            signal.signal(sig, handler)


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--cli", default="./zstack-cli")
    parser.add_argument("--config", required=True, help="pre-authenticated CLI config; copied to a temporary directory")
    parser.add_argument("--sample-size", type=int, choices=range(1, 4), default=3)
    parser.add_argument("--timeout", type=int, metavar="SECONDS", default=30, help="per-command deadline, 1..120 (default: 30)")
    parser.add_argument("--suite-timeout", type=int, metavar="SECONDS", default=600, help="complete CLI run deadline, 1..3600 (default: 600)")
    args = parser.parse_args(argv)
    if not 1 <= args.timeout <= 120 or not 1 <= args.suite_timeout <= 3600:
        parser.error("timeout must be 1..120 and suite-timeout must be 1..3600 seconds")
    binary = shutil.which(args.cli)
    if not binary or not Path(args.config).is_file():
        parser.error("CLI executable or input config not found")
    try:
        with interruption_handlers(), tempfile.TemporaryDirectory(prefix="zstack-readonly-") as directory:
            config = Path(directory) / "config.yaml"
            config.touch(mode=0o600)
            config.write_bytes(Path(args.config).read_bytes())
            cli = CLI(binary, config, args.timeout, args.suite_timeout)
            validate_auth(cli)
            checks = run_suite(cli, args.sample_size)
    except CheckFailure as error:
        checks = [{"check": "configuration preflight", "status": "failed", "reason": str(error)}]
    except RunInterrupted as error:
        print(json.dumps({"interrupted": True, "signal": error.signum}))
        return 128 + error.signum
    except OSError:
        parser.error("unable to read or isolate configuration")
    totals = {status: sum(c["status"] == status for c in checks) for status in ("passed", "failed", "skipped")}
    print(json.dumps({"summary": totals, "checks": checks}, indent=2))
    return 1 if totals["failed"] else 0


if __name__ == "__main__":
    raise SystemExit(main())
