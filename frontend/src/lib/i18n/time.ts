import { getCurrentLanguage } from "./i18n"

const UNITS: [Intl.RelativeTimeFormatUnit, number][] = [
    ["year", 31536000],
    ["month", 2592000],
    ["day", 86400],
    ["hour", 3600],
    ["minute", 60],
]

export function formatRelativeTime(iso: string) {
    const formatter = new Intl.RelativeTimeFormat(getCurrentLanguage(), { numeric: "auto" })
    const seconds = Math.round((new Date(iso).getTime() - Date.now()) / 1000)

    for (const [unit, size] of UNITS) {
        if (Math.abs(seconds) >= size) {
            return formatter.format(Math.round(seconds / size), unit)
        }
    }

    return formatter.format(seconds, "second")
}

export function formatDate(iso: string) {
    return new Date(iso).toLocaleDateString(getCurrentLanguage(), { year: "numeric", month: "long", day: "numeric" })
}