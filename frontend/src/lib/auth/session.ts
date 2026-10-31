const SESSION_COOKIE = 'Low-Code-Bot-Platform-Session'
const LEGACY_TOKEN_COOKIES = ['Low-Code-Bot-Platform-Auth-Token', 'Low-Code-Bot-Platform-Refresh-Token']
const SESSION_COOKIE_MAX_AGE = 60 * 60 * 24 * 60

function expireCookie(name: string) {
    const secure = window.location.protocol === 'https:' ? '; Secure' : ''
    const expired = `${name}=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; Path=/`

    document.cookie = `${expired}; SameSite=Lax${secure}`
    document.cookie = `${expired}${secure}`

    const parts = window.location.hostname.split('.')
    for (let i = 0; i < parts.length - 1; i++) {
        const domain = parts.slice(i).join('.')
        document.cookie = `${expired}; Domain=${domain}; SameSite=Lax${secure}`
        document.cookie = `${expired}; Domain=.${domain}; SameSite=Lax${secure}`
    }
}

// Сами токены лежат в HttpOnly куках сервера, здесь только флаг "пользователь вошёл" для UI
export function saveAuthSession() {
    if (typeof document === 'undefined') {
        return
    }

    const secure = window.location.protocol === 'https:' ? '; Secure' : ''
    document.cookie = `${SESSION_COOKIE}=1; Path=/; Max-Age=${SESSION_COOKIE_MAX_AGE}; SameSite=Lax${secure}`
    expireCookie(LEGACY_TOKEN_COOKIES[1])
}

export function clearAuthSession() {
    if (typeof document === 'undefined') {
        return
    }

    expireCookie(SESSION_COOKIE)
    LEGACY_TOKEN_COOKIES.forEach(expireCookie)
}

export function hasAuthSession() {
    if (typeof document === 'undefined') {
        return false
    }

    return document.cookie
        .split('; ')
        .some((cookie) => cookie === `${SESSION_COOKIE}=1`)
}

export function getRedirectTarget(search: string) {
    const redirectTarget = new URLSearchParams(search).get('redirect')

    return redirectTarget?.startsWith('/') && !redirectTarget.startsWith('//') ? redirectTarget : '/users/dashboard/index.php'
}