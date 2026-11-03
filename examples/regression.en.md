# Regression runners

[English](regression.en.md) | [简体中文](regression.md) | [Examples](README.en.md)

Run from the repository root. session.yaml is a dedicated authenticated configuration and should stay out of version control. Account login must use save-password=false. The runners use that file's current context. Confirm the endpoint first.

```sh
make build
python3 scripts/readonly_smoke.py --cli ./zstack-cli --config ./session.yaml > readonly-report.json
make verify
```

Make verify runs local checks. The following write profiles create and delete test resources, requiring an authorized test environment. Use a new journal for each creation run.

```sh
python3 scripts/write_e2e.py --cli ./zstack-cli --config ./session.yaml --journal ./metadata-run.json --execute
python3 scripts/write_e2e.py --cli ./zstack-cli --config ./session.yaml --journal ./compute-run.json --execute --profile compute --image ROOT_IMAGE_UUID --l3-network L3_UUID
```

After failure, preserve evidence and wait for asynchronous operations to settle, then clean the same journal.

```sh
python3 scripts/write_e2e.py --cli ./zstack-cli --config ./session.yaml --journal ./compute-run.json --execute --cleanup-only
```

Use one process per journal. Ownership mismatches, unknown creation outcomes and root-volume leftovers remain unresolved. Preserve journal states for recovery. The script retains the input session. Log out afterward only when the verification owns that session.
