import TextConfig from "../../../lib/configs/site/text.configs"
import { passwordStrength } from "../../../lib/auth/passwordStrength"

type StrengthProps = {
    value: string
}

function Strength({ value }: StrengthProps) {
    const score = passwordStrength(value)
    const labels = ['', TextConfig.StrengthWeak, TextConfig.StrengthFair, TextConfig.StrengthGood, TextConfig.StrengthStrong]

    return (
        <div className="auth-strength" aria-live="polite">
            <div className="auth-strength-bars">
                {[1, 2, 3, 4].map((step) => (
                    <span key={step} className={step <= score ? 'on' : ''} />
                ))}
            </div>
            <span className="auth-strength-label">{labels[score]}</span>
        </div>
    )
}
export default Strength