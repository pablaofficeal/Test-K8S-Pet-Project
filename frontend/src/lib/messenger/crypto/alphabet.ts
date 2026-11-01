import { toBase64Url } from './base64url.ts'
import { compareCodePoints, toBuffer, utf8 } from './bytes.ts'
import type { Alphabet } from './types.ts'

const SECTIONS = ['US', 'RU', 'symbols']

// Канонический JSON как у json.dumps(sort_keys=True, separators=(",", ":"), ensure_ascii=False)
function canonical(value: unknown): string {
    if (Array.isArray(value)) {
        return `[${value.map(canonical).join(',')}]`
    }

    if (value !== null && typeof value === 'object') {
        const record = value as Record<string, unknown>

        return `{${Object.keys(record).sort(compareCodePoints).map((key) => `${JSON.stringify(key)}:${canonical(record[key])}`).join(',')}}`
    }

    return JSON.stringify(value)
}

// Загружает словарь (encryption.json) и проверяет его так же строго, как прототип
export async function loadAlphabet(data: unknown): Promise<Alphabet> {
    if (data === null || typeof data !== 'object' || Array.isArray(data)) {
        throw new Error('Словарь должен быть JSON-объектом')
    }

    const record = data as Record<string, unknown>
    const forward = new Map<string, string>()
    const tokens = new Set<string>()

    for (const name of SECTIONS) {
        const section = record[name]

        if (section === null || typeof section !== 'object' || Array.isArray(section) || Object.keys(section).length === 0) {
            throw new Error(`В алфавите нет секции ${name}`)
        }

        for (const [character, token] of Object.entries(section as Record<string, unknown>)) {
            if (Array.from(character).length !== 1) {
                throw new Error(`Ключ ${JSON.stringify(character)} должен быть одним символом`)
            }

            if (typeof token !== 'string' || token === '') {
                throw new Error(`Пустой токен у символа ${JSON.stringify(character)}`)
            }

            if (/\s/u.test(token)) {
                throw new Error(`Токен ${JSON.stringify(token)} не должен содержать пробелы`)
            }

            if (forward.has(character)) {
                throw new Error(`Символ ${JSON.stringify(character)} встречается дважды`)
            }

            if (tokens.has(token)) {
                throw new Error(`Токен ${JSON.stringify(token)} встречается дважды`)
            }

            forward.set(character, token)
            tokens.add(token)
        }
    }

    const digest = await crypto.subtle.digest('SHA-256', toBuffer(utf8(canonical(data))))

    return { forward, hash: toBase64Url(new Uint8Array(digest)) }
}

export interface DualAlphabets {
    first: Alphabet
    second: Alphabet
}

// Два словаря должны описывать одни и те же символы, отличаться и не делить токены
export async function loadDualAlphabets(firstData: unknown, secondData: unknown): Promise<DualAlphabets> {
    const first = await loadAlphabet(firstData)
    const second = await loadAlphabet(secondData)
    const firstChars = [...first.forward.keys()].sort(compareCodePoints)
    const secondChars = [...second.forward.keys()].sort(compareCodePoints)

    if (firstChars.length !== secondChars.length || firstChars.some((char, index) => char !== secondChars[index])) {
        throw new Error('Алфавиты должны содержать одинаковые символы')
    }

    if (first.hash === second.hash) {
        throw new Error('Алфавиты совпадают, нужны два разных файла')
    }

    const firstTokens = new Set(first.forward.values())

    for (const token of second.forward.values()) {
        if (firstTokens.has(token)) {
            throw new Error('Токены алфавитов пересекаются, они должны быть разными')
        }
    }

    return { first, second }
}

// Порядок символов задаёт индексы для сдвига: sorted() по кодовым точкам
export function alphabetOrder(forward: Map<string, string>): { characters: string[]; tokens: string[]; indexes: Map<string, number> } {
    const characters = [...forward.keys()].sort(compareCodePoints)

    return {
        characters,
        tokens: characters.map((character) => forward.get(character) as string),
        indexes: new Map(characters.map((character, index) => [character, index])),
    }
}