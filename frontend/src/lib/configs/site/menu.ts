import type TextConfig from './text.configs'
import { ROUTES } from './routes'

export type TextKey = keyof typeof TextConfig
export type MenuBadge = 'beta' | 'new'

export interface MenuLinkDef {
    id: string
    segment: string
    to: string
    labelKey: TextKey
    icon: string[]
    badge?: MenuBadge
}

export interface MenuGroupDef {
    id: string
    labelKey: TextKey
    icon: string[]
    badge?: MenuBadge
    children: MenuLinkDef[]
}

export interface MenuSectionDef {
    id: string
    titleKey?: TextKey
    items: (MenuLinkDef | MenuGroupDef)[]
}

export function isMenuLinkActive(pathname: string, link: MenuLinkDef) {
    return pathname.split('/')[3] === link.segment
}

// Пункты бокового меню: разделы -> ссылки или сворачиваемые группы продуктов.
// segment - часть адреса /users/<uuid>/<segment>, по ней подсвечивается активный пункт
export const MENU: MenuSectionDef[] = [
    {
        id: 'account',
        titleKey: 'SidebarSectionAccount',
        items: [
            { id: 'dashboard', segment: 'dashboard', to: ROUTES.DASHBOARD, labelKey: 'DashboardTitle', icon: ['M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h7v7h-7z'] },
            { id: 'profile', segment: 'profile', to: ROUTES.PROFILE, labelKey: 'ProfileTitle', icon: ['M12 4a4 4 0 1 0 0 8 4 4 0 0 0 0-8z', 'M4 21a8 8 0 0 1 16 0'] },
            { id: 'settings', segment: 'settings', to: ROUTES.SETTINGS, labelKey: 'SettingsTitle', icon: ['M4 6h10M18 6h2M4 12h2M10 12h10M4 18h12M20 18h0', 'M16 4a2 2 0 1 0 0 4 2 2 0 0 0 0-4z', 'M8 10a2 2 0 1 0 0 4 2 2 0 0 0 0-4z', 'M18 16a2 2 0 1 0 0 4 2 2 0 0 0 0-4z'] },
        ],
    },
    {
        id: 'products',
        titleKey: 'SidebarSectionProducts',
        items: [
            {
                id: 'todo',
                labelKey: 'TodoTitle',
                icon: ['M9 11l3 3L22 4', 'M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11'],
                badge: 'beta',
                children: [
                    { id: 'todo-projects', segment: 'todo', to: ROUTES.TODO, labelKey: 'TodoProjects', icon: ['M3 7h18M3 12h18M3 17h18'] },
                ],
            },
        ],
    },
]