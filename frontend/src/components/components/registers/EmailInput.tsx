import textConfig from "../../../lib/configs/site/text.configs"

type EmailProps = {
    disabled?: boolean
    onChange: (value: string) => void
    value: string
}

function RegisterEmail({ disabled = false, onChange, value }: EmailProps) {
    return (
        <div className="auth-field">
            <label htmlFor="email">{textConfig.RegisterWorkEmail}</label>
            <input
                id="email"
                type="email"
                value={value}
                onChange={(event) => onChange(event.target.value)}
                autoComplete="email"
                placeholder={textConfig.RegisterWorkEmailPlaceholder}
                maxLength={254}
                disabled={disabled}
                required
            />
        </div>
    )
}
export default RegisterEmail