import { Link } from 'react-router-dom'
import Brand from './brand'
import TextConfig from '../../lib/configs/site/text.configs'

type AuthHeaderProps = {
    mode: 'login' | 'register'
}

function AuthHeader({ mode }: AuthHeaderProps) {
    const isLogin = mode === 'login'

    return (
        <header className="auth-header">
            <Brand />
            <p>
                {isLogin ? TextConfig.AuthNoAccount : TextConfig.HaveAccount}{' '}
                <Link to={isLogin ? '/Start/create/my/account/register.php' : '/Start/log/in/my/account/login.php'}>
                    {isLogin ? TextConfig.AuthCreateShort : TextConfig.NavLogin}
                </Link>
            </p>
        </header>
    )
}
export default AuthHeader