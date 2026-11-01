import { concat, positionBytes, toBuffer, utf8 } from './bytes.ts'
import { ALGORITHM, ENVELOPE_TYPE, ENVELOPE_VERSION } from './types.ts'

export async function hmacKey(raw: Uint8Array): Promise<CryptoKey> {
    return crypto.subtle.importKey('raw', toBuffer(raw), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign'])
}

export async function hmacSign(key: CryptoKey, data: Uint8Array): Promise<Uint8Array> {
    return new Uint8Array(await crypto.subtle.sign('HMAC', key, toBuffer(data)))
}

// Общий секрет X25519 -> HKDF-SHA256 на 64 байта: первые 32 - ключ маски, вторые 32 - ключ MAC.
// В info входят тип, версия, алгоритм, отправитель, получатель и хэши обоих словарей
export async function sessionKeys(
    privateKey: CryptoKey,
    publicKey: CryptoKey,
    salt: Uint8Array,
    senderId: string,
    recipientId: string,
    alphabetHash: string,
    alphabetHash2: string,
): Promise<{ maskKey: Uint8Array; macKey: Uint8Array }> {
    const secret = new Uint8Array(await crypto.subtle.deriveBits({ name: 'X25519', public: publicKey }, privateKey, 256))
    const material = await crypto.subtle.importKey('raw', toBuffer(secret), 'HKDF', false, ['deriveBits'])
    const separator = '\u0000'
    const info = utf8(
        `${ENVELOPE_TYPE}/v${ENVELOPE_VERSION}${separator}${ALGORITHM}${separator}${senderId}${separator}${recipientId}${separator}${alphabetHash}${separator}${alphabetHash2}`,
    )
    const derived = new Uint8Array(
        await crypto.subtle.deriveBits({ name: 'HKDF', hash: 'SHA-256', salt: toBuffer(salt), info: toBuffer(info) }, material, 512),
    )

    return { maskKey: derived.slice(0, 32), macKey: derived.slice(32) }
}

// Поток маски для одной позиции: HMAC-SHA256 в режиме счётчика, как _mask_stream в прототипе
export async function maskStream(key: CryptoKey, nonce: Uint8Array, position: number, length: number): Promise<Uint8Array> {
    const prefix = utf8('alphabet-mask-v2\u0000')
    const out = new Uint8Array(Math.ceil(length / 32) * 32)

    for (let counter = 0; counter * 32 < length; counter += 1) {
        out.set(await hmacSign(key, concat(prefix, nonce, positionBytes(position, 8), positionBytes(counter, 4))), counter * 32)
    }

    return out.slice(0, length)
}