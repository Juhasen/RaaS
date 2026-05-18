# Implementation Plan: 008-frontend-init

**Branch**: `008-frontend-init` | **Date**: 2026-05-18 | **Spec**: [specs/008-frontend-init/spec.md](specs/008-frontend-init/spec.md)
**Input**: Feature specification from `/specs/008-frontend-init/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Scaffold Angular routes, shared layout, and placeholder pages for listing creation and listing management in the existing `ui/` app, including a user-friendly not found route, with no backend integration or auth.

## Technical Context

**Language/Version**: TypeScript 5.9 (Angular 21.2)  
**Primary Dependencies**: Angular, Angular Router, Angular SSR (app already configured), RxJS  
**Storage**: N/A  
**Testing**: Angular CLI unit tests via `ng test` (Vitest runner)  
**Target Platform**: Web browsers (SSR-capable build)  
**Project Type**: Web application (Angular)  
**Performance Goals**: Fast initial render for placeholder content; no heavy client work  
**Constraints**: No backend or auth; use existing `ui/` app structure; routes must render inside shared layout  
**Scale/Scope**: Two listing pages + not found page + shared layout shell

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- PASS: UI remains simple/responsive; no backend integration (API gateway rule not exercised).
- PASS: No new service/data coupling introduced.

## Project Structure

### Documentation (this feature)

```text
specs/008-frontend-init/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command) (none expected)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
ui/
├── src/
│   ├── app/
│   │   ├── app.ts
│   │   ├── app.html
│   │   ├── app.routes.ts
│   │   ├── layout/
│   │   │   ├── listing-shell.component.ts
│   │   │   ├── listing-shell.component.html
│   │   │   └── listing-shell.component.css
│   │   ├── pages/
│   │   │   ├── listing-create/
│   │   │   │   ├── listing-create.component.ts
│   │   │   │   ├── listing-create.component.html
│   │   │   │   └── listing-create.component.css
│   │   │   ├── listing-manage/
│   │   │   │   ├── listing-manage.component.ts
│   │   │   │   ├── listing-manage.component.html
│   │   │   │   └── listing-manage.component.css
│   │   │   └── not-found/
│   │   │       ├── not-found.component.ts
│   │   │       ├── not-found.component.html
│   │   │       └── not-found.component.css
│   │   └── app.css
│   └── styles.css
└── package.json
```

**Structure Decision**: Use the existing Angular app under `ui/` with standalone components and route configuration in `app.routes.ts`, adding a shared layout shell and placeholder pages under `src/app`.
