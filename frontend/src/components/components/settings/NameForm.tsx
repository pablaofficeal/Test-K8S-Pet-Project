import { useState, type FormEvent } from 'react'
import TextConfig from "../../../lib/configs/site/text.configs"
import { updateName } from "../../../lib/auth/profile"
import { useProfileUpdate } from "../../../lib/auth/useProfileUpdate"

type NameFormProps = {
    current: string
    onSaved: (username: string) => void
}

function NameForm({ current, onSaved }: NameFormProps) {
    const [value, setValue] = useState(current)
    const { run, pending, error, saved, resetSaved } = useProfileUpdate(updateName)
    const next = value.trim()

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        const result = await run(next)

        if (result) {
            setValue(result.username)
            onSaved(result.username)
        }
    }

    return (
        <form className="settings-form" onSubmit={handleSubmit}>
            <label htmlFor="settings-name">{TextConfig.SettingsName}</label>
            <input
                id="settings-name"
                type="text"
                value={value}
                onChange={(event) => { setValue(event.target.value); resetSaved() }}
                autoComplete="username"
                maxLength={64}
                disabled={pending}
                required
            />
            <p className="settings-hint">{TextConfig.SettingsNameHint}</p>
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
export default NameForm