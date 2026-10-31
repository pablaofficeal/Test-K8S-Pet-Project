import textConfig from "../../../lib/configs/site/text.configs"

type PasswordProps = {
  disabled?: boolean
  onChange: (value: string) => void
  value: string
}

function LoginPassword({ disabled = false, onChange, value }: PasswordProps) {
  return (
    <div className="login-field">
      <label className="login-bubble" htmlFor="password">{textConfig.LoginAskPassword}</label>
      <div className="login-input">
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><rect x="4" y="11" width="16" height="10" rx="2" /><path d="M8 11V7a4 4 0 0 1 8 0v4" /></svg>
        <input
          id="password"
          type="password"
          value={value}
          onChange={(event) => onChange(event.target.value)}
          autoComplete="current-password"
          placeholder={textConfig.ExempelPassword}
          disabled={disabled}
          required
        />
      </div>
    </div>
  )
}
export default LoginPassword