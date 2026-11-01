import { Link } from 'react-router-dom'
import TextConfig from "../../../lib/configs/site/text.configs"
import { ROUTES } from "../../../lib/configs/site/routes"
import Avatar from "./Avatar"
import { formatDate } from "../../../lib/i18n/date"
import type { ProfileResponse } from "../../../lib/auth/profile"

type ProfileSideProps = {
    profile: ProfileResponse
}

function ProfileSide({ profile }: ProfileSideProps) {
    return (
        <aside className="profile-side">
            <Avatar uuid={profile.uuid} username={profile.username} avatarKey={profile.avatar_key} />
            <Link to={ROUTES.SETTINGS} className="profile-edit">{TextConfig.ProfileEdit}</Link>
            <h1 className="profile-name">{profile.username}</h1>
            <p className="profile-handle">{profile.email}</p>
            <div className="profile-badges">
                <span>{TextConfig.DashboardUserId} {profile.uuid.slice(0, 8)}</span>
                <span className={profile.email_verified ? 'ok' : 'warn'}>{profile.email_verified ? TextConfig.VerifyBadgeOk : TextConfig.VerifyBadgeNo}</span>
            </div>
            <p className="profile-joined">{TextConfig.DashboardCreated}: <time dateTime={profile.created_at}>{formatDate(profile.created_at)}</time></p>
        </aside>
    )
}
export default ProfileSide