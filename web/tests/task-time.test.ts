import {describe,it,expect} from 'vitest'
import {taskElapsed} from '../src/services/task-time'
describe('task elapsed time',()=>{
  const task={createdAt:'2026-09-12T00:00:00Z',updatedAt:'2026-09-12T00:01:05Z',state:'SUCCEEDED'}
  it('freezes confirmed terminal durations and includes waiting for active tasks',()=>{
    const now=Date.parse('2026-09-12T02:00:00Z')
    expect(taskElapsed(task,now)).toBe('0小时 1分 5秒')
    expect(taskElapsed({...task,state:'WAITING_NODE'},now)).toBe('2小时 0分 0秒')
  })
  it('does not fabricate a duration for missing or inconsistent timestamps',()=>{
    expect(taskElapsed({...task,updatedAt:undefined},0)).toBe('时间未记录')
    expect(taskElapsed({...task,updatedAt:'2026-09-11T00:00:00Z'},0)).toBe('时间记录不一致')
  })
})
