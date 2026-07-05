# grocery-compare

A personal PWA for comparing a weekly shop across Tesco, Aldi, Asda, and Lidl — factoring in promotions, Clubcard/Asda Rewards personalized pricing, and a manually-entered employee gift-card discount — and recommending either the cheapest single store or a worthwhile two-store split.

This is a **Phase 1** build: all pricing data (shelf prices, promotions, Clubcard/Rewards prices, gift-card discounts) is entered manually through the app. See [docs/architecture.md](docs/architecture.md) for the roadmap toward automated scraping (Phase 2) and session-based authenticated pricing (Phase 3), and for why those are deliberately not built yet.

## Stack

- **Backend**: Go (chi router) + embedded SQLite, serving both the JSON API and the built frontend from a single binary.
- **Frontend**: React + TypeScript + Vite, built as an installable PWA (`vite-plugin-pwa`).
- **Deployment**: Helm chart (`charts/app`) + Jenkins (`homelabpipe` shared library), matching this homelab's house style.

## Local development

```bash
# Terminal 1 — API (SQLite file created automatically at apps/api/data/grocery.db)
cd apps/api
go run .                          # listens on :8080

# Terminal 2 — frontend (proxies /api/* to :8080 by default)
cd apps/web
npm install
npm run dev                       # listens on :5173
```

Open http://localhost:5173. Create a shopping list, add an item, map it to a couple of stores with prices (via each item's "edit store mappings" page), then compare.

### Running tests

```bash
cd apps/api
go test ./...
```

The most important tests live in `internal/compare/engine_test.go` (the comparison/recommendation algorithm) and `internal/compare/service_test.go` (the same logic wired through a real SQLite database).

### Building the production image

```bash
docker build -t grocery-compare:local .
docker run -p 8080:8080 -v grocery-data:/data grocery-compare:local
```

## Deployment

Deployed via Jenkins (`Jenkinsfile`, using the `homelabpipe` shared library) to the homelab's MicroK8s cluster, namespace `grocery-compare-ns`, at `grocery.atarnet.org`. See [docs/architecture.md](docs/architecture.md) for the one-time manual setup steps (GitHub repo, Jenkins job, DNS) and the hard constraint that this chart **must** stay at a single replica.
