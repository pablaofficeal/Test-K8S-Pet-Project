import { fromBase64Url, toBase64Url } from './base64url.ts'
import { equalBytes, hex, toBuffer } from './bytes.ts'
import { KEY_VERSION } from './types.ts'
import type { PrivateKeyDoc, PublicKeyDoc } from './types.ts'

export async function keyIdOf(publicRaw: Uint8Array): Promise<string> {
    const digest = await crypto.subtle.digest('SHA-256', toBuffer(publicRaw))

    return hex(new Uint8Array(digest)).slice(0, 32)
}

export async function generateKeyPair(): Promise<{ privateDoc: PrivateKeyDoc; publicDoc: PublicKeyDoc }> {
    const pair = (await crypto.subtle.generateKey({ name: 'X25519' }, true, ['deriveBits'])) as CryptoKeyPair
    const jwk = await crypto.subtle.exportKey('jwk', pair.privateKey)
    const publicRaw = new Uint8Array(await crypto.subtle.exportKey('raw', pair.publicKey))
    const keyId = await keyIdOf(publicRaw)
    const publicDoc: PublicKeyDoc = { version: KEY_VERSION, algorithm: 'X25519', key_id: keyId, public_key: toBase64Url(publicRaw) }
    const privateRaw = fromBase64Url(jwk.d ?? '', 'private_key')

    return { privateDoc: { ...publicDoc, private_key: toBase64Url(privateRaw) }, publicDoc }
}

export interface LoadedPrivateKey {
    key: CryptoKey
    keyId: string
    publicRaw: Uint8Array
}

export async function loadPrivateKey(doc: PrivateKeyDoc): Promise<LoadedPrivateKey> {
    const keys = Object.keys(doc).sort().join(',')

    if (keys !== 'algorithm,key_id,private_key,public_key,version' || doc.version !== KEY_VERSION || doc.algorithm !== 'X25519') {
        throw new Error('Неподдерживаемый формат приватного ключа')
    }

    const privateRaw = fromBase64Url(doc.private_key, 'private_key')
    const publicRaw = fromBase64Url(doc.public_key, 'public_key')

    if (privateRaw.length !== 32 || publicRaw.length !== 32) {
        throw new Error('X25519-ключ должен содержать 32 байта')
    }

    const key = await crypto.subtle.importKey(
        'jwk',
        { kty: 'OKP', crv: 'X25519', d: doc.private_key, x: doc.public_key, ext: true },
        { name: 'X25519' },
        false,
        ['deriveBits'],
    )
    const keyId = await keyIdOf(publicRaw)

    if (keyId !== doc.key_id) {
        throw new Error('Приватный ключ повреждён')
    }

    return { key, keyId, publicRaw }
}

export interface LoadedPublicKey {
    key: CryptoKey
    keyId: string
}

export async function loadPublicKey(doc: PublicKeyDoc): Promise<LoadedPublicKey> {
    const keys = Object.keys(doc).sort().join(',')

    if (keys !== 'algorithm,key_id,public_key,version' || doc.version !== KEY_VERSION || doc.algorithm !== 'X25519') {
        throw new Error('Неподдерживаемый формат публичного ключа')
    }

    const publicRaw = fromBase64Url(doc.public_key, 'public_key')

    if (publicRaw.length !== 32) {
        throw new Error('X25519-ключ должен содержать 32 байта')
    }

    const keyId = await keyIdOf(publicRaw)

    if (!equalBytes(new TextEncoder().encode(keyId), new TextEncoder().encode(doc.key_id))) {
        throw new Error('Публичный ключ повреждён')
    }

    const key = await crypto.subtle.importKey('raw', toBuffer(publicRaw), { name: 'X25519' }, false, [])

    return { key, keyId }
}