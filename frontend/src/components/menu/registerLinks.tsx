import { Link } from 'react-router-dom'
import RegisterText from './registerText'
import { ROUTES } from '../../lib/configs/site/routes'


function RegisterLinks() {
    return (
        <Link to={ROUTES.REGISTER} className="main-menu-cta">
            <RegisterText />
        </Link>
    )
}
export default RegisterLinks