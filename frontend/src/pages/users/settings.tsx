import Shell from '../../components/menu/shell'
import NameForm from '../../components/components/settings/NameForm'
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
                    <div className="settings-list">
                        <NameForm current={profile.username} onSaved={(username) => setProfile({ ...profile, username })} />
                        <EmailForm current={profile.email} onSaved={(email) => setProfile({ ...profile, email })} />
                    </div>
                )}
            </main>
        </Shell>
    )
}
export default Settings