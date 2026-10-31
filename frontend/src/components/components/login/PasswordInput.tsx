import { useState } from 'react'
import textConfig from "../../../lib/configs/site/text.configs"

type PasswordProps = {
    disabled?: boolean
    onChange: (value: string) => void
    value: string
}

function LoginPassword({ disabled = false, onChange, value }: PasswordProps) {
    const [visible, setVisible] = useState(false)

    return (
        <div className="auth-field">
            <div className="auth-field-head">
                <label htmlFor="password">{textConfig.Password}</label>
                <button type="button" className="auth-link-button" onClick={() => setVisible(!visible)}>
                    {visible ? textConfig.LoginHidePassword : textConfig.LoginShowPassword}
                </button>
            </div>
            <input
                id="password"
                type={visible ? 'text' : 'password'}
                value={value}
                onChange={(event) => onChange(event.target.value)}
                autoComplete="current-password"
                placeholder={textConfig.ExempelPassword}
                disabled={disabled}
                required
            />
        </div>
    )
}
export default LoginPassword