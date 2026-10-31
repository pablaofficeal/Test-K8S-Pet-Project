import { useEffect, useState, type MouseEvent } from 'react'
import { useLocation } from 'react-router-dom'
import Brand from './brand'
import NavItems from './navItems'
import LoginLinks from './loginLinks'
import RegisterLinks from './registerLinks'
import DashboardLinks from './dashboardLinks'
import UserLinks from './userLinks'
import LogoutButton from './logoutButton'
import LanguageSwitcher from './languageSwitcher'
import TextConfig from '../../lib/configs/site/text.configs'
import { ROUTES } from '../../lib/configs/site/routes'
import { hasAuthSession } from '../../lib/auth/session'
import '../../assets/css/menu.css'

function MainMenu() {
    const { pathname } = useLocation()
    // Меню открыто только на той странице, где его открыли, поэтому при переходе оно закрывается само
    const [openedOn, setOpenedOn] = useState<string | null>(null)
    const open = openedOn === pathname
    const label = open ? TextConfig.MenuClose : TextConfig.MenuOpen

    useEffect(() => {
        if (!open) {
            return
        }

        function handleKeyDown(event: KeyboardEvent) {
            if (event.key === 'Escape') {
                setOpenedOn(null)
            }
        }

        document.addEventListener('keydown', handleKeyDown)
        return () => document.removeEventListener('keydown', handleKeyDown)
    }, [open])

    function handlePanelClick(event: MouseEvent<HTMLDivElement>) {
        if ((event.target as HTMLElement).closest('a')) {
            setOpenedOn(null)
        }
    }

    return (
        <header className="main-menu">
            <Brand />
            <button
                type="button"
                className="main-menu-burger"
                aria-expanded={open}
                aria-controls="main-menu-panel"
                aria-label={label}
                title={label}
                onClick={() => setOpenedOn(open ? null : pathname)}
            >
                {open
                    ? <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M18 6 6 18M6 6l12 12" /></svg>
                    : <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M4 7h16M4 12h16M4 17h16" /></svg>}
            </button>
            <div id="main-menu-panel" className={open ? 'main-menu-panel open' : 'main-menu-panel'} onClick={handlePanelClick}>
                {pathname === ROUTES.HOME && (
                    <nav className="main-menu-links">
                        <NavItems />
                    </nav>
                )}
                <div className="main-menu-action">
                    <LanguageSwitcher />
                    {hasAuthSession() ? <>{pathname.startsWith('/users/') ? <UserLinks /> : <DashboardLinks />}<LogoutButton /></> : <><LoginLinks /><RegisterLinks /></>}
                </div>
            </div>
        </header>
    )
}
export default MainMenu