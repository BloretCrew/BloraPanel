import { chromium } from '@playwright/test'
import { readFileSync, mkdirSync, writeFileSync } from 'node:fs'

const fixture = JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS, 'utf8'))
const out = '/data/instances/blora-panel/docs/acceptance/screenshots/2026-09-20-material-modern'
mkdirSync(out, { recursive: true })
const browser = await chromium.launch({ executablePath: '/usr/bin/chromium-browser', args: ['--no-sandbox', '--disable-dev-shm-usage'] })
const page = await browser.newPage({ ignoreHTTPSErrors: true, viewport: { width: 1600, height: 1000 }, deviceScaleFactor: 3, locale: 'zh-CN' })
const measurements = []
async function measure(state) {
  await page.waitForTimeout(250)
  const result = await page.evaluate(() => {
    const frame = document.querySelector('.taskbar').getBoundingClientRect()
    const icons = [...document.querySelectorAll('.taskbar .application-art')]
    const first = icons[0].getBoundingClientRect(), last = icons.at(-1).getBoundingClientRect()
    const points = path => [...path.getAttribute('d').matchAll(/[ML](-?[\d.]+) (-?[\d.]+)/g)].map(match => new DOMPoint(Number(match[1]), Number(match[2])).matrixTransform(path.getScreenCTM()))
    const outer = points(document.querySelector('.dock-surface path'))
    const distance = (point, a, b) => {
      const dx = b.x - a.x, dy = b.y - a.y, length = dx * dx + dy * dy
      const t = length ? Math.max(0, Math.min(1, ((point.x - a.x) * dx + (point.y - a.y) * dy) / length)) : 0
      return Math.hypot(point.x - a.x - t * dx, point.y - a.y - t * dy)
    }
    // Measure rendered contours, including both ends' curved portions. This
    // does not import the implementation's shape or layout calculation.
    const distances = [icons[0], icons.at(-1)].flatMap((icon, index) => {
      const rect = icon.getBoundingClientRect()
      return points(icon.querySelector('.app-surface'))
        .filter(point => index ? point.x >= rect.x + rect.width / 2 : point.x <= rect.x + rect.width / 2)
        .map(point => Math.min(...outer.map((a, i) => distance(point, a, outer[(i + 1) % outer.length]))))
    })
    return {
      viewport: innerWidth, iconSize: first.width, dockWidth: frame.width, dockHeight: frame.height,
      margins: { left: first.left - frame.left, top: first.top - frame.top, bottom: frame.bottom - first.bottom, right: frame.right - last.right },
      curveDistance: { min: Math.min(...distances), max: Math.max(...distances), samples: distances.length },
      iconAspectError: Math.max(...icons.map(icon => { const rect = icon.getBoundingClientRect(); return Math.abs(rect.width - rect.height) })),
      horizontalOverflow: Math.max(0, -frame.left, frame.right - innerWidth),
    }
  })
  measurements.push({ state, ...result })
  console.log(JSON.stringify({ state, ...result }))
}
try {
  await page.goto(fixture.url)
  await page.getByRole('textbox', { name: '账号', exact: true }).fill(fixture.admin.name)
  await page.getByLabel('密码', { exact: true }).fill(fixture.admin.password)
  await page.getByRole('button', { name: '进入工作区 →', exact: true }).click()
  await page.locator('.desktop').waitFor()
  await page.mouse.move(1150, 80)
  await measure('desktop')
  await page.locator('.taskbar').screenshot({ path: `${out}/01-dock.png` })
  const dock = await page.locator('.taskbar').boundingBox()
  await page.screenshot({ path: `${out}/02-right-corner.png`, clip: { x: dock.x + dock.width - 150, y: dock.y, width: 150, height: dock.height } })
  await page.screenshot({ path: `${out}/03-left-corner.png`, clip: { x: dock.x, y: dock.y, width: 150, height: dock.height } })
  await page.locator('.taskbar > .dock-item').last().hover()
  await measure('hover')
  await page.locator('.taskbar').getByRole('button', { name: '实例中心', exact: true }).click()
  await page.locator('.app-window').getByRole('button', { name: '关闭窗口', exact: true }).click()
  await page.mouse.move(1150, 80)
  await measure('with-history')
  await page.locator('.taskbar').screenshot({ path: `${out}/04-dock-with-history.png` })
  for (const width of [1280, 760, 390, 320]) {
    await page.setViewportSize({ width, height: 844 })
    await measure(`viewport-${width}`)
    if (width === 390) await page.locator('.taskbar').screenshot({ path: `${out}/05-mobile-dock.png` })
  }
  writeFileSync(`${out}/geometry.json`, JSON.stringify(measurements, null, 2) + '\n')
} finally { await browser.close() }
