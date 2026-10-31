import TextConfig from "../../../lib/configs/site/text.configs"

type CodeInputProps = {
    disabled?: boolean
    onChange: (value: string) => void
    value: string
}

function CodeInput({ disabled = false, onChange, value }: CodeInputProps) {
    return (
        <div className="auth-field">
            <label htmlFor="code">{TextConfig.VerifyCodeLabel}</label>
            <input
                id="code"
                className="verify-code"
                type="text"
                inputMode="numeric"
                pattern="[0-9]{6}"
                value={value}
                onChange={(event) => onChange(event.target.value.replace(/\D/g, '').slice(0, 6))}
                autoComplete="one-time-code"
                placeholder="000000"
                maxLength={6}
                disabled={disabled}
                autoFocus
                required
            />
        </div>
    )
}
export default CodeInput