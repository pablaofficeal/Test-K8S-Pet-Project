import { loadDualAlphabets, type DualAlphabets } from './crypto/alphabet.ts'

const bundled = import.meta.glob<unknown>('./dictionaries/*.json', { eager: true, import: 'default' })
const CUSTOM_KEY = 'Cadence-Cloud-Messenger-Dictionaries'
const MAX_DICTIONARY_BYTES = 1024 * 1024

interface CustomDictionaries {
    first: unknown
    second: unknown
}

function readCustom(): CustomDictionaries | null {
    try {
        const raw = localStorage.getItem(CUSTOM_KEY)

        return raw ? (JSON.parse(raw) as CustomDictionaries) : null
    } catch {
        return null
    }
}

export async function bundledAlphabets(): Promise<DualAlphabets> {
    return loadDualAlphabets(bundled['./dictionaries/encryption.json'], bundled['./dictionaries/encryption2.json'])
}

// Словари пользователя (если загрузил свои) важнее встроенных; битые молча заменяем встроенными
export async function activeAlphabets(): Promise<{ alphabets: DualAlphabets; custom: boolean }> {
    const custom = readCustom()

    if (custom) {
        try {
            return { alphabets: await loadDualAlphabets(custom.first, custom.second), custom: true }
        } catch {
            clearCustomDictionaries()
        }
    }

    return { alphabets: await bundledAlphabets(), custom: false }
}

// Сначала проверяем оба словаря так же строго, как при шифровании, и только потом сохраняем
export async function saveCustomDictionaries(firstText: string, secondText: string): Promise<void> {
    if (firstText.length > MAX_DICTIONARY_BYTES || secondText.length > MAX_DICTIONARY_BYTES) {
        throw new Error('Словарь больше 1 МБ')
    }

    const first = JSON.parse(firstText) as unknown
    const second = JSON.parse(secondText) as unknown

    await loadDualAlphabets(first, second)
    localStorage.setItem(CUSTOM_KEY, JSON.stringify({ first, second }))
}

export function clearCustomDictionaries() {
    try {
        localStorage.removeItem(CUSTOM_KEY)
    } catch {
        // хранилище недоступно, удалять нечего
    }
}