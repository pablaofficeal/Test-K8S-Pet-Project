import TextConfig from "../../../lib/configs/site/text.configs"

type ButtonsProps = {
    disabled?: boolean
    isLoading?: boolean
}

function RegisterButtons({ disabled = false, isLoading = false }: ButtonsProps) {
    return (
        <button type="submit" className="login-submit" disabled={disabled}>
            <span>{isLoading ? `${TextConfig.RegistersButtons} ...` : TextConfig.RegistersButtons}</span>
            <span className="login-submit-arrow">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M5 12h14M13 6l6 6-6 6" /></svg>
            </span>
        </button>
    )
}
export default RegisterButtons