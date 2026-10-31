import { useEffect, useState, type FormEvent } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import AuthHeader from '../../components/menu/authHeader'
import CodeInput from '../../components/components/verify/CodeInput'
import ResendButton from '../../components/components/verify/ResendButton'
import VerifyButtons from '../../components/components/verify/Buttons'
import TextConfig from '../../lib/configs/site/text.configs'
import { ROUTES, loginPathWithRedirect } from '../../lib/configs/site/routes'
import { describeError } from '../../lib/auth/api'
import { UnauthorizedError } from '../../lib/auth/profile'
import { clearAuthSession, getRedirectTarget } from '../../lib/auth/session'
import { confirmCode, getVerifyStatus, resendCode, VerifyError } from '../../lib/auth/verify'
import '../../assets/css/auth.css'

function Verify() {
    const navigate = useNavigate()
    const location = useLocation()
    const [email, setEmail] = useState('')
    const [code, setCode] = useState('')
    const [resendIn, setResendIn] = useState(0)
    const [loaded, setLoaded] = useState(false)
    const [error, setError] = useState('')
    const [notice, setNotice] = useState('')
    const [isLoading, setIsLoading] = useState(false)
    const [isSending, setIsSending] = useState(false)

    useEffect(() => {
        let cancelled = false

        async function load() {
            try {
                const status = await getVerifyStatus()

                if (cancelled) {
                    return
                }

                if (status.verified) {
                    navigate(getRedirectTarget(location.search), { replace: true })
                    return
                }

                setEmail(status.email)
                setResendIn(status.resend_in)
                setLoaded(true)

                // Код мог ещё не дойти из очереди, а на этой странице пользователь может оказаться и без письма
                if (status.expires_in === 0 && status.resend_in === 0) {
                    const result = await resendCode()

                    if (!cancelled) {
                        setResendIn(result.resendIn)
                    }
                }
            } catch (err) {
                if (cancelled) {
                    return
                }

                if (err instanceof UnauthorizedError) {
                    clearAuthSession()
                    navigate(loginPathWithRedirect(ROUTES.VERIFY), { replace: true })
                    return
                }

                setLoaded(true)
                setError(err instanceof VerifyError ? err.message : describeError(err, TextConfig.UnexpectedError))
            }
        }

        load()

        return () => {
            cancelled = true
        }
    }, [location.search, navigate])

    useEffect(() => {
        if (resendIn <= 0) {
            return
        }

        const timer = window.setTimeout(() => setResendIn(resendIn - 1), 1000)

        return () => window.clearTimeout(timer)
    }, [resendIn])

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        setError('')
        setNotice('')
        setIsLoading(true)

        try {
            await confirmCode(code)
            navigate(getRedirectTarget(location.search), { replace: true })
        } catch (err) {
            if (err instanceof UnauthorizedError) {
                clearAuthSession()
                navigate(loginPathWithRedirect(ROUTES.VERIFY), { replace: true })
                return
            }

            setError(err instanceof VerifyError ? err.message : describeError(err, TextConfig.UnexpectedError))
        } finally {
            setIsLoading(false)
        }
    }

    async function handleResend() {
        setError('')
        setNotice('')
        setIsSending(true)

        try {
            const result = await resendCode()

            if (result.verified) {
                navigate(getRedirectTarget(location.search), { replace: true })
                return
            }

            setCode('')
            setResendIn(result.resendIn)
            setNotice(TextConfig.VerifyResent)
        } catch (err) {
            if (err instanceof UnauthorizedError) {
                clearAuthSession()
                navigate(loginPathWithRedirect(ROUTES.VERIFY), { replace: true })
                return
            }

            if (err instanceof VerifyError && err.retryAfter > 0) {
                setResendIn(err.retryAfter)
            }
            setError(err instanceof VerifyError ? err.message : describeError(err, TextConfig.UnexpectedError))
        } finally {
            setIsSending(false)
        }
    }

    return (
        <div className="auth-page">
            <AuthHeader mode="verify" />
            <main className="auth-main">
                <form className="auth-form" onSubmit={handleSubmit}>
                    <h1>{TextConfig.VerifyTitle}</h1>
                    <p className="auth-intro">{TextConfig.VerifyIntro} <strong>{email}</strong></p>
                    <CodeInput value={code} onChange={setCode} disabled={isLoading || !loaded} />
                    {error && <p className="auth-error" role="alert">{error}</p>}
                    {notice && <p className="auth-notice" role="status">{notice}</p>}
                    <VerifyButtons disabled={isLoading || !loaded || code.length !== 6} isLoading={isLoading} />
                    <div className="auth-divider"><span>{TextConfig.OR}</span></div>
                    <ResendButton seconds={resendIn} sending={isSending} onResend={handleResend} />
                </form>
            </main>
        </div>
    )
}
export default Verify