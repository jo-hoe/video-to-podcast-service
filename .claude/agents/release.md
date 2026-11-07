---
name: release
description: Release a new version of video-to-podcast-service. Supports two release types — full (new image + chart) and chart-only (chart bump, existing image). Run in foreground (run_in_background: false) so step progress is visible in real time.
allowedTools:
  - Read
  - Edit
  - Bash(git *)
  - Bash(gh *)
  - Bash(helm *)
  - Bash(grep *)
  - Bash(go test *)
  - Bash(docker *)
---

## Release process for video-to-podcast-service

Follow these steps in order. Do not skip steps. After each step, report completion with a one-line status so the user can track progress.

### Step 1 — Determine version and release type

Report: `[Step 1/6] Determining version and release type...`

Check the current versions and what has changed since the last tag:
```bash
grep -E '^version:|^appVersion:' charts/video-to-podcast-service/Chart.yaml
git tag --sort=-v:refname | head -3
git log $(git tag --sort=-v:refname | head -1)..HEAD --oneline
git diff $(git tag --sort=-v:refname | head -1)..HEAD -- charts/video-to-podcast-service/ internal/ main.go go.mod go.sum Dockerfile
```

**Determine release type from the diff:**
- **No release needed** — only non-app, non-chart files changed (e.g. `.claude/`, docs, CI config). Stop here and report: `[Step 1/6] No release needed — no app or chart changes since last tag.`
- **Full release** — app code changed (`internal/`, `main.go`, `go.mod`, `go.sum`, `Dockerfile`). Bumps both `version` and `appVersion`. Pushes a new semver tag to trigger `image.yml`.
- **Chart-only release** — only chart templates/values changed (`charts/`), no app code changes. Bumps only `version`, keeps `appVersion`. No new tag pushed.

Report: `[Step 1/6] ✓ Release type: <full|chart-only>, chart version: <new-version>, appVersion: <app-version>`

### Step 2 — Bump Chart.yaml

Report: `[Step 2/6] Bumping Chart.yaml...`

**Full release:** update both fields:
```yaml
version: <new-version>
appVersion: "<new-version>"
```

**Chart-only release:** update only `version`, leave `appVersion` unchanged:
```yaml
version: <new-version>
appVersion: "<current-app-version>"   # unchanged
```

Report: `[Step 2/6] ✓ Chart.yaml updated`

### Step 3 — Run integration tests locally

Report: `[Step 3/6] Running integration tests...`

Run the full test suite via Docker (includes CGO, ffmpeg, yt-dlp, and bgutil sidecar):
```bash
docker compose -f docker-compose.test.yml up --build --abort-on-container-exit
```

Check the exit code of the `test` container. If any test fails, report the failure and stop. Do not proceed to commit or push with failing tests.

Clean up after:
```bash
docker compose -f docker-compose.test.yml down
```

Report: `[Step 3/6] ✓ All tests passed`

### Step 4 — Commit, push, and (for full releases) tag

Report: `[Step 4/6] Committing and pushing...`

**First, check for uncommitted code changes and commit them if present:**
```bash
git status --short
```

If there are any modified or untracked files outside of `charts/` (i.e. app code, tests, config), stage and commit them before the chart bump commit:
```bash
git add <each modified or untracked file>
git commit -m "feat: <short summary of the changes>"
```

Use `git diff --cached` and `git log --oneline -5` to write an accurate commit message reflecting what actually changed.

Then commit the chart bump:
```bash
git add charts/video-to-podcast-service/Chart.yaml
git commit -m "chore: bump chart and appVersion to <new-version>"
git push origin main
```

**Full release only** — also push the semver tag to trigger image build:
```bash
git tag v<new-version>
git push origin v<new-version>
```

If push fails due to remote changes, rebase first:
```bash
git fetch origin && git rebase origin/main
```
Then re-push (and re-tag if needed).

Report: `[Step 4/6] ✓ Pushed main` (and `+ tag v<new-version>` for full releases)

### Step 5 — Babysit CI

Report: `[Step 5/6] Waiting for CI (timeout: 10 minutes)...`

Poll every 30 seconds, up to 20 times. On each poll:
```bash
gh run list --repo jo-hoe/video-to-podcast-service --limit 8
```

**Full release** — track all three:
- `test` — triggered by main push
- `Release Image` — triggered by the semver tag
- `Release Charts` — triggered by Chart.yaml change on main

**Chart-only release** — track only two:
- `test` — triggered by main push
- `Release Charts` — triggered by Chart.yaml change on main

Report each poll as: `[Step 5/6] Poll <n>/20 — test: <status>, image: <status|n/a>, chart: <status>`

Stop as soon as all tracked workflows show `completed`. If any shows `failure`, fetch logs immediately:
```bash
gh run view <id> --log-failed
```
Then report the failure and stop.

If 20 polls pass without completion: `[Step 5/6] ✗ Timeout after 10 minutes` and stop.

Report: `[Step 5/6] ✓ All workflows completed successfully`

### Step 6 — Verify and confirm

Report: `[Step 6/6] Verifying published artifacts...`

Always verify the chart:
```bash
helm show chart oci://ghcr.io/jo-hoe/charts/video-to-podcast-service --version <new-version>
```

For full releases also verify the image tag via the workflow success (image API requires read:packages scope which may not be available).

If chart verification fails, report the error and stop.

Report: `[Step 6/6] ✓ Release complete`

Confirm:
- Chart: `oci://ghcr.io/jo-hoe/charts/video-to-podcast-service --version <new-version>`
- Image: `ghcr.io/jo-hoe/video-to-podcast-service:v<app-version>` (full release) or `unchanged at v<current-app-version>` (chart-only)
