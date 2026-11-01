import { useState } from 'react'
import TextConfig from "../../../lib/configs/site/text.configs"
import type { useMessenger } from "../../../lib/messenger/useMessenger"

type ChatSettingsProps = {
    messenger: ReturnType<typeof useMessenger>
}

function ChatSettings({ messenger }: ChatSettingsProps) {
    const [passphrase, setPassphrase] = useState('')
    const [first, setFirst] = useState<File | null>(null)
    const [second, setSecond] = useState<File | null>(null)
    const state = messenger.state

    if (state.phase !== 'ready') {
        return null
    }

    function handleExport() {
        if (state.phase !== 'ready') {
            return
        }

        const url = URL.createObjectURL(new Blob([JSON.stringify(state.privateDoc, null, 2)], { type: 'application/json' }))
        const link = document.createElement('a')
        link.href = url
        link.download = 'private.json'
        link.click()
        URL.revokeObjectURL(url)
    }

    async function handleApply() {
        if (first && second) {
            await messenger.applyDictionaries(await first.text(), await second.text())
            setFirst(null)
            setSecond(null)
        }
    }

    return (
        <details className="msg-settings">
            <summary>{TextConfig.MsgDictionaries} · {TextConfig.MsgKeyId}</summary>
            <div className="msg-settings-body">
                <p className="msg-muted">{TextConfig.MsgKeyId}: <code>{state.publicDoc.key_id}</code></p>
                <p className="msg-warning">{TextConfig.MsgBrowserKeyWarning}</p>
                <div className="msg-row">
                    <button type="button" className="todo-button compact secondary" onClick={handleExport}>{TextConfig.MsgExportKey}</button>
                </div>
                <div className="msg-row">
                    <input
                        type="password"
                        value={passphrase}
                        onChange={(event) => setPassphrase(event.target.value)}
                        placeholder={TextConfig.MsgPassphrase}
                        aria-label={TextConfig.MsgPassphrase}
                        autoComplete="new-password"
                    />
                    <button type="button" className="todo-button compact" onClick={() => messenger.saveBackup(passphrase)} disabled={messenger.pending || passphrase.length < 8}>
                        {TextConfig.MsgBackupCreate}
                    </button>
                </div>
                <p className="msg-muted">{messenger.customDictionaries ? TextConfig.MsgDictionaryCustom : TextConfig.MsgDictionaryDefault}</p>
                <label className="msg-file">
                    <span>{TextConfig.MsgDictionaryOne}</span>
                    <input type="file" accept=".json,application/json" onChange={(event) => setFirst(event.target.files?.[0] ?? null)} />
                </label>
                <label className="msg-file">
                    <span>{TextConfig.MsgDictionaryTwo}</span>
                    <input type="file" accept=".json,application/json" onChange={(event) => setSecond(event.target.files?.[0] ?? null)} />
                </label>
                <div className="msg-row">
                    <button type="button" className="todo-button compact" onClick={handleApply} disabled={messenger.pending || !first || !second}>{TextConfig.MsgDictionaryApply}</button>
                    {messenger.customDictionaries && (
                        <button type="button" className="todo-button compact secondary" onClick={() => messenger.resetDictionaries()} disabled={messenger.pending}>{TextConfig.MsgDictionaryReset}</button>
                    )}
                </div>
                {messenger.notice && <p className="msg-notice" role="status">{messenger.notice}</p>}
                {messenger.error && <p className="dashboard-error" role="alert">{messenger.error}</p>}
            </div>
        </details>
    )
}
export default ChatSettings