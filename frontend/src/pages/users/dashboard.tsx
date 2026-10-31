import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import Shell from '../../components/menu/shell'
import TextConfig from '../../lib/configs/site/text.configs'
import { ROUTES, loginPathWithRedirect, userDashboardPath } from '../../lib/configs/site/routes'
import { describeError } from '../../lib/auth/api'
import { getProfile, UnauthorizedError, type ProfileResponse } from '../../lib/auth/profile'
import { clearAuthSession } from '../../lib/auth/session'
import { getCurrentLanguage } from '../../lib/i18n/i18n'
import '../assets/css/dashboard.css'

function Dashboard() {
    const { uuid } = useParams()
    const navigate = useNavigate()
    const [profile, setProfile] = useState<ProfileResponse | null>(null)
    const [error, setError] = useState('')

    useEffect(() => {
        let cancelled = false

        getProfile().then(
            (data) => {
                if (cancelled) {
                    return
                }

                setProfile(data)

                // uuid в адресе и в куке - только для красоты, настоящий берём из ответа API
                if (data.uuid !== uuid) {
                    navigate(userDashboardPath(data.uuid), { replace: true })
                }
            },
            (err) => {
                if (cancelled) {
                    return
                }

                if (err instanceof UnauthorizedError) {
                    clearAuthSession()
                    navigate(loginPathWithRedirect(ROUTES.DASHBOARD), { replace: true })
                    return
                }

                setError(describeError(err, TextConfig.DashboardLoadFailed))
            },
        )

        return () => {
            cancelled = true
        }
    }, [uuid, navigate])

    return (
        <Shell dashboard>
            <main className="dashboard">
                <h1>{TextConfig.DashboardTitle}</h1>
                {error && <p className="dashboard-error" role="alert">{error}</p>}
                {!profile && !error && <p className="dashboard-muted">{TextConfig.DashboardLoading}</p>}
                {profile && (
                    <>
                        <p className="dashboard-muted">{TextConfig.DashboardWelcome} {profile.username}</p>
                        <dl className="dashboard-card">
                            <div>
                                <dt>{TextConfig.SettingsName}</dt>
                                <dd>{profile.username}</dd>
                            </div>
                            <div>
                                <dt>{TextConfig.SettingsEmail}</dt>
                                <dd>{profile.email}</dd>
                            </div>
                            <div>
                                <dt>{TextConfig.DashboardUserId}</dt>
                                <dd><code>{profile.uuid}</code></dd>
                            </div>
                            <div>
                                <dt>{TextConfig.DashboardCreated}</dt>
                                <dd><time dateTime={profile.created_at}>{new Date(profile.created_at).toLocaleDateString(getCurrentLanguage(), { dateStyle: 'long' })}</time></dd>
                            </div>
                        </dl>
                    </>
                )}
            </main>
        </Shell>
    )
}
export default Dashboard