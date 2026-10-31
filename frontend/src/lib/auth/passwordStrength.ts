// Оценка только для подсказки в форме: 0 - пусто, 4 - максимум
export function passwordStrength(password: string): number {
    if (!password) {
        return 0
    }

    let score = 0

    if (password.length >= 8) {
        score += 1
    }
    if (password.length >= 12) {
        score += 1
    }
    if (/[a-z]/.test(password) && /[A-Z]/.test(password) && /\d/.test(password)) {
        score += 1
    }
    if (/[^A-Za-z0-9]/.test(password)) {
        score += 1
    }

    return Math.max(score, 1)
}