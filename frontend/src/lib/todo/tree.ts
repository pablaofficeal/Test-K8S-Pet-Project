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

// Поиск и фильтр: оставляем подходящие задачи вместе с их родителями, чтобы дерево не развалилось
export function filterTasks(tasks: TodoTask[], query: string, hideDone: boolean): TodoTask[] {
    const needle = query.trim().toLowerCase()

    if (needle === '' && !hideDone) {
        return tasks
    }

    const byId = new Map(tasks.map((task) => [task.id, task]))
    const keep = new Set<string>()

    for (const task of tasks) {
        const visible = !hideDone || task.status !== 'done'
        const found = needle === ''
            || task.title.toLowerCase().includes(needle)
            || task.description.toLowerCase().includes(needle)
            || task.tags.some((tag) => tag.toLowerCase().includes(needle))

        if (!visible || !found) {
            continue
        }

        let current: TodoTask | undefined = task

        while (current && !keep.has(current.id)) {
            keep.add(current.id)
            current = current.parent_id === null ? undefined : byId.get(current.parent_id)
        }
    }

    return tasks.filter((task) => keep.has(task.id))
}