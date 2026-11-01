import { useState } from 'react'
import { Link } from 'react-router-dom'
import MenuIcon from './menuIcon'
import MenuBadge from './menuBadge'
import SidebarLink from './sidebarLink'
import TextConfig from '../../lib/configs/site/text.configs'
import { isMenuLinkActive, type MenuGroupDef } from '../../lib/configs/site/menu'

type SidebarGroupProps = {
    group: MenuGroupDef
    pathname: string
    collapsed: boolean
}

function SidebarGroup({ group, pathname, collapsed }: SidebarGroupProps) {
    const active = group.children.some((child) => isMenuLinkActive(pathname, child))
    const [open, setOpen] = useState(active)
    const label = TextConfig[group.labelKey]

    if (collapsed) {
        return (
            <Link to={group.children[0].to} title={label} className={active ? 'active' : ''}>
                <MenuIcon paths={group.icon} />
                <span>{label}</span>
            </Link>
        )
    }

    return (
        <div className="sidebar-group">
            <button type="button" className={active ? 'sidebar-group-toggle active' : 'sidebar-group-toggle'} onClick={() => setOpen(!open)} aria-expanded={open} title={label}>
                <MenuIcon paths={group.icon} />
                <span>{label}</span>
                {group.badge && <MenuBadge badge={group.badge} />}
                <svg className={open ? 'sidebar-chevron open' : 'sidebar-chevron'} viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="m6 9 6 6 6-6" /></svg>
            </button>
            {open && (
                <div className="sidebar-group-items">
                    {group.children.map((child) => (
                        <SidebarLink key={child.id} link={child} active={isMenuLinkActive(pathname, child)} sub />
                    ))}
                </div>
            )}
        </div>
    )
}
export default SidebarGroup