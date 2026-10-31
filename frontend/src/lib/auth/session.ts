import { ROUTES } from '../configs/site/routes'

// Куку выставляет сервер при входе. Она читается из JS и нужна только для URL и отображения,
// доверять ей нельзя: права проверяет шлюз по HttpOnly access-токену
const USER_COOKIE = 'Low-Code-Bot-Platform-User-Id'
const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

export function getUserId() {
    if (typeof document === 'undefined') {
        return null
    }

    const prefix = `${USER_COOKIE}=`
    const raw = document.cookie.split('; ').find((cookie) => cookie.startsWith(prefix))?.slice(prefix.length)
    const value = raw ? decodeURIComponent(raw) : ''

    return UUID_PATTERN.test(value) ? value : null
}

export function hasAuthSession() {
    return getUserId() !== null
}

export function clearAuthSession() {
    if (typeof document === 'undefined') {
        return
    }

    document.cookie = `${USER_COOKIE}=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; Path=/; SameSite=Lax`
}

export function getRedirectTarget(search: string) {
    const redirectTarget = new URLSearchParams(search).get('redirect')

    return redirectTarget?.startsWith('/') && !redirectTarget.startsWith('//') ? redirectTarget : ROUTES.DASHBOARD
}