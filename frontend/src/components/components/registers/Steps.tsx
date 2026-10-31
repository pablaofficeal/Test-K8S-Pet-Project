import TextConfig from "../../../lib/configs/site/text.configs"

function Steps() {
    const steps = [TextConfig.RegisterStepAccount, TextConfig.RegisterStepRecovery, TextConfig.RegisterStepDevice]

    return (
        <ol className="auth-steps">
            {steps.map((step, index) => (
                <li key={step} className={index === 0 ? 'active' : ''} aria-current={index === 0 ? 'step' : undefined}>
                    <b>{index + 1}</b>
                    <span>{step}</span>
                </li>
            ))}
        </ol>
    )
}
export default Steps