import { fromBase64Url, toBase64Url } from './base64url.ts'
import { concat, equalBytes, utf8 } from './bytes.ts'
import { hmacKey, hmacSign, sessionKeys } from './keySchedule.ts'
import { loadPrivateKey, loadPublicKey } from './keys.ts'
import { addDualNoise, removeDualNoise } from './noise.ts'
import type { DualAlphabets } from './alphabet.ts'
import { ALGORITHM, ENVELOPE_TYPE, ENVELOPE_VERSION, MAX_TEXT_BYTES } from './types.ts'
import type { Envelope, PrivateKeyDoc, PublicKeyDoc, RandomBytes } from './types.ts'

const REQUIRED_KEYS = [
    'type', 'version', 'algorithm', 'alphabet_hash', 'alphabet_hash2', 'sender_key_id',
    'recipient_key_id', 'salt', 'nonce', 'ciphertext', 'auth_tag',
]
const TAMPERED = 'Неверные ключи или сообщение было изменено'

const systemRandom: RandomBytes = (length) => crypto.getRandomValues(new Uint8Array(length))

// Метаданные для MAC: те же поля и тот же канонический JSON, что в прототипе
function metadata(envelope: Envelope): Uint8Array {
    const names = ['type', 'version', 'algorithm', 'alphabet_hash', 'alphabet_hash2', 'sender_key_id', 'recipient_key_id', 'salt', 'nonce'] as const
    const data: Record<string, unknown> = {}

    for (const name of [...names].sort()) {
        data[name] = envelope[name]
    }

    return utf8(JSON.stringify(data))
}

export function validateDescriptor(value: unknown): Envelope {
    if (value === null || typeof value !== 'object' || Array.isArray(value)) {
        throw new Error('Нарушен дескриптор сообщения: ожидается объект')
    }

    const envelope = value as Record<string, unknown>

    if (Object.keys(envelope).length !== REQUIRED_KEYS.length || REQUIRED_KEYS.some((key) => !(key in envelope))) {
        throw new Error('Нарушен дескриптор сообщения: лишний или недостающий ключ')
    }

    if (envelope.type !== ENVELOPE_TYPE) {
        throw new Error('Нарушен дескриптор сообщения: неверный type')
    }

    if (envelope.version !== ENVELOPE_VERSION) {
        throw new Error('Устаревшая или неподдерживаемая версия сообщения, расшифровка запрещена')
    }

    if (envelope.algorithm !== ALGORITHM) {
        throw new Error('Алгоритм сообщения запрещён политикой, возможен даунгрейд')
    }

    for (const field of ['alphabet_hash', 'alphabet_hash2', 'salt', 'nonce', 'auth_tag']) {
        const text = envelope[field]

        if (typeof text !== 'string' || text === '' || /\s/u.test(text)) {
            throw new Error(`Нарушен дескриптор сообщения: поле ${field}`)
        }
    }

    for (const field of ['sender_key_id', 'recipient_key_id']) {
        const text = envelope[field]

        if (typeof text !== 'string' || text.length !== 32) {
            throw new Error('Нарушен дескриптор сообщения: неверный key_id')
        }
    }

    if (typeof envelope.ciphertext !== 'string' || envelope.ciphertext === '') {
        throw new Error('Нарушен дескриптор сообщения: поле ciphertext')
    }

    return envelope as unknown as Envelope
}

export interface EncryptInput {
    text: string
    alphabets: DualAlphabets
    privateDoc: PrivateKeyDoc
    peerPublicDoc: PublicKeyDoc
    random?: RandomBytes
}

export async function encryptText({ text, alphabets, privateDoc, peerPublicDoc, random = systemRandom }: EncryptInput): Promise<Envelope> {
    if (text === '') {
        throw new Error('Пустой текст шифровать запрещено')
    }

    if (utf8(text).length > MAX_TEXT_BYTES) {
        throw new Error('Текст слишком большой')
    }

    const sender = await loadPrivateKey(privateDoc)
    const recipient = await loadPublicKey(peerPublicDoc)
    const salt = random(32)
    const nonce = random(16)
    const envelope: Envelope = {
        type: ENVELOPE_TYPE,
        version: ENVELOPE_VERSION,
        algorithm: ALGORITHM,
        alphabet_hash: alphabets.first.hash,
        alphabet_hash2: alphabets.second.hash,
        sender_key_id: sender.keyId,
        recipient_key_id: recipient.keyId,
        salt: toBase64Url(salt),
        nonce: toBase64Url(nonce),
        ciphertext: '',
        auth_tag: '',
    }
    const { maskKey, macKey } = await sessionKeys(sender.key, recipient.key, salt, sender.keyId, recipient.keyId, alphabets.first.hash, alphabets.second.hash)
    const ciphertext = await addDualNoise(text, alphabets.first, alphabets.second, await hmacKey(maskKey), nonce)
    const tag = await hmacSign(await hmacKey(macKey), concat(metadata({ ...envelope, ciphertext }), new Uint8Array([0]), utf8(ciphertext)))

    return validateDescriptor({ ...envelope, ciphertext, auth_tag: toBase64Url(tag) })
}

export interface DecryptInput {
    envelope: unknown
    alphabets: DualAlphabets
    privateDoc: PrivateKeyDoc
    peerPublicDoc: PublicKeyDoc
    // recipient - мы получатель (как в прототипе); sender - перечитываем своё отправленное сообщение.
    // Общий секрет и ключи в обе стороны одинаковые, поэтому формат конверта не меняется
    role?: 'recipient' | 'sender'
}

// Проверяет дескриптор, ключи и MAC и только потом снимает шум. Любая подмена даёт одну и ту же ошибку
export async function decryptEnvelope({ envelope: raw, alphabets, privateDoc, peerPublicDoc, role = 'recipient' }: DecryptInput): Promise<string> {
    const envelope = validateDescriptor(raw)

    if (envelope.alphabet_hash !== alphabets.first.hash) {
        throw new Error('Сообщение создано с другим encryption.json')
    }

    if (envelope.alphabet_hash2 !== alphabets.second.hash) {
        throw new Error('Сообщение создано с другим encryption2.json')
    }

    const own = await loadPrivateKey(privateDoc)
    const other = await loadPublicKey(peerPublicDoc)
    const senderId = role === 'recipient' ? other.keyId : own.keyId
    const recipientId = role === 'recipient' ? own.keyId : other.keyId

    if (envelope.recipient_key_id !== recipientId) {
        throw new Error(role === 'recipient' ? 'Сообщение предназначено для другого приватного ключа' : 'Сообщение отправлено другому получателю')
    }

    if (envelope.sender_key_id !== senderId) {
        throw new Error(role === 'recipient' ? 'Указан неверный публичный ключ отправителя' : 'Сообщение отправлено другим ключом')
    }

    let salt: Uint8Array
    let nonce: Uint8Array
    let tag: Uint8Array

    try {
        salt = fromBase64Url(envelope.salt, 'salt')
        nonce = fromBase64Url(envelope.nonce, 'nonce')
        tag = fromBase64Url(envelope.auth_tag, 'auth_tag')
    } catch {
        throw new Error(TAMPERED)
    }

    if (salt.length !== 32 || nonce.length !== 16 || tag.length !== 32 || utf8(envelope.ciphertext).length > MAX_TEXT_BYTES * 20) {
        throw new Error(TAMPERED)
    }

    const { maskKey, macKey } = await sessionKeys(own.key, other.key, salt, senderId, recipientId, alphabets.first.hash, alphabets.second.hash)
    const expected = await hmacSign(await hmacKey(macKey), concat(metadata(envelope), new Uint8Array([0]), utf8(envelope.ciphertext)))

    if (!equalBytes(tag, expected)) {
        throw new Error(TAMPERED)
    }

    return removeDualNoise(envelope.ciphertext, alphabets.first, alphabets.second, await hmacKey(maskKey), nonce)
}