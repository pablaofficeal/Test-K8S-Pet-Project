import { fromBase64Url, toBase64Url } from './crypto/base64url.ts'
import { toBuffer, utf8 } from './crypto/bytes.ts'
import type { PrivateKeyDoc } from './crypto/types.ts'

export const BACKUP_ITERATIONS = 600_000
const MAX_ITERATIONS = 5_000_000
const AAD = utf8('cadence-messenger-key-backup-v1')
export const WRONG_PASSPHRASE = 'Неверная парольная фраза или копия повреждена'

async function wrapKeyFor(passphrase: string, salt: Uint8Array, iterations: number): Promise<CryptoKey> {
    const base = await crypto.subtle.importKey('raw', toBuffer(utf8(passphrase)), 'PBKDF2', false, ['deriveKey'])

    return crypto.subtle.deriveKey(
        { name: 'PBKDF2', hash: 'SHA-256', salt: toBuffer(salt), iterations },
        base,
        { name: 'AES-GCM', length: 256 },
        false,
        ['encrypt', 'decrypt'],
    )
}

// Копия приватного ключа для сервера: PBKDF2-SHA256 -> AES-256-GCM. Сервер хранит непрозрачную строку и прочитать её не может
export async function wrapKey(doc: PrivateKeyDoc, passphrase: string, iterations = BACKUP_ITERATIONS): Promise<string> {
    if (passphrase.length < 8) {
        throw new Error('Парольная фраза короче 8 символов')
    }

    const salt = crypto.getRandomValues(new Uint8Array(16))
    const iv = crypto.getRandomValues(new Uint8Array(12))
    const key = await wrapKeyFor(passphrase, salt, iterations)
    const encrypted = await crypto.subtle.encrypt({ name: 'AES-GCM', iv: toBuffer(iv), additionalData: toBuffer(AAD) }, key, toBuffer(utf8(JSON.stringify(doc))))

    return JSON.stringify({
        v: 1,
        kdf: 'PBKDF2-SHA256',
        iter: iterations,
        salt: toBase64Url(salt),
        iv: toBase64Url(iv),
        ct: toBase64Url(new Uint8Array(encrypted)),
    })
}

export async function unwrapKey(blob: string, passphrase: string): Promise<PrivateKeyDoc> {
    try {
        const parsed = JSON.parse(blob) as Record<string, unknown>
        const iterations = parsed.iter

        if (parsed.v !== 1 || parsed.kdf !== 'PBKDF2-SHA256' || typeof iterations !== 'number' || !Number.isInteger(iterations) || iterations < 1 || iterations > MAX_ITERATIONS) {
            throw new Error('format')
        }

        const salt = fromBase64Url(String(parsed.salt), 'salt')
        const iv = fromBase64Url(String(parsed.iv), 'iv')
        const ciphertext = fromBase64Url(String(parsed.ct), 'ct')
        const key = await wrapKeyFor(passphrase, salt, iterations)
        const plain = await crypto.subtle.decrypt({ name: 'AES-GCM', iv: toBuffer(iv), additionalData: toBuffer(AAD) }, key, toBuffer(ciphertext))

        return JSON.parse(new TextDecoder().decode(plain)) as PrivateKeyDoc
    } catch {
        throw new Error(WRONG_PASSPHRASE)
    }
}