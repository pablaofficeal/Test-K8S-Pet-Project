import { useEffect, useRef, useState, type FormEvent, type KeyboardEvent } from 'react'
import TextConfig from "../../../lib/configs/site/text.configs"
import type { DualAlphabets } from "../../../lib/messenger/crypto/alphabet.ts"
import type { PrivateKeyDoc } from "../../../lib/messenger/crypto/types.ts"
import { useThread, type ChatMessage } from "../../../lib/messenger/useThread"
import { useConfirm } from "../../../lib/todo/useConfirm"
import ConfirmDialog from "../todo/ConfirmDialog"
import MessageBubble from "./MessageBubble"

type ThreadProps = {
    me: string
    peerUuid: string
    peerName: string
    privateDoc: PrivateKeyDoc
    alphabets: DualAlphabets
}

function Thread({ me, peerUuid, peerName, privateDoc, alphabets }: ThreadProps) {
    const { messages, peerKey, loaded, error, sending, send, remove } = useThread({ me, peerUuid, privateDoc, alphabets })
    const confirm = useConfirm()
    const [text, setText] = useState('')
    const endRef = useRef<HTMLLIElement>(null)

    useEffect(() => {
        endRef.current?.scrollIntoView({ block: 'end' })
    }, [messages.length, loaded])

    async function handleSubmit(event?: FormEvent<HTMLFormElement>) {
        event?.preventDefault()
        const value = text.trim()

        if (value === '' || sending) {
            return
        }

        if (await send(value)) {
            setText('')
        }
    }

    function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
        if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault()
            handleSubmit()
        }
    }

    async function handleDelete(message: ChatMessage) {
        if (await confirm.ask(TextConfig.MsgDeleteConfirm)) {
            remove(message.id)
        }
    }

    return (
        <section className="msg-thread">
            <header className="msg-thread-header">
                <h2>{peerName}</h2>
                <span className="msg-lock">{TextConfig.MsgEncrypted}</span>
            </header>
            <ul className="msg-messages">
                {!loaded && <li className="msg-muted">{TextConfig.DashboardLoading}</li>}
                {messages.map((message) => (
                    <MessageBubble key={message.id} message={message} onDelete={handleDelete} />
                ))}
                <li ref={endRef} aria-hidden="true" />
            </ul>
            {peerKey === null && loaded && <p className="msg-warning">{TextConfig.MsgNoKeyUser}</p>}
            {error && <p className="dashboard-error" role="alert">{error}</p>}
            <form className="msg-composer" onSubmit={handleSubmit}>
                <textarea
                    value={text}
                    onChange={(event) => setText(event.target.value)}
                    onKeyDown={handleKeyDown}
                    placeholder={TextConfig.MsgTypeMessage}
                    aria-label={TextConfig.MsgTypeMessage}
                    rows={2}
                    disabled={peerKey === null}
                />
                <button type="submit" className="todo-button" disabled={sending || text.trim() === '' || peerKey === null}>
                    {sending ? TextConfig.MsgSending : TextConfig.MsgSend}
                </button>
            </form>
            <ConfirmDialog message={confirm.message} onAnswer={confirm.answer} />
        </section>
    )
}
export default Thread