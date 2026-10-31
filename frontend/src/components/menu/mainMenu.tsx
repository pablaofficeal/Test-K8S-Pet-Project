import Logo from './logo'
import LoginLinks from './loginLinks'
import RegisterLinks from './registerLinks'
import HomeLinks from './homeLinks'
import DashboardLinks from './dashboardLinks'
import LogoutButton from './logoutButton'
import LanguageSwitcher from './languageSwitcher'
import { hasAuthSession } from '../../lib/auth/session'
import '../../assets/css/menu.css'

function MainMenu() {
    return (
        <header className="main-menu">
            <Logo />
            <nav className="main-menu-links">
                <HomeLinks />
            </nav>
            <div className="main-menu-action">
                <LanguageSwitcher />
                {hasAuthSession() ? <><DashboardLinks /><LogoutButton /></> : <><LoginLinks /><RegisterLinks /></>}
            </div>
        </header>
    )
}
export default MainMenu