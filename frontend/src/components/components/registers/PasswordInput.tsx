import textConfig from "../../../lib/configs/site/text.configs"
import { PASSWORD_MAX_LENGTH, PASSWORD_MIN_LENGTH } from "../../../lib/auth/passwordRules"

type PasswordProps = {
    disabled?: boolean
    onChange: (value: string) => void
    value: string
}

function RegisterPassword({ disabled = false, onChange, value }: PasswordProps) {
    return (
        <div className="auth-field">
            <label htmlFor="password">{textConfig.Password}</label>
            <input
                id="password"
                type="password"
                value={value}
                onChange={(event) => onChange(event.target.value)}
                autoComplete="new-password"
                placeholder={textConfig.ExempelPassword}
                minLength={PASSWORD_MIN_LENGTH}
                maxLength={PASSWORD_MAX_LENGTH}
                disabled={disabled}
                required
            />
        </div>
    )
}
export default RegisterPassword