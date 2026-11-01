import TextConfig from "../../../lib/configs/site/text.configs"
import type { ChatMessage } from "../../../lib/messenger/useThread"

type MessageBubbleProps = {
    message: ChatMessage
    onDelete: (message: ChatMessage) => void
}

function MessageBubble({ message, onDelete }: MessageBubbleProps) {
    const time = new Date(message.createdAt).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })

    return (
        <li className={message.mine ? 'msg-bubble mine' : 'msg-bubble'}>
            {message.text !== null ? (
                <p className="msg-text">{message.text}</p>
            ) : (
                <p className="msg-text broken" title={message.error ?? undefined}>{TextConfig.MsgCannotDecrypt}</p>
            )}
            <span className="msg-meta">
                {time}
                {message.mine && (
                    <button type="button" onClick={() => onDelete(message)} title={TextConfig.MsgDeleteMessage} aria-label={TextConfig.MsgDeleteMessage}>×</button>
                )}
            </span>
        </li>
    )
}
export default MessageBubble