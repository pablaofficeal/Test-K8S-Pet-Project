import type { PrivateKeyDoc } from './crypto/types.ts'

// MVP: приватный ключ лежит в localStorage этого браузера. Это небезопасно (XSS, расширения, чужой профиль),
// зато просто. Позже заменим хранилищем устройства
const KEY_PREFIX = 'Cadence-Cloud-Messenger-Key:'

export function loadStoredKey(uuid: string): PrivateKeyDoc | null {
    try {
        const raw = localStorage.getItem(`${KEY_PREFIX}${uuid}`)

        return raw ? (JSON.parse(raw) as PrivateKeyDoc) : null
    } catch {
        return null
    }
}

export function saveStoredKey(uuid: string, doc: PrivateKeyDoc): boolean {
    try {
        localStorage.setItem(`${KEY_PREFIX}${uuid}`, JSON.stringify(doc))

        return true
    } catch {
        return false
    }
}

export function clearStoredKey(uuid: string) {
    try {
        localStorage.removeItem(`${KEY_PREFIX}${uuid}`)
    } catch {
        // хранилище недоступно, удалять нечего
    }
}