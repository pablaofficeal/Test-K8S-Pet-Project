import TextConfig from '../../lib/configs/site/text.configs'
import type { MenuBadge as MenuBadgeType } from '../../lib/configs/site/menu'

type MenuBadgeProps = {
    badge: MenuBadgeType
}

function MenuBadge({ badge }: MenuBadgeProps) {
    return <em className={`sidebar-badge ${badge}`}>{badge === 'beta' ? TextConfig.BadgeBeta : TextConfig.BadgeNew}</em>
}
export default MenuBadge