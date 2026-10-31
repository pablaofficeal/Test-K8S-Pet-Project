import { Link } from 'react-router-dom'
import TextConfig from '../../lib/configs/site/text.configs'
import { ROUTES } from '../../lib/configs/site/routes'

function UserLinks() {
    return (
        <>
            <Link to={ROUTES.DASHBOARD}>{TextConfig.DashboardTitle}</Link>
            <Link to={ROUTES.PROFILE}>{TextConfig.ProfileTitle}</Link>
            <Link to={ROUTES.SETTINGS}>{TextConfig.SettingsTitle}</Link>
        </>
    )
}
export default UserLinks