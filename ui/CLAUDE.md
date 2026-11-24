# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## About

yaba-ui is a personal finance management dashboard (budgeting, expense tracking, payment methods, credit card rewards). It connects to a yaba backend server at `http://localhost:9222`.

## Commands

```bash
npm start          # Dev server on port 3000
npm run build      # Production build
npm test           # Run Vitest test suite
npm run lint       # ESLint check
npm run lint:fix   # ESLint auto-fix
npm run prettier   # Format with Prettier
```

## Architecture

**Stack:** React 18 + Vite, Material-UI v5, React Router v6, Apollo Client (GraphQL), Formik + Yup, ApexCharts, dayjs.

**API layer:** Three mechanisms in use:
- Apollo Client with `useQuery`/`useMutation` for GraphQL — queries/mutations defined in `src/api/graph.jsx`
- `fetch('/graphql')` with `credentials: 'include'` — used directly in `src/pages/profile/index.jsx`
- Axios + SWR for REST calls to `/api` endpoints

Dev proxy routes `/graphql` and `/api` to `http://localhost:9222`.

**State management:**
- Apollo Client for server state
- `BudgetContext` (`src/pages/budget/BudgetContext.jsx`) for budget page local state
- Auth reducer in `src/contexts/` for authentication state

**Routing:** `src/routes/` — routes use `React.lazy()` + a `Loadable` wrapper for code splitting. Two layouts: `DashboardLayout` (authenticated pages) and `MinimalLayout` (login).

**Theme:** MUI theme configured in `src/themes/` with palette, typography, shadows, and component overrides.

**Import aliases:** `src/` maps to the project root `src/` directory (configured in `jsconfig.json` and `vite.config.mjs`).

## Key Files

| File | Purpose |
|------|---------|
| `src/api/graph.jsx` | All GraphQL queries and mutations |
| `src/routes/index.jsx` | Top-level router with lazy-loaded pages |
| `src/pages/budget/BudgetContext.jsx` | Budget state context |
| `src/themes/index.jsx` | MUI theme provider |
| `src/menu-items/dashboard.jsx` | Sidebar navigation definitions |
| `vite.config.mjs` | Build config, dev proxy, port |
