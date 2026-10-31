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

    return (
        <aside className="sidebar">
            <Brand />
            <button type="button" className="sidebar-toggle" onClick={onToggle} aria-label={toggleLabel} title={toggleLabel}>
                <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="m15 18-6-6 6-6" /></svg>
            </button>
            <nav className="sidebar-links">
                <Link to={ROUTES.DASHBOARD} title={TextConfig.DashboardTitle} className={pathname.endsWith('/dashboard') ? 'active' : ''}>
                    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><rect x="3" y="3" width="7" height="7" rx="1" /><rect x="14" y="3" width="7" height="7" rx="1" /><rect x="3" y="14" width="7" height="7" rx="1" /><rect x="14" y="14" width="7" height="7" rx="1" /></svg>
                    <span>{TextConfig.DashboardTitle}</span>
                </Link>
                <LanguageSwitcher />
            </nav>
            <LogoutButton />
        </aside>
    )
}
export default Sidebar