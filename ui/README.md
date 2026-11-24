# yaba-ui

UI for YABA. Built into `ui/dist` and embedded in the Go binary.

## Setup

From the repo root, `make dev` starts Postgres, the Go server (:9222), and this dev server (:3000).

## Pages

| Route | Description |
|-------|-------------|
| `/dashboard` | Budget vs spending charts, category breakdown, date range selector |
| `/budget` | Create and edit budgets with income sources and expense categories |
| `/expenditure` | Browse, filter, upload, and export transactions |
| `/browse` | Search reward cards and add them as payment methods |
| `/payment-methods` | Manage payment methods and their linked reward cards |
| `/profile` | Update account email and password |

## Commands

```bash
yarn start          # Dev server (port 3000, proxies /graphql and /api to localhost:9222)
yarn build          # Production build
yarn test           # Run Vitest test suite
yarn lint           # ESLint
yarn lint:fix       # ESLint auto-fix
yarn prettier       # Format with Prettier
```
