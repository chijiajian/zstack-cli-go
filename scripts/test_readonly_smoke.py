import contextlib
import io
import json
import os
import signal
import sys
import time
import subprocess
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import readonly_smoke as smoke


class FakeCLI:
    def __init__(self, empty=False):
        self.empty = empty
        self.calls = []

    def run(self, args):
        self.calls.append(args)
        if args[0] == "describe":
            record = {"uuid": args[2]}
            return json.dumps(record if args[1] == "job" else {"resource": record})
        resource = args[1]
        key = smoke.RESOURCES[resource]
        rows = [{key: i if key == "id" else f"{i:032x}"} for i in (1, 2)]
        if self.empty:
            rows = []
        if "--sort" in args and args[args.index("--sort") + 1].startswith("-"):
            rows.reverse()
        if "--count" in args:
            return json.dumps(len(rows))
        if "--start" in args:
            rows = rows[int(args[args.index("--start") + 1]):]
        if "--all-pages" not in args:
            rows = rows[:int(args[args.index("--limit") + 1])]
        fmt = args[args.index("-o") + 1]
        if fmt == "json":
            return json.dumps(rows)
        value = rows[0][key]
        if fmt == "yaml":
            return f"- {key}: {value}\n"
        if fmt == "text":
            return f"[map[{key}:{value}]]\n"
        return f"| {key.upper()} |\n| {value} |\n"


class SmokeTests(unittest.TestCase):
    def test_success_and_readonly_commands(self):
        cli = FakeCLI()
        checks = smoke.run_suite(cli)
        self.assertFalse([c for c in checks if c["status"] == "failed"])
        self.assertTrue(all(args[0] in ("get", "describe") for args in cli.calls))
        for args in cli.calls:
            if "--all-pages" in args:
                self.assertIn("-q", args)
                self.assertEqual(args[args.index("--limit") + 1], "1")
        self.assertEqual(sum(c["status"] == "skipped" for c in checks), 0)

    def test_builtin_alarm_ids_are_opaque_but_query_safe(self):
        self.assertTrue(smoke.valid_identity("uuid", "33198a88f22e4d19b5ff8ebaebb6ujm7", "alarms"))
        self.assertTrue(smoke.valid_identity("uuid", "065f9609dce141bb952c80f729f58af4", "alarms"))
        self.assertTrue(smoke.valid_identity("uuid", "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6", "tags"))
        self.assertFalse(smoke.valid_identity("uuid", "nonhex", "zones"))
        for value in ["x,y", "x?=y", "", " ", None, "a" * 129]:
            self.assertFalse(smoke.valid_identity("uuid", value, "alarms"))

    def test_full_query_coverage_and_global_config_identity(self):
        expected = set("zones clusters hosts instances volumes images primary-storages long-jobs cdroms disk-offerings instance-offerings image-storages image-groups l2-networks l3-networks ip-ranges management-nodes nics snapshots tags vips eips virtual-router-offerings virtual-routers vm-scripts vm-script-records accounts users roles policies projects organizations alarms security-groups security-group-rules port-forwarding-rules load-balancers load-balancer-listeners load-balancer-server-groups events audits global-configs".split())
        self.assertEqual(set(smoke.RESOURCES), expected)
        self.assertEqual(smoke.RESOURCES["global-configs"], "id")
        checks = smoke.run_suite(FakeCLI())
        self.assertIn({"check": "global-configs: pagination", "status": "passed"}, checks)

    def test_pagination_mismatch_fails_and_continues(self):
        class BrokenPaging(FakeCLI):
            def run(self, args):
                if args[1] == "zones" and "--all-pages" in args:
                    return "[]"
                return super().run(args)
        checks = smoke.run_suite(BrokenPaging())
        failed = [c["check"] for c in checks if c["status"] == "failed"]
        self.assertEqual(failed, ["zones: pagination"])
        self.assertTrue(any(c["check"] == "audits: pagination" and c["status"] == "passed" for c in checks))

    def test_consistently_missing_pages_fail(self):
        class Missing(FakeCLI):
            def run(self, args):
                if args[1] == "zones" and "-q" in args:
                    return "0" if "--count" in args else "[]"
                return super().run(args)
        checks = smoke.run_suite(Missing())
        self.assertTrue(any(c["check"] == "zones: pagination" and c["status"] == "failed" for c in checks))

    def test_ignored_sort_fails(self):
        class Unsorted(FakeCLI):
            def run(self, args):
                return super().run(["+uuid" if arg == "-uuid" else arg for arg in args])
        checks = smoke.run_suite(Unsorted())
        self.assertTrue(any(c["check"] == "zones: pagination" and c["status"] == "failed" for c in checks))

    def test_non_json_garbage_fails(self):
        class Garbage(FakeCLI):
            def run(self, args):
                if "-o" in args and args[args.index("-o") + 1] != "json":
                    return "unexpected error text"
                return super().run(args)
        checks = smoke.run_suite(Garbage())
        self.assertTrue(any(c["check"] == "zones: fields/table" and c["status"] == "failed" for c in checks))

    def test_empty_environment_is_explicitly_skipped(self):
        cli = FakeCLI(empty=True)
        checks = smoke.run_suite(cli)
        self.assertEqual(len(cli.calls), len(smoke.RESOURCES))
        self.assertEqual(sum(c["status"] == "skipped" for c in checks), len(smoke.RESOURCES))

    def test_bad_response_is_failure_not_skip(self):
        class BadCLI:
            def run(self, args):
                return '{"not":"a list"}'
        checks = smoke.run_suite(BadCLI())
        self.assertTrue(all(c["status"] == "failed" for c in checks))

    def test_malformed_identity_is_failure(self):
        class BadCLI:
            def run(self, args):
                return '[{"uuid": null, "id": true, "name": ""}]'
        checks = smoke.run_suite(BadCLI())
        self.assertTrue(all(c["status"] == "failed" for c in checks))

    def test_mutations_rejected_before_subprocess(self):
        with patch.object(smoke.subprocess, "run") as process:
            for args in (["login", "server"], ["volume", "resize", "id"], ["get", "unknown"]):
                with self.assertRaises(smoke.CheckFailure):
                    smoke.CLI("cli", "config").run(args)
            process.assert_not_called()

    def test_cli_failures_do_not_leak_output(self):
        result = smoke.subprocess.CompletedProcess([], 1, "secret", "secret")
        with patch.object(smoke.subprocess, "run", return_value=result):
            with self.assertRaisesRegex(smoke.CheckFailure, "status 1") as error:
                smoke.CLI("cli", "config").run(["get", "zones"])
            self.assertNotIn("secret", str(error.exception))

    def test_success_exit_with_query_error_text_is_failure(self):
        result = subprocess.CompletedProcess([], 0, "[]", "secret error text")
        with patch.object(smoke.subprocess, "run", return_value=result):
            with self.assertRaisesRegex(smoke.CheckFailure, "unexpected diagnostic"):
                smoke.CLI("cli", "config").run(["get", "zones"])

    def test_duplicate_identities_and_boolean_count_fail(self):
        class Duplicate(FakeCLI):
            def run(self, args):
                if args[1] == "zones":
                    return json.dumps([{"uuid": "1" * 32}] * 2)
                return super().run(args)
        self.assertTrue(any(c["check"] == "zones: sample" and c["status"] == "failed" for c in smoke.run_suite(Duplicate())))
        class BooleanCount(FakeCLI):
            def run(self, args):
                if "--count" in args:
                    return "true"
                return super().run(args)
        self.assertTrue(any(c["check"] == "zones: pagination" and c["status"] == "failed" for c in smoke.run_suite(BooleanCount())))

    def test_boolean_id_cannot_match_integer_projection(self):
        with self.assertRaises(smoke.CheckFailure):
            smoke.validate_projection('[{"id": true}]', "json", "id", 1)

    def test_cli_unicode_table_borders_are_valid(self):
        value = "0123456789abcdef0123456789abcdef"
        table = f"┌──────────────────────────────────┐\n│               UUID               │\n├──────────────────────────────────┤\n│ {value} │\n└──────────────────────────────────┘\n"
        smoke.validate_projection(table, "table", "uuid", value)
        with self.assertRaises(smoke.CheckFailure):
            smoke.validate_projection(table + "unexpected", "table", "uuid", value)

    def test_invalid_deadlines_rejected_before_configuration(self):
        for flag, value in [("--timeout", "0"), ("--timeout", "121"), ("--suite-timeout", "-1"), ("--suite-timeout", "3601")]:
            with contextlib.redirect_stderr(io.StringIO()), self.assertRaises(SystemExit) as error:
                smoke.main(["--config", "absent", flag, value])
            self.assertEqual(error.exception.code, 2)

    def test_broken_relation_is_not_hidden_as_permission_skip(self):
        class BrokenRelation(FakeCLI):
            def run(self, args):
                output = super().run(args)
                if args[:2] == ["describe", "instance"]:
                    data = json.loads(output)
                    data["warnings"] = ["host unavailable: resource identity mismatch"]
                    return json.dumps(data)
                return output
        checks = smoke.run_suite(BrokenRelation())
        self.assertTrue(any(c["check"] == "instances: detail identity" and c["status"] == "failed" for c in checks))

    def test_saved_password_rejected_before_queries(self):
        config = {"selected-context": "test", "contexts": {"test": {"auth_type": "account", "password": "*****", "session_uuid": "*****"}}}
        with patch.object(smoke, "document", return_value=config):
            with self.assertRaisesRegex(smoke.CheckFailure, "save-password=false"):
                smoke.validate_auth(None)
            config["contexts"]["test"]["password"] = ""
            smoke.validate_auth(None)

    def test_related_warnings_do_not_fail_main_identity(self):
        class Warnings(FakeCLI):
            def run(self, args):
                result = super().run(args)
                if args[:2] == ["describe", "instance"]:
                    data = json.loads(result)
                    data["warnings"] = ["host unavailable: GET host: HTTP 403: denied"]
                    return json.dumps(data)
                return result
        checks = smoke.run_suite(Warnings())
        self.assertFalse([c for c in checks if c["status"] == "failed"])
        self.assertTrue(any(c["check"] == "instances: related resources" and c["status"] == "skipped" for c in checks))

    def test_expired_suite_never_launches_more_processes(self):
        cli = smoke.CLI("cli", "config")
        cli.deadline = 0
        with patch.object(smoke.subprocess, "run") as process:
            with self.assertRaisesRegex(smoke.CheckFailure, "suite deadline"):
                cli.run(["get", "zones"])
            process.assert_not_called()

    def test_inherited_credentials_do_not_override_config(self):
        result = subprocess.CompletedProcess([], 0, "[]", "")
        with patch.dict(os.environ, {"ZSTACK_CONFIG": "original", "ZSTACK_ACCESS_KEY_SECRET": "secret"}), patch.object(smoke.subprocess, "run", return_value=result) as process:
            smoke.CLI("cli", "isolated").run(["get", "zones"])
            env = process.call_args.kwargs["env"]
            self.assertEqual(env["ZSTACK_CONFIG"], "isolated")
            self.assertNotIn("ZSTACK_ACCESS_KEY_SECRET", env)

    @unittest.skipUnless(os.name == "posix", "POSIX executable fixture")
    def test_timeout_kills_and_reaps_child(self):
        with tempfile.TemporaryDirectory() as directory:
            child = Path(directory) / "child"
            pid_file = Path(directory) / "pid"
            child.write_text(f"#!{sys.executable}\nimport os,time\nopen({str(pid_file)!r},'w').write(str(os.getpid()))\ntime.sleep(30)\n")
            child.chmod(0o700)
            # Leave time for a cold interpreter to start on a busy CI runner.
            # The child still sleeps 30s, so this must exercise timeout cleanup.
            cli = smoke.CLI(str(child), "isolated", timeout=1, suite_timeout=5)
            with self.assertRaises(smoke.CheckFailure):
                cli.run(["get", "zones"])
            self.assertTrue(pid_file.exists())
            with self.assertRaises(ProcessLookupError):
                os.kill(int(pid_file.read_text()), 0)

    @unittest.skipUnless(os.name == "posix", "POSIX signal semantics")
    def test_sigterm_cleans_config_and_child(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / "source.yaml"
            source.write_text("original")
            marker = root / "started"
            child = root / "cli"
            config_view = {"selected-context": "test", "contexts": {"test": {"auth_type": "account", "password": "", "session_uuid": "*****"}}}
            child.write_text(f"#!{sys.executable}\nimport sys,os,time,json\nif sys.argv[1]=='config':\n print({json.dumps(config_view)!r})\nelse:\n open({str(marker)!r},'w').write(json.dumps([os.environ['ZSTACK_CONFIG'],os.getpid()]))\n time.sleep(30)\n")
            child.chmod(0o700)
            process = subprocess.Popen([sys.executable, str(Path(smoke.__file__)), "--cli", str(child), "--config", str(source)], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            try:
                deadline = time.monotonic() + 5
                while not marker.exists() and time.monotonic() < deadline and process.poll() is None:
                    time.sleep(0.01)
                self.assertTrue(marker.exists())
                config_path, child_pid = json.loads(marker.read_text())
                process.send_signal(signal.SIGTERM)
                output, errors = process.communicate(timeout=5)
                self.assertEqual(process.returncode, 128 + signal.SIGTERM, errors)
                self.assertTrue(json.loads(output)["interrupted"])
                self.assertFalse(Path(config_path).exists())
                self.assertEqual(source.read_text(), "original")
                with self.assertRaises(ProcessLookupError):
                    os.kill(child_pid, 0)
            finally:
                if process.poll() is None:
                    process.kill()
                process.communicate()

    def test_main_isolates_and_removes_config(self):
        with tempfile.TemporaryDirectory() as directory:
            source = Path(directory) / "source.yaml"
            source.write_text("original")
            paths = []

            def suite(cli, sample_size):
                paths.append(cli.config)
                self.assertNotEqual(source, cli.config)
                self.assertEqual(cli.config.stat().st_mode & 0o777, 0o600)
                cli.config.write_text("refreshed")
                return [{"check": "example", "status": "failed", "reason": "test"}]

            with patch.object(smoke.shutil, "which", return_value="/fake/cli"), patch.object(smoke, "run_suite", suite), patch.object(smoke, "validate_auth"), contextlib.redirect_stdout(io.StringIO()):
                self.assertEqual(smoke.main(["--config", str(source)]), 1)
            self.assertEqual(source.read_text(), "original")
            self.assertFalse(paths[0].exists())


if __name__ == "__main__":
    unittest.main()
