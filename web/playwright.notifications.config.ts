import {readFileSync} from 'node:fs'
import {defineConfig} from '@playwright/test'

process.env.PLAYWRIGHT_NO_COPY_PROMPT='1'
if(!process.env.BLORA_E2E_CREDENTIALS||!process.env.DISPLAY||!process.env.DBUS_SESSION_BUS_ADDRESS||process.env.BLORA_NATIVE_NOTIFICATION_E2E!=='1')throw new Error('Native notification tests require an isolated display, session bus and real fixture credentials')
const {url}=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS,'utf8'))
export default defineConfig({
  testDir:'./tests/platform',testMatch:'native-notifications.spec.ts',workers:1,timeout:60000,
  use:{baseURL:url,ignoreHTTPSErrors:true,headless:false,viewport:{width:1280,height:800},
    launchOptions:{args:['--no-sandbox','--disable-dev-shm-usage']},trace:'off',screenshot:'off'},
  reporter:'list',outputDir:process.env.BLORA_NATIVE_OUTPUT||'/tmp/blora-native-notifications',
})
