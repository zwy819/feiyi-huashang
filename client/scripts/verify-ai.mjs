import { chromium } from 'playwright-core'
import { mkdir } from 'node:fs/promises'
import path from 'node:path'

const chrome = process.env.CHROME_PATH || 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe'
const base = process.env.APP_URL || 'http://localhost:5173'
const outDir = path.resolve('test-results')
await mkdir(outDir, { recursive: true })

const browser = await chromium.launch({
  executablePath: chrome,
  headless: true,
  args: ['--disable-gpu', '--no-sandbox'],
})
const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } })
const result = {}

try {
  await page.goto(base + '/', { waitUntil: 'networkidle' })
  const homeTitle = await page.title()

  await page.getByRole('button', { name: '打开织语 AI 导览' }).click()
  const dialog = page.getByRole('dialog', { name: '织语 AI 导览' })
  await dialog.waitFor({ state: 'visible' })
  await dialog.getByRole('button', { name: /苗族百褶裙/ }).click()
  await page.waitForTimeout(2500)
  console.log('DIALOG_TEXT:' + (await dialog.innerText()).slice(0, 1200))
  await dialog.getByRole('button', { name: /来源\s*\d+/ }).waitFor({ state: 'visible', timeout: 30000 })
  await page.waitForTimeout(800)

  result.home = {
    title: homeTitle,
    citationChips: await dialog.locator('[data-cite]').count(),
    sourceButton: await dialog.getByRole('button', { name: /来源\s*\d+/ }).first().innerText(),
    excerpt: (await dialog.innerText()).slice(0, 500),
  }
  await page.screenshot({ path: path.join(outDir, 'ai-panel-home.png') })

  await dialog.getByRole('button', { name: /来源\s*\d+/ }).first().click()
  await page.getByText('回答来源', { exact: true }).waitFor({ state: 'visible' })
  result.sources = {
    drawerOpen: true,
    locatableHotspots: await page.getByText('3D 可定位', { exact: true }).count(),
  }
  await page.screenshot({ path: path.join(outDir, 'ai-sources.png') })
  await page.getByRole('button', { name: '关闭来源' }).click()

  await page.goto(base + '/detail/163', { waitUntil: 'networkidle' })
  await page.getByRole('button', { name: '问织语' }).click()
  const detailDialog = page.getByRole('dialog', { name: '织语 AI 导览' })
  await detailDialog.waitFor({ state: 'visible' })
  await detailDialog.getByRole('textbox', { name: '输入你的问题' }).fill('这条裙子的裙摆工艺有什么特点？')
  await detailDialog.getByRole('button', { name: '发送', exact: true }).click()

  const action = detailDialog.getByRole('button', { name: /在 3D 中查看/ })
  await action.waitFor({ state: 'visible', timeout: 30000 })
  const actionLabel = await action.innerText()
  await action.click()
  await page.getByText('百褶工艺', { exact: true }).first().waitFor({ state: 'visible', timeout: 10000 })
  await page.waitForTimeout(800)

  result.detail = {
    actionLabel,
    annotationCardVisible: await page.getByText('百褶裙的裙褶多达数百道', { exact: false }).first().isVisible(),
    url: page.url(),
  }
  await page.screenshot({ path: path.join(outDir, 'ai-3d-highlight.png') })

  const mobile = await browser.newPage({ viewport: { width: 390, height: 844 }, isMobile: true })
  await mobile.goto(base + '/', { waitUntil: 'networkidle' })
  await mobile.getByRole('button', { name: '打开织语 AI 导览' }).click()
  const mobileDialog = mobile.getByRole('dialog', { name: '织语 AI 导览' })
  await mobileDialog.waitFor({ state: 'visible' })
  const box = await mobileDialog.boundingBox()
  await mobile.screenshot({ path: path.join(outDir, 'ai-panel-mobile.png') })
  result.mobile = {
    viewport: '390x844',
    panelWidth: box?.width,
    panelHeight: box?.height,
    fullScreen: box ? box.width >= 389 && box.height >= 843 : false,
  }
  await mobile.close()

  console.log(JSON.stringify({ ok: true, result }, null, 2))
} catch (error) {
  try {
    await page.screenshot({ path: path.join(outDir, 'ai-verify-failure.png') })
    result.failureText = (await page.locator('[role="dialog"]').innerText()).slice(0, 1200)
  } catch {}
  console.error(JSON.stringify({ ok: false, error: String(error), result }, null, 2))
  process.exitCode = 1
} finally {
  await browser.close()
}
