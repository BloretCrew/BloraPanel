import { chromium } from '@playwright/test'
import { readFileSync, writeFileSync, mkdirSync, existsSync, renameSync } from 'node:fs'

const fixture = JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS, 'utf8'))
const stage = process.env.BLORA_LAYOUT_STAGE
if (!['before', 'after'].includes(stage)) throw new Error('BLORA_LAYOUT_STAGE must be before or after')
const out = '/data/instances/blora-panel/docs/acceptance/screenshots/2026-09-20-material-modern'
mkdirSync(out, { recursive: true })
const browser = await chromium.launch({ executablePath: '/usr/bin/chromium-browser', args: ['--no-sandbox', '--disable-dev-shm-usage'] })
const page = await browser.newPage({ ignoreHTTPSErrors: true, viewport: { width: 1600, height: 1000 }, locale: 'zh-CN' })
const measurements = {}
async function measure(name, selector) {
  const target = page.locator(selector)
  await target.first().waitFor()
  // Motion uses a JS-driven transform; take the layout after its 150ms entry
  // animation, rather than comparing unrelated frames of that animation.
  await page.waitForFunction(() => {
    const launcher = document.querySelector('.launcher')
    return !launcher || Math.abs(new DOMMatrixReadOnly(getComputedStyle(launcher).transform).m42) < .001
  })
  measurements[name] = await target.evaluateAll(items => items.map(item => {
    const { x, y, width, height } = item.getBoundingClientRect()
    return { x, y, width, height }
  }))
}
try {
  await page.goto(fixture.url)
  if (stage === 'before' && process.env.BLORA_LAYOUT_BASELINE_CSS) {
    // Replay the exact styles saved before this change without rebuilding or
    // changing the running product. The current DOM and app behaviour match.
    const content = readFileSync(process.env.BLORA_LAYOUT_BASELINE_CSS, 'utf8')
    const style = await page.addStyleTag({ content })
    await style.evaluate(element => {
      const prioritize = rules => {
        for (const rule of rules) {
          if (rule.selectorText && rule.selectorText !== ':root') rule.selectorText = `#app :is(${rule.selectorText})`
          if (rule.cssRules) prioritize(rule.cssRules)
        }
      }
      prioritize(element.sheet.cssRules)
    })
  }
  await page.getByRole('textbox', { name: '账号', exact: true }).fill(fixture.admin.name)
  await page.getByLabel('密码', { exact: true }).fill(fixture.admin.password)
  await page.getByRole('button', { name: '进入工作区 →', exact: true }).click()
  await page.locator('.desktop').waitFor()
  await measure('topbar', '.topbar')
  await measure('dock', '.taskbar')
  await measure('desktopIcons', '.desktop-apps .desktop-icon')
  await page.locator('.launcher-button').click()
  await measure('launcher', '.launcher')
  await measure('launcherIcons', '.launcher-grid > button')
  await page.locator('.launcher').getByRole('button', { name: '实例中心', exact: true }).click()
  await page.locator('.instance-card').first().waitFor()
  await measure('window', '.app-window.focused')
  await measure('titlebar', '.app-window.focused .window-titlebar')
  await measure('tabs', '.app-window.focused .view-tabs')
  await measure('sidebar', '.app-window.focused .resource-sidebar')
  await measure('content', '.app-window.focused .resource-main')
  await page.setViewportSize({ width: 390, height: 844 })
  await page.waitForTimeout(300)
  await measure('mobileWindow', '.app-window.focused')
  await measure('mobileTopbar', '.topbar')
  await measure('mobileDock', '.taskbar')
  if (stage === 'before' && existsSync(`${out}/layout-before.json`) && !existsSync(`${out}/layout-before-animated.json`)) {
    renameSync(`${out}/layout-before.json`, `${out}/layout-before-animated.json`)
    if (existsSync(`${out}/layout-after.json`)) renameSync(`${out}/layout-after.json`, `${out}/layout-after-animated.json`)
    if (existsSync(`${out}/layout-comparison.json`)) renameSync(`${out}/layout-comparison.json`, `${out}/layout-comparison-animated.json`)
  }
  writeFileSync(`${out}/layout-${stage}.json`, JSON.stringify(measurements, null, 2) + '\n')
  if (stage === 'after') {
    const baseline = JSON.parse(readFileSync(`${out}/layout-before.json`, 'utf8'))
    const differences = []
    for (const [name, rects] of Object.entries(measurements)) {
      if (rects.length !== baseline[name]?.length) differences.push(`${name}: item count changed`)
      rects.forEach((rect, i) => Object.entries(rect).forEach(([key, value]) => {
        if (Math.abs(value - (baseline[name]?.[i]?.[key] ?? Infinity)) > .5) differences.push(`${name}[${i}].${key}: ${baseline[name]?.[i]?.[key]} → ${value}`)
      }))
    }
    writeFileSync(`${out}/layout-comparison.json`, JSON.stringify({ groups: Object.keys(measurements).length, toleranceCssPx: .5, differences }, null, 2) + '\n')
    if (differences.length) throw new Error(differences.join('\n'))
    console.log('Layout unchanged: 13 desktop/mobile groups, within 0.5 CSS px')
  } else console.log('Baseline recorded: 13 desktop/mobile groups')
} finally { await browser.close() }
