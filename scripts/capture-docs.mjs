import { mkdirSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import puppeteer from 'puppeteer-core'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const outDir = resolve(root, 'docs')
mkdirSync(outDir, { recursive: true })

const browser = await puppeteer.launch({
  executablePath: process.env.CHROME || 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe',
  headless: true,
  defaultViewport: { width: 1440, height: 900, deviceScaleFactor: 2 },
  args: ['--hide-scrollbars'],
})

const page = await browser.newPage()
await page.goto('http://localhost:5173/login', { waitUntil: 'networkidle0', timeout: 20000 })
await page.waitForSelector('.hero')
await page.screenshot({ path: resolve(outDir, 'login.png'), fullPage: true })

await page.click('button.btn-ghost')
await page.waitForSelector('.grid', { timeout: 15000 })
await page.waitForSelector('.list li')
await new Promise((r) => setTimeout(r, 600))
await page.screenshot({ path: resolve(outDir, 'today.png'), fullPage: true })

const minutesBtn = await page.$('.minutes .btn')
if (minutesBtn) {
  await minutesBtn.click()
  await page.waitForSelector('.result', { timeout: 10000 })
  await new Promise((r) => setTimeout(r, 400))
  await page.screenshot({ path: resolve(outDir, 'minutes.png'), fullPage: true })
}

await browser.close()
console.log('wrote docs/login.png docs/today.png docs/minutes.png')
