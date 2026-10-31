import { Link, useLocation } from 'react-router-dom'
import Brand from './brand'
import LogoutButton from './logoutButton'
import LanguageSwitcher from './languageSwitcher'
import TextConfig from '../../lib/configs/site/text.configs'
import { ROUTES } from '../../lib/configs/site/routes'
import '../../assets/css/sidebar.css'

type SidebarProps = {
    collapsed: boolean
    onToggle: () => void
}

function Sidebar({ collapsed, onToggle }: SidebarProps) {
    const { pathname } = useLocation()
    const toggleLabel = collapsed ? TextConfig.SidebarExpand : TextConfig.SidebarCollapse
    const links = [
        {
            to: ROUTES.DASHBOARD,
            suffix: '/dashboard',
            label: TextConfig.DashboardTitle,
            icon: <path d="M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h7v7h-7z" />,
        },
        {
            to: ROUTES.PROFILE,
            suffix: '/profile',
            label: TextConfig.ProfileTitle,
            icon: <><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" /></>,
        },
        {
            to: ROUTES.SETTINGS,
            suffix: '/settings',
            label: TextConfig.SettingsTitle,
            icon: <><path d="M4 6h10M18 6h2M4 12h2M10 12h10M4 18h12M20 18h0" /><circle cx="16" cy="6" r="2" /><circle cx="8" cy="12" r="2" /><circle cx="18" cy="18" r="2" /></>,
        },
    ]

    return (
        <aside className="sidebar">
            <Brand />
            <button type="button" className="sidebar-toggle" onClick={onToggle} aria-label={toggleLabel} title={toggleLabel}>
                <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="m15 18-6-6 6-6" /></svg>
            </button>
            <nav className="sidebar-links">
                {links.map((link) => (
                    <Link key={link.to} to={link.to} title={link.label} className={pathname.endsWith(link.suffix) ? 'active' : ''}>
                        <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">{link.icon}</svg>
                        <span>{link.label}</span>
                    </Link>
                ))}
                <LanguageSwitcher />
            </nav>
            <LogoutButton />
        </aside>
    )
}
export default Sidebar