Custom Stonework & Masonry Estimation — Phase 1

This repository implements the phase-1 scaffold for an estimation API.

Structure

- `cmd/estimator/` — application entrypoint (minimal scaffold)
- `domain/` — domain models and small helpers (types and validation)
- `service/` — use-case interfaces and future implementations
- `store/` — data access interfaces (material catalog)
- `handler/` — HTTP handler wiring (thin adapters)

Next steps (phase 2 options)
- Implement `service` estimation logic (surface area, tonnage, mortar)
- Add a concrete `store` implementation (in-memory or DB-backed)
- Implement HTTP handlers and routes in `handler`
