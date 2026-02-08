# Changelog

## [3.0.0] - 2026-02-08

### Added
- External configuration via JSON file (`demo/config.json`), loaded from `CONFIG_PATH` env var with hardcoded fallback
- Health endpoints: `GET /healthz` (liveness) and `GET /readyz` (readiness)
- Graceful shutdown with `signal.NotifyContext` for SIGINT/SIGTERM
- Unit test suite (`demo/src/main_test.go`) with 11 tests covering all endpoints and config loading
- Kubernetes liveness and readiness probes in deployment template
- Kubernetes resource requests/limits (cpu: 50m/200m, memory: 64Mi/128Mi)
- Kubernetes ConfigMap for externalizing service configuration
- Go test and vet steps in GitHub Actions CI pipeline
- `golangci-lint` pre-commit hook
- `check-yaml` pre-commit hook (previously commented out)
- Auto-refresh polling (every 30s) in frontend status page
- Visual green/red circle status indicators in frontend
- `<meta charset="utf-8">` and viewport meta tag in `index.html`
- CA certificates in Docker image for HTTPS health checks

### Fixed
- **Critical**: Closure bug in service handler loop — all `/status/{id}` handlers previously resolved to the last service due to captured loop variable
- Response body leak: `resp.Body.Close()` now called on all paths (including non-200 responses)
- Proper HTTP error responses from `/status` endpoint (previously swallowed errors)

### Changed
- Bumped version from 2.0.0 to 3.0.0
- Updated `go.mod` from Go 1.22.1 to Go 1.23
- Switched from default `http.ServeMux` to `http.NewServeMux()` for testability
- Pinned Docker builder image to `golang:1.23-alpine`
- Status values changed from `"up /\"` / `"down \/"` to `"up"` / `"down"`
- Updated `docker/login-action` from v1 to v3 in GitHub Actions
- Removed duplicate Docker build step in CI (now tags and pushes existing image to GHCR)
- Fixed duplicate step names in GitHub Actions workflow
- Updated pre-commit hook versions: gitleaks v8.16.1 -> v8.21.2, pre-commit-hooks v2.3.0 -> v5.0.0, black 22.10.0 -> 24.10.0

### Removed
- `status:` section from Kubernetes deployment template (runtime state doesn't belong in a template)
- `automountServiceAccountToken: true` replaced with `false` (not needed)
- Removed `enableServiceLinks` and `shareProcessNamespace` from K8s template

## [2.0.0]

### Added
- Initial Go web server with service status monitoring
- Frontend with SVG animation (Vivus.js) and touch/keyboard navigation
- Dockerfile with multi-stage build (scratch base)
- Kubernetes deployment template
- GitHub Actions CI/CD pipeline for GKE and GHCR
- Pre-commit hooks (gitleaks, end-of-file-fixer, trailing-whitespace, black)
