# Local Master and trusted reverse-proxy request sources — 2026-10-09

The user requested local operation without requiring a bound domain, plus original-source detection behind Nginx. This change is on current source **after v0.1.0-beta.1**, not in those immutable release archives. No production services, domain, firewall or existing state were modified.

## Behavior

- Master still binds `127.0.0.1:8443` by default. `--origin` is now optional; absent a pinned origin, direct requests must use a loopback host (`localhost` or loopback IPv4/IPv6) and real TLS. Browser Origin must match the request's HTTPS origin. Generated certificates already cover localhost, 127.0.0.1 and ::1.
- `--trusted-proxies` accepts explicitly configured IPs/CIDRs; the empty default trusts nobody, including loopback peers. The address of the actual immediate connection determines trust, never a forwarding header.
- Trusted `X-Forwarded-Host` / `X-Forwarded-Proto` describe the original HTTPS origin. Malformed, duplicate/list-valued or non-HTTPS origin headers are rejected. HTTPS port 443 is omitted to match browser `URL.origin`; other ports are preserved.
- Trusted XFF is parsed as a bounded IP chain and walked from the nearest proxy toward the client, stopping at the first untrusted hop. `X-Real-IP` is the fallback when XFF is absent. Untrusted forwarding headers are ignored. Login rate limiting uses the resolved client IP.
- API, event WSS, native terminal WSS, instance logs and container logs all use the same resolved origin. A pinned `--origin` retains the previous fixed-origin policy. Backend TLS is still required; proxy headers never fabricate `r.TLS`.
- The Nginx example overwrites original-host/protocol/client-IP headers, preserves WSS upgrades and verifies the local backend's generated certificate using `proxy_ssl_name localhost`. The external certificate is independently configured on Nginx for its entry IP/hostname. Frontend and API stay on one HTTPS origin.

## Verification

| Command/check | Actual result |
| --- | --- |
| `go test ./internal/master -run 'TestRequestSourceTrustBoundary\|TestProxySourceRealTLSLoginAndWebSocket' -count=1` | Initial sandbox setup could not resolve standard-library `net/http/httputil`; same command with normal host access passed in 0.039s |
| `go test -race ./internal/master ./cmd/master -count=1` | PASS, exit 0; Master 272.171s, CLI package has no tests. This full run preceded the final default-port normalization; the targeted suite below covers that final delta |
| Final `go test -race ./internal/master -run 'TestRequestSourceTrustBoundary\|TestProxySourceRealTLSLoginAndWebSocket' -count=1` | PASS, exit 0, 1.373s; includes default HTTPS port and default trust-disabled checks |
| Final `make check build windows` | PASS, exit 0: Go vet, Linux Master/Daemon and Windows amd64 Master/Daemon builds |
| Documentation checks | 98 local README/operations links exist; JSON blocks parse, shell quoting balances, diff whitespace check clean |

The targeted boundary suite covers local IPv4/IPv6, arbitrary direct-host rejection, default local-peer distrust, untrusted spoofed headers, explicit proxy trust, forged XFF prefixes, trusted hops, IPv6 client IPs, pinned origin, insecure protocol, host lists/credentials/invalid ports, malformed IP chains, duplicate headers and absence of actual TLS.

The integration check starts a real HTTPS Master and a separate HTTPS reverse proxy forwarding the same headers as the Nginx example. It verifies login/session cookie behavior, authenticated event WebSocket upgrade, cross-origin HTTP and WSS rejection, client IP attribution in the login limiter, and continued local direct access. It uses fresh private SQLite state, an ephemeral test-only account and owned loopback listeners, closed at test cleanup. It is a Go reverse-proxy integration, **not an executed Nginx deployment**.

No frontend implementation, dependencies, database migrations or Daemon protocol changed. Windows evidence is cross-compilation for these changes, not a new device run. The prior beta acceptance, earlier device measurements and E08 limits retain their own source identities; they are not relabeled for this change. For deployment commands and version scope see [English README](../../../README.md#production-startup), [中文 README](../../../README.zh-CN.md#生产环境启动) and [operations](../../operations/OPERATIONS.md).
