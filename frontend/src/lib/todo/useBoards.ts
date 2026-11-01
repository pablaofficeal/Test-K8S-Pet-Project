import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import TextConfig from '../configs/site/text.configs'
import { ROUTES, loginPathWithRedirect } from '../configs/site/routes'
import { describeError } from '../auth/api'
import { UnauthorizedError } from '../auth/profile'
import { clearAuthSession } from '../auth/session'
import { createBoard, deleteBoard, getBoards, importBoard } from './api'
import type { BoardSummary } from './types'

// Список проектов: загрузка, создание, импорт markdown и удаление
export function useBoards() {
    const navigate = useNavigate()
    const [boards, setBoards] = useState<BoardSummary[] | null>(null)
    const [error, setError] = useState('')
    const [pending, setPending] = useState(false)

    const fail = useCallback((err: unknown, fallback: string) => {
        if (err instanceof UnauthorizedError) {
            clearAuthSession()
            navigate(loginPathWithRedirect(ROUTES.TODO), { replace: true })
            return
        }

        setError(describeError(err, fallback))
    }, [navigate])

    useEffect(() => {
        let cancelled = false

        getBoards().then(
            (data) => {
                if (!cancelled) {
                    setBoards(data)
                }
            },
            (err) => {
                if (!cancelled) {
                    fail(err, TextConfig.TodoLoadFailed)
                }
            },
        )

        return () => {
            cancelled = true
        }
    }, [fail])

    async function run<T>(action: () => Promise<T>): Promise<T | null> {
        setPending(true)
        setError('')

        try {
            return await action()
        } catch (err) {
            fail(err, TextConfig.TodoActionFailed)

            return null
        } finally {
            setPending(false)
        }
    }

    async function create(title: string) {
        const board = await run(() => createBoard(title))

        if (board) {
            setBoards((current) => [{ ...board, nodes_count: 0, tasks_total: 0, tasks_done: 0 }, ...(current ?? [])])
        }

        return board
    }

    async function importFile(markdown: string) {
        return run(() => importBoard(markdown))
    }

    async function remove(id: string) {
        const done = await run(() => deleteBoard(id).then(() => true))

        if (done) {
            setBoards((current) => (current ?? []).filter((board) => board.id !== id))
        }
    }

    return { boards, error, pending, create, importFile, remove }
}