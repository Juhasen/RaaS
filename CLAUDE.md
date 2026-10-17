# RaaS

Microservices on a local Kubernetes cluster (namespace `raas`). Read the guideline file for the stack you are touching before changing a service; follow it so the services stay consistent.

## Guidelines

- **Go** (`/listing`, `/media`, `/booking`): [guidelines/GO.md](guidelines/GO.md)
- **Java** (`/payment`, `/review`, `/favorites`): [guidelines/JAVA.md](guidelines/JAVA.md)
- **Python** (`/notification`, `/user`, `/analytics`): no guideline file; follow PEP 8 and match the surrounding code.
- **Frontend** (`/ui`): [guidelines/TYPESCRIPT.md](guidelines/TYPESCRIPT.md) and [guidelines/ANGULAR.md](guidelines/ANGULAR.md)

## Workflow

Use the `deliver-change` skill for features, fixes and infrastructure work. Specs live in `.planning/`.

## Agent skills

### Issue tracker

GitHub Issues via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Default vocabulary (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
