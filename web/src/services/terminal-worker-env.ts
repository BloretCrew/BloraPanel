// xterm 6.0.0's shared task queue probes `window.requestIdleCallback` even
// when its headless parser runs in a Worker. There is no DOM dependency: use
// the Worker global for that feature probe, which selects its timer fallback.
// This module must run before the headless module is evaluated.
Object.defineProperty(globalThis,'window',{value:globalThis,configurable:true})
