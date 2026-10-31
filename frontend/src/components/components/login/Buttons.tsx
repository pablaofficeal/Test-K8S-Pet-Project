import TextConfig from "../../../lib/configs/site/text.configs"

type ButtonsProps = {
    disabled?: boolean
    isLoading?: boolean
}

function LoginButtons({ disabled = false, isLoading = false }: ButtonsProps) {
    return (
        <button type="submit" className="auth-submit" disabled={disabled}>
            {isLoading ? TextConfig.LoginSubmitting : TextConfig.NavLogin}
        </button>
    )
}
export default LoginButtons