# Selected application artwork

User-approved `h04` material with the backing tile retained, exported unchanged
from the code renderer on 2026-09-22. No image generation or image editing was used.

- Source geometry: `web/src/app-host/classic-icon-layers.ts`.
- Material: `h04` in `web/src/app-host/icon-optics.ts`.
- Renderer: `web/src/app-host/icon-optics-renderer.ts`.
- Export: `web/.local/export-selected-icons.mjs` with a local Vite server on 5173.
- Parameters: `light=[0,0]`, `bare=false`, `padded=true`, `probe=plain`.
- Each PNG is 288 × 288 with transparency; all 14 images total 185,447 bytes.
- The shared SVG wrapper preserves the existing padded/unpadded viewBox and Dock geometry.

The default interface loads these assets without initializing WebGL. Interactive
material candidates remain development-only. After an approved material change,
regenerate this set and verify the production build with
`web/.local/verify-selected-icons.mjs` against a local Vite preview on 5174.
