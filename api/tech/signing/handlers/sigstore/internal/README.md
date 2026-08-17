# sigstore handler — vendored cosign helpers

Stop-gap fork of the Fulcio/OIDC signing glue that cosign v3.1.2 removed in
upstream commit `0fc9811` ("Remove unused signing code"). The OCM sigstore
handler in `../handler.go` depended on the removed `fulcio.NewSigner` call
chain, so the minimum surface it reaches is preserved here verbatim from
cosign **v3.1.1**.

- Issue: https://github.com/open-component-model/ocm/issues/2056
- Cosign v3.1.2 release notes: https://github.com/sigstore/cosign/releases/tag/v3.1.2

## Layout

| Path | Upstream source (cosign v3.1.1) | Notes |
|------|---------------------------------|-------|
| `fulcio/fulcio.go` | `cmd/cosign/cli/fulcio/fulcio.go` | Removed upstream in v3.1.2. `GetRoots`/`GetIntermediates` dropped (unused; needed the internal `fulcioroots` package). |
| `auth/auth.go` | `internal/auth/auth.go` | Behind Go's `internal/` visibility, so forked. `RetrieveIDToken` + related dead symbols dropped. |
| `ui/{env,log,prompt}.go` | `internal/ui/*` | Behind Go's `internal/` visibility. Byte-identical except unused `WithEnv`/`RunWithTestCtx`/`Warnf` dropped. |

Public cosign packages still imported directly from upstream (no fork
needed): `cmd/cosign/cli/options`, `cmd/cosign/cli/sign/privacy`,
`pkg/providers`, `pkg/cosign`.