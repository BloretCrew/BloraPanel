import {readFileSync} from 'node:fs'
import {defineConfig} from '@playwright/test'
// This pinned Playwright version collects an AI aria page snapshot separately
// from trace and screenshot. Real account tests must not persist field values.
process.env.PLAYWRIGHT_NO_COPY_PROMPT='1'
if(!process.env.BLORA_E2E_CREDENTIALS)throw new Error('需要主代理启动真实双节点 fixture，并设置 BLORA_E2E_CREDENTIALS 私有文件路径')
const {url}=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS,'utf8'))
const requestedBrowser=process.env.BLORA_BROWSER||'chromium'
if(!['chromium','firefox','webkit'].includes(requestedBrowser))throw new Error(`不支持的 BLORA_BROWSER: ${requestedBrowser}`)
const browserName=requestedBrowser as 'chromium'|'firefox'|'webkit'
// Local runners may select a system Chromium explicitly. In the official
// Playwright container no such system path exists, so let the pinned package
// resolve its matching downloaded browser instead.
const chromiumLaunch=browserName==='chromium'?{launchOptions:{...(process.env.BLORA_CHROMIUM?{executablePath:process.env.BLORA_CHROMIUM}:{}),args:['--no-sandbox','--disable-dev-shm-usage',...(process.env.BLORA_PERF_SOFTWARE_COMPOSITING==='1'?['--disable-gpu']:[])]}}:{}
export default defineConfig({testDir:'./tests/real',timeout:60000,fullyParallel:false,workers:1,use:{baseURL:url,ignoreHTTPSErrors:true,viewport:{width:1440,height:960},browserName,...chromiumLaunch,trace:'off',screenshot:'off'},outputDir:'real-test-results',reporter:[['list'],['html',{outputFolder:'real-playwright-report',open:'never'}]]})
