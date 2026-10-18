# RaaS

Microservices on a local Kubernetes cluster (namespace `raas`). Read the guideline file for the stack you are touching before changing a service; follow it so the services stay consistent.

## Guidelines

- **Go** (`/listing`, `/media`, `/booking`): [guidelines/GO.md](guidelines/GO.md)
- **Java** (`/payment`, `/review`, `/favorites`): [guidelines/JAVA.md](guidelines/JAVA.md)
- **Python** (`/notification`, `/user`, `/analytics`): no guideline file; follow PEP 8 and match the surrounding code.
- **Frontend** (`/ui`): [guidelines/TYPESCRIPT.md](guidelines/TYPESCRIPT.md) and [guidelines/ANGULAR.md](guidelines/ANGULAR.md)

## Workflow

Use the `deliver-change` skill for features, fixes and infrastructure work. Specs live in `.changes/`.

## Agent skills

### Issue tracker

GitHub Issues via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Default vocabulary (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

Rules:
- `graphify-out/` is local and gitignored. If it is missing, skip these rules; build it with `/graphify .` and install the rebuild hook with `graphify hook install`.
- When graphify-out/graph.json exists, run `graphify query "<question>"` before your first grep, search or source-file read for a codebase question, and put this rule in every subagent prompt that explores code (except the cold adversarial reviewer in `deliver-change`, which gets only the spec path and the repo). Read raw files only for lines the graph or an earlier read already located. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
