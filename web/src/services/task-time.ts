import type {Task} from './api'
export const terminalTask=(state:string)=>['SUCCEEDED','FAILED','CANCELLED','INTERRUPTED'].includes(state)
export function taskElapsed(task:Pick<Task,'state'|'createdAt'|'updatedAt'>,now:number):string{
  const start=Date.parse(task.createdAt),end=terminalTask(task.state)?Date.parse(task.updatedAt||''):now
  if(!Number.isFinite(start)||!Number.isFinite(end)||start<=0||end<=0)return '时间未记录'
  if(end<start)return '时间记录不一致'
  const seconds=Math.floor((end-start)/1000)
  return `${Math.floor(seconds/3600)}小时 ${Math.floor(seconds/60)%60}分 ${seconds%60}秒`
}
export function taskTimestamp(value?:string){const time=Date.parse(value||'');return Number.isFinite(time)&&time>0?new Date(time).toLocaleString():'未记录'}
