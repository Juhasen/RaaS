---
description: "Task list for API Gateway Entry Point & Auth Context"
---

# Tasks: API Gateway Entry Point & Auth Context

**Input**: Design documents from `/specs/006-api-gateway-auth/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Not included (not explicitly requested in the feature specification)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

 - [X] T001 Create gateway container entrypoint and baseline image config in gateway/Dockerfile
 - [X] T002 [P] Create base routing config files in gateway/config/routes.yaml and gateway/config/public-routes.yaml
 - [X] T003 [P] Create initial ownership policy file in gateway/policies/ownership.yaml
 - [X] T004 [P] Add gateway deployment/service manifest stub in k8s/apps/gateway.yaml

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete


 - [X] T005 Define shared gateway runtime settings (config sources, request id propagation, timeouts) in gateway/config/gateway.yaml
 - [X] T006 Define standard error response mapping for auth and routing failures in gateway/config/errors.yaml

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Single Entry Point for the UI (Priority: P1) MVP

**Goal**: Provide one stable entry point that routes requests by path and method without response reshaping.

**Independent Test**: Send multiple public and protected requests through the entry point and confirm the correct backend responses.

### Implementation for User Story 1


 - [X] T007 [US1] Populate full routing map (method + path -> backend) in gateway/config/routes.yaml
 - [X] T008 [P] [US1] Populate explicit public allowlist in gateway/config/public-routes.yaml
 - [X] T009 [US1] Configure protect-by-default routing behavior and passthrough proxying in gateway/config/gateway.yaml
 - [X] T010 [US1] Wire gateway service exposure and config mounts for the UI entry point in k8s/apps/gateway.yaml

**Checkpoint**: User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Protected Access Uses Valid Identity (Priority: P1)

**Goal**: Enforce valid JWT access on protected routes and inject trusted identity context downstream.

**Independent Test**: Attempt the same protected request with a valid token and with an invalid or missing token.

### Implementation for User Story 2

 - [X] T011 [US2] Configure JWT validation inputs (issuer, audience, keys/JWKS) in gateway/config/auth.yaml
 - [X] T012 [P] [US2] Define identity context injection and header stripping rules for x-raas-* in gateway/config/auth-context.yaml
 - [X] T013 [US2] Attach auth requirements to protected routes in gateway/config/routes.yaml

**Checkpoint**: User Story 2 should be fully functional and testable independently

---

## Phase 5: User Story 3 - Role-Appropriate Access (Priority: P2)

**Goal**: Enforce role-restricted routes and ownership checks for protected resources.

**Independent Test**: Attempt a host-only action using a guest role and confirm access is denied.

### Implementation for User Story 3

 - [X] T014 [P] [US3] Define role-based policy rules for host and guest in gateway/policies/roles.yaml
 - [X] T015 [US3] Map required roles per route in gateway/config/routes.yaml
 - [X] T016 [US3] Define ownership verification rules (resource type, id source, verifier endpoint) in gateway/policies/ownership.yaml
 - [X] T017 [US3] Configure ownership preflight settings (timeout, error mapping) in gateway/config/gateway.yaml and gateway/config/errors.yaml

**Checkpoint**: User Story 3 should be fully functional and testable independently

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T018 Align configuration docs and env var list in specs/006-api-gateway-auth/quickstart.md
- [ ] T019 [P] Validate header propagation and ownership contract notes in specs/006-api-gateway-auth/contracts/gateway-auth-context.md
- [ ] T020 Run the smoke tests listed in specs/006-api-gateway-auth/quickstart.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: Depend on Foundational phase completion
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Foundational (Phase 2) only
- **User Story 2 (P1)**: Depends on Foundational (Phase 2) only
- **User Story 3 (P2)**: Depends on User Story 2 (identity context) and Foundational (Phase 2)

### Within Each User Story

- Configuration and policy files before deployment wiring
- Auth context before role or ownership enforcement
- Story complete before moving to the next priority

### Parallel Opportunities

- Setup tasks marked [P] can run in parallel
- Foundational tasks can proceed sequentially due to shared files
- After Foundational completes, User Stories 1 and 2 can proceed in parallel
- Within each story, tasks marked [P] can run in parallel

---

## Parallel Example: User Story 1

```bash
Task: "Populate full routing map (method + path -> backend) in gateway/config/routes.yaml"
Task: "Populate explicit public allowlist in gateway/config/public-routes.yaml"
```

## Parallel Example: User Story 2

```bash
Task: "Configure JWT validation inputs (issuer, audience, keys/JWKS) in gateway/config/auth.yaml"
Task: "Define identity context injection and header stripping rules for x-raas-* in gateway/config/auth-context.yaml"
```

## Parallel Example: User Story 3

```bash
Task: "Define role-based policy rules for host and guest in gateway/policies/roles.yaml"
Task: "Define ownership verification rules (resource type, id source, verifier endpoint) in gateway/policies/ownership.yaml"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. Validate User Story 1 independently

### Incremental Delivery

1. Complete Setup + Foundational
2. Add User Story 1 -> validate independently -> deploy/demo
3. Add User Story 2 -> validate independently -> deploy/demo
4. Add User Story 3 -> validate independently -> deploy/demo

### Parallel Team Strategy

1. Team completes Setup + Foundational together
2. After Foundational:
   - Developer A: User Story 1
   - Developer B: User Story 2
3. After User Story 2:
   - Developer C: User Story 3
