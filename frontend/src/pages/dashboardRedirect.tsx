import { Navigate } from 'react-router-dom'
import { getUserId } from '../lib/auth/session'
import { ROUTES, loginPathWithRedirect, userDashboardPath } from '../lib/configs/site/routes'

function DashboardRedirect() {
    const uuid = getUserId()

    return <Navigate to={uuid ? userDashboardPath(uuid) : loginPathWithRedirect(ROUTES.DASHBOARD)} replace />
}
export default DashboardRedirect