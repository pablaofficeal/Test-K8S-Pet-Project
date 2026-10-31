import textConfig from "../../../lib/configs/site/text.configs"

type EmailProps = {
    disabled?: boolean
    onChange: (value: string) => void
    value: string
}

function LoginEmail({ disabled = false, onChange, value }: EmailProps) {
    return (
        <div className="auth-field">
            <label htmlFor="email">{textConfig.LoginEmailLabel}</label>
            <input
                id="email"
                type="email"
                value={value}
                onChange={(event) => onChange(event.target.value)}
                autoComplete="email"
                placeholder={textConfig.RegisterWorkEmailPlaceholder}
                disabled={disabled}
                required
            />
        </div>
    )
}
export default LoginEmail