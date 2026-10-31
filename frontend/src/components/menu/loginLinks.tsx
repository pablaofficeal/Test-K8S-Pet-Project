import { Link } from 'react-router-dom'
import LoginText from './loginText'


function LoginLinks() {
    return (
        <Link to="/Start/log/in/my/account/login.php">
            <LoginText />
        </Link>
    )
}
export default LoginLinks