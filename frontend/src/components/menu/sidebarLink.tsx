import { Link } from 'react-router-dom'
import MenuIcon from './menuIcon'
import MenuBadge from './menuBadge'
import TextConfig from '../../lib/configs/site/text.configs'
import type { MenuLinkDef } from '../../lib/configs/site/menu'

type SidebarLinkProps = {
    link: MenuLinkDef
    active: boolean
    sub?: boolean
}

function SidebarLink({ link, active, sub = false }: SidebarLinkProps) {
    const label = TextConfig[link.labelKey]
    const classes = [active ? 'active' : '', sub ? 'sidebar-sublink' : ''].filter(Boolean).join(' ')

    return (
        <Link to={link.to} title={label} className={classes}>
            {!sub && <MenuIcon paths={link.icon} />}
            <span>{label}</span>
            {link.badge && <MenuBadge badge={link.badge} />}
        </Link>
    )
}
export default SidebarLink