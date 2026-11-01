import { getCurrentLanguage } from "../i18n/i18n"

// Дедлайн приходит как дата в UTC, поэтому показываем её в UTC, иначе в западных поясах день съезжает назад
export function formatDeadline(iso: string) {
    return new Date(iso).toLocaleDateString(getCurrentLanguage(), { dateStyle: 'medium', timeZone: 'UTC' })
}