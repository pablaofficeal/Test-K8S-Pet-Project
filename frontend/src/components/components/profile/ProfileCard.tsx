import TextConfig from "../../../lib/configs/site/text.configs"
import { getCurrentLanguage } from "../../../lib/i18n/i18n"
import type { ProfileResponse } from "../../../lib/auth/profile"

type ProfileCardProps = {
    profile: ProfileResponse
}

function ProfileCard({ profile }: ProfileCardProps) {
    return (
        <dl className="profile-card">
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
    )
}
export default ProfileCard