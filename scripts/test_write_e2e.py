import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import write_e2e as e2e


class FakeCLI:
    def __init__(self):
        self.resources = {}
        self.writes = []
        self.fail_create = False

    def document(self, args):
        if args[0] == "get":
            identifier = args[args.index("-q") + 1].split("=", 1)[1]
            return [self.resources[identifier]] if identifier in self.resources else []
        if args[0] == "create":
            identifier = args[args.index("--resource-uuid") + 1]
            self.writes.append(args)
            self.resources[identifier] = {"uuid": identifier, "name": args[2], "state": "Enabled"}
            if self.fail_create:
                raise e2e.Failure("output failed after successful create")
            return self.resources[identifier]
        raise AssertionError(args)

    def run(self, args):
        self.writes.append(args)
        if args[0] == "update":
            self.resources[args[2]]["description"] = args[args.index("--description") + 1]
        elif args[0] == "project" and args[1] == "recover":
            self.resources[args[2]]["state"] = "Enabled"
        elif args[0] == "delete" and args[1] == "project":
            self.resources[args[2]]["state"] = "Deleted"
        elif args[0] in ("delete", "expunge"):
            self.resources.pop(args[2], None)
        else:
            raise AssertionError(args)
        return ""


class WriteE2ETests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.path = Path(self.directory.name) / "journal.json"
        self.journal = e2e.Journal(self.path, endpoint="test:8080")
        self.cli = FakeCLI()
        self.runner = e2e.Runner(self.cli, self.journal)

    def test_metadata_lifecycle_and_reverse_cleanup(self):
        e2e.run_metadata(self.runner, self.directory.name)
        names = [r["kind"] for r in self.journal.data["resources"]]
        self.assertEqual(len(names), 7)
        before_cleanup = len(self.cli.writes)
        self.assertEqual(self.runner.cleanup(), [])
        self.assertFalse(self.cli.resources)
        self.assertEqual([a[1] for a in self.cli.writes[before_cleanup:] if a[0] == "delete"], list(reversed(names)))
        self.assertTrue(all(r["state"] == "cleaned" for r in json.loads(self.path.read_text())["resources"]))
        self.assertEqual(self.path.stat().st_mode & 0o777, 0o600)

    def test_successful_create_with_broken_output_is_still_cleaned(self):
        self.cli.fail_create = True
        with self.assertRaises(e2e.Failure):
            self.runner.create("account")
        self.assertEqual(self.runner.cleanup(), [])
        self.assertFalse(self.cli.resources)

    def test_unknown_creation_is_not_claimed_clean(self):
        item = self.journal.plan("volume")
        item["state"] = "submitted"
        self.journal.save()
        errors = self.runner.cleanup()
        self.assertEqual(len(errors), 1)
        self.assertEqual(item["state"], "submitted")
        self.assertFalse(self.cli.writes)

    def test_foreign_resource_name_is_never_deleted(self):
        item = self.journal.plan("account")
        item["state"] = "submitted"
        self.cli.resources[item["uuid"]] = {"uuid": item["uuid"], "name": "existing-business-account"}
        self.assertEqual(len(self.runner.cleanup()), 1)
        self.assertFalse(self.cli.writes)

    def test_unsubmitted_intent_never_mutates(self):
        item = self.journal.plan("volume")
        self.cli.resources[item["uuid"]] = {"uuid": item["uuid"], "name": item["name"]}
        self.assertEqual(self.runner.cleanup(), [])
        self.assertFalse(self.cli.writes)

    def test_cleanup_continues_after_independent_failure(self):
        first = self.runner.create("account")
        second = self.runner.create("user")
        self.cli.resources[second["uuid"]]["name"] = "foreign"
        errors = self.runner.cleanup()
        self.assertEqual(len(errors), 1)
        self.assertNotIn(first["uuid"], self.cli.resources)
        self.assertIn(second["uuid"], self.cli.resources)

    def test_cleanup_binds_endpoint_and_validates_journal(self):
        with self.assertRaises(e2e.Failure):
            e2e.Journal(self.path, True, "other:8080")
        item = self.journal.plan("account")
        item["name"] = "unowned"
        self.journal.save()
        with self.assertRaises(e2e.Failure):
            e2e.Journal(self.path, True, "test:8080")

    def test_existing_journal_is_never_overwritten(self):
        before = self.path.read_bytes()
        with self.assertRaises(FileExistsError):
            e2e.Journal(self.path, endpoint="test:8080")
        self.assertEqual(before, self.path.read_bytes())

    def test_explicit_execute_required(self):
        with patch.object(e2e.subprocess, "run") as run:
            with self.assertRaises(SystemExit):
                e2e.main(["--config", "config", "--journal", str(self.path)])
            run.assert_not_called()

    def test_runner_errors_do_not_leak_credentials(self):
        result = e2e.subprocess.CompletedProcess([], 3, "secret", "secret")
        with patch.object(e2e.subprocess, "run", return_value=result):
            with self.assertRaises(e2e.Failure) as caught:
                e2e.CLI("cli", "config").run(["create", "account", "owned"])
            self.assertNotIn("secret", str(caught.exception))


class ComputeCLI(FakeCLI):
    def document(self, args):
        if args[0] == "describe":
            return {"resource": {"size": self.resources[args[2]]["size"]}}
        if args[:2] == ["snapshot", "create"]:
            identifier = args[args.index("--resource-uuid") + 1]
            self.resources[identifier] = {"uuid": identifier, "name": args[args.index("--name") + 1], "volumeUuid": args[2]}
            return self.resources[identifier]
        result = super().document(args)
        if args[0] == "create":
            if args[1] == "instance":
                result.update(state="Stopped", rootVolumeUuid="a" * 32)
            if args[1] == "volume":
                result.update(status="Ready", size=1024 ** 3)
        return result

    def run(self, args):
        if args[0] in ("volume", "instance", "snapshot"):
            self.writes.append(args)
            record = self.resources[args[2]]
            if args[1] == "attach":
                record["vmInstanceUuid"] = args[args.index("--instance") + 1]
            elif args[1] == "detach":
                record["vmInstanceUuid"] = ""
            elif args[1] == "resize":
                record["size"] = 2 * 1024 ** 3
            elif args[1] == "recover":
                record.update(state="Stopped", status="Ready")
            return ""
        if args[0] == "delete" and args[1] in ("volume", "instance"):
            self.writes.append(args)
            self.resources[args[2]].update(state="Destroyed", status="Deleted")
            return ""
        return super().run(args)


class ComputeE2ETests(unittest.TestCase):
    def test_full_compute_flow_and_cleanup(self):
        with tempfile.TemporaryDirectory() as directory:
            journal = e2e.Journal(Path(directory) / "journal.json", endpoint="test")
            cli = ComputeCLI()
            runner = e2e.Runner(cli, journal)
            e2e.run_compute(runner, "b" * 32, "c" * 32)
            self.assertIn("instance-recover", journal.data["compute_checks"])
            self.assertEqual(runner.cleanup(), [])
            self.assertFalse(cli.resources)

    def test_failed_flow_detaches_only_from_owned_instance(self):
        with tempfile.TemporaryDirectory() as directory:
            journal = e2e.Journal(Path(directory) / "journal.json", endpoint="test")
            cli = ComputeCLI()
            runner = e2e.Runner(cli, journal)
            vm = runner.create("instance")
            volume = runner.create("volume")
            cli.resources[volume["uuid"]]["vmInstanceUuid"] = vm["uuid"]
            self.assertEqual(runner.cleanup(), [])
            self.assertIn(["volume", "detach", volume["uuid"]], cli.writes)

    def test_foreign_attachment_is_not_detached(self):
        with tempfile.TemporaryDirectory() as directory:
            journal = e2e.Journal(Path(directory) / "journal.json", endpoint="test")
            cli = ComputeCLI()
            runner = e2e.Runner(cli, journal)
            volume = runner.create("volume")
            cli.resources[volume["uuid"]]["vmInstanceUuid"] = "foreign-vm"
            errors = runner.cleanup()
            self.assertEqual(len(errors), 1)
            self.assertFalse([a for a in cli.writes if a[0] in ("volume", "delete", "expunge")])


class CleanupBoundaryTests(unittest.TestCase):
    def test_missing_vm_does_not_hide_orphaned_root_volume(self):
        with tempfile.TemporaryDirectory() as directory:
            journal = e2e.Journal(Path(directory) / "journal.json", endpoint="test")
            cli = ComputeCLI()
            runner = e2e.Runner(cli, journal)
            vm = runner.create("instance")
            vm["root_volume"] = "a" * 32
            del cli.resources[vm["uuid"]]
            cli.resources["a" * 32] = {"uuid": "a" * 32, "name": "root"}
            errors = runner.cleanup()
            self.assertEqual(len(errors), 1)
            self.assertNotEqual(vm["state"], "cleaned")

    def test_malformed_record_does_not_authorize_mutation(self):
        with tempfile.TemporaryDirectory() as directory:
            journal = e2e.Journal(Path(directory) / "journal.json", endpoint="test")
            item = journal.plan("account")
            item["state"] = "submitted"
            cli = FakeCLI()
            cli.document = lambda args: [None]
            self.assertEqual(len(e2e.Runner(cli, journal).cleanup()), 1)
            self.assertFalse(cli.writes)

    def test_license_rejection_can_be_verified_absent(self):
        with tempfile.TemporaryDirectory() as directory:
            journal = e2e.Journal(Path(directory) / "journal.json", endpoint="test")
            cli = FakeCLI()
            original = cli.document
            def document(args):
                if args[0] == "create":
                    raise e2e.Failure("license", rejected=True)
                return original(args)
            cli.document = document
            runner = e2e.Runner(cli, journal)
            with self.assertRaises(e2e.Failure):
                runner.create("project")
            self.assertEqual(journal.data["resources"][0]["state"], "rejected")
            self.assertEqual(runner.cleanup(), [])


class QARecoveryTests(unittest.TestCase):
    def test_recovery_requires_observed_state_transition(self):
        for kind in ("project", "volume", "instance"):
            for fault in ("delete_noop", "recover_noop", "missing_state"):
                with self.subTest(kind=kind, fault=fault), tempfile.TemporaryDirectory() as directory:
                    journal = e2e.Journal(Path(directory) / "journal.json", endpoint="test")
                    cli = ComputeCLI()
                    runner = e2e.Runner(cli, journal)
                    item = runner.create(kind)
                    original = cli.run
                    def run(args):
                        if fault == "delete_noop" and args[0] == "delete":
                            return ""
                        if args[:2] == [kind, "recover"]:
                            if fault == "recover_noop":
                                return ""
                            if fault == "missing_state":
                                cli.resources[item["uuid"]].pop("status" if kind == "volume" else "state", None)
                                return ""
                        return original(args)
                    cli.run = run
                    with self.assertRaises(e2e.Failure):
                        runner.recover(item)
                    self.assertNotIn("recovery_verified", item)

    def test_failed_create_discovers_root_before_cleanup(self):
        with tempfile.TemporaryDirectory() as directory:
            journal = e2e.Journal(Path(directory) / "journal.json", endpoint="test")
            cli = ComputeCLI()
            runner = e2e.Runner(cli, journal)
            item = journal.plan("instance")
            item["state"] = "submitted"
            root = "a" * 32
            cli.resources[item["uuid"]] = {"uuid": item["uuid"], "name": item["name"], "state": "Stopped", "rootVolumeUuid": root}
            cli.resources[root] = {"uuid": root, "name": "root"}
            errors = runner.cleanup()
            self.assertEqual(len(errors), 1)
            self.assertNotEqual(item["state"], "cleaned")
            self.assertEqual(json.loads(journal.path.read_text())["resources"][0]["root_volume"], root)

    def test_malformed_diagnostic_is_a_sanitized_failure(self):
        for diagnostic in (None, [], "secret", {"apiCode": []}, {"apiCode": 42}):
            result = e2e.subprocess.CompletedProcess([], 8, "secret", json.dumps({"error": diagnostic}))
            with self.subTest(diagnostic=diagnostic), patch.object(e2e.subprocess, "run", return_value=result):
                with self.assertRaises(e2e.Failure) as caught:
                    e2e.CLI("cli", "config").run(["delete", "account", "owned"])
                self.assertNotIn("secret", str(caught.exception))
