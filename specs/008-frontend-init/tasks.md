---

description: "Task list for 008-frontend-init"
---

# Tasks: 008-frontend-init

**Input**: Design documents from `/specs/008-frontend-init/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md

**Tests**: Not requested for this feature.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Create listing feature directories under ui/src/app/layout/ and ui/src/app/pages/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core routing and shell primitives that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Create listing shell component in ui/src/app/layout/listing-shell.component.ts, ui/src/app/layout/listing-shell.component.html, ui/src/app/layout/listing-shell.component.css with a RouterOutlet wrapper
- [x] T003 Create not found component in ui/src/app/pages/not-found/not-found.component.ts, ui/src/app/pages/not-found/not-found.component.html, ui/src/app/pages/not-found/not-found.component.css
- [x] T004 Define base route tree with ListingShellComponent and NotFoundComponent in ui/src/app/app.routes.ts (child routes to be added in US1)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Navigate core listing areas (Priority: P1) 🎯 MVP

**Goal**: Navigate between listing creation and listing management pages using defined routes and visible navigation links

**Independent Test**: Load the app, visit `/listing/create` and `/listing/manage`, and navigate between them via visible links inside the shared layout

### Implementation for User Story 1

- [x] T005 [P] [US1] Create listing create component in ui/src/app/pages/listing-create/listing-create.component.ts, ui/src/app/pages/listing-create/listing-create.component.html, ui/src/app/pages/listing-create/listing-create.component.css
- [x] T006 [P] [US1] Create listing manage component in ui/src/app/pages/listing-manage/listing-manage.component.ts, ui/src/app/pages/listing-manage/listing-manage.component.html, ui/src/app/pages/listing-manage/listing-manage.component.css
- [x] T007 [US1] Wire `/listing/create` and `/listing/manage` child routes to the new components in ui/src/app/app.routes.ts
- [x] T008 [US1] Add navigation links between listing pages in ui/src/app/layout/listing-shell.component.html and link styles in ui/src/app/layout/listing-shell.component.css

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Consistent layout shell (Priority: P2)

**Goal**: Provide a consistent layout shell around all listing pages

**Independent Test**: Verify that header/navigation elements appear consistently on both listing pages

### Implementation for User Story 2

- [x] T009 [US2] Refine shared layout structure (header, nav, main) in ui/src/app/layout/listing-shell.component.html
- [x] T010 [US2] Align listing page templates to shared layout spacing classes in ui/src/app/pages/listing-create/listing-create.component.html and ui/src/app/pages/listing-manage/listing-manage.component.html
- [x] T011 [US2] Add shared layout styling rules (spacing, max-width, responsive tweaks) in ui/src/app/layout/listing-shell.component.css

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Placeholder content clarity (Priority: P3)

**Goal**: Provide clear placeholder messaging on listing pages

**Independent Test**: Verify each listing page shows a descriptive title and short message

### Implementation for User Story 3

- [x] T012 [P] [US3] Add title and description copy in ui/src/app/pages/listing-create/listing-create.component.html
- [x] T013 [P] [US3] Add title and description copy in ui/src/app/pages/listing-manage/listing-manage.component.html
- [x] T014 [US3] Add not found page copy and CTA link back to `/listing/create` in ui/src/app/pages/not-found/not-found.component.html

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T015 Update ui/README.md with the new listing routes and not found behavior
- [x] T016 Validate quickstart scenarios and adjust specs/008-frontend-init/quickstart.md if the steps differ

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Builds on the shared layout shell
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - No dependencies on other stories

### Within Each User Story

- Create components before wiring routes
- Routes before navigation links
- Layout structure before styling tweaks
- Placeholder copy after page templates exist

---

## Parallel Example: User Story 1

```bash
Task: "Create listing create component in ui/src/app/pages/listing-create/listing-create.component.ts, ui/src/app/pages/listing-create/listing-create.component.html, ui/src/app/pages/listing-create/listing-create.component.css"
Task: "Create listing manage component in ui/src/app/pages/listing-manage/listing-manage.component.ts, ui/src/app/pages/listing-manage/listing-manage.component.html, ui/src/app/pages/listing-manage/listing-manage.component.css"
```

---

## Parallel Example: User Story 3

```bash
Task: "Add title and description copy in ui/src/app/pages/listing-create/listing-create.component.html"
Task: "Add title and description copy in ui/src/app/pages/listing-manage/listing-manage.component.html"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Demo (MVP)
3. Add User Story 2 → Test independently → Demo
4. Add User Story 3 → Test independently → Demo
5. Finish Polish tasks

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
3. Stories complete and integrate independently
