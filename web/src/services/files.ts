import {api} from './api'
export interface FileEntry {name:string;path:string;kind?:string;isDir:boolean;size:number;modified?:string;version?:string}
export interface DirectoryPage {items:FileEntry[];version:string;total:number;nextOffset:number}
export interface DirectoryRequest {path:string;offset:number;search:string;sort:string;order:string}
// A node may reduce a requested page to fit the bulk channel's frame budget.
// Fill only the visible block, following its acknowledged nextOffset/version.
export async function readDirectoryBlock(endpoint:string,request:DirectoryRequest,signal:AbortSignal):Promise<DirectoryPage>{
  const items:FileEntry[]=[];let offset=request.offset,version='',total=0,nextOffset=offset
  do{
    const query=new URLSearchParams({...request,offset:String(offset),limit:String(100-items.length),version})
    const page=await api<DirectoryPage>(`${endpoint}?${query}`,{signal})
    if(version&&page.version!==version)throw new Error('目录已改变，请刷新后重新选择文件')
    version=page.version;total=page.total??page.items.length
    if(page.items.length>100-items.length)throw new Error('节点目录分页超出请求范围')
    items.push(...page.items);nextOffset=page.nextOffset??0
    if(nextOffset<=0||nextOffset>=total)break
    if(nextOffset!==offset+page.items.length||nextOffset<=offset)throw new Error('节点目录分页未按已返回的条目推进')
    offset=nextOffset
  }while(items.length<100)
  return{items,version,total,nextOffset}
}
