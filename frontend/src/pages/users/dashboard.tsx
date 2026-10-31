import { Link } from 'react-router-dom'
import Shell from '../../components/menu/shell'
import VerifyBanner from '../../components/components/profile/VerifyBanner'
import TextConfig from '../../lib/configs/site/text.configs'
import { ROUTES, userDashboardPath } from '../../lib/configs/site/routes'
import { useProfile } from '../../lib/auth/useProfile'
import '../../assets/css/dashboard.css'

function Dashboard() {
    const { profile, error } = useProfile(userDashboardPath, ROUTES.DASHBOARD)

    return (
        <Shell dashboard>
            <main className="dashboard">
                <h1>{TextConfig.DashboardTitle}</h1>
                {error && <p className="dashboard-error" role="alert">{error}</p>}
                {!profile && !error && <p className="dashboard-muted">{TextConfig.DashboardLoading}</p>}
                {profile && (
                    <>
                        <p className="dashboard-muted">{TextConfig.DashboardWelcome} {profile.username}</p>
                        {!profile.email_verified && <VerifyBanner />}
                        <div className="dashboard-links">
                            <Link to={ROUTES.PROFILE}>
                                <strong>{TextConfig.ProfileTitle}</strong>
                                <span>{TextConfig.ProfileDescription}</span>
                            </Link>
                            <Link to={ROUTES.SETTINGS}>
                                <strong>{TextConfig.SettingsTitle}</strong>
                                <span>{TextConfig.SettingsDescription}</span>
                            </Link>
                        </div>
                    </>
                )}
            </main>
        </Shell>
    )
}
export default Dashboard