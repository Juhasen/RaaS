---
name: deliver-change
description: Deliver a change to RaaS end to end in an isolated git worktree - write a spec under .planning, get it attacked by an adversarial review, build it test-first, redeploy locally, then open a PR after the user approves. Covers service features, bug fixes and infrastructure work (cluster additions such as monitoring or storage). Use whenever the user asks to build, add, change or ship something in this repo, even if they do not say "deliver".
argument-hint: <change-name> <service-name|infrastructure>
disable-model-invocation: true
---

# Deliver change

Arguments: `$0` = change name (kebab-case; used for the branch, the spec folder and the PR), `$1` = target: a service directory (`listing`, `media`, `booking`, `payment`, `review`, `favorites`, `notification`, `user`, `analytics`, `gateway`, `ui`) or `infrastructure` for cluster-level work. If either is missing, ask for it.

Run the steps in order and stop at the first failure. The git, docker and kubectl commands are exact; do not alter them.

Why this shape: the spec is written and attacked before any code exists, because a wrong spec is the most expensive bug. The work happens in a worktree so the user's main checkout stays untouched while this runs.

Branch prefix: `feat/` for a new capability, `fix/` for a bug, `infra/` when `$1` is `infrastructure`. Call it `<prefix>/$0` below.

## 1. Worktree

Start from an up-to-date `main` in a fresh worktree next to the repo. Do not touch the main checkout's files.

```sh
git fetch origin
git worktree add "../RaaS-$0" -b "<prefix>/$0" origin/main
cd "../RaaS-$0"
```

Do all remaining work inside that worktree. If the branch or folder already exists, stop and ask; it may be someone's unfinished work.

## 2. Spec

Planning docs live in `.planning/`:

- `.planning/services/<service>/$0/spec.md` for a service
- `.planning/infrastructure/$0/spec.md` for cluster work

Create the folder if missing. `.planning/proposals/` holds unscheduled ideas; if a proposal covers this change, read it first and use it as input.

Read `CLAUDE.md` and the guideline file it names for the target, plus `GLOSSARY.md` and `docs/adr/` if they exist. Then call the Skill tool with `grilling` to close the open questions with the user, one at a time, before writing.

Write `spec.md` with the `to-spec` template: Problem Statement, Solution, User Stories, Implementation Decisions, Testing Decisions, Out of Scope, Further Notes. Do not publish it to the issue tracker; the file in `.planning` is the spec. Follow `to-spec` for test seams: prefer existing seams, as few as possible, and confirm them with the user.

For infrastructure the "user" is whoever runs or observes the cluster. Stories read "As an operator, I want ..., so that ...". Testing Decisions name the check that proves it works (pods Ready, a metric scraped, a bucket reachable), not a unit test.

## 3. Adversarial review

Spawn a fresh subagent that has not seen the conversation. Give it only the spec path and the repo, and ask it to attack the spec: missing acceptance criteria, contradictions with existing code or ADRs, unstated assumptions, failure modes, scope creep, anything untestable. It reports findings; it does not edit.

Fix the spec for every valid finding and note rejected ones with a reason in Further Notes. Then show the user the spec and wait for approval. Do not write code before the spec is approved.

## 4. Implement

- **Service:** call the Skill tool with `tdd` and build at the seams the spec names. Format and test only the target service, from inside its directory:

| Service has | Format | Test |
|---|---|---|
| `go.mod` | `gofmt -w .` | `go test ./...` |
| `pom.xml` | - | `./mvnw test` |
| `requirements.txt` | - | `pytest` |
| `ui` | - | `npm run test -- --watch=false` |

- **Infrastructure:** write manifests under `k8s/`, matching the layout already there (`k8s/infra/` for shared components in namespace `raas`). `tdd` does not fit manifests; the spec's check is the test. Dry-run first: `kubectl apply --dry-run=server -R -f <path>`.

Fix failures before continuing. Commit to the branch in small conventional commits (`type(scope): short description`).

## 5. Redeploy locally

- **Service:**

```sh
powershell -File scripts/redeploy.ps1 -Service $1
```

The script builds `raas/<service>:latest` and restarts its deployment (`<service>-deployment`, gateway: `gateway`).

- **Infrastructure:** `kubectl apply -R -f <path>`, then `kubectl -n raas get pods` and `kubectl -n raas rollout status` for each new workload. Run the spec's check and report what you saw.

## 6. Review and approval

Call the Skill tool with `code-review` on the branch and fix what it raises. Then summarize what changed, the test and rollout results, and any spec deviations, and wait for the user to approve. Do not push or open a PR without approval.

## 7. Push and PR

After approval:

1. `git status`, check for files that should not be committed, commit anything left.
2. `git push -u origin HEAD`.
3. Write the PR body with the Skill tool `pr`, linking the spec path. Open it:

```sh
gh pr create --title "<prefix>: $0" --body "<description>" --base main --head "<prefix>/$0" --assignee @me
```

Report the PR URL. Leave the worktree in place until the user confirms the PR is merged, then `git worktree remove "../RaaS-$0"`.
