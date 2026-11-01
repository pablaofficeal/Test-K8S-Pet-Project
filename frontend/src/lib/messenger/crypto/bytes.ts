// Мелкие операции над байтами. Для WebCrypto отдаём ArrayBuffer-копии: так не мешают общие буферы и дженерики Uint8Array

export function toBuffer(bytes: Uint8Array): ArrayBuffer {
    return bytes.slice().buffer as ArrayBuffer
}

export function utf8(text: string): Uint8Array {
    return new TextEncoder().encode(text)
}

export function concat(...parts: Uint8Array[]): Uint8Array {
    const out = new Uint8Array(parts.reduce((sum, part) => sum + part.length, 0))
    let offset = 0

    for (const part of parts) {
        out.set(part, offset)
        offset += part.length
    }

    return out
}

export function positionBytes(value: number, size: number): Uint8Array {
    const out = new Uint8Array(size)
    let rest = value

    for (let index = size - 1; index >= 0; index -= 1) {
        out[index] = rest % 256
        rest = Math.floor(rest / 256)
    }

    return out
}

// 4 байта big-endian как беззнаковое число
export function uint32(bytes: Uint8Array, start: number): number {
    return ((bytes[start] << 24) | (bytes[start + 1] << 16) | (bytes[start + 2] << 8) | bytes[start + 3]) >>> 0
}

export function uint16(bytes: Uint8Array, start: number): number {
    return (bytes[start] << 8) | bytes[start + 1]
}

// Сравнение за постоянное время для MAC
export function equalBytes(a: Uint8Array, b: Uint8Array): boolean {
    if (a.length !== b.length) {
        return false
    }

    let diff = 0

    for (let index = 0; index < a.length; index += 1) {
        diff |= a[index] ^ b[index]
    }

    return diff === 0
}

export function hex(bytes: Uint8Array): string {
    return Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('')
}

// Порядок как у sorted() в Python: по кодовым точкам Unicode, а не по UTF-16 единицам
export function compareCodePoints(a: string, b: string): number {
    const left = Array.from(a)
    const right = Array.from(b)
    const length = Math.min(left.length, right.length)

    for (let index = 0; index < length; index += 1) {
        const difference = (left[index].codePointAt(0) ?? 0) - (right[index].codePointAt(0) ?? 0)

        if (difference !== 0) {
            return difference
        }
    }

    return left.length - right.length
}