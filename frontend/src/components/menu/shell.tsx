import { useState, type ReactNode } from 'react'
import MainMenu from './mainMenu'
import Sidebar from './sidebar'
import { hasAuthSession } from '../../lib/auth/session'
import { isSidebarCollapsed, saveSidebarCollapsed } from '../../lib/auth/sidebar'
import '../../assets/css/sidebar.css'

type ShellProps = {
    children: ReactNode
    dashboard?: boolean
}

function Shell({ children, dashboard = false }: ShellProps) {
    const [collapsed, setCollapsed] = useState(isSidebarCollapsed)

    function handleToggle() {
        saveSidebarCollapsed(!collapsed)
        setCollapsed(!collapsed)
    }

    if (!dashboard || !hasAuthSession()) {
        return (
            <div className="main-pages">
                <MainMenu />
                {children}
            </div>
        )
    }

    return (
        <div className={collapsed ? 'dash-page collapsed' : 'dash-page'}>
            <Sidebar collapsed={collapsed} onToggle={handleToggle} />
            <div className="dash-content">
                <MainMenu />
                {children}
            </div>
        </div>
    )
}
export default Shell