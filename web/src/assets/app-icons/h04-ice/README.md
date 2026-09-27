# Cold white / ice blue appearance preview

Exported on 2026-09-23 using the existing h04 renderer and the opt-in palette in
`web/src/app-host/icon-ice-palette.ts`. Original glyph geometry, backing contour,
material parameters and optical shell remain unchanged.

- Reproduce with `node .local/export-selected-icons.mjs --ice` from `web/`, with Vite on 5173.
- Parameters: h04, lighting `[0,0]`, `bare=false`, `padded=true`, `probe=plain`.
- 14 PNG assets, 288 × 288 each, 182,848 bytes total.
- Displayed only by the development preview `/?appearance=ice`.
- The standard h04 assets remain the default. These files come directly from
  the code renderer; no image generation, hand-painted changes or diagonal sheen.
