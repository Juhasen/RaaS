# Research: 008-frontend-init

## Decisions

### 1) Angular routing with a shared layout shell

- **Decision**: Use Angular Router with a top-level listing shell component that renders child routes for listing creation and listing management, plus a catch-all not found route.
- **Rationale**: The app already uses Angular standalone components and an empty `Routes` array. A layout shell keeps navigation and structure consistent across pages.
- **Alternatives considered**: Direct routes without a layout shell (rejected because it would duplicate navigation and layout markup on each page).

### 2) Standalone placeholder page components

- **Decision**: Implement listing creation, listing management, and not found pages as standalone components with static title/description content.
- **Rationale**: Matches the current Angular app setup and keeps the feature scope limited to placeholders.
- **Alternatives considered**: Inline templates in route config (rejected for maintainability and future expansion).

### 3) Minimal styling scoped to components

- **Decision**: Keep styling minimal and scoped to component styles and existing global `styles.css`.
- **Rationale**: Avoids large design changes while providing clear, consistent placeholders.
- **Alternatives considered**: Introduce a design system or global theme (rejected due to scope and time).
