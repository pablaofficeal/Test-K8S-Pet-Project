import { useCallback, useEffect, useRef, useState } from 'react'
import { deleteMessage, getMessages, getUserKey, markRead, sendMessage, type StoredMessage } from './api'
import { decryptEnvelope, encryptText } from './crypto/envelope.ts'
import type { DualAlphabets } from './crypto/alphabet.ts'
import type { PrivateKeyDoc, PublicKeyDoc } from './crypto/types.ts'
import TextConfig from '../configs/site/text.configs'
import { messageOf } from './errors'
import { AlphabetCharacterError } from './crypto/noise.ts'

const POLL_MS = 3000

export interface ChatMessage {
    id: number
    mine: boolean
    createdAt: string
    read: boolean
    text: string | null
    error: string | null
}

type ThreadKeys = {
    me: string
    peerUuid: string
    privateDoc: PrivateKeyDoc
    alphabets: DualAlphabets
}

// Переписка с одним собеседником: расшифровываем в браузере, шифруем перед отправкой, опрашиваем новые сообщения.
// Сервер видит только конверты
export function useThread({ me, peerUuid, privateDoc, alphabets }: ThreadKeys) {
    const [messages, setMessages] = useState<ChatMessage[]>([])
    const [peerKey, setPeerKey] = useState<PublicKeyDoc | null>(null)
    const [loaded, setLoaded] = useState(false)
    const [error, setError] = useState('')
    const [sending, setSending] = useState(false)
    const lastId = useRef(0)
    const peerKeyRef = useRef<PublicKeyDoc | null>(null)

    const decrypt = useCallback(async (stored: StoredMessage, key: PublicKeyDoc | null): Promise<ChatMessage> => {
        const mine = stored.sender_uuid === me
        const base = { id: stored.id, mine, createdAt: stored.created_at, read: stored.read_at !== null }

        if (!key) {
            return { ...base, text: null, error: TextConfig.MsgNoKeyUser }
        }

        try {
            const text = await decryptEnvelope({ envelope: stored.envelope, alphabets, privateDoc, peerPublicDoc: key, role: mine ? 'sender' : 'recipient' })

            return { ...base, text, error: null }
        } catch (err) {
            return { ...base, text: null, error: err instanceof Error ? err.message : TextConfig.MsgCannotDecrypt }
        }
    }, [me, alphabets, privateDoc])

    const append = useCallback(async (stored: StoredMessage[]) => {
        if (stored.length === 0) {
            return
        }

        const decrypted = await Promise.all(stored.map((item) => decrypt(item, peerKeyRef.current)))
        lastId.current = Math.max(lastId.current, ...stored.map((item) => item.id))
        setMessages((current) => {
            const known = new Set(current.map((item) => item.id))

            return [...current, ...decrypted.filter((item) => !known.has(item.id))]
        })

        if (stored.some((item) => item.sender_uuid !== me && item.read_at === null)) {
            markRead(peerUuid).catch(() => undefined)
        }
    }, [decrypt, me, peerUuid])

    useEffect(() => {
        let cancelled = false
        lastId.current = 0
        setMessages([])
        setLoaded(false)
        setError('')

        async function start() {
            try {
                const key = await getUserKey(peerUuid)
                peerKeyRef.current = key
                setPeerKey(key)
                const stored = await getMessages(peerUuid)

                if (!cancelled) {
                    await append(stored)
                    setLoaded(true)
                }
            } catch (err) {
                if (!cancelled) {
                    setError(messageOf(err, TextConfig.MsgLoadFailed))
                    setLoaded(true)
                }
            }
        }

        start()

        const timer = window.setInterval(async () => {
            if (document.visibilityState !== 'visible') {
                return
            }

            try {
                const fresh = await getMessages(peerUuid, lastId.current)

                if (!cancelled) {
                    await append(fresh)
                }
            } catch {
                // сеть моргнула, следующий опрос догонит
            }
        }, POLL_MS)

        return () => {
            cancelled = true
            window.clearInterval(timer)
        }
    }, [peerUuid, append])

    async function send(text: string): Promise<boolean> {
        const key = peerKeyRef.current

        if (!key) {
            setError(TextConfig.MsgNoKeyUser)
            return false
        }

        setSending(true)
        setError('')

        try {
            const envelope = await encryptText({ text, alphabets, privateDoc, peerPublicDoc: key })
            await sendMessage(peerUuid, envelope)
            await append(await getMessages(peerUuid, lastId.current))

            return true
        } catch (err) {
            setError(err instanceof AlphabetCharacterError
                ? `${TextConfig.MsgCharacterNotAllowed}: ${err.character === '\n' ? '\\n' : err.character}`
                : messageOf(err, TextConfig.MsgSendFailed))
            getUserKey(peerUuid).then((fresh) => {
                peerKeyRef.current = fresh
                setPeerKey(fresh)
            }, () => undefined)

            return false
        } finally {
            setSending(false)
        }
    }

    async function remove(id: number) {
        try {
            await deleteMessage(id)
            setMessages((current) => current.filter((item) => item.id !== id))
        } catch (err) {
            setError(messageOf(err, TextConfig.MsgSendFailed))
        }
    }

    return { messages, peerKey, loaded, error, sending, send, remove }
}