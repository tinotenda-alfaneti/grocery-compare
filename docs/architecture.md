# Architecture

## Why this exists

Tesco, Aldi, Asda, and Lidl publish no pricing API. Clubcard/Asda Rewards personalized prices only exist behind a logged-in session. An employee gift-card discount (e.g. a perks site like JPMorgan Discounts) sits behind corporate SSO — automating against that is a different, higher-stakes risk than a retailer account and is deliberately never attempted (see Phase boundaries below).

## Deployable unit

One container, matching the only pattern proven in this homelab (one Deployment, one Service, one Ingress):

```
Browser (installed PWA)
   │ HTTPS, grocery.atarnet.org
   ▼
Ingress → Service (ClusterIP :80→:8080) → Deployment (1 replica)
   grocery-compare container:
     ├─ serves /api/*        (chi router, comparison engine)
     ├─ serves /*  static    (Vite build, SPA fallback)
     └─ SQLite file on PVC (/data/grocery.db)
```

The Go binary serves both the JSON API and the built React static assets (SPA fallback to `index.html`), exactly how `homelabsite` and `ebook-reader` serve their frontends in this homelab — one image, one container, no separate frontend service.

## Storage: SQLite, not Postgres

- Single personal user, near-zero write concurrency.
- Matches `homelabsite`'s embedded-SQLite pattern for simple single-instance Go apps (CGO via `mattn/go-sqlite3`, statically linked — see the `Dockerfile`).
- One file on a PVC: trivial to back up, no separate database pod to keep alive.

**Hard constraint: `replicaCount` must stay `1`, and `autoscaling.enabled` must stay `false`.** A second replica writing the same SQLite file on a `ReadWriteOnce` PVC will corrupt data or fail to schedule. `charts/app/templates/hpa.yaml` hard-fails the Helm chart if `autoscaling.enabled` is ever set to `true`, so this can't be silently turned on later.

## Comparison algorithm

Implemented in `apps/api/internal/compare/engine.go` (pure functions, no DB) and wired to SQLite in `service.go`. Given a shopping list:

1. For each store, resolve each item's **effective price**: shelf price → overridden by an active promo → further reduced by an active Clubcard/Rewards price if lower.
2. Sum per store, apply that store's gift-card discount **once, to the summed subtotal** (it's a payment-method discount, not a per-item one).
3. Pick the cheapest store that **fully covers** the list (falling back to partial coverage, flagged, only if no store covers everything).
4. Try every store pair for a two-store split, likewise preferring the pair that covers the most items before comparing price — a pair that stocks nothing looks "free" otherwise, which was a real bug caught during testing (see `TestBestTwoStoreSplit_IgnoresZeroCoveragePairs`).
5. Recommend the split only if it saves at least `Settings.secondStopMinSavingPence` (an absolute £ amount, not a percentage — the cost of a second trip is roughly fixed regardless of basket size).

## Data model

See [DATA_MODEL.md](DATA_MODEL.md). The core idea: you maintain your own list of "canonical items" (e.g. "porridge oats"), and manually map each one to the specific product at each store you actually buy it from (`ProductMapping`). There's no automatic cross-store product matching — Aldi/Lidl stock own-brand equivalents rather than the same branded SKUs as Tesco/Asda, so that's a real entity-resolution problem this app deliberately doesn't attempt to solve. Comparisons only ever run over items you've explicitly mapped.

## Phasing

- **Phase 1 (this build)**: manual entry of everything — prices, promos, Clubcard/Rewards prices, gift-card discount %. Fully functional on its own.
- **Phase 2 (roadmap, not built)**: public-page scrapers per retailer as K8s CronJobs. They wouldn't touch the SQLite file directly (MicroK8s `ReadWriteOnce` PVCs are node-local; a CronJob pod on a different node couldn't mount it anyway) — instead they'd call small internal write endpoints on the running API, reusing the same `PriceObservation`/`PromoObservation` write paths as manual entry, tagged `source: scraped`. `ProductMapping.product_url`, already captured during manual mapping, becomes the scrape target — no catalog-wide crawling needed to start. Suggested order: Aldi/Lidl first (simpler sites, lighter bot-defense), Asda/Tesco last. Keep scrape frequency low (a few times a week).
- **Phase 3 (roadmap, not built)**: session-based Clubcard/Rewards scraping. You'd log in once yourself in a real browser; the session cookie gets captured manually (DevTools → Application → Cookies) and stored as a Kubernetes Secret — never an automated login replay, and never your raw credentials. On any 401/403, the session is marked stale and the app falls back to the base/promo price; manual entry remains the permanent fallback.
- **Never in scope**: automating anything against the employee gift-card discount site. It's behind corporate SSO — a different risk category from a retailer account — so that value is manual-entry-only, permanently.

## CI/CD

`Jenkinsfile` is a self-contained inline pipeline (Checkout → install kubectl/helm → verify cluster → prepare namespace/secrets → Kaniko build → Trivy scan → Helm deploy → verify), modeled directly on `ebook-reader`'s proven, currently-working Jenkinsfile. It deliberately does **not** use the `homelabpipe` shared library — that library exists as a repo in this homelab but isn't registered in Jenkins yet (Manage Jenkins → Global Pipeline Libraries), so depending on it would make this pipeline unable to run at all. Revisit switching to `@Library('homelabpipe') _` once that library is actually installed and proven on another repo first.

The Kaniko/Trivy job manifests live in `ci/kubernetes/{kaniko.yaml,trivy.yaml}` as `__CONTEXT_URL__`/`__IMAGE_DEST__` templates, substituted by the Jenkinsfile at build time — same pattern as `ebook-reader`.

## One-time manual setup (not covered by code)

- Create the GitHub repo as **public** (the pipeline clones over a plain unauthenticated `git://github.com/...` URL, same as `ebook-reader`).
- Wire a Jenkins multibranch job to it.
- Add the `grocery.atarnet.org` DNS record.
- Confirm `dockerhub-creds` exists in `test-ns` for the namespace-copy step the pipeline performs.

## Other things worth knowing

- All money is stored and computed as **integer pence**, never floats.
- Effective dates (`effective_from`/`effective_to`) are plain `YYYY-MM-DD` strings compared as UTC dates — a few hours of slop around midnight is a non-issue for weekly promo cycles.
- `Store.included_in_comparisons` lets you exclude a store you never actually visit from all comparisons.
- The optional PIN lock (`internal/auth`) is a convenience gate, not real security — no rate limiting or lockout.
