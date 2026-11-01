import type { TodoTask } from './types'

export interface TaskTreeNode {
    task: TodoTask
    children: TaskTreeNode[]
}

export interface TaskCounts {
    done: number
    total: number
}

export function buildTaskTree(tasks: TodoTask[], nodeId: string): TaskTreeNode[] {
    const own = tasks.filter((task) => task.node_id === nodeId).sort((a, b) => a.sort_order - b.sort_order)
    const byParent = new Map<string | null, TodoTask[]>()

    for (const task of own) {
        const list = byParent.get(task.parent_id) ?? []
        list.push(task)
        byParent.set(task.parent_id, list)
    }

    function build(parentId: string | null): TaskTreeNode[] {
        return (byParent.get(parentId) ?? []).map((task) => ({ task, children: build(task.id) }))
    }

    return build(null)
}

export function countTasks(nodes: TaskTreeNode[]): TaskCounts {
    let done = 0
    let total = 0

    for (const node of nodes) {
        total += 1
        done += node.task.status === 'done' ? 1 : 0
        const inner = countTasks(node.children)
        done += inner.done
        total += inner.total
    }

    return { done, total }
}

export function collectTaskIds(tasks: TodoTask[], rootId: string): Set<string> {
    const ids = new Set<string>([rootId])
    let grew = true

    while (grew) {
        grew = false

        for (const task of tasks) {
            if (task.parent_id !== null && ids.has(task.parent_id) && !ids.has(task.id)) {
                ids.add(task.id)
                grew = true
            }
        }
    }

    return ids
}