// One continuous corner for app artwork and its surrounding surfaces. These
// cubics join with continuous curvature, including zero curvature at the edges.
// Material You changes the colour planes, while preserving the established
// icon dimensions and constant-distance Dock inset.
type Point = readonly [number, number]
const segments: readonly (readonly [Point, Point, Point, Point])[] = [
  [[0, 0], [.36, 0], [.68, 0], [.84, .16]],
  [[.84, .16], [1, .32], [1, .64], [1, 1]],
]
export const APP_CORNER_RATIO = 30 / 80

// Cache the curve and unit outward normals. Offsetting along these normals
// keeps the whole corner equidistant; scaling another squircle would not.
const corner = segments.flatMap(([a, b, c, d], segment) =>
  Array.from({ length: 49 }, (_, step) => {
    const t = step / 48, u = 1 - t
    const x = u ** 3 * a[0] + 3 * u ** 2 * t * b[0] + 3 * u * t ** 2 * c[0] + t ** 3 * d[0]
    const y = u ** 3 * a[1] + 3 * u ** 2 * t * b[1] + 3 * u * t ** 2 * c[1] + t ** 3 * d[1]
    const dx = 3 * u ** 2 * (b[0] - a[0]) + 6 * u * t * (c[0] - b[0]) + 3 * t ** 2 * (d[0] - c[0])
    const dy = 3 * u ** 2 * (b[1] - a[1]) + 6 * u * t * (c[1] - b[1]) + 3 * t ** 2 * (d[1] - c[1])
    const speed = Math.hypot(dx, dy)
    return { x, y, nx: dy / speed, ny: -dx / speed }
  }).slice(segment ? 1 : 0),
)

/** A continuous rectangle, with a true normal offset from its base contour. */
export function continuousRectPath(width: number, height: number, extent: number, offset = 0, x = 0, y = 0): string {
  const r = Math.min(extent, width / 2, height / 2)
  const points: Point[] = []
  for (let side = 0; side < 4; side++) {
    for (const point of corner) {
      const px = r * point.x + offset * point.nx, py = r * point.y + offset * point.ny
      const rotated: Point = side === 0 ? [width - r + px, py]
        : side === 1 ? [width - py, height - r + px]
        : side === 2 ? [r - px, height - py]
        : [py, r - px]
      points.push([rotated[0] + x, rotated[1] + y])
    }
  }
  return points.map(([px, py], i) => `${i ? 'L' : 'M'}${px.toFixed(3)} ${py.toFixed(3)}`).join(' ') + 'Z'
}

export const APP_SURFACE_PATH = continuousRectPath(80, 80, 80 * APP_CORNER_RATIO, 0, 8, 8)
