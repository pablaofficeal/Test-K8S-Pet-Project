import { useRef, useState } from 'react'
import TextConfig from "../../../lib/configs/site/text.configs"
import type { useMessenger } from "../../../lib/messenger/useMessenger"

type KeySetupProps = {
    messenger: ReturnType<typeof useMessenger>
}

function KeySetup({ messenger }: KeySetupProps) {
    const fileRef = useRef<HTMLInputElement>(null)
    const [passphrase, setPassphrase] = useState('')
    const state = messenger.state

    if (state.phase !== 'setup') {
        return null
    }

    async function handleFile(file: File | undefined) {
        if (file) {
            await messenger.importKey(await file.text())
        }
    }

    const canRestore = state.hasBackup && state.reason !== 'new'

    return (
        <section className="msg-setup">
            <h2>{TextConfig.MsgSetupTitle}</h2>
            <p className="msg-muted">{TextConfig.MsgSetupIntro}</p>
            <p className="msg-warning">{TextConfig.MsgBrowserKeyWarning}</p>
            {state.reason === 'mismatch' && (
                <>
                    <p className="dashboard-error" role="alert">{TextConfig.MsgKeyMismatch}</p>
                    <button type="button" className="todo-button" onClick={() => messenger.useLocalKey()} disabled={messenger.pending}>{TextConfig.MsgUseLocalKey}</button>
                </>
            )}
            {canRestore && (
                <div className="msg-row">
                    <input
                        type="password"
                        value={passphrase}
                        onChange={(event) => setPassphrase(event.target.value)}
                        placeholder={TextConfig.MsgPassphrase}
                        aria-label={TextConfig.MsgPassphrase}
                        autoComplete="current-password"
                    />
                    <button type="button" className="todo-button" onClick={() => messenger.restoreBackup(passphrase)} disabled={messenger.pending || passphrase === ''}>
                        {TextConfig.MsgRestoreBackup}
                    </button>
                </div>
            )}
            <div className="msg-row">
                <button type="button" className="todo-button" onClick={() => messenger.generate()} disabled={messenger.pending}>
                    {messenger.pending ? TextConfig.MsgGenerating : TextConfig.MsgGenerate}
                </button>
                <button type="button" className="todo-button secondary" onClick={() => fileRef.current?.click()} disabled={messenger.pending}>{TextConfig.MsgImportKey}</button>
                <input ref={fileRef} type="file" accept=".json,application/json" hidden onChange={(event) => { handleFile(event.target.files?.[0]); event.target.value = '' }} />
            </div>
            {state.reason !== 'new' && <p className="msg-muted">{TextConfig.MsgRotateWarning}</p>}
            {messenger.error && <p className="dashboard-error" role="alert">{messenger.error}</p>}
        </section>
    )
}
export default KeySetup