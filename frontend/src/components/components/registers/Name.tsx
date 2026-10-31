import textConfig from "../../../lib/configs/site/text.configs"

type NameProps = {
    disabled?: boolean
    onChange: (value: string) => void
    value: string
}

function RegisterName({ disabled = false, onChange, value }: NameProps) {
    return (
        <div className="auth-field">
            <label htmlFor="name">{textConfig.RegisterTeamName}</label>
            <input
                id="name"
                type="text"
                value={value}
                onChange={(event) => onChange(event.target.value)}
                autoComplete="organization"
                placeholder={textConfig.RegisterTeamNamePlaceholder}
                maxLength={64}
                disabled={disabled}
                required
            />
        </div>
    )
}
export default RegisterName