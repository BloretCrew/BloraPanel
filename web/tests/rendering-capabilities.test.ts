import {describe,it,expect} from 'vitest'
import {nativePaintFallback} from '../src/services/rendering-capabilities'

describe('native renderer fallback',()=>{
 it('rejects unavailable and named software contexts without changing identified hardware',()=>{
  for(const renderer of [undefined,'SwiftShader Device','llvmpipe (LLVM)','Microsoft Basic Render Driver'])expect(nativePaintFallback('Firefox/155.0',renderer)).toBe(true)
  for(const renderer of ['AMD Radeon','Intel UHD Graphics','Apple GPU',''])expect(nativePaintFallback('Firefox/155.0',renderer)).toBe(false)
 })
 it('treats the fixed Linux WebKit identifier as unqualified, preserving other platforms and engines',()=>{
  expect(nativePaintFallback('X11; Linux x86_64 AppleWebKit/605.1.15','Apple GPU')).toBe(true)
  expect(nativePaintFallback('Macintosh AppleWebKit/605.1.15','Apple GPU')).toBe(false)
  expect(nativePaintFallback('X11; Linux x86_64 AppleWebKit/537.36 Chrome/153.0','Apple GPU')).toBe(false)
  expect(nativePaintFallback('X11; Linux x86_64 AppleWebKit/605.1.15','AMD Radeon')).toBe(false)
 })
 it('handles a Safari UA from a Linux WebKit port without changing actual Apple hardware',()=>{
  const safari='Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Safari/605.1.15'
  expect(nativePaintFallback(safari,'Apple GPU','Linux x86_64')).toBe(true)
  expect(nativePaintFallback(safari,'Apple GPU','MacIntel')).toBe(false)
  expect(nativePaintFallback(safari,'AMD Radeon','Linux x86_64')).toBe(false)
 })
})
