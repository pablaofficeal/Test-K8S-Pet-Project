import TextConfig from '../configs/site/text.configs'

// 32 символа (без I, O, 0, 1), поэтому остаток от деления байта не даёт перекоса
const ALPHABET = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789'
const GROUPS = 12
const GROUP_LENGTH = 4

export function generateRecoveryCode(): string[] {
    const bytes = crypto.getRandomValues(new Uint8Array(GROUPS * GROUP_LENGTH))
    const chars = Array.from(bytes, (byte) => ALPHABET[byte % ALPHABET.length])

    return Array.from({ length: GROUPS }, (_, index) => chars.slice(index * GROUP_LENGTH, (index + 1) * GROUP_LENGTH).join(''))
}

export function formatRecoveryCode(groups: string[]) {
    return groups.join('-')
}

export function downloadRecoveryCode(groups: string[]) {
    const text = `${TextConfig.RecoveryFileTitle}\n\n${formatRecoveryCode(groups)}\n`
    const url = URL.createObjectURL(new Blob([text], { type: 'text/plain;charset=utf-8' }))
    const link = document.createElement('a')

    link.href = url
    link.download = TextConfig.RecoveryFileName
    link.click()
    URL.revokeObjectURL(url)
}