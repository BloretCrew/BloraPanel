import {describe,it,expect} from 'vitest'
import type {useDesktop} from '../src/desktop/store'
import {receiveTaskNotification} from '../src/services/task-notifications'
describe('task notification checkpoints',()=>{
  it('does not acknowledge successful notification storage when the recovery journal failed',()=>{
    const state={preferences:{taskEventCursor:10,notifications:[]},views:{}}
    const desktop={state,recovery:{status:{protected:false,message:'journal failed'}},commit:()=>{}} as unknown as ReturnType<typeof useDesktop>
    expect(()=>receiveTaskNotification(desktop,{sequence:11,task:{taskId:'t',state:'SUCCEEDED',action:'file.mkdir'}})).toThrow('journal failed')
  })
  it('commits notification and cursor together and never recreates dismissed replay',()=>{
    const state={preferences:{taskEventCursor:10,notifications:[] as unknown[]},views:{}}
    const batches:any[][]=[]
    const desktop={state,commit:(ops:any[])=>{batches.push(ops);for(const op of ops)(state.preferences as any)[op.path[1]]=op.value}} as unknown as ReturnType<typeof useDesktop>
    const event={sequence:11,task:{taskId:'t',state:'FAILED',action:'instance.start'}}
    expect(receiveTaskNotification(desktop,event)).toBe(true)
    expect(batches).toHaveLength(1);expect(batches[0]).toHaveLength(2)
    expect(state.preferences.taskEventCursor).toBe(11)
    state.preferences.notifications=[]
    expect(receiveTaskNotification(desktop,event)).toBe(false)
    expect(state.preferences.notifications).toHaveLength(0)
  })
})
