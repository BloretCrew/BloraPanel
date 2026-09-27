// This file is the only executable code loaded directly by an extension
// sandbox document. The parent supplies the already authenticated bundle via
// postMessage; the bundle itself is then imported from a Blob URL created in
// this opaque iframe, never from the parent document.
(() => {
  const pending = new Map()
  let sequence = 0
  const call = (method, payload) => new Promise((resolve, reject) => {
    const id = String(++sequence)
    pending.set(id, {resolve, reject})
    parent.postMessage({type: 'blora-extension-call', id, method, payload}, '*')
  })
  addEventListener('message', async event => {
    if (event.data?.type === 'blora-extension-result') {
      const item = pending.get(event.data.id)
      if (!item) return
      pending.delete(event.data.id)
      if (event.data.ok) item.resolve(event.data.value)
      else item.reject(new Error(event.data.value || '宿主调用失败'))
      return
    }
    if (event.data?.type !== 'blora-extension-init' || typeof event.data.bundle !== 'string') return
    const url = URL.createObjectURL(new Blob([event.data.bundle], {type: 'text/javascript'}))
    try {
      const module = await import(url)
      const original = await call('state.capture', null)
      const targetVersion = event.data.stateSchemaVersion || 1
      if (original.schemaVersion !== targetVersion) {
        if (typeof module.migrate !== 'function') throw new Error('扩展未提供状态迁移；原始现场已保留')
        let timer
        try {
          const candidate = await Promise.race([
            Promise.resolve().then(() => module.migrate(JSON.parse(JSON.stringify(original)), targetVersion)),
            new Promise((_, reject) => {timer = setTimeout(() => reject(new Error('扩展迁移超时；原始现场已保留')), 10000)}),
          ])
          await call('state.migrate', {from: original, to: candidate})
        } finally {clearTimeout(timer)}
      }
      if (module.start) await module.start(call)
    } catch (error) {
      document.body.insertAdjacentText('beforeend', String(error))
      parent.postMessage({type:'blora-extension-load-error',message:String(error)},'*')
    } finally {
      URL.revokeObjectURL(url)
    }
  })
  parent.postMessage({type: 'blora-extension-ready'}, '*')
})()
