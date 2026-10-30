#!/usr/bin/env python3
"""Check installed CLI help capabilities only; never authenticate or call a cloud."""
import argparse
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tempfile

BASE = {
    (): ("context", "error-format"),
    ("config", "view"): ("output",),
    ("doctor",): ("offline", "request-timeout"),
    ("get", "instances"): ("all-pages", "fields", "reply-with-count", "request-timeout"),
    ("describe", "instance"): ("request-timeout",),
}
PROFILES = {
    "query": {},
    "lifecycle": {
        ("wait", "long-job"): ("for", "timeout"),
        ("volume", "resize"): ("size", "dry-run"),
        ("instance", "recover"): ("dry-run",),
        ("volume", "recover"): ("dry-run",),
        ("job", "resume"): ("yes", "request-timeout"),
        ("job", "clean"): ("yes",),
    },
    "iam": {
        ("create", "account"): ("password-file", "resource-uuid", "dry-run"),
        ("create", "policy"): ("statements-file",),
        ("create", "project"): ("organization",),
        ("update", "project"): ("description",),
        ("delete", "project"): ("yes", "delete-mode"),
        ("project", "recover"): ("yes",),
        ("expunge", "project"): ("yes",),
    },
}


def check(binary, profile, timeout=10):
    report = {"compatible": False, "profile": profile, "version": None, "checks": [],
              "scope": "Local help capabilities only; not authentication or API compatibility."}
    resolved = shutil.which(binary)
    if not resolved:
        report["error"] = "CLI executable unavailable"
        return report
    resolved = str(Path(resolved).resolve())
    with tempfile.TemporaryDirectory(prefix="zstack-skill-check-") as directory:
        env = {k: v for k, v in os.environ.items() if not k.startswith("ZSTACK_")}
        env["ZSTACK_CONFIG"] = str(Path(directory) / "config.yaml")

        def run(args):
            return subprocess.run([resolved, *args], cwd=directory, env=env,
                                  capture_output=True, text=True, timeout=timeout, check=False)

        try:
            result = run(["version"])
            match = re.fullmatch(r"zstack-cli version: ([A-Za-z0-9._+\-]+), commit: ([A-Za-z0-9._+\-]+)\s*", result.stdout)
            if result.returncode != 0 or not match:
                report["error"] = "CLI version could not be identified"
                return report
            report["version"] = {"version": match[1], "commit": match[2]}
            for command, flags in {**BASE, **PROFILES[profile]}.items():
                result = run([*command, "--help"])
                path = " ".join(["zstack-cli", *command])
                usage = re.search(r"Usage:\n(.*?)(?:\n\n|\Z)", result.stdout, re.S)
                identified = usage is not None and re.search(r"^\s+" + re.escape(path) + r"(?:\s|$)", usage[1], re.M)
                supported = set(re.findall(r"^\s+(?:-\w,\s+)?--([a-z][a-z0-9-]*)(?=\s|$)", result.stdout, re.M))
                missing = ["--" + flag for flag in flags if flag not in supported]
                report["checks"].append({"command": path, "passed": bool(result.returncode == 0 and identified and not missing),
                                         "missingFlags": missing})
        except (OSError, UnicodeError, subprocess.TimeoutExpired):
            report["error"] = "CLI help check failed or timed out"
            return report
    report["compatible"] = all(item["passed"] for item in report["checks"])
    return report


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--cli", default="zstack-cli", help="Installed CLI binary or path")
    parser.add_argument("--profile", choices=tuple(PROFILES), default="query")
    parser.add_argument("--timeout", type=int, default=10, help="Per-process timeout in seconds (1..30)")
    args = parser.parse_args()
    if not 1 <= args.timeout <= 30:
        parser.error("timeout must be 1..30 seconds")
    report = check(args.cli, args.profile, args.timeout)
    print(json.dumps(report, ensure_ascii=False, indent=2))
    return 0 if report["compatible"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
