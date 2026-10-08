# Local regression checks

Run these checks on the source you intend to hand off. They build and test the
project; they do not create a release, publish a package or deploy a service.

## Build and unit checks

```sh
make check build windows
go test -race -p 1 ./... -count=1
cd web
npm run build
npm test
```

`make windows` verifies the Windows build, not Windows runtime behavior. Existing
real-device results and their fixed revisions are recorded in the acceptance
reports. Opt-in tests requiring extra environments may skip in the Go suite;
such skips do not count as passed environment checks.

## Ordinary browser checks

Install the matching Playwright browsers, then run from `web`:

```sh
npx playwright install chromium firefox webkit
export BLORA_CHROMIUM="$(node -e 'process.stdout.write(require("playwright").chromium.executablePath())')"
BLORA_E2E_FRESH_SERVER=1 npm run test:e2e -- --workers=1
BLORA_BROWSER=firefox BLORA_E2E_FRESH_SERVER=1 npm run test:e2e -- --workers=1
BLORA_BROWSER=webkit BLORA_E2E_FRESH_SERVER=1 npm run test:e2e -- --workers=1
```

The explicit Chromium path selects the installed matching full browser rather
than the default system executable. Choose a free `BLORA_E2E_PORT` if the default
port is in use; a fresh runner must not stop or reuse another service.

The default config selects `tests/browser`. Chromium's ordinary window-position
checks use native DPR 1 and 1.25, with unchanged whole-screen pixel comparisons
and eight stable captures per scene. Recovery-worker checks exercise real
IndexedDB transactions, worker failures and immediate reload protection. Mock
API cases verify client behavior; they are not real backend acceptance evidence.

## Real local functional regression

On Linux with Go, npm, Python 3 and an available Docker Engine:

```sh
bash scripts/local-functional-validation.sh
```

The runner installs locked dependencies and builds the SDK/reference packages
and current source, starts an isolated HTTPS Master and two
Daemons, and runs all real functional scenarios serially. It uses the matching
Playwright container and local Docker/Compose, including the bounded 121-sample
monitoring scenario. The fixture includes the actual 10,000-file archive and
16MiB source required by the browser-close/continuation case. It selects the
matching full Chromium binary: the matching headless shell reports native
notifications as denied even after a permission grant in this environment.
No notification constructor or permission result is fabricated. Port 9443 must
be free. It stops only its named container
and fixture process. Owned resource names keep its Docker objects separate from
existing services; do not run global Docker prune/cleanup commands.

Structured evidence and private fixture state stay under a unique `.local`
directory. Publish only sanitized summaries; credentials, databases, raw logs
and real browser snapshots must remain private. The runner returns nonzero for
failed, skipped or flaky functional cases and for an unconfirmed fixture exit.

E08's extreme mixed-workload performance scenario is measured separately. Its
original 50ms target and actual failures remain in the
[accepted bounded closeout](../acceptance/reports/e08-bounded-closeout-2026-10-08.md).
The functional runner excludes that one separately recorded scenario; it does
not change its assertions or claim its old target passed.

## Rejected rendering experiments

From `web`, explicit diagnostics remain runnable:

```sh
npx playwright test --config playwright.experiments.config.ts --workers=1
```

`tests/experiments` retains comparisons for rejected or unadopted rendering
prototypes, including the bar-shadow, alpha-partition, grouped-shadow and
terminal-row adapters that explicitly install test-only implementations. Some
fail strict equivalence checks. Their failure
records remain reviewable; their code is not imported by the shipped UI and
they do not substitute for default product regressions. Diagnostic helpers
under `tests/helpers` and explicit real performance diagnostics have the same
boundary. Their prior failures remain in the optical closeout report. Actual
production window-position, Dock, partially covered shadows and terminal
viewport guards remain in the ordinary suite with their original zero-pixel
assertions; a candidate's failure is not a reason to move a product guard.

See the [acceptance matrix](../acceptance/ACCEPTANCE_MATRIX.md) for current scope,
environment-specific evidence and the explicitly deferred final release.
