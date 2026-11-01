export type TaskStatus = 'todo' | 'in_progress' | 'done'
export type TaskPriority = 'low' | 'normal' | 'high'

export interface TodoBoard {
    id: string
    owner_uuid: string
    title: string
    created_at: string
    updated_at: string
}

export interface TodoNode {
    id: string
    board_id: string
    title: string
    sort_order: number
    pos_x: number
    pos_y: number
}

export interface TodoTask {
    id: string
    board_id: string
    node_id: string
    parent_id: string | null
    title: string
    description: string
    status: TaskStatus
    priority: TaskPriority
    tags: string[]
    deadline: string | null
    sort_order: number
}

export interface BoardView {
    board: TodoBoard
    nodes: TodoNode[]
    tasks: TodoTask[]
}

export interface TaskInput {
    title: string
    description: string
    status: TaskStatus
    priority: TaskPriority
    tags: string[]
    deadline: string
}

export interface BoardActions {
    renameNode: (id: string, title: string) => Promise<boolean>
    moveNode: (id: string, x: number, y: number) => void
    deleteNode: (id: string) => Promise<boolean>
    createTask: (nodeId: string, parentId: string | null, input: TaskInput) => Promise<boolean>
    editTask: (id: string, patch: Partial<TaskInput>) => Promise<boolean>
    deleteTask: (id: string) => Promise<boolean>
}

export type TaskFormState = { mode: 'create'; parentId: string | null } | { mode: 'edit'; taskId: string } | null