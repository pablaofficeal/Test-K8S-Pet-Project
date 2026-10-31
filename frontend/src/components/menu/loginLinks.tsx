import { Link } from 'react-router-dom'
import LoginText from './loginText'
import { ROUTES } from '../../lib/configs/site/routes'


function LoginLinks() {
    return (
        <Link to={ROUTES.LOGIN}>
            <LoginText />
        </Link>
    )
}
export default LoginLinks