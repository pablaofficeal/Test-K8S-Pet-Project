import { getCurrentLanguage } from "./i18n"

export function formatDate(iso: string) {
    return new Date(iso).toLocaleDateString(getCurrentLanguage(), { dateStyle: 'long' })
}