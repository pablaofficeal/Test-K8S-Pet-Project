import TextConfig from "../../../lib/configs/site/text.configs"

type ButtonsProps = {
    disabled?: boolean
    isLoading?: boolean
}

function VerifyButtons({ disabled = false, isLoading = false }: ButtonsProps) {
    return (
        <button type="submit" className="auth-submit" disabled={disabled}>
            {isLoading ? TextConfig.VerifySubmitting : TextConfig.VerifySubmit}
        </button>
    )
}
export default VerifyButtons