import Shell from '../../components/menu/shell'
import ProfileCard from '../../components/components/profile/ProfileCard'
import TextConfig from '../../lib/configs/site/text.configs'
import { ROUTES, userProfilePath } from '../../lib/configs/site/routes'
import { useProfile } from '../../lib/auth/useProfile'
import '../../assets/css/dashboard.css'
import '../../assets/css/profile.css'

function Profile() {
    const { profile, error } = useProfile(userProfilePath, ROUTES.PROFILE)

    return (
        <Shell dashboard>
            <main className="dashboard">
                <h1>{TextConfig.ProfileTitle}</h1>
                <p className="dashboard-muted">{TextConfig.ProfileDescription}</p>
                {error && <p className="dashboard-error" role="alert">{error}</p>}
                {!profile && !error && <p className="dashboard-muted">{TextConfig.DashboardLoading}</p>}
                {profile && <ProfileCard profile={profile} />}
            </main>
        </Shell>
    )
}
export default Profile