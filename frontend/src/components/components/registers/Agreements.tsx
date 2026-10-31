import TextConfig from "../../../lib/configs/site/text.configs"

type AgreementsProps = {
    accepted: boolean
    onAcceptedChange: (value: boolean) => void
    disabled?: boolean
}

function Agreements({ accepted, onAcceptedChange, disabled = false }: AgreementsProps) {
    return (
        <div className="auth-agreements">
            <label className="auth-check">
                <input type="checkbox" checked={accepted} onChange={(event) => onAcceptedChange(event.target.checked)} disabled={disabled} />
                <span className="auth-check-box" aria-hidden="true" />
                <span>
                    {TextConfig.IAgreeToThe} <em>{TextConfig.TermsOfService}</em> {TextConfig.RegisterAnd} <em>{TextConfig.PrivacyPolicy}</em>
                </span>
            </label>
        </div>
    )
}
export default Agreements