import Shell from '../../components/menu/shell'
import AvatarForm from '../../components/components/settings/AvatarForm'
import NameForm from '../../components/components/settings/NameForm'
import VerifyBanner from '../../components/components/profile/VerifyBanner'
import EmailForm from '../../components/components/settings/EmailForm'
import TextConfig from '../../lib/configs/site/text.configs'
import { ROUTES, userSettingsPath } from '../../lib/configs/site/routes'
import { useProfile } from '../../lib/auth/useProfile'
import '../../assets/css/dashboard.css'
import '../../assets/css/profile.css'

function Settings() {
    const { profile, setProfile, error } = useProfile(userSettingsPath, ROUTES.SETTINGS)

    return (
        <Shell dashboard>
            <main className="dashboard">
                <h1>{TextConfig.SettingsTitle}</h1>
                <p className="dashboard-muted">{TextConfig.SettingsDescription}</p>
                {error && <p className="dashboard-error" role="alert">{error}</p>}
                {!profile && !error && <p className="dashboard-muted">{TextConfig.DashboardLoading}</p>}
                {profile && (
                    <>
                        {!profile.email_verified && <VerifyBanner />}
                        <div className="settings-list">
                            <AvatarForm uuid={profile.uuid} username={profile.username} current={profile.avatar_key} onChanged={(avatar_key) => setProfile({ ...profile, avatar_key })} />
                            <NameForm current={profile.username} onSaved={(username) => setProfile({ ...profile, username })} />
                            <EmailForm current={profile.email} onSaved={(email) => setProfile({ ...profile, email, email_verified: false })} />
                        </div>
                    </>
                )}
            </main>
        </Shell>
    )
}
export default Settings