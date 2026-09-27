import {describe,expect,it} from 'vitest'
import {documentLanguage} from '../src/services/document-language'

describe('documentLanguage',()=>{
  const cases:Array<[string,string]>=[
    ['src/main.go','go'],['src/view.tsx','typescript'],['settings.JSON','json'],['compose.yaml','yaml'],['config.toml','ini'],['.env.production','ini'],['Dockerfile.dev','dockerfile'],['README.md','markdown'],['startup.sh','shell'],['payload.unknown','plaintext'],['notes','plaintext']
  ]
  it.each(cases)('maps %s to the registered Monaco grammar %s',(path,expected)=>{
    expect(documentLanguage(path)).toBe(expected)
  })
})
