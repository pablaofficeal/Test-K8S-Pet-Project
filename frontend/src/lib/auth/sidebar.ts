const SIDEBAR_COOKIE = 'Low-Code-Bot-Platform-Sidebar'
const SIDEBAR_COOKIE_MAX_AGE = 60 * 60 * 24 * 365

export function isSidebarCollapsed() {
    if (typeof document === 'undefined') {
        return false
    }

    return document.cookie
        .split('; ')
        .some((cookie) => cookie === `${SIDEBAR_COOKIE}=collapsed`)
}

export function saveSidebarCollapsed(collapsed: boolean) {
    if (typeof document === 'undefined') {
        return
    }

    const secure = window.location.protocol === 'https:' ? '; Secure' : ''
    document.cookie = `${SIDEBAR_COOKIE}=${collapsed ? 'collapsed' : 'open'}; Path=/; Max-Age=${SIDEBAR_COOKIE_MAX_AGE}; SameSite=Lax${secure}`
}