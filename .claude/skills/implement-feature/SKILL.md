---
name: implement-feature
description: Implement a feature in one RaaS service on a new feat/ branch, test and redeploy it locally, then commit, push and open a pull request after the user approves. Use when the user asks to build a feature end to end.
argument-hint: <feature_name> <service_name>
disable-model-invocation: true
---

# Implement feature

Arguments: `$0` = feature name (used for the branch and PR), `$1` = service directory (`listing`, `media`, `booking`, `payment`, `review`, `favorites`, `notification`, `user`, `analytics`, `gateway` or `ui`). If either is missing, ask for it.

Run the steps in order and stop at the first failure. The git, docker and kubectl commands are exact; do not alter them.

## 1. Branch

The working tree must be clean and the branch comes from an up-to-date `main`:

```sh
test -z "$(git status --porcelain)" || { echo "Working tree is not clean. Commit or stash changes first."; exit 1; }
git fetch origin && git checkout main && git pull origin main
git checkout -b "feat/$0"
```

## 2. Implement

Implement "$0" in `$1`. Read `agent.md` first and follow the guideline file it names for that service.

## 3. Format and test

Format only the service you changed, then run its tests:

| Service has | Format | Test |
|---|---|---|
| `go.mod` | `gofmt -w .` | `go test ./...` |
| `pom.xml` | - | `./mvnw test` |
| `requirements.txt` | - | `pytest` |
| `ui` | - | `npm run test -- --watch=false` |

Run each from inside `$1`. Fix failures before continuing.

## 4. Redeploy locally

```sh
powershell -File scripts/redeploy.ps1 -Service $1
```

The script builds `raas/<service>:latest` and restarts its deployment (`<service>-deployment`, gateway: `gateway`).

## 5. Ask for approval

Summarize what changed and the test and rollout results, then wait for the user to approve. Do not commit, push or open a PR without that approval.

## 6. Commit, push, PR

After approval:

1. `git add .`, then check `git status` for files that should not be committed.
2. Commit with a concise conventional message: `type(scope): short description`.
3. `git push -u origin HEAD`.
4. Open the PR with a structured body (Summary, What was changed, How to test, Risks or notes):

```sh
gh pr create --title "feat: $0" --body "<description>" --base main --head "feat/$0" --assignee @me
```

Report the PR URL when done.
