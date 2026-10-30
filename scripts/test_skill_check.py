"""Contract tests for the independently installable skill helper."""
import importlib.util
import os
from pathlib import Path
import subprocess
import unittest
from unittest.mock import patch

SCRIPT = Path(__file__).resolve().parents[1] / "skills/zstack-operations/scripts/check_cli.py"
spec = importlib.util.spec_from_file_location("skill_check", SCRIPT)
helper = importlib.util.module_from_spec(spec)
spec.loader.exec_module(helper)


class SkillCheckTests(unittest.TestCase):
    def fake_run(self, args, **kwargs):
        self.assertNotIn("ZSTACK_PASSWORD", kwargs["env"])
        self.assertNotIn("ZSTACK_ACCESS_KEY_SECRET", kwargs["env"])
        self.assertEqual(Path(kwargs["env"]["ZSTACK_CONFIG"]).parent, Path(kwargs["cwd"]))
        self.assertFalse(Path(kwargs["env"]["ZSTACK_CONFIG"]).exists())
        self.assertGreater(kwargs["timeout"], 0)
        if args[1:] == ["version"]:
            return subprocess.CompletedProcess(args, 0, "zstack-cli version: dev, commit: none\n", "")
        self.assertEqual(args[-1], "--help")
        command = tuple(args[1:-1])
        flags = {**helper.BASE, **helper.PROFILES[self.profile]}[command]
        text = "Usage:\n  " + " ".join(["zstack-cli", *command]) + " [flags]\n\nFlags:\n"
        text += "\n".join("      --" + flag + " string   help" for flag in flags)
        return subprocess.CompletedProcess(args, 0, text, "")

    def invoke(self, run, profile="query"):
        self.profile = profile
        with patch.dict(os.environ, {"ZSTACK_PASSWORD": "secret", "ZSTACK_ACCESS_KEY_SECRET": "secret"}), patch.object(helper.shutil, "which", return_value="/tmp/cli with spaces"), patch.object(helper.subprocess, "run", side_effect=run):
            return helper.check("cli", profile)

    def test_profiles_use_only_isolated_help_and_version(self):
        for profile in helper.PROFILES:
            with self.subTest(profile=profile):
                self.assertTrue(self.invoke(self.fake_run, profile)["compatible"])

    def test_parent_help_does_not_pass_for_missing_child(self):
        def run(args, **kwargs):
            result = self.fake_run(args, **kwargs)
            if args[1:-1] == ["describe", "instance"]:
                result.stdout = result.stdout.replace("zstack-cli describe instance", "zstack-cli describe")
            return result
        self.assertFalse(self.invoke(run)["compatible"])

    def test_flag_prefix_does_not_satisfy_required_flag(self):
        def run(args, **kwargs):
            result = self.fake_run(args, **kwargs)
            result.stdout = result.stdout.replace("--error-format ", "--error-format-extra ")
            return result
        self.assertFalse(self.invoke(run)["compatible"])

    def test_timeout_and_failed_version_are_sanitized(self):
        for failure in (subprocess.TimeoutExpired("secret", 1, output="secret"), OSError("secret")):
            with self.subTest(failure=type(failure).__name__):
                report = self.invoke(lambda *args, **kwargs: (_ for _ in ()).throw(failure))
                self.assertFalse(report["compatible"])
                self.assertNotIn("secret", str(report))
        report = self.invoke(lambda *args, **kwargs: subprocess.CompletedProcess([], 1, "secret", "secret"))
        self.assertFalse(report["compatible"])
        self.assertNotIn("secret", str(report))

    def test_missing_binary_never_executes(self):
        with patch.object(helper.shutil, "which", return_value=None), patch.object(helper.subprocess, "run") as run:
            self.assertFalse(helper.check("absent", "query")["compatible"])
            run.assert_not_called()
