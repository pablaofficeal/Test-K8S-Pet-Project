import { useState, type FormEvent } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import AuthHeader from '../../components/menu/authHeader'
import RegisterEmail from '../../components/components/registers/EmailInput'
import RegisterName from '../../components/components/registers/Name'
import RegisterPassword from '../../components/components/registers/PasswordInput'
import Strength from '../../components/components/registers/Strength'
import Agreements from '../../components/components/registers/Agreements'
import RegisterButtons from '../../components/components/registers/Buttons'
import TextConfig from '../../lib/configs/site/text.configs'
import { describeError } from '../../lib/auth/api'
import { login } from '../../lib/auth/login'
import { register } from '../../lib/auth/registers'
import { getRedirectTarget } from '../../lib/auth/session'
import '../../assets/css/auth.css'

function Register() {
    const navigate = useNavigate()
    const location = useLocation()
    const [email, setEmail] = useState('')
    const [name, setName] = useState('')
    const [password, setPassword] = useState('')
    const [accepted, setAccepted] = useState(false)
    const [error, setError] = useState('')
    const [isLoading, setIsLoading] = useState(false)

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        setError('')
        setIsLoading(true)

        try {
            await register(email.trim(), password, name.trim())
            await login(email.trim(), password)
            navigate(getRedirectTarget(location.search), { replace: true })
        } catch (err) {
            setError(describeError(err, TextConfig.RegisterFailed))
        } finally {
            setIsLoading(false)
        }
    }

    return (
        <div className="auth-page">
            <AuthHeader mode="register" />
            <main className="auth-main auth-main-wide">
                <form className="auth-form" onSubmit={handleSubmit}>
                    <h1>{TextConfig.RegisterTitle}</h1>
                    <p className="auth-intro">{TextConfig.RegisterIntro}</p>
                    <div className="auth-row">
                        <RegisterEmail value={email} onChange={setEmail} disabled={isLoading} />
                        <RegisterName value={name} onChange={setName} disabled={isLoading} />
                    </div>
                    <RegisterPassword value={password} onChange={setPassword} disabled={isLoading} />
                    <Strength value={password} />
                    <Agreements accepted={accepted} onAcceptedChange={setAccepted} disabled={isLoading} />
                    {error && <p className="auth-error" role="alert">{error}</p>}
                    <RegisterButtons disabled={isLoading || !accepted} isLoading={isLoading} />
                    <p className="auth-note">{TextConfig.RegisterFooterNote}</p>
                </form>
            </main>
        </div>
    )
}
export default Register