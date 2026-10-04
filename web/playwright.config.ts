import {defineConfig} from '@playwright/test'
const requestedBrowser=process.env.BLORA_BROWSER||'chromium'
if(!['chromium','firefox','webkit'].includes(requestedBrowser))throw new Error(`不支持的 BLORA_BROWSER: ${requestedBrowser}`)
const browserName=requestedBrowser as 'chromium'|'firefox'|'webkit'
const requestedPort=process.env.BLORA_E2E_PORT
const serverPort=requestedPort===undefined?5173:Number(requestedPort)
if((requestedPort!==undefined&&!/^\d+$/.test(requestedPort))||!Number.isInteger(serverPort)||serverPort<1||serverPort>65535)throw new Error('BLORA_E2E_PORT must be an integer from 1 to 65535')
const serverUrl=`http://127.0.0.1:${serverPort}`
// The Windows WebKit followup checks the shipped bundle. Compile it before
// running tests; on-demand dev transforms must not consume the 45s case limit.
// Keep development serving available for suites that explicitly import /src.
const serverCommand=process.env.BLORA_E2E_PREVIEW==='1'
  ?`node node_modules/vite/bin/vite.js preview --host 127.0.0.1 --port ${serverPort} --strictPort --outDir .local/browser-test-build`
  :`npm run dev -- --mode test --port ${serverPort} --strictPort`
// Bound independent browser processes for this workspace's shared CPU/memory.
// Mixed-window performance acceptance is a separate same-page scenario.
export default defineConfig({testDir:'./tests/browser',timeout:45000,fullyParallel:false,workers:2,use:{baseURL:serverUrl,viewport:{width:1440,height:960},browserName,launchOptions:browserName==='chromium'?{...(process.env.BLORA_CHROMIUM?{executablePath:process.env.BLORA_CHROMIUM}:process.platform==='linux'?{executablePath:'/usr/bin/chromium-browser'}:{}),args:['--no-sandbox','--disable-dev-shm-usage']}:{},trace:'retain-on-failure',screenshot:'only-on-failure'},webServer:{command:serverCommand,url:serverUrl,reuseExistingServer:process.env.BLORA_E2E_FRESH_SERVER!=='1'},reporter:[['list'],['html',{open:'never'}]]})
