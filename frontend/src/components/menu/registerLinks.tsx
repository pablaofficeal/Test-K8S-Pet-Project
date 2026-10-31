import { Link } from 'react-router-dom'
import RegisterText from './registerText'


function RegisterLinks() {
    return (
        <Link to="/Start/create/my/account/register.php" className="main-menu-cta">
            <RegisterText />
        </Link>
    )
}
export default RegisterLinks