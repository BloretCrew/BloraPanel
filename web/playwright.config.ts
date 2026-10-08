import {defineConfig} from '@playwright/test'
import {browserLaunchOptions} from './playwright-browser'
const requestedBrowser=process.env.BLORA_BROWSER||'chromium'
if(!['chromium','firefox','webkit'].includes(requestedBrowser))throw new Error(`不支持的 BLORA_BROWSER: ${requestedBrowser}`)
const browserName=requestedBrowser as 'chromium'|'firefox'|'webkit'
const requestedPort=process.env.BLORA_E2E_PORT
const serverPort=requestedPort===undefined?5173:Number(requestedPort)
if((requestedPort!==undefined&&!/^\d+$/.test(requestedPort))||!Number.isInteger(serverPort)||serverPort<1||serverPort>65535)throw new Error('BLORA_E2E_PORT must be an integer from 1 to 65535')
const serverUrl=`http://127.0.0.1:${serverPort}`
const nativeScale=process.env.BLORA_E08_NATIVE_DEVICE_SCALE
if(nativeScale!==undefined&&(!['1','1.25'].includes(nativeScale)||browserName!=='chromium'))throw new Error('Native device-scale control requires Chromium and DPR 1 or 1.25')
const nativeProjects=browserName==='chromium'&&nativeScale===undefined&&process.env.BLORA_E08_EMULATED_SCALE_ONLY!=='1'
 ?[{name:'chromium',grepInvert:/DPR /},...[1,1.25].map(scale=>({name:`chromium-native-dpr-${scale}`,grep:scale===1?/DPR 1(?![.\d])/:/DPR 1\.25(?![.\d])/,use:{deviceScaleFactor:scale,launchOptions:browserLaunchOptions(browserName,scale)}}))]
 :undefined
// The Windows WebKit followup checks the shipped bundle. Compile it before
// running tests; on-demand dev transforms must not consume the 45s case limit.
// Keep development serving available for suites that explicitly import /src.
const serverCommand=process.env.BLORA_E2E_PREVIEW==='1'
  ?`node node_modules/vite/bin/vite.js preview --host 127.0.0.1 --port ${serverPort} --strictPort --outDir .local/browser-test-build`
  :`npm run dev -- --mode test --port ${serverPort} --strictPort`
// Bound independent browser processes for this workspace's shared CPU/memory.
// Mixed-window performance acceptance is a separate same-page scenario.
export default defineConfig({testDir:'./tests/browser',timeout:45000,fullyParallel:false,workers:2,projects:nativeProjects,use:{baseURL:serverUrl,viewport:{width:1440,height:960},browserName,launchOptions:browserLaunchOptions(browserName,nativeScale),trace:'retain-on-failure',screenshot:'only-on-failure'},webServer:{command:serverCommand,url:serverUrl,reuseExistingServer:process.env.BLORA_E2E_FRESH_SERVER!=='1'},reporter:[['list'],['html',{open:'never'}]]})
