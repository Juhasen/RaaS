---
name: deliver-change
description: Deliver a change to RaaS end to end in an isolated git worktree - write a spec under .changes, get it attacked by an adversarial review, build it test-first, redeploy locally, then open a PR after the user approves. Covers service features, bug fixes and infrastructure work (cluster additions such as monitoring or storage). Use whenever the user asks to build, add, change or ship something in this repo, even if they do not say "deliver".
argument-hint: <issue-url|issue-number|description> [service-name|infrastructure]
---

# Deliver change

Arguments: `$0` = what to deliver: a GitHub issue URL or number, or a description pasted in the chat. `$1` (optional) = target: a service directory (`listing`, `media`, `booking`, `payment`, `review`, `favorites`, `notification`, `user`, `analytics`, `gateway`, `ui`) or `infrastructure` for cluster-level work.

Resolve the input first:

- **Issue URL or number:** fetch it as `docs/agents/issue-tracker.md` describes (`gh issue view <n> --json number,title,body,labels,comments`). The issue is the starting requirement; keep its number for the PR.
- **Pasted description:** use it as the requirement. No issue number.
- If `$1` is missing, infer the target from the requirement; ask if it is not obvious.

Derive `<name>` from the issue title or description: kebab-case, at most five words. It names the branch, the spec folder and the PR.
Call the target `<target>` below (`$1`, or the one you inferred).

Run the steps in order and stop at the first failure. The git, docker and kubectl commands are exact; do not alter them.

Why this shape: the spec is written and attacked before any code exists, because a wrong spec is the most expensive bug. The work happens in a worktree so the user's main checkout stays untouched while this runs.

Branch prefix: `feat/` for a new capability, `fix/` for a bug, `infra/` when `<target>` is `infrastructure`. Call it `<prefix>/<name>` below.

## 1. Worktree

Start from an up-to-date `main` in a fresh worktree next to the repo. Do not touch the main checkout's files.

```sh
git fetch origin
git worktree add "../RaaS-<name>" -b "<prefix>/<name>" origin/main
cd "../RaaS-<name>"
```

Do all remaining work inside that worktree. If the branch or folder already exists, stop and ask; it may be someone's unfinished work.

## 2. Spec

Planning docs live in `.changes/`:

- `.changes/services/<service>/<name>/spec.md` for a service
- `.changes/infrastructure/<name>/spec.md` for cluster work

Create the folder if missing. `.changes/proposals/` holds unscheduled ideas; if a proposal covers this change, read it first and use it as input.

Read the issue body and comments if there is an issue. Read `CLAUDE.md` and the guideline file it names for the target, plus `GLOSSARY.md` and `docs/adr/` if they exist. Then call the Skill tool with `grilling` to close the open questions with the user, one at a time, before writing.

Write `spec.md` with the `to-spec` template: Problem Statement, Solution, User Stories, Implementation Decisions, Testing Decisions, Out of Scope, Further Notes. Do not publish it to the issue tracker; the file in `.changes` is the spec. Follow `to-spec` for test seams: prefer existing seams, as few as possible, and confirm them with the user.

Testing Decisions must also say which levels of verification this change needs: unit/integration tests, a check on the local deployment, and any manual test only a human can do (a UI flow, a Grafana dashboard looking right). Step 6 reports against that list.

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
powershell -File scripts/redeploy.ps1 -Service <target>
```

The script builds `raas/<service>:latest` and restarts its deployment (`<service>-deployment`, gateway: `gateway`).

- **Infrastructure:** `kubectl apply -R -f <path>`, then `kubectl -n raas get pods` and `kubectl -n raas rollout status` for each new workload. Run the spec's check and report what you saw.

## 6. Validation report

Write `validation.md` next to `spec.md`. Keep it short; its job is to say what was actually checked, so nobody has to guess. Fill it from real command output, not from memory:

```md
# Validation: <name>

| Level | What was run | Result |
|---|---|---|
| Unit / integration | `go test ./...` | pass, 42 tests |
| Deployment | redeploy, `kubectl -n raas get pods`, curl of the new endpoint | pods Ready, 200 OK |
| Manual | what the user must try, or what you tried in the browser | done / pending the user |

## Not verified
Anything in the spec's Testing Decisions that was skipped or could not be run, with the reason.
```

Mark a level `n/a` with a reason when the change does not need it (an infra change has no unit tests). Never write `pass` for something you did not run. Manual checks you cannot do yourself stay `pending the user`, and you tell the user exactly what to click or run.

## 7. Review and approval

Call the Skill tool with `code-review` and pass an explicit target in `args`: the branch, the worktree path and `git diff origin/main...<prefix>/<name>`, and say to ignore uncommitted changes in the main checkout. Never call it with no argument: the session cwd (and the shell cwd after every command) is the main checkout, so a bare call reviews the wrong files. When the report arrives, compare its file paths with `git diff --name-only origin/main...HEAD` run in the worktree; if they differ, discard it and run the review again. Fix what it raises. Then summarize what changed, point to `validation.md`, and any spec deviations, and wait for the user to approve. Do not push or open a PR without approval.

## 8. Push and PR

After approval:

1. `git status`, check for files that should not be committed, commit anything left.
2. `git push -u origin HEAD`.
3. Write the PR body with the Skill tool `pr`, linking the spec and `validation.md` paths and summarizing the validation table. If the work came from an issue, add `Closes #<n>`. Open it:

```sh
gh pr create --title "<prefix>: <name>" --body "<description>" --base main --head "<prefix>/<name>" --assignee @me
```

Report the PR URL. Leave the worktree in place until the user confirms the PR is merged, then `git worktree remove "../RaaS-<name>"`.
