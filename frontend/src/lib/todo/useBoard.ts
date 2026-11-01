import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import TextConfig from '../configs/site/text.configs'
import { ROUTES, loginPathWithRedirect } from '../configs/site/routes'
import { describeError } from '../auth/api'
import { UnauthorizedError } from '../auth/profile'
import { clearAuthSession } from '../auth/session'
import * as api from './api'
import { collectTaskIds } from './tree'
import { useConfirm } from './useConfirm'
import type { BoardActions, BoardView, TaskInput } from './types'

// Состояние одного проекта (канваса): блоки, задачи и все изменения.
// Перестановка блоков и перенос карточек сразу применяются локально, на сервер уходит один запрос при отпускании
export function useBoard(boardId: string) {
    const navigate = useNavigate()
    const [view, setView] = useState<BoardView | null>(null)
    const [error, setError] = useState('')
    const confirm = useConfirm()

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
        const node = await run(() => api.createNode(boardId, title))

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
        confirm: confirm.ask,

        async renameNode(id: string, title: string) {
            const done = await run(() => api.renameNode(id, title).then(() => true))

            if (done) {
                setView((current) => current && { ...current, nodes: current.nodes.map((node) => (node.id === id ? { ...node, title } : node)) })
            }

            return done === true
        },

        reorderNodes(order: string[]) {
            setView((current) => {
                if (!current) {
                    return current
                }

                const byId = new Map(current.nodes.map((node) => [node.id, node]))

                return { ...current, nodes: order.flatMap((id) => byId.get(id) ?? []) }
            })
            run(() => api.reorderNodes(boardId, order)).then((result) => {
                if (result === null) {
                    api.getBoard(boardId).then(setView, () => undefined)
                }
            })
        },

        moveTask(taskId: string, nodeId: string, index: number) {
            setView((current) => {
                if (!current) {
                    return current
                }

                const card = current.tasks.find((task) => task.id === taskId)

                if (!card) {
                    return current
                }

                const moving = collectTaskIds(current.tasks, taskId)
                const siblings = current.tasks
                    .filter((task) => task.node_id === nodeId && task.parent_id === null && task.id !== taskId)
                    .sort((a, b) => a.sort_order - b.sort_order)
                const position = Math.min(index, siblings.length)
                const ranks = new Map<string, number>()

                siblings.splice(position, 0, card)
                siblings.forEach((task, rank) => ranks.set(task.id, rank))

                return {
                    ...current,
                    tasks: current.tasks.map((task) => {
                        if (moving.has(task.id)) {
                            return { ...task, node_id: nodeId, sort_order: ranks.get(task.id) ?? task.sort_order }
                        }

                        return ranks.has(task.id) ? { ...task, sort_order: ranks.get(task.id) as number } : task
                    }),
                }
            })
            run(() => api.moveTask(taskId, nodeId, index)).then((result) => {
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

    return { view, error, actions, addNode, renameBoard, removeBoard, exportMarkdown, confirmMessage: confirm.message, answerConfirm: confirm.answer }
}