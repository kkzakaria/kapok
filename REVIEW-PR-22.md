# Review: PR #22 — Kubernetes Deployment & Helm Chart Generation (Epic 4)

## Summary

This PR implements the `kapok deploy` command with cloud-provider detection, Helm chart generation (umbrella + 3 subcharts), TLS/cert-manager support, HPA, and KEDA autoscaling. +1181/-12 lines across 16 files.

## Verdict: Approve with suggestions

The architecture is solid — clean separation between orchestration (`internal/deploy`), k8s concerns (`internal/k8s`), and CLI wiring. The testability via interfaces (`CloudDetector`, `CommandRunner`) is well done. Tests pass (k8s: OK, deploy: OK).

---

## Issues

### 1. Undefined `--context` flag read (Bug)
`cmd/kapok/cmd/deploy.go:63` — `runDeploy` reads `cmd.Flags().GetString("context")` but this flag is never registered in `init()`. This will always return an empty string silently. Either add the flag or use the config's `KubernetesConfig.Context`.

### 2. Temp directory leak when not dry-run
`internal/deploy/deploy.go:37-41` — When `OutputDir` is empty, a temp dir is created but never cleaned up after `helm upgrade --install` completes. Add a `defer os.RemoveAll(outputDir)` when the dir is auto-created.

### 3. No `--wait` timeout surfaced to user
`internal/deploy/deploy.go:58` — The hardcoded `--timeout 10m` for `helm upgrade --install` is reasonable but not configurable. Consider exposing it as a flag or at least documenting it.

### 4. KEDA template uses double-escaped Helm syntax inconsistently
`internal/k8s/keda.go` uses `{{ "{{ .Release.Name }}" }}` (double-escaped) while `internal/k8s/hpa.go` and `templates.go` use raw `{{ .Release.Name }}`. This works because of how each template is written, but the inconsistency is confusing. A comment explaining why KEDA/TLS need double-escaping would help.

### 5. Cloud detection false positives
`internal/k8s/cloud.go:79` — `strings.Contains(ctx, "eks")` matches any context containing "eks" (e.g. "my-deksktop-cluster"). Similarly "aks" matches "makes-cluster". Consider more specific patterns or word-boundary matching.

### 6. ConfigMap is identical for all subcharts
All 3 subcharts get the same ConfigMap with identical env vars. Each service likely needs different config. Fine as scaffolding but should be noted as a TODO.

### 7. No secrets management
The generated charts have no Secret resources for database passwords, JWT secrets, etc. Should be addressed before production use.

## Minor / Nits

- `internal/k8s/helm.go:94-95`: `DeploymentYAMLFmt` takes 6 `%s` args for the same name — consider `strings.ReplaceAll` instead of positional `Sprintf`.
- `pkg/config/config.go`: New `Domain`, `TLS`, `KEDA` fields in `KubernetesConfig` are added but not used by the deploy command (reads from flags). Wire them as defaults or remove.
- No test verifying generated YAML is valid Kubernetes manifests (not blocking, good to add later).
