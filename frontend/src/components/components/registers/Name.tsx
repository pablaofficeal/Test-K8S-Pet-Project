import textConfig from "../../../lib/configs/site/text.configs"

type NameProps = {
  disabled?: boolean
  onChange: (value: string) => void
  value: string
}

function RegisterName({ disabled = false, onChange, value }: NameProps) {
  return (
    <div className="login-field">
      <label className="login-bubble" htmlFor="name">{textConfig.RegisterAskName}</label>
      <div className="login-input">
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" /></svg>
        <input
          id="name"
          type="text"
          value={value}
          onChange={(event) => onChange(event.target.value)}
          autoComplete="username"
          placeholder={textConfig.ExempelName}
          maxLength={64}
          disabled={disabled}
          required
        />
      </div>
    </div>
  )
}
export default RegisterName