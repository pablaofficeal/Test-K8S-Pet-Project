import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import TextConfig from '../configs/site/text.configs'
import { ROUTES, loginPathWithRedirect } from '../configs/site/routes'
import { describeError } from '../auth/api'
import { UnauthorizedError } from '../auth/profile'
import { clearAuthSession } from '../auth/session'
import * as api from './api'
import { nextNodePosition } from './grid'
import { collectTaskIds } from './tree'
import type { BoardActions, BoardView, TaskInput } from './types'

// Состояние одного проекта (канваса): блоки, задачи и все изменения.
// Перемещение блока сразу применяется локально, на сервер уходит один запрос при отпускании
export function useBoard(boardId: string) {
    const navigate = useNavigate()
    const [view, setView] = useState<BoardView | null>(null)
    const [error, setError] = useState('')

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

        api.getBoard(boardId).then(
            (data) => {
                if (!cancelled) {
                    setView(data)
                }
            },
            (err) => {
                if (!cancelled) {
                    fail(err, TextConfig.TodoBoardLoadFailed)
                }
            },
        )

        return () => {
            cancelled = true
        }
    }, [boardId, fail])

    async function run<T>(action: () => Promise<T>): Promise<T | null> {
        setError('')

        try {
            return await action()
        } catch (err) {
            fail(err, TextConfig.TodoActionFailed)

            return null
        }
    }

    async function renameBoard(title: string) {
        const done = await run(() => api.renameBoard(boardId, title).then(() => true))

        if (done) {
            setView((current) => current && { ...current, board: { ...current.board, title } })
        }

        return done === true
    }

    async function addNode(title: string) {
        const position = nextNodePosition((view?.nodes ?? []).map((node) => node.pos_x))
        const node = await run(() => api.createNode(boardId, title, position.x, position.y))

        if (node) {
            setView((current) => current && { ...current, nodes: [...current.nodes, node] })
        }

        return node !== null
    }

    async function removeBoard() {
        const done = await run(() => api.deleteBoard(boardId).then(() => true))

        return done === true
    }

    async function exportMarkdown() {
        return run(() => api.exportBoard(boardId))
    }

    const actions: BoardActions = {
        async renameNode(id: string, title: string) {
            const done = await run(() => api.renameNode(id, title).then(() => true))

            if (done) {
                setView((current) => current && { ...current, nodes: current.nodes.map((node) => (node.id === id ? { ...node, title } : node)) })
            }

            return done === true
        },

        moveNode(id: string, x: number, y: number) {
            setView((current) => current && { ...current, nodes: current.nodes.map((node) => (node.id === id ? { ...node, pos_x: x, pos_y: y } : node)) })
            run(() => api.moveNode(id, x, y)).then((result) => {
                if (result === null) {
                    api.getBoard(boardId).then(setView, () => undefined)
                }
            })
        },

        async deleteNode(id: string) {
            const done = await run(() => api.deleteNode(id).then(() => true))

            if (done) {
                setView((current) => current && {
                    ...current,
                    nodes: current.nodes.filter((node) => node.id !== id),
                    tasks: current.tasks.filter((task) => task.node_id !== id),
                })
            }

            return done === true
        },

        async createTask(nodeId: string, parentId: string | null, input: TaskInput) {
            const task = await run(() => api.createTask(nodeId, parentId, input))

            if (task) {
                setView((current) => current && { ...current, tasks: [...current.tasks, task] })
            }

            return task !== null
        },

        async editTask(id: string, patch: Partial<TaskInput>) {
            const task = await run(() => api.editTask(id, patch))

            if (task) {
                setView((current) => current && { ...current, tasks: current.tasks.map((item) => (item.id === id ? task : item)) })
            }

            return task !== null
        },

        async deleteTask(id: string) {
            const done = await run(() => api.deleteTask(id).then(() => true))

            if (done) {
                setView((current) => {
                    if (!current) {
                        return current
                    }

                    const removed = collectTaskIds(current.tasks, id)

                    return { ...current, tasks: current.tasks.filter((task) => !removed.has(task.id)) }
                })
            }

            return done === true
        },
    }

    return { view, error, actions, addNode, renameBoard, removeBoard, exportMarkdown }
}