import {describe,it,expect} from 'vitest'
import {offsetWindowRenderer} from '../src/desktop/window-position'

describe('native software window positioning',()=>{
 it('keeps supported hardware and unidentified contexts on their original transform renderer',()=>{
  for(const ua of ['Firefox/155.0','AppleWebKit/605.1.15'])for(const renderer of ['', 'Apple GPU','AMD Radeon','Intel UHD Graphics'])expect(offsetWindowRenderer(ua,renderer)).toBe(false)
 })
 it('uses native offsets for software or unavailable contexts without applying to Chromium',()=>{
  for(const ua of ['Firefox/155.0','AppleWebKit/605.1.15'])for(const renderer of [undefined,'llvmpipe (LLVM)','SwiftShader Device','Microsoft Basic Render Driver'])expect(offsetWindowRenderer(ua,renderer)).toBe(true)
  for(const ua of ['AppleWebKit/537.36 Chrome/153.0 Safari/537.36','AppleWebKit/537.36 Edg/153.0','unknown'])expect(offsetWindowRenderer(ua,undefined)).toBe(false)
 })
 it('does not mistake Linux WebKit\'s fixed identifier for a hardware qualification',()=>{
  expect(offsetWindowRenderer('Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/605.1.15 Safari/605.1.15','Apple GPU')).toBe(true)
  expect(offsetWindowRenderer('Mozilla/5.0 (Macintosh) AppleWebKit/605.1.15 Safari/605.1.15','Apple GPU')).toBe(false)
  expect(offsetWindowRenderer('Mozilla/5.0 (Macintosh) AppleWebKit/605.1.15 Safari/605.1.15','Apple GPU','Linux x86_64')).toBe(true)
  expect(offsetWindowRenderer('Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/153.0','Apple GPU')).toBe(false)
 })
})
