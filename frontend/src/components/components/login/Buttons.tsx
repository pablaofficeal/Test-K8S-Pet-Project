import TextConfig from "../../../lib/configs/site/text.configs"

type ButtonsProps = {
    disabled?: boolean
    isLoading?: boolean
}

function LoginButtons({ disabled = false, isLoading = false }: ButtonsProps) {
    return (
        <>
            <button type="submit" className="auth-submit" disabled={disabled}>
                {isLoading ? TextConfig.LoginSubmitting : TextConfig.LoginSubmit}
            </button>
            <div className="auth-divider"><span>{TextConfig.OR}</span></div>
            <button type="button" className="auth-secondary" disabled={disabled}>
                <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><circle cx="8" cy="15" r="4" /><path d="m11 12 9-9M16 7l3 3M14 9l2 2" /></svg>
                <span>{TextConfig.LoginDeviceKey}</span>
            </button>
        </>
    )
}
export default LoginButtons