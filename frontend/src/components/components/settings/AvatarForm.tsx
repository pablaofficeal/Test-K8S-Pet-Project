import { useRef, useState, type ChangeEvent } from 'react'
import TextConfig from "../../../lib/configs/site/text.configs"
import { deleteAvatar, uploadAvatar } from "../../../lib/auth/profile"
import { useProfileUpdate } from "../../../lib/auth/useProfileUpdate"
import Avatar from "../profile/Avatar"

type AvatarFormProps = {
    uuid: string
    username: string
    current: string | null
    onChanged: (avatarKey: string | null) => void
}

const MAX_AVATAR_BYTES = 5 * 1024 * 1024
const AVATAR_TYPES = ['image/jpeg', 'image/png', 'image/webp', 'image/gif']

function AvatarForm({ uuid, username, current, onChanged }: AvatarFormProps) {
    const inputRef = useRef<HTMLInputElement>(null)
    const [localError, setLocalError] = useState('')
    const upload = useProfileUpdate<{ avatar_key: string }, File>(uploadAvatar)
    const remove = useProfileUpdate<{ deleted: true }, void>(deleteAvatar)
    const pending = upload.pending || remove.pending
    const error = localError || upload.error || remove.error

    async function handleChange(event: ChangeEvent<HTMLInputElement>) {
        const file = event.target.files?.[0]
        event.target.value = ''
        remove.resetSaved()
        setLocalError('')

        if (!file) {
            return
        }

        if (!AVATAR_TYPES.includes(file.type)) {
            setLocalError(TextConfig.SettingsAvatarBadType)
            return
        }

        if (file.size > MAX_AVATAR_BYTES) {
            setLocalError(TextConfig.SettingsAvatarTooLarge)
            return
        }

        const result = await upload.run(file)

        if (result) {
            onChanged(result.avatar_key)
        }
    }

    async function handleRemove() {
        upload.resetSaved()
        setLocalError('')
        const result = await remove.run()

        if (result) {
            onChanged(null)
        }
    }

    return (
        <div className="settings-form settings-avatar-form">
            <span className="settings-label">{TextConfig.SettingsAvatar}</span>
            <div className="settings-avatar-row">
                <Avatar uuid={uuid} username={username} avatarKey={current} className="settings-avatar" />
                <div className="settings-avatar-body">
                    <p className="settings-hint">{TextConfig.SettingsAvatarHint}</p>
                    <div className="settings-actions">
                        <button type="button" onClick={() => inputRef.current?.click()} disabled={pending}>
                            {upload.pending ? TextConfig.SettingsSaving : TextConfig.SettingsAvatarChoose}
                        </button>
                        {current && (
                            <button type="button" className="settings-secondary" onClick={handleRemove} disabled={pending}>
                                {TextConfig.SettingsAvatarRemove}
                            </button>
                        )}
                        {(upload.saved || remove.saved) && <span className="settings-saved" role="status">{TextConfig.SettingsSaved}</span>}
                    </div>
                </div>
            </div>
            <input ref={inputRef} type="file" accept={AVATAR_TYPES.join(',')} onChange={handleChange} hidden />
            {error && <p className="settings-error" role="alert">{error}</p>}
        </div>
    )
}
export default AvatarForm