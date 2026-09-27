import {defineConfig} from '@playwright/test'
// Bound independent browser processes for this workspace's shared CPU/memory.
// Mixed-window performance acceptance is a separate same-page scenario.
export default defineConfig({testDir:'./tests/browser',timeout:45000,fullyParallel:false,workers:2,use:{baseURL:'http://127.0.0.1:5173',viewport:{width:1440,height:960},launchOptions:{executablePath:process.env.BLORA_CHROMIUM || '/usr/bin/chromium-browser',args:['--no-sandbox','--disable-dev-shm-usage']},trace:'retain-on-failure',screenshot:'only-on-failure'},webServer:{command:'npm run dev -- --mode test',url:'http://127.0.0.1:5173',reuseExistingServer:true},reporter:[['list'],['html',{open:'never'}]]})
