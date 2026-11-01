import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import TextConfig from "../../../lib/configs/site/text.configs"
import { userMessengerThreadPath } from "../../../lib/configs/site/routes"
import { searchUsers, type Conversation, type UserHit } from "../../../lib/messenger/api"

type ConversationListProps = {
    uuid: string
    activePeer: string | undefined
    conversations: Conversation[]
}

function ConversationList({ uuid, activePeer, conversations }: ConversationListProps) {
    const navigate = useNavigate()
    const [query, setQuery] = useState('')
    const [hits, setHits] = useState<UserHit[] | null>(null)

    useEffect(() => {
        const text = query.trim()

        if (text.length < 2) {
            setHits(null)
            return
        }

        let cancelled = false
        const timer = window.setTimeout(() => {
            searchUsers(text).then(
                (found) => {
                    if (!cancelled) {
                        setHits(found)
                    }
                },
                () => {
                    if (!cancelled) {
                        setHits([])
                    }
                },
            )
        }, 300)

        return () => {
            cancelled = true
            window.clearTimeout(timer)
        }
    }, [query])

    function open(hit: UserHit) {
        setQuery('')
        setHits(null)
        navigate(userMessengerThreadPath(uuid, hit.uuid), { state: { username: hit.username } })
    }

    return (
        <aside className="msg-list">
            <input
                type="search"
                className="msg-search"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder={TextConfig.MsgSearchUsers}
                aria-label={TextConfig.MsgSearchUsers}
            />
            {hits !== null ? (
                <ul className="msg-results">
                    {hits.length === 0 && <li className="msg-muted">{TextConfig.MsgNoUsersFound}</li>}
                    {hits.map((hit) => (
                        <li key={hit.uuid}>
                            <button type="button" onClick={() => open(hit)}>
                                <span>{hit.username}</span>
                                {hit.key_id === null && <small>{TextConfig.MsgNoKeyUser}</small>}
                            </button>
                        </li>
                    ))}
                </ul>
            ) : (
                <ul className="msg-conversations">
                    {conversations.length === 0 && <li className="msg-muted">{query.trim().length === 1 ? TextConfig.MsgSearchHint : TextConfig.MsgNoChats}</li>}
                    {conversations.map((item) => (
                        <li key={item.peer_uuid}>
                            <Link
                                to={userMessengerThreadPath(uuid, item.peer_uuid)}
                                state={{ username: item.username }}
                                className={item.peer_uuid === activePeer ? 'active' : ''}
                            >
                                <span>{item.username || item.peer_uuid.slice(0, 8)}</span>
                                {item.unread > 0 && <em>{item.unread}</em>}
                            </Link>
                        </li>
                    ))}
                </ul>
            )}
        </aside>
    )
}
export default ConversationList