import { useLocation } from 'react-router-dom'
import Brand from './brand'
import LogoutButton from './logoutButton'
import LanguageSwitcher from './languageSwitcher'
import SidebarGroup from './sidebarGroup'
import SidebarLink from './sidebarLink'
import TextConfig from '../../lib/configs/site/text.configs'
import { MENU, isMenuLinkActive } from '../../lib/configs/site/menu'
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
                {MENU.map((section) => (
                    <div key={section.id} className="sidebar-section">
                        {section.titleKey && <p className="sidebar-section-title">{TextConfig[section.titleKey]}</p>}
                        {section.items.map((item) => (
                            'children' in item
                                ? <SidebarGroup key={item.id} group={item} pathname={pathname} collapsed={collapsed} />
                                : <SidebarLink key={item.id} link={item} active={isMenuLinkActive(pathname, item)} />
                        ))}
                    </div>
                ))}
                <LanguageSwitcher />
            </nav>
            <LogoutButton />
        </aside>
    )
}
export default Sidebar