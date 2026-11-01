import { useCallback, useRef, useState } from 'react'

// Подтверждение действия через диалог: ask() возвращает промис, который закрывается ответом пользователя
export function useConfirm() {
    const [message, setMessage] = useState<string | null>(null)
    const resolver = useRef<((ok: boolean) => void) | null>(null)

    const ask = useCallback((text: string) => new Promise<boolean>((resolve) => {
        resolver.current?.(false)
        resolver.current = resolve
        setMessage(text)
    }), [])

    const answer = useCallback((ok: boolean) => {
        resolver.current?.(ok)
        resolver.current = null
        setMessage(null)
    }, [])

    return { message, ask, answer }
}