export const ROUTES = {
    HOME: '/',
    LOGIN: '/login',
    REGISTER: '/register',
    DASHBOARD: '/dashboard',
    PROFILE: '/profile',
    SETTINGS: '/settings',
    VERIFY: '/verify',
    TODO: '/todo',
    MESSENGER: '/messenger',
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

export function userTodoPath(uuid: string) {
    return `/users/${uuid}/todo`
}

export function userTodoBoardPath(uuid: string, boardId: string) {
    return `/users/${uuid}/todo/${boardId}`
}

export function userMessengerPath(uuid: string) {
    return `/users/${uuid}/messenger`
}

export function userMessengerThreadPath(uuid: string, peerUuid: string) {
    return `/users/${uuid}/messenger/${peerUuid}`
}

export function loginPathWithRedirect(target: string) {
    return `${ROUTES.LOGIN}?redirect=${encodeURIComponent(target)}`
}

export function verifyPathWithRedirect(target: string) {
    return `${ROUTES.VERIFY}?redirect=${encodeURIComponent(target)}`
}