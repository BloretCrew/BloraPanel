import {defineConfig} from '@playwright/test'
const requestedBrowser=process.env.BLORA_BROWSER||'chromium'
if(!['chromium','firefox','webkit'].includes(requestedBrowser))throw new Error(`不支持的 BLORA_BROWSER: ${requestedBrowser}`)
const browserName=requestedBrowser as 'chromium'|'firefox'|'webkit'
// Bound independent browser processes for this workspace's shared CPU/memory.
// Mixed-window performance acceptance is a separate same-page scenario.
export default defineConfig({testDir:'./tests/browser',timeout:45000,fullyParallel:false,workers:2,use:{baseURL:'http://127.0.0.1:5173',viewport:{width:1440,height:960},browserName,launchOptions:browserName==='chromium'?{...(process.env.BLORA_CHROMIUM?{executablePath:process.env.BLORA_CHROMIUM}:process.platform==='linux'?{executablePath:'/usr/bin/chromium-browser'}:{}),args:['--no-sandbox','--disable-dev-shm-usage']}:{},trace:'retain-on-failure',screenshot:'only-on-failure'},webServer:{command:'npm run dev -- --mode test',url:'http://127.0.0.1:5173',reuseExistingServer:process.env.BLORA_E2E_FRESH_SERVER!=='1'},reporter:[['list'],['html',{open:'never'}]]})
