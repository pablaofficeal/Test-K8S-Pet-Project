// Пока аватарка не загружена, рисуем заглушку: первая буква имени и градиент от uuid
export function avatarLetter(name: string) {
    return (Array.from(name.trim())[0] ?? '?').toUpperCase()
}

export function avatarGradient(uuid: string) {
    let hue = 0

    for (const char of uuid) {
        hue = (hue * 31 + char.charCodeAt(0)) % 360
    }

    return `linear-gradient(135deg, hsl(${hue} 55% 40%), hsl(${(hue + 55) % 360} 60% 20%))`
}