import { Navigate } from 'react-router-dom'
import { getUserId } from '../../lib/auth/session'
import { loginPathWithRedirect } from '../../lib/configs/site/routes'

type UserRedirectProps = {
    from: string
    to: (uuid: string) => string
}

// /dashboard, /profile, /settings -> /users/<uuid>/..., либо на вход, если сессии нет
function UserRedirect({ from, to }: UserRedirectProps) {
    const uuid = getUserId()

    return <Navigate to={uuid ? to(uuid) : loginPathWithRedirect(from)} replace />
}
export default UserRedirect