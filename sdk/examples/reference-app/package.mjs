import {readFile, writeFile} from 'node:fs/promises'
import {createHash} from 'node:crypto'
import {build} from 'esbuild'
import {execFileSync} from 'node:child_process'
import {fileURLToPath} from 'node:url'
const version=process.argv[2]||'0.3.0',schema=Number(process.argv[3]||1),output=process.argv[4]||'reference.blora-extension.json'
if(!/^\d+\.\d+\.\d+$/.test(version)||![1,2,3].includes(schema)||!/^[-a-zA-Z0-9.]+\.json$/.test(output))throw new Error('invalid package version, schema or output filename')

await build({entryPoints: [fileURLToPath(new URL('./src/index.ts', import.meta.url))],
  bundle: true, platform: 'browser', format: 'esm', target: 'es2022',
  define:{REFERENCE_VERSION:JSON.stringify(version),REFERENCE_SCHEMA:String(schema)},
  outfile: fileURLToPath(new URL('./dist/index.js', import.meta.url))})

execFileSync('go',['build','-buildvcs=false','-trimpath','-o',fileURLToPath(new URL('./dist/backend.wasm',import.meta.url)),fileURLToPath(new URL('./backend/main.go',import.meta.url))],
  {env:{...process.env,GOOS:'wasip1',GOARCH:'wasm',CGO_ENABLED:'0'},stdio:'inherit'})
const frontend = await readFile(new URL('./dist/index.js', import.meta.url),'utf8')
const backend = await readFile(new URL('./dist/backend.wasm', import.meta.url))
const payload = Buffer.from('BLORA-BUNDLE-1\n'+JSON.stringify({frontend,backend:backend.toString('base64')}))
const manifest = {
  appId: 'example.reference', packageVersion: version, hostApiVersion: 1,
  title: '参考扩展', icon: '✦', color: '#8ec5ff', permissions: [],
  entrypoints: ['overview'], resourceHandlers: ['instance', 'node'],
  capabilities: ['window.open', 'window.move', 'window.close', 'shortcut.create', 'data.read', 'data.write', 'resource.read', 'resource.write', 'notification.publish', 'task.create'],
  dependencies: {}, windowPolicy: 'multiple',
  tabPolicy: {types: ['overview', 'resource'], movable: true}, stateSchemaVersion: schema,dataSchemaVersion:schema,
}
const pkg = {manifest, sha256: createHash('sha256').update(payload).digest('hex'), payload: payload.toString('base64')}
await writeFile(new URL('./'+output, import.meta.url), JSON.stringify(pkg, null, 2) + '\n', {mode: 0o600})
console.log('wrote '+output)
