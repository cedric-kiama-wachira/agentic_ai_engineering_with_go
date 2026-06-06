# Agentic AI Engineering with Go

> Who said AI applications have to be developed in Python?

Production-grade agentic AI systems, built in Go.

## Prerequisites

- Go 1.26.1
- Git with SSH access to this repository

## Getting Started

```bash
git clone git@github.com:cedric-kiama-wachira/agentic_ai_engineering_with_go.git
cd agentic_ai_engineering_with_go
cp .env.example .env   # then fill in your local values
```

## Branching Model

This repo follows **Git Flow**. Do **not** commit directly to `main` or `develop`.

| Branch       | Purpose                          | Branches from | Merges into        |
|--------------|----------------------------------|---------------|--------------------|
| `main`       | Production-ready, tagged releases| —             | —                  |
| `develop`    | Integration branch               | `main`        | `main` (via PR)    |
| `feature/*`  | New work                         | `develop`     | `develop` (via PR) |
| `bugfix/*`   | Non-urgent fixes                 | `develop`     | `develop` (via PR) |
| `hotfix/*`   | Urgent production fixes          | `main`        | `main` + `develop` |

All changes land through reviewed pull requests.

## Contributing

See `CONTRIBUTING.md` (added in a later step).

## License

Apache-2.0 — see `LICENSE`.

