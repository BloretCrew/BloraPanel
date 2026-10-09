/** Return a v4 UUID in both secure and ordinary HTTP browser contexts. */
export type UUIDProvider = Pick<Crypto, 'getRandomValues'> & Partial<Pick<Crypto, 'randomUUID'>>

export function createUUID(provider: UUIDProvider): string {
  if (typeof provider.randomUUID === 'function') return provider.randomUUID.call(provider)

  // `getRandomValues` is available in insecure contexts, unlike `randomUUID`.
  const bytes = provider.getRandomValues(new Uint8Array(16))
  bytes[6] = (bytes[6] & 0x0f) | 0x40
  bytes[8] = (bytes[8] & 0x3f) | 0x80
  const hex = Array.from(bytes, value => value.toString(16).padStart(2, '0')).join('')
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`
}

export function randomUUID(): string {
  return createUUID(globalThis.crypto)
}
