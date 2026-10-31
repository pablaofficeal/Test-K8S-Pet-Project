import TextConfig from "../../../lib/configs/site/text.configs"

type ResendButtonProps = {
    seconds: number
    sending: boolean
    onResend: () => void
}

function ResendButton({ seconds, sending, onResend }: ResendButtonProps) {
    const waiting = seconds > 0

    return (
        <button type="button" className="auth-secondary" onClick={onResend} disabled={waiting || sending}>
            {waiting ? `${TextConfig.VerifyResendIn} ${seconds} ${TextConfig.VerifySecondsShort}` : TextConfig.VerifyResend}
        </button>
    )
}
export default ResendButton