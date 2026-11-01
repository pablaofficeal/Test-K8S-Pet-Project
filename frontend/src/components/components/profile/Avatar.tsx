import { useState } from 'react'
import { avatarGradient, avatarLetter } from "../../../lib/auth/avatar"
import { avatarUrl } from "../../../lib/auth/profile"

type AvatarProps = {
    uuid: string
    username: string
    avatarKey: string | null
    className?: string
}

function Avatar({ uuid, username, avatarKey, className = 'profile-avatar' }: AvatarProps) {
    const [failedKey, setFailedKey] = useState<string | null>(null)
    const showImage = avatarKey !== null && failedKey !== avatarKey

    return (
        <div className={className} style={{ background: avatarGradient(uuid) }} aria-hidden="true">
            {showImage ? <img src={avatarUrl(avatarKey)} alt="" onError={() => setFailedKey(avatarKey)} /> : avatarLetter(username)}
        </div>
    )
}
export default Avatar