import { getCurrentLanguage } from "../i18n/i18n"

// Дедлайн приходит как дата в UTC, поэтому показываем её в UTC, иначе в западных поясах день съезжает назад
export function formatDeadline(iso: string) {
    return new Date(iso).toLocaleDateString(getCurrentLanguage(), { dateStyle: 'medium', timeZone: 'UTC' })
}

export type DeadlineState = 'done' | 'overdue' | 'soon' | 'later'

// Сравниваем даты как строки YYYY-MM-DD в локальном часовом поясе пользователя
export function deadlineState(iso: string, done: boolean): DeadlineState {
    if (done) {
        return 'done'
    }

    const today = new Date().toLocaleDateString('en-CA')
    const date = iso.slice(0, 10)

    if (date < today) {
        return 'overdue'
    }

    const limit = new Date()
    limit.setDate(limit.getDate() + 2)

    return date <= limit.toLocaleDateString('en-CA') ? 'soon' : 'later'
}