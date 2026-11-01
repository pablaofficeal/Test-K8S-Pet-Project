import { useLocation, useParams } from 'react-router-dom'
import Shell from '../../components/menu/shell'
import ChatSettings from '../../components/components/messenger/ChatSettings'
import ConversationList from '../../components/components/messenger/ConversationList'
import KeySetup from '../../components/components/messenger/KeySetup'
import Thread from '../../components/components/messenger/Thread'
import TextConfig from '../../lib/configs/site/text.configs'
import { getUserId } from '../../lib/auth/session'
import { useConversations } from '../../lib/messenger/useConversations'
import { useMessenger } from '../../lib/messenger/useMessenger'
import '../../assets/css/dashboard.css'
import '../../assets/css/todo.css'
import '../../assets/css/messenger.css'

function Messenger() {
    const { peerUuid } = useParams()
    const location = useLocation()
    const uuid = getUserId() ?? ''
    const messenger = useMessenger()
    const state = messenger.state
    const { conversations } = useConversations(state.phase === 'ready')
    const known = conversations.find((item) => item.peer_uuid === peerUuid)
    const passed = (location.state as { username?: string } | null)?.username
    const peerName = known?.username || passed || (peerUuid ?? '').slice(0, 8)

    return (
        <Shell dashboard>
            <main className="dashboard messenger-page">
                <div className="msg-header">
                    <h1>{TextConfig.MessengerTitle}</h1>
                    {state.phase === 'ready' && <ChatSettings messenger={messenger} />}
                </div>
                {state.phase === 'loading' && !messenger.error && <p className="dashboard-muted">{TextConfig.MsgSetupLoading}</p>}
                {state.phase === 'loading' && messenger.error && <p className="dashboard-error" role="alert">{messenger.error}</p>}
                {state.phase === 'setup' && <KeySetup messenger={messenger} />}
                {state.phase === 'ready' && messenger.alphabets && (
                    <div className="msg-layout">
                        <ConversationList uuid={uuid} activePeer={peerUuid} conversations={conversations} />
                        {peerUuid ? (
                            <Thread key={peerUuid} me={uuid} peerUuid={peerUuid} peerName={peerName} privateDoc={state.privateDoc} alphabets={messenger.alphabets} />
                        ) : (
                            <p className="msg-empty">{TextConfig.MsgPickChat}</p>
                        )}
                    </div>
                )}
            </main>
        </Shell>
    )
}
export default Messenger