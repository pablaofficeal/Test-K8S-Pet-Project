import { API_CONFIG } from "../configs/api/Api.config"
import { apiFetch, toApiError } from "../auth/api"
import { UnauthorizedError } from "../auth/profile"
import type { BoardView, TaskInput, TodoBoard, TodoNode, TodoTask } from "./types"

const TODO = API_CONFIG.TODO

async function request(path: string, init: RequestInit = {}) {
    const response = await apiFetch(path, init)

    if (response.status === 401) {
        throw new UnauthorizedError()
    }

    if (!response.ok) {
        throw await toApiError(response)
    }

    return response
}

function jsonInit(method: string, body: unknown): RequestInit {
    return {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
    }
}

function normalizeTask(task: TodoTask): TodoTask {
    return { ...task, tags: task.tags ?? [] }
}

export async function getBoards(): Promise<TodoBoard[]> {
    const data = await (await request(TODO.BOARDS)).json()

    return data.boards
}

export async function getBoard(id: string): Promise<BoardView> {
    const data: BoardView = await (await request(`${TODO.BOARD}/${id}`)).json()

    return { ...data, tasks: data.tasks.map(normalizeTask) }
}

export async function createBoard(title: string): Promise<TodoBoard> {
    return (await request(TODO.CREATE_BOARD, jsonInit("POST", { title }))).json()
}

export async function importBoard(markdown: string): Promise<TodoBoard> {
    const response = await request(TODO.IMPORT_BOARD, {
        method: "POST",
        headers: { "Content-Type": "text/markdown" },
        body: markdown,
    })
    const data = await response.json()

    return data.board
}

export async function renameBoard(id: string, title: string) {
    await request(`${TODO.EDIT_BOARD}/${id}`, jsonInit("PATCH", { title }))
}

export async function deleteBoard(id: string) {
    await request(`${TODO.DELETE_BOARD}/${id}`, { method: "DELETE" })
}

export async function exportBoard(id: string): Promise<string> {
    return (await request(`${TODO.EXPORT_BOARD}/${id}`)).text()
}

export async function createNode(boardId: string, title: string, x: number, y: number): Promise<TodoNode> {
    return (await request(`${TODO.CREATE_BOARD}/${boardId}/node`, jsonInit("POST", { title, pos_x: x, pos_y: y }))).json()
}

export async function renameNode(id: string, title: string) {
    await request(`${TODO.EDIT_NODE}/${id}`, jsonInit("PATCH", { title }))
}

export async function moveNode(id: string, x: number, y: number) {
    await request(`${TODO.MOVE_NODE}/${id}`, jsonInit("PATCH", { x, y }))
}

export async function deleteNode(id: string) {
    await request(`${TODO.DELETE_NODE}/${id}`, { method: "DELETE" })
}

export async function createTask(nodeId: string, parentId: string | null, input: TaskInput): Promise<TodoTask> {
    const response = await request(`${TODO.CREATE_TASK}/${nodeId}/task`, jsonInit("POST", { ...input, parent_id: parentId }))

    return normalizeTask(await response.json())
}

export async function editTask(id: string, patch: Partial<TaskInput>): Promise<TodoTask> {
    return normalizeTask(await (await request(`${TODO.EDIT_TASK}/${id}`, jsonInit("PATCH", patch))).json())
}

export async function deleteTask(id: string) {
    await request(`${TODO.DELETE_TASK}/${id}`, { method: "DELETE" })
}