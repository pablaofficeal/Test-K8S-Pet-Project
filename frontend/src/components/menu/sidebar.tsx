import { NavLink, useLocation } from 'react-router-dom'
import Logo from './logo'
import LogoutButton from './logoutButton'
import LanguageSwitcher from './languageSwitcher'
import TextConfig from '../../lib/configs/site/text.configs'
import '../../assets/css/sidebar.css'

type SidebarProps = {
    collapsed: boolean
    onToggle: () => void
}

function Sidebar({ collapsed, onToggle }: SidebarProps) {
    const { pathname } = useLocation()
    const toggleLabel =collapsed ? TextConfig.SidebarExpand : TextConfig.SidebarCollapse

    return (
        <aside className="sidebar">
            <Logo />
            <button type="button" className="sidebar-toggle" onClick={onToggle} aria-label={toggleLabel} title={toggleLabel}>
                <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="m15 18-6-6 6-6" /></svg>
            </button>
            <nav className="sidebar-links">
                <NavLink to="/users/dashboard/index.php" title={TextConfig.DashboardTitle}>
                    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><rect x="3" y="3" width="7" height="7" rx="1" /><rect x="14" y="3" width="7" height="7" rx="1" /><rect x="3" y="14" width="7" height="7" rx="1" /><rect x="14" y="14" width="7" height="7" rx="1" /></svg>
                    <span>{TextConfig.DashboardTitle}</span>
                </NavLink>
                <NavLink to="/users/bots/index.php" title={TextConfig.MyBotsTitle} className={({ isActive }) => (isActive || pathname.startsWith('/users/bots') ? 'active' : '')}>
                    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><rect x="4" y="8" width="16" height="12" rx="3" /><path d="M12 8V4" /><circle cx="9" cy="14" r="1" /><circle cx="15" cy="14" r="1" /></svg>
                    <span>{TextConfig.MyBotsTitle}</span>
                </NavLink>
                <NavLink to="/users/limits/index.php" title={TextConfig.SidebarLimits}>
                    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M4 20V10M10 20V4M16 20v-7M22 20H2" /></svg>
                    <span>{TextConfig.SidebarLimits}</span>
                </NavLink>
                <NavLink to="/users/settings/index.php" title={TextConfig.SettingsTitle}>
                    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" /><circle cx="12" cy="12" r="3" /></svg>
                    <span>{TextConfig.SettingsTitle}</span>
                </NavLink>
                <LanguageSwitcher />
            </nav>
            <LogoutButton />
        </aside>
    )
}
export default Sidebar