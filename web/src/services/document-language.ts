const extensions:Record<string,string>={
  abap:'abap',bat:'bat',bicep:'bicep',c:'cpp',cc:'cpp',cpp:'cpp',css:'css',cs:'csharp',dart:'dart',dockerfile:'dockerfile',go:'go',graphql:'graphql',h:'cpp',hpp:'cpp',html:'html',htm:'html',ini:'ini',java:'java',js:'javascript',jsx:'javascript',mjs:'javascript',cjs:'javascript',json:'json',jsonc:'json',kt:'kotlin',kts:'kotlin',less:'less',lua:'lua',md:'markdown',markdown:'markdown',php:'php',ps1:'powershell',py:'python',python:'python',r:'r',rb:'ruby',rs:'rust',scss:'scss',sh:'shell',bash:'shell',zsh:'shell',sql:'sql',swift:'swift',tf:'hcl',toml:'ini',ts:'typescript',tsx:'typescript',xml:'xml',svg:'xml',yaml:'yaml',yml:'yaml'
}

export function documentLanguage(path?:string){
  const name=(path||'').replaceAll('\\','/').split('/').at(-1)?.toLowerCase()||''
  if(name==='dockerfile'||name.startsWith('dockerfile.'))return 'dockerfile'
  if(name==='.env'||name.startsWith('.env.'))return 'ini'
  const extension=name.includes('.')?name.slice(name.lastIndexOf('.')+1):name
  return extensions[extension]||'plaintext'
}
