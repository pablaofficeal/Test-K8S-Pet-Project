import { useState, type FormEvent } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import AuthHeader from '../../components/menu/authHeader'
import LoginEmail from '../../components/components/login/EmailInput'
import LoginPassword from '../../components/components/login/PasswordInput'
import Remember from '../../components/components/login/Remember'
import LoginButtons from '../../components/components/login/Buttons'
import DevicePanel from '../../components/components/login/DevicePanel'
import TextConfig from '../../lib/configs/site/text.configs'
import { describeError } from '../../lib/auth/api'
import { login } from '../../lib/auth/login'
import { getRedirectTarget, saveAuthSession } from '../../lib/auth/session'
import '../assets/css/auth.css'

function Login() {
    const navigate = useNavigate()
    const location = useLocation()
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [error, setError] = useState('')
    const [isLoading, setIsLoading] = useState(false)

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        setError('')
        setIsLoading(true)

        try {
            await login(email.trim(), password)
            saveAuthSession()
            navigate(getRedirectTarget(location.search), { replace: true })
        } catch (err) {
            setError(describeError(err, TextConfig.LoginFailed))
        } finally {
            setIsLoading(false)
        }
    }

    return (
        <div className="auth-page">
            <AuthHeader mode="login" />
            <main className="auth-login">
                <form className="auth-form" onSubmit={handleSubmit}>
                    <h1>{TextConfig.LoginTitle}</h1>
                    <p className="auth-intro">{TextConfig.LoginIntro}</p>
                    <LoginEmail value={email} onChange={setEmail} disabled={isLoading} />
                    <LoginPassword value={password} onChange={setPassword} disabled={isLoading} />
                    <Remember />
                    {error && <p className="auth-error" role="alert">{error}</p>}
                    <LoginButtons disabled={isLoading} isLoading={isLoading} />
                </form>
                <DevicePanel />
            </main>
        </div>
    )
}
export default Login