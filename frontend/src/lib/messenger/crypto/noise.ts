import { alphabetOrder } from './alphabet.ts'
import { maskStream } from './keySchedule.ts'
import { uint16, uint32 } from './bytes.ts'
import type { Alphabet } from './types.ts'

export class AlphabetCharacterError extends Error {
    character: string

    constructor(character: string, position: number) {
        super(`Символ ${JSON.stringify(character)} в позиции ${position + 1} отсутствует в алфавите`)
        this.name = 'AlphabetCharacterError'
        this.character = character
    }
}

// Каждый символ превращается в группу из 4-7 токенов: один настоящий (со сдвигом из потока маски) и шум
export async function addDualNoise(text: string, first: Alphabet, second: Alphabet, maskKey: CryptoKey, nonce: Uint8Array): Promise<string> {
    const order1 = alphabetOrder(first.forward)
    const order2 = alphabetOrder(second.forward)
    const size = order1.tokens.length

    if (size === 0 || size !== order2.tokens.length) {
        throw new Error('Алфавиты пустые или разного размера')
    }

    const combined = [...order1.tokens, ...order2.tokens]
    const output: string[] = []
    let position = 0

    for (const character of text) {
        const index1 = order1.indexes.get(character)
        const index2 = order2.indexes.get(character)

        if (index1 === undefined || index2 === undefined) {
            throw new AlphabetCharacterError(character, position)
        }

        const stream = await maskStream(maskKey, nonce, position, 80)
        const noiseCount = 3 + (stream[0] % 4)
        const groupSize = noiseCount + 1
        const realPosition = uint16(stream, 1) % groupSize
        const shift1 = uint32(stream, 4) % size
        const shift2 = uint32(stream, 8) % size
        const realFlag = (stream[12] ^ stream[13] ^ stream[14] ^ (position & 0xff)) & 1
        const realToken = realFlag === 0 ? order1.tokens[(index1 + shift1) % size] : order2.tokens[(index2 + shift2) % size]
        const group: string[] = []
        const used = new Set<string>([realToken])

        for (let noise = 0; noise < noiseCount; noise += 1) {
            let tokenIndex = uint32(stream, 16 + noise * 4) % combined.length
            let token = combined[tokenIndex]
            let probe = 0

            while (used.has(token) && probe < combined.length) {
                tokenIndex = (tokenIndex + 1) % combined.length
                token = combined[tokenIndex]
                probe += 1
            }

            used.add(token)
            group.push(token)
        }

        group.splice(realPosition, 0, realToken)
        output.push(...group)
        position += 1
    }

    return output.join(' ')
}

export async function removeDualNoise(ciphertext: string, first: Alphabet, second: Alphabet, maskKey: CryptoKey, nonce: Uint8Array): Promise<string> {
    const order1 = alphabetOrder(first.forward)
    const order2 = alphabetOrder(second.forward)
    const size = order1.tokens.length

    if (size === 0 || size !== order2.tokens.length) {
        throw new Error('Алфавиты пустые или разного размера')
    }

    const tokenIndex1 = new Map(order1.tokens.map((token, index) => [token, index]))
    const tokenIndex2 = new Map(order2.tokens.map((token, index) => [token, index]))
    const tokens = ciphertext.split(/\s+/).filter((token) => token !== '')
    const plaintext: string[] = []
    let offset = 0
    let position = 0

    while (offset < tokens.length) {
        const stream = await maskStream(maskKey, nonce, position, 80)
        const noiseCount = 3 + (stream[0] % 4)
        const groupSize = noiseCount + 1

        if (offset + groupSize > tokens.length) {
            throw new Error('Повреждена структура зашифрованного текста')
        }

        const realPosition = uint16(stream, 1) % groupSize
        const shift1 = uint32(stream, 4) % size
        const shift2 = uint32(stream, 8) % size
        const realFlag = (stream[12] ^ stream[13] ^ stream[14] ^ (position & 0xff)) & 1
        const realToken = tokens[offset + realPosition]
        const inFirst = tokenIndex1.get(realToken)
        const inSecond = tokenIndex2.get(realToken)

        if (inFirst === undefined && inSecond === undefined) {
            throw new Error(`Токен ${JSON.stringify(realToken)} отсутствует в алфавитах`)
        }

        if (realFlag === 0 ? inFirst === undefined : inSecond === undefined) {
            throw new Error('Повреждена структура: алфавит токена не совпал')
        }

        const plainIndex = realFlag === 0
            ? (((inFirst as number) - shift1) % size + size) % size
            : (((inSecond as number) - shift2) % size + size) % size

        plaintext.push(order1.characters[plainIndex])
        offset += groupSize
        position += 1
    }

    return plaintext.join('')
}