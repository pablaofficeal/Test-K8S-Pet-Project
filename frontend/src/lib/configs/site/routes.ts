export const ROUTES = {
    HOME: '/',
    LOGIN: '/login',
    REGISTER: '/register',
    DASHBOARD: '/dashboard',
}

export function userDashboardPath(uuid: string) {
    return `/users/${uuid}/dashboard`
}

export function loginPathWithRedirect(target: string) {
    return `${ROUTES.LOGIN}?redirect=${encodeURIComponent(target)}`
}