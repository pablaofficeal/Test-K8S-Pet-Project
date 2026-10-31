import { Link } from 'react-router-dom'
import Brand from './brand'
import LogoutButton from './logoutButton'
import TextConfig from '../../lib/configs/site/text.configs'
import { ROUTES } from '../../lib/configs/site/routes'

type AuthHeaderProps = {
    mode: 'login' | 'register' | 'verify'
}

function AuthHeader({ mode }: AuthHeaderProps) {
    if (mode === 'verify') {
        return (
            <header className="auth-header">
                <Brand />
                <LogoutButton />
            </header>
        )
    }

    const isLogin = mode === 'login'

    return (
        <header className="auth-header">
            <Brand />
            <p>
                {isLogin ? TextConfig.AuthNoAccount : TextConfig.HaveAccount}{' '}
                <Link to={isLogin ? ROUTES.REGISTER : ROUTES.LOGIN}>
                    {isLogin ? TextConfig.AuthCreateShort : TextConfig.NavLogin}
                </Link>
            </p>
        </header>
    )
}
export default AuthHeader