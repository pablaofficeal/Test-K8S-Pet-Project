import Shell from '../../components/menu/shell'
import ProfileSide from '../../components/components/profile/ProfileSide'
import ProfileCard from '../../components/components/profile/ProfileCard'
import VerifyBanner from '../../components/components/profile/VerifyBanner'
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
                {!profile && <h1>{TextConfig.ProfileTitle}</h1>}
                {error && <p className="dashboard-error" role="alert">{error}</p>}
                {!profile && !error && <p className="dashboard-muted">{TextConfig.DashboardLoading}</p>}
                {profile && (
                    <div className="profile-layout">
                        <ProfileSide profile={profile} />
                        <section className="profile-main">
                            <p className="profile-label">{TextConfig.ProfileTitle}</p>
                            <h2 className="profile-heading">{TextConfig.ProfileAccountTitle}</h2>
                            {!profile.email_verified && <VerifyBanner />}
                            <ProfileCard profile={profile} />
                        </section>
                    </div>
                )}
            </main>
        </Shell>
    )
}
export default Profile