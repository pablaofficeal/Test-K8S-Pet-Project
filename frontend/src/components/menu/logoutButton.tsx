import { useNavigate } from 'react-router-dom'
import TextConfig from '../../lib/configs/site/text.configs'
import { clearAuthSession } from '../../lib/auth/session'
import { logout } from '../../lib/auth/logout'

function LogoutButton() {
    const navigate = useNavigate()

    async function handleLogout() {
        await logout()
        clearAuthSession()
        navigate('/Start/log/in/my/account/login.php', { replace: true })
    }

    return (
        <button type="button" className="logout-button" onClick={handleLogout}>
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9" /></svg>
            <span>{TextConfig.DashboardLogout}</span>
        </button>
    )
}
export default LogoutButton