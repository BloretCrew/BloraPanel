import type {ResourceRef} from '../app-host/types'
export interface RoleTemplate {roleId:string;name:string;actions:string[];nodeOnly:boolean;builtin:boolean;revision:number}
export interface Grant {userId:string;resource:ResourceRef;action:string}
export const permissionLabels:Record<string,string>={
  'instance.read':'查看实例','instance.create':'创建实例','instance.start':'启动实例','instance.stop':'停止实例','instance.restart':'重启实例','instance.kill':'强制结束实例','instance.configure':'修改实例配置',
  'file.read':'读取文件','file.write':'修改文件','terminal.read':'查看终端','terminal.input':'终端输入','node.read':'查看节点','host.manage':'管理宿主机',
  'backup.create':'创建备份','backup.restore':'恢复备份','schedule.create':'创建计划','app.use':'使用扩展应用',
}
export const nodeActions=new Set(['instance.create','host.manage','node.read'])
export function passwordProblem(value:string,confirmation:string){const bytes=new TextEncoder().encode(value).byteLength;if(bytes<12||bytes>72)return '新密码必须为12～72字节';if(value!==confirmation)return '两次输入的新密码不一致';return ''}
