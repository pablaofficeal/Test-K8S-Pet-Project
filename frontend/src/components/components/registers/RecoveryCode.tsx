import { useState } from 'react'
import TextConfig from "../../../lib/configs/site/text.configs"
import { downloadRecoveryCode, formatRecoveryCode } from "../../../lib/auth/recoveryCode"

type RecoveryCodeProps = {
    groups: string[]
    onRegenerate: () => void
    disabled?: boolean
}

function RecoveryCode({ groups, onRegenerate, disabled = false }: RecoveryCodeProps) {
    const [copied, setCopied] = useState(false)

    async function handleCopy() {
        try {
            await navigator.clipboard.writeText(formatRecoveryCode(groups))
            setCopied(true)
            window.setTimeout(() => setCopied(false), 2000)
        } catch {
            setCopied(false)
        }
    }

    return (
        <section className="auth-recovery">
            <header>
                <h2>{TextConfig.RecoveryTitle}</h2>
                <span>{TextConfig.RecoveryCreatedHere}</span>
            </header>
            <ul className="auth-recovery-code">
                {groups.map((group, index) => (
                    <li key={index}>{group}</li>
                ))}
            </ul>
            <div className="auth-recovery-actions">
                <button type="button" onClick={handleCopy} disabled={disabled}>{copied ? TextConfig.RecoveryCopied : TextConfig.RecoveryCopy}</button>
                <button type="button" onClick={() => downloadRecoveryCode(groups)} disabled={disabled}>{TextConfig.RecoveryDownload}</button>
                <button type="button" onClick={onRegenerate} disabled={disabled}>{TextConfig.RecoveryRegenerate}</button>
            </div>
            <p className="auth-recovery-warning">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="9" /><path d="M12 7v6M12 16.5h.01" /></svg>
                <span>{TextConfig.RecoveryWarning}</span>
            </p>
        </section>
    )
}
export default RecoveryCode