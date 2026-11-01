export function toBase64Url(bytes: Uint8Array): string {
    let binary = ''

    for (const byte of bytes) {
        binary += String.fromCharCode(byte)
    }

    return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

export function fromBase64Url(value: string, name: string): Uint8Array {
    if (!/^[A-Za-z0-9_-]*$/.test(value) || value.length % 4 === 1) {
        throw new Error(`Некорректное поле ${name}`)
    }

    const padded = value.replace(/-/g, '+').replace(/_/g, '/') + '='.repeat((4 - (value.length % 4)) % 4)
    const binary = atob(padded)

    return Uint8Array.from(binary, (char) => char.charCodeAt(0))
}