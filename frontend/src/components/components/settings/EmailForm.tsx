import { useState, type FormEvent } from 'react'
import TextConfig from "../../../lib/configs/site/text.configs"
import { updateEmail } from "../../../lib/auth/profile"
import { useProfileUpdate } from "../../../lib/auth/useProfileUpdate"

type EmailFormProps = {
    current: string
    onSaved: (email: string) => void
}

function EmailForm({ current, onSaved }: EmailFormProps) {
    const [value, setValue] = useState(current)
    const { run, pending, error, saved, resetSaved } = useProfileUpdate(updateEmail)
    const next = value.trim()

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        const result = await run(next)

        if (result) {
            setValue(result.email)
            onSaved(result.email)
        }
    }

    return (
        <form className="settings-form" onSubmit={handleSubmit}>
            <label htmlFor="settings-email">{TextConfig.SettingsEmail}</label>
            <input
                id="settings-email"
                type="email"
                value={value}
                onChange={(event) => { setValue(event.target.value); resetSaved() }}
                autoComplete="email"
                maxLength={254}
                disabled={pending}
                required
            />
            <p className="settings-hint">{TextConfig.SettingsEmailHint}</p>
            {error && <p className="settings-error" role="alert">{error}</p>}
            <div className="settings-actions">
                <button type="submit" disabled={pending || next === '' || next === current}>
                    {pending ? TextConfig.SettingsSaving : TextConfig.SettingsSave}
                </button>
                {saved && <span className="settings-saved" role="status">{TextConfig.SettingsSaved}</span>}
            </div>
        </form>
    )
}
export default EmailForm