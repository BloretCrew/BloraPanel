<p align="center">
  <img src="docs/assets/blora-mark.svg" width="80" height="80" alt="Blora logo">
</p>

<h1 align="center">Blora Panel</h1>

<p align="center">
  <strong>A desktop workspace for your servers.</strong><br>
  Manage nodes, instances, files, terminals, containers, and backups from one desktop.
</p>

<p align="center">
  <strong>English</strong> · <a href="README.zh-CN.md">简体中文</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&amp;logo=go&amp;logoColor=white" alt="Go 1.25">
  <img src="https://img.shields.io/badge/Vue-3-42B883?style=flat-square&amp;logo=vuedotjs&amp;logoColor=white" alt="Vue 3">
  <img src="https://img.shields.io/badge/TypeScript-6-3178C6?style=flat-square&amp;logo=typescript&amp;logoColor=white" alt="TypeScript 6">
  <img src="https://img.shields.io/badge/Status-In_development-64748B?style=flat-square" alt="In development">
</p>

<p align="center">
  <a href="#quick-start">Quick start</a> ·
  <a href="#production-startup">Production startup</a> ·
  <a href="#features">Features</a> ·
  <a href="#documentation">Documentation</a> ·
  <a href="sdk/README.md">Extension SDK</a>
</p>

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="web/ui-screenshots/all-appearance-20260926/mineral-dark-on-launcher.png">
    <img src="web/ui-screenshots/all-appearance-20260926/mineral-light-on-launcher.png" width="100%" alt="Blora Panel desktop with the application launcher open">
  </picture>
</p>

<p align="center">
  <sub>Light and dark appearance · Six color palettes · Independent translucent materials</sub><br>
  <sub>Screenshots show the current Chinese UI with isolated demo data.</sub>
</p>

<details>
<summary>View both light and dark screenshots</summary>

| Light | Dark |
| :---: | :---: |
| ![Light desktop](web/ui-screenshots/all-appearance-20260926/mineral-light-on-launcher.png) | ![Dark desktop](web/ui-screenshots/all-appearance-20260926/mineral-dark-on-launcher.png) |

</details>

## Features

Blora Panel brings server management into a Material Design 3–inspired desktop, with Go services handling operations independently of the browser.

| Area | What you can do |
| :--- | :--- |
| **Desktop workspace** | Open multiple windows and tabs, move views between windows, create resource shortcuts, and recover your workspace after a refresh. |
| **Instances & nodes** | View authorized resources across nodes, control instance lifecycles, manage quotas, enroll nodes, and rotate node identities. |
| **Files & editor** | Browse files, transfer data between nodes, and edit with Monaco, including recovery of unsaved text and undo history. |
| **Terminals & monitoring** | Use PTY terminals, inspect bounded log archives, follow metrics, and track persistent tasks. |
| **Docker & Compose** | Manage containers and Compose projects through authorized node operations, with task tracking and recovery. |
| **Backups & schedules** | Create backups, review restore plans, configure schedules, and run administrator-defined application hooks. |
| **Users & extensions** | Apply resource permissions and role templates, install and upgrade apps, and build extensions with the standalone SDK. |

Closing a window does not stop its instance. Opening a resource shortcut does not start it. Translucency, color palette, and light/dark appearance are independent settings.

## Quick start

**Tested toolchain:** Linux amd64, Go **1.25.9**, Node.js **24.15.0**, and npm **11.12.1**. Use GNU Make; release packaging also requires Python **3.9+**. Dependency versions are pinned in `go.mod`, `go.sum`, and the npm lockfiles.

```sh
git clone https://github.com/BloretCrew/BloraPanel.git
cd BloraPanel

make build
make web
make sdk
make fixture
```

| Command | Result |
| :--- | :--- |
| `make build` | Builds `dist/blora-master` and `dist/blora-daemon`. |
| `make web` | Installs locked frontend dependencies and builds `web/dist`. |
| `make sdk` | Builds the SDK and three reference extension packages. |
| `make fixture` | Starts a local Master and two Daemons with isolated test accounts and resources. |

Open the URL printed by the fixture; the default is **`https://127.0.0.1:9443`**. It creates a local test certificate and a private `.local/fixture-*` directory. Random administrator and member credentials are written to `browser-credentials.json` with mode `0600`; there is no reusable default password.

The member account has restricted access to two test instances and no instance-creation or host-shell permission. Instances start in the stopped state. Press **Ctrl+C** to stop the fixture's resources and services; its private logs remain available for diagnosis.

<details>
<summary><strong>Connect your own nodes</strong></summary>

### Initialize the Master

Create a private password file with a local editor, restrict it to mode `0600`, and use a password of 12–72 bytes. Keep credentials out of shell arguments and the repository.

```sh
./dist/blora-master --state-dir .local/master --init --password-file /absolute/private/initial-password
./dist/blora-master --state-dir .local/master --listen 127.0.0.1:8443 --origin https://localhost:8443 --static-dir web/dist
```

Initialization generates a local TLS certificate and refuses to overwrite existing accounts. For your management domain, use `--tls-cert` and `--tls-key`; `--origin` must match the browser's actual HTTPS origin.

### Enroll a Daemon

Sign in, generate a one-time enrollment ticket in the node application, and save it as a private file on the node. Create a separate configuration for each Daemon:

```json
{
  "stateDir": "/absolute/private/blora-node-a",
  "masterUrl": "https://localhost:8443",
  "caFile": "/absolute/private/master-ca.crt",
  "enrollmentFile": "/absolute/private/node-a.enrollment",
  "allowPGIDFallback": false
}
```

```sh
./dist/blora-daemon --config /absolute/private/node-a.json
```

Each Daemon needs its own state directory and identity. `caFile` must trust the management certificate; local testing can use the Master's generated `tls.crt`. An enrolled node reuses its identity rather than consuming another ticket.

Native Linux instances require an administrator-delegated, writable `cgroupRoot`. Administrators running trusted workloads can explicitly enable `allowPGIDFallback`; process groups do not provide host isolation for ordinary users. Isolated user workloads require a configured `dockerEndpoint`, a prepared image, and the appropriate permissions.

See the [operations guide](docs/operations/OPERATIONS.md) for state directories, health checks, consistent backups, upgrades, and rollback. [Administration evidence](docs/acceptance/reports/administration.md) covers quotas, roles, account changes, node maintenance, and key rotation. Rotation preserves the node ID and resource associations; log helpers and Windows Job keepers run independently of open browser views.

</details>

## Production startup

Use the matching [beta release archives](https://github.com/BloretCrew/BloraPanel/releases/tag/v0.1.0-beta.1), verify `SHA256SUMS`, and extract each component into its own directory. These commands start the real services without fixture accounts or demo resources. The current release remains a beta pending the owner's hands-on evaluation.

The examples use `panel.example.com` and Linux paths: extract Master to `/opt/blora/master`, Daemon to `/opt/blora/daemon` on each managed machine, and optionally the standalone Web archive to `/opt/blora/web`. Replace the domain and paths. Prepare dedicated `blora-master` / `blora-daemon` accounts and private state/config directories owned by the corresponding account (directories `0700`, secret files `0600`). Give Master read access to its TLS key and both services access to their required resources. Use a trusted certificate whose names include your management domain; the automatically generated localhost certificate is for local testing.

Point the management domain at the entry host and allow the chosen HTTPS port (`8443` for direct Master hosting, `443` for the Nginx example) from browsers and nodes. Managed applications' business ports are configured separately; these examples expose only the management entry.

### 1. Master — initialize once, then start

Create `/etc/blora/initial-password` with a local editor (12–72 bytes, readable only by the Master account). Initialization exits after creating the administrator; **do not put `--init` in the ongoing service command**.

```sh
sudo -u blora-master /opt/blora/master/blora-master \
  --state-dir /var/lib/blora/master --init \
  --password-file /etc/blora/initial-password

sudo -u blora-master /opt/blora/master/blora-master \
  --state-dir /var/lib/blora/master \
  --listen 0.0.0.0:8443 --origin https://panel.example.com:8443 \
  --static-dir /opt/blora/master/web/dist \
  --tls-cert /etc/blora/tls/fullchain.pem --tls-key /etc/blora/tls/privkey.pem \
  --extensions-dir /var/lib/blora/master/extensions
```

Open **`https://panel.example.com:8443`** and sign in as `admin` using your initial password. Master serves the bundled frontend, so a separate frontend server is optional. Remove the initial password file after successful initialization; preserve the state directory across restarts. `--extensions-dir` enables persistent extension installation; `blora-master --help` lists optional catalog-source and trusted-publisher-key flags.

### 2. Daemon — one per managed machine

In the node application, generate a one-time enrollment ticket and save it privately on the target machine as `/etc/blora/node.enrollment`. Save the following as `/etc/blora/node.json`:

```json
{
  "stateDir": "/var/lib/blora/node",
  "masterUrl": "https://panel.example.com:8443",
  "enrollmentFile": "/etc/blora/node.enrollment",
  "allowPGIDFallback": false
}
```

```sh
sudo -u blora-daemon /opt/blora/daemon/blora-daemon --config /etc/blora/node.json
```

Publicly trusted certificates use the system CA store. For a private CA, add `"caFile": "/etc/blora/master-ca.pem"` containing the trusted **CA certificate**, not a private key. After the node appears online, remove the ticket file and `enrollmentFile` configuration entry; later starts reuse the identity in `stateDir`. Each node needs its own state directory and ticket. Daemon makes outbound HTTPS/WSS connections; it needs no inbound management listening port. To manage the Master machine itself, run a Daemon there too.

This minimal configuration brings the node online. Linux native instances additionally require an administrator-prepared writable `cgroupRoot`; Docker/Compose requires an accessible `dockerEndpoint` and its runtime dependencies. Choose the service account's permissions for the intended node capabilities. Leave `allowPGIDFallback` disabled for normal deployment; it is an explicit trusted-workload fallback, not container isolation. See [operations](docs/operations/OPERATIONS.md) for platform capability and state/backup requirements.

### 3. Frontend — bundled or separately hosted

The frontend is a compiled static website executed in the browser. **With the Master command above, it is already available; there is no third application process to start.** Installing Node.js or running `npm run dev` / `vite preview` is not required on the production host.

For separate hosting, extract the matching standalone Web archive directly into `/opt/blora/web` (`index.html` is at that directory's root). Use the provided [Nginx configuration](docs/operations/examples/nginx.conf) inside Nginx's `http {}` context. It serves the website at **`https://panel.example.com`**, proxies `/api/` including node/browser WebSockets and `/healthz` to Master, and verifies backend TLS. Keep the frontend and API on the **same HTTPS origin**: the current frontend uses relative `/api/v1` URLs and same-origin cookies; it has no separate production API-base setting.

For this alternative, replace the Master's ongoing command with:

```sh
sudo -u blora-master /opt/blora/master/blora-master \
  --state-dir /var/lib/blora/master \
  --listen 127.0.0.1:8443 --origin https://panel.example.com --static-dir= \
  --tls-cert /etc/blora/tls/fullchain.pem --tls-key /etc/blora/tls/privkey.pem \
  --extensions-dir /var/lib/blora/master/extensions

# After installing and configuring Nginx, validate before starting it:
sudo nginx -t
sudo systemctl enable --now nginx
```

In the Nginx example, replace the domain/certificate paths and prepare `/etc/blora/backend-ca.pem` as a CA bundle that trusts Master's backend certificate. That certificate must also cover `panel.example.com`; Master always serves HTTPS, even behind a proxy. Change each Daemon's `masterUrl` to `https://panel.example.com` and use the CA for the browser-facing certificate if private. If Nginx is already running, reload it after a successful configuration check. [Nginx's WebSocket guide](https://nginx.org/en/docs/http/websocket.html) and [proxy TLS documentation](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_ssl_verify) explain the forwarding and certificate settings.

<details>
<summary><strong>Keep services running on Linux / start on Windows</strong></summary>

For Linux, the example [Master unit](docs/operations/examples/blora-master.service) and [Daemon unit](docs/operations/examples/blora-daemon.service) use the paths/accounts above. Create those accounts and directories first, initialize Master manually, then copy each unit to `/etc/systemd/system/` on its respective host. For separate frontend hosting, adjust Master's `ExecStart` to the loopback address, external origin without `:8443`, and `--static-dir=` as above. Do not start a unit and a foreground copy against the same state directory.

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now blora-master.service  # On the Master host
sudo systemctl enable --now blora-daemon.service  # On each node, when configured
sudo systemctl status blora-master.service       # Or blora-daemon.service
sudo journalctl -u blora-master.service -f        # Or blora-daemon.service
```

The Daemon unit deliberately stops only the Daemon process, preserving independently owned runs/log helpers. This is not a stop-all-resources operation: explicitly stop instances and PTYs and wait for task completion before a consistent backup or upgrade. The sample does not provision a delegated cgroup or install Docker.

On Windows, extract Master to `C:\Blora\master` and Daemon to `C:\Blora\daemon`, create private directories with ACLs restricted to their runtime accounts, and provide the domain certificate/key. Run Master in PowerShell:

```powershell
# Once only, after creating the private initial-password file:
& 'C:\Blora\master\blora-master.exe' --state-dir 'C:\Blora\state\master' --init --password-file 'C:\Blora\private\initial-password'

# Ongoing service process; built frontend is bundled with Master:
& 'C:\Blora\master\blora-master.exe' --state-dir 'C:\Blora\state\master' --listen '0.0.0.0:8443' --origin 'https://panel.example.com:8443' --static-dir 'C:\Blora\master\web\dist' --tls-cert 'C:\Blora\private\fullchain.pem' --tls-key 'C:\Blora\private\privkey.pem' --extensions-dir 'C:\Blora\state\master\extensions'
```

On the node, use the same Daemon JSON keys, replace `stateDir` / `enrollmentFile` / optional `caFile` with Windows paths such as `C:/Blora/state/node` and `C:/Blora/private/node.enrollment`, then start:

```powershell
& 'C:\Blora\daemon\blora-daemon.exe' --config 'C:\Blora\private\node.json'
```

Windows executables are console applications; they are not native SCM services and cannot be installed directly with `sc create`. For unattended startup, configure your chosen service wrapper/supervisor to launch these exact programs and preserve their state and independently owned helper processes. The new beta Windows packages have not been rerun on a device; earlier validation retains its recorded source identity.

</details>

Confirm `https://<management-origin>/healthz` succeeds with certificate verification, then check node-online status, permissions and intended resource operations in the desktop. `/healthz` checks the Master database, not all nodes/resources. These are deployment instructions; no production host or service has been changed by this documentation update.

## Architecture

| Component | Responsibility |
| :--- | :--- |
| **Frontend / Desktop** · Vue 3 / TypeScript | Runs in the browser: windows, tabs, default apps, extension hosting, user interaction and local workspace recovery. Distributed as static files, served by Master or a web server. |
| **Master** · Go | Central management service: accounts/sessions, server-side authorization, node enrollment, resource indexes, persistent task coordination and the API/WSS gateway. Usually one per deployment; it can also serve the frontend. |
| **Daemon** · Go | Execution service on each managed machine: local process lifecycle, PTYs, files, Docker/Compose, backups, monitoring and task execution/reporting. Only resources on a machine with a Daemon can be managed there. |

Nodes establish outbound management connections to the Master. HTTPS and WSS carry management data; managed services keep their own business networking. Resource authorization is enforced on the server, and extensions do not receive host command execution by default.

The normal management path is **browser frontend → Master → target Daemon**. The SDK is an extension-development toolkit, not a fourth running service. Closing the browser does not stop server-side instances, schedules or tasks; an offline node remains offline/unknown in the UI rather than being reported as stopped.

```text
cmd/        Master, Daemon, fixtures, and helper entry points
internal/   Backend services, protocol, storage, and platform adapters
web/        Vue desktop, default applications, and browser tests
sdk/        Standalone app SDK and reference extension
scripts/    Packaging and validation tools
docs/       Design, API, operations, and acceptance records
```

## Development

```sh
make check
make test
npm --prefix web test
npm --prefix web run test:e2e
```

The mock browser runner defaults to `/usr/bin/chromium-browser`; set `BLORA_CHROMIUM` to another Chromium executable when needed. Race tests require a supported native platform and a C toolchain. Tests that depend on Docker, delegated cgroups, or other platform capabilities have separate prerequisites and evidence.

<details>
<summary><strong>Real browser and fault validation</strong></summary>

With the local fixture running, use its private credential file:

```sh
BLORA_E2E_CREDENTIALS=/absolute/private/fixture/browser-credentials.json npm --prefix web run test:e2e:real
```

The real runner uses Playwright's installed browsers by default; `BLORA_CHROMIUM` can select a system Chromium. The [local regression guide](docs/operations/LOCAL_REGRESSION.md) includes a current-source Linux runner that builds the project, owns its HTTPS fixture and runs the complete real functional suite. Private credentials and raw evidence stay outside version control. See the [desktop report](docs/acceptance/reports/desktop-2026-09-09.md) for individual scenario setup.

**Sustained mixed load.** Start `./dist/blora-devfixture --performance`, then run in a second terminal:

```sh
BLORA_E2E_CREDENTIALS=/absolute/private/fixture/browser-credentials.json BLORA_PERF_SOAK_SECONDS=3600 npm --prefix web run test:e2e:real -- tests/real/performance.spec.ts
```

This exercises eight windows, two continuously producing PTYs, a directory with 10,000 entries, and repeated cross-node copies. Do not replace the frontend build or run competing heavy tests during measurement. Linux resource sampling is available through `python3 scripts/performance-resources.py .local/fixture-ID --samples 62 --interval 60`. Stop the sampler and the owned fixture after the test. Current results, including failed performance thresholds, are recorded in the [performance report](docs/acceptance/reports/performance-current-2026-09-26.md).

**Isolated disk-full validation.** This requires private Linux user/mount namespaces and uses a 1 MiB tmpfs in an owned test directory:

```sh
BLORA_TEST_ENOSPC=1 go test -race ./internal/master -run 'ENOSPC' -count=1 -v
```

**Isolated firewall validation.** This additionally needs private network namespaces, firewalld, firewall-cmd, nft, ip, and dbus-broker-launch:

```sh
BLORA_TEST_FIREWALL_NAMESPACE=1 go test -race ./internal/master -run '^TestPrivateFirewalldApplyAndRestore$' -count=1 -v
```

These fault tests only operate on their own isolated resources. A skipped test is not a passing result. See the [disk-full report](docs/acceptance/reports/enospc-2026-09-19.md) and [firewall report](docs/acceptance/reports/firewall-private-2026-09-19.md).

</details>

<details>
<summary><strong>Containers, backup hooks, and app registries</strong></summary>

Build the isolated instance image:

```sh
docker build -f Dockerfile.isolated -t blora/isolated:local .
```

The image uses a non-root user and includes the fixed-path exec helper. The adapter restricts privileges, resources, and mounts; PTY cleanup checks process identity before ending its own exec. See the [container terminal report](docs/acceptance/reports/terminal-container-2026-09-09.md).

Linux Daemons reconcile owned leftover Compose CLI processes on startup. If `recover Compose CLI ownership` fails, preserve `containers/cli-runs/`, investigate record integrity and process inspection permissions, then restart. New Engine mutations remain blocked until ownership is resolved; CLI exit alone does not prove remote rollback. See the [Compose recovery report](docs/acceptance/reports/compose-cli-recovery-2026-09-13.md).

Open the backup application from an instance's **Backups & schedules** entry. Restores use version checks, an immutable plan, explicit overwrite confirmation, and persistent tasks. Optional hooks are administrator-defined argv mappings in the Daemon configuration:

```json
{
  "backupRoot": "/absolute/private/node-a-backups",
  "backupHookCommands": {
    "backup.save.before": ["/absolute/private/bin/app-save", "--instance"],
    "backup.save.after": ["/absolute/private/bin/app-resume", "--instance"]
  }
}
```

Hooks do not run through a shell. Resource identity and hook ID are passed through `BLORA_BACKUP_*` environment variables; an unconfigured policy reports unavailable capability.

Use `--extensions-catalog /absolute/private/catalog` for a local registry, or `--extensions-catalog-url https://registry.example.invalid/blora/` for an HTTPS registry. These options are mutually exclusive; remote registries reject redirects. Enabled apps load authenticated bundles in an opaque-origin `sandbox="allow-scripts"` iframe and use the host's authorized capability and task bridge. See the [SDK documentation](sdk/README.md) for package construction, signatures, installation, upgrades, rollback, and data migration.

</details>

<details>
<summary><strong>Build release archives</strong></summary>

After the dependency setup above:

```sh
BLORA_VERSION=development-YYYYMMDD make package
cd dist/releases/development-YYYYMMDD
sha256sum -c SHA256SUMS
```

Packaging creates six archives: Linux and Windows Master/Daemon, a standalone frontend, and the SDK. Master archives include the frontend; Daemon archives include isolated-image build inputs; SDK archives include reference source, packages, and signing tools for both platforms. See the included `START.txt` for SDK build instructions.

Each archive also contains a per-file `MANIFEST.json`. Fixed timestamps, ordering, and modes make packaging reproducible for identical inputs; a conflicting existing version is rejected. Source or documentation changes require a new version. Archives exclude runtime identities and `node_modules`. See the [RC24 release report](docs/acceptance/reports/release-rc24-2026-09-27.md) for actual smoke, restore, and compatible rollback results.

`make package` collects upstream licenses and notices offline from compiled Go dependencies and locked npm packages. Missing license material stops packaging. If Go license files live outside `GOROOT`, set `BLORA_GO_LICENSE_DIR`; Fedora's `/usr/share/licenses/golang` is recognized by default.

</details>

## Project status

**Blora Panel is in active development.** Implementation and verification are tracked separately in the [acceptance matrix](docs/acceptance/ACCEPTANCE_MATRIX.md).

The first public build is **[v0.1.0-beta.1](https://github.com/BloretCrew/BloraPanel/releases/tag/v0.1.0-beta.1)**, a prerelease for initial hands-on evaluation. It has automated and agent-run validation, but has not yet been personally tested by the project owner. Stable-release readiness is not claimed. The [beta acceptance report](docs/acceptance/reports/beta-0.1.0-2026-10-08.md) records package/source integrity, extracted-package startup and restore, compatible upgrades/rollback, and verified remote attachments.

[Non-release source closeout](docs/acceptance/reports/non-release-closeout-2026-10-08.md) covers backend race checks, builds, 123 frontend units, all 51 real functional scenarios across a complete run and independent follow-ups, recovery repairs and strict native rendering checks. Earlier failures and the limits of combined evidence remain visible. Beta package validation followed that source closeout; the stable release awaits hands-on evaluation.

[Ordinary Windows real-device validation](docs/acceptance/reports/windows-device-passed-2026-10-06.md) passed all fourteen checks; earlier native Job/ConPTY and visible-WebKit results retain their fixed-source scope. The new beta Windows archives have not been rerun on a Windows device. The [current remaining-work record](docs/execution/NON_UI_REMAINING_2026-09-26.md) distinguishes the delivered beta from pending stable-release evaluation. Extra physical-failure and deployment combinations were waived for this delivery, rather than marked as passed.

[E08 optimization is closed under the user's adjusted scope](docs/acceptance/reports/e08-bounded-closeout-2026-10-08.md): Chromium measured 38.8ms, Firefox 66ms, and WebKit 89–132ms, with its last repetition at 95ms. The original universal 50ms target remains unmet; these source-specific stress results do not guarantee every run stays below 100ms.

Real systemd service/timer lifecycles and delegated cgroup process control have [isolated Linux container evidence](docs/acceptance/reports/systemd-cgroup-2026-09-29.md), including CPU throttling and bounded memory/PID exhaustion. [Native Linux notification rendering and mouse clicks](docs/acceptance/reports/native-notifications-2026-09-29.md) are also verified with X11/Dunst. Each report covers its stated environment; the [platform guide](docs/operations/PLATFORM_VALIDATION.md) provides reproducible runners. Production deployment and other platform combinations are not certified by local regression.

The repository includes source, tests, dependency lockfiles, documentation, and the two README screenshots. Credentials, runtime databases, installed dependencies, release archives, generated reference packages, and the remaining historical screenshot galleries stay outside version control.

## Documentation

Detailed engineering documents and the SDK guide are currently maintained in Chinese.

| Guide | Contents |
| :--- | :--- |
| [Operations](docs/operations/OPERATIONS.md) | State layout, health checks, backups, upgrades, and rollback. |
| [Platform validation](docs/operations/PLATFORM_VALIDATION.md) | Windows, systemd, remote Engine, and fault-validation procedures. |
| [Local regression](docs/operations/LOCAL_REGRESSION.md) | Current-source builds, ordinary browser checks, real functional validation and opt-in experiments. |
| [API specification](docs/api/openapi.yaml) | Requests, authorization, idempotency, tasks, WebSockets, and structured errors. |
| [Extension SDK](sdk/README.md) | App manifests, sandbox capabilities, signatures, and lifecycle contracts. |
| [Acceptance matrix](docs/acceptance/ACCEPTANCE_MATRIX.md) | Implemented, verified, pending, and environment-dependent requirements. |
| [Execution progress](docs/execution/PROGRESS.md) | Latest engineering decisions, evidence, and remaining work. |
| [Technical architecture](docs/plan/Blora-01-技术架构.md) | Master/Daemon architecture and core technical contracts. |
| [Features & interaction](docs/plan/Blora-02-功能与交互.md) | Product scope, applications, and workspace behavior. |

The API's `sessionCookie`, `X-CSRF-Token`, and `Idempotency-Key` describe checks performed by the actual Master. Resource-specific response bodies are defined by their respective applications.

## Third-party notices

Blora Panel is licensed under the **GNU General Public License version 3 only** (SPDX: `GPL-3.0-only`); see [LICENSE](LICENSE). Third-party components retain their own licenses. Their original texts and notices are recorded in the [dependency inventory](docs/licenses/inventory.json) and [third-party notices](docs/licenses/THIRD-PARTY-NOTICES.txt), and accompany the release archives.
