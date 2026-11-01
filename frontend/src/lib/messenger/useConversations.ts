import { useCallback, useEffect, useState } from 'react'
import { getConversations, type Conversation } from './api'

const POLL_MS = 5000

// Список чатов: загружаем сразу и обновляем раз в несколько секунд, пока вкладка видна
export function useConversations(enabled: boolean) {
    const [conversations, setConversations] = useState<Conversation[]>([])

    const reload = useCallback(async () => {
        try {
            setConversations(await getConversations())
        } catch {
            // следующий опрос попробует ещё раз
        }
    }, [])

    useEffect(() => {
        if (!enabled) {
            return
        }

        reload()
        const timer = window.setInterval(() => {
            if (document.visibilityState === 'visible') {
                reload()
            }
        }, POLL_MS)

        return () => window.clearInterval(timer)
    }, [enabled, reload])

    return { conversations, reload }
}