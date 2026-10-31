import textConfig from "../../../lib/configs/site/text.configs"

type EmailProps = {
  disabled?: boolean
  onChange: (value: string) => void
  value: string
}

function LoginEmail({ disabled = false, onChange, value }: EmailProps) {
  return (
    <div className="login-field">
      <label className="login-bubble" htmlFor="email">{textConfig.LoginAskEmail}</label>
      <div className="login-input">
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><rect x="2" y="4" width="20" height="16" rx="2" /><path d="m22 7-10 6L2 7" /></svg>
        <input
          id="email"
          type="email"
          value={value}
          onChange={(event) => onChange(event.target.value)}
          autoComplete="email"
          placeholder={textConfig.ExempelEmail}
          disabled={disabled}
          required
        />
      </div>
    </div>
  )
}
export default LoginEmail