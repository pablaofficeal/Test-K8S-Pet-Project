export const ROUTES = {
    HOME: '/',
    LOGIN: '/login',
    REGISTER: '/register',
    DASHBOARD: '/dashboard',
    PROFILE: '/profile',
    SETTINGS: '/settings',
}

export function userDashboardPath(uuid: string) {
    return `/users/${uuid}/dashboard`
}

export function userProfilePath(uuid: string) {
    return `/users/${uuid}/profile`
}

export function userSettingsPath(uuid: string) {
    return `/users/${uuid}/settings`
}

export function loginPathWithRedirect(target: string) {
    return `${ROUTES.LOGIN}?redirect=${encodeURIComponent(target)}`
}