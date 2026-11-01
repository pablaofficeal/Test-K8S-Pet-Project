import { useRef, useState, type FormEvent, type PointerEvent } from 'react'
import TextConfig from "../../../lib/configs/site/text.configs"
import { NODE_WIDTH, snapToGrid } from "../../../lib/todo/grid"
import { countTasks, type TaskTreeNode } from "../../../lib/todo/tree"
import type { BoardActions, TaskFormState, TodoNode } from "../../../lib/todo/types"
import TaskForm from "./TaskForm"
import TaskItem from "./TaskItem"

type NodeBlockProps = {
    node: TodoNode
    tree: TaskTreeNode[]
    actions: BoardActions
    fits: (id: string, x: number, y: number, width: number, height: number) => boolean
}

type DragState = {
    x: number
    y: number
    blocked: boolean
}

type DragStart = {
    pointerX: number
    pointerY: number
    originX: number
    originY: number
}

function NodeBlock({ node, tree, actions, fits }: NodeBlockProps) {
    const blockRef = useRef<HTMLElement>(null)
    const startRef = useRef<DragStart | null>(null)
    const [drag, setDrag] = useState<DragState | null>(null)
    const [form, setForm] = useState<TaskFormState>(null)
    const [renaming, setRenaming] = useState(false)
    const [title, setTitle] = useState(node.title)
    const counts = countTasks(tree)

    function handlePointerDown(event: PointerEvent<HTMLDivElement>) {
        if (event.button !== 0 || (event.target as HTMLElement).closest('button, input')) {
            return
        }

        event.currentTarget.setPointerCapture(event.pointerId)
        startRef.current = { pointerX: event.clientX, pointerY: event.clientY, originX: node.pos_x, originY: node.pos_y }
        setDrag({ x: node.pos_x, y: node.pos_y, blocked: false })
    }

    function handlePointerMove(event: PointerEvent<HTMLDivElement>) {
        const start = startRef.current
        const block = blockRef.current

        if (!start || !block) {
            return
        }

        const x = snapToGrid(start.originX + event.clientX - start.pointerX)
        const y = snapToGrid(start.originY + event.clientY - start.pointerY)
        setDrag({ x, y, blocked: !fits(node.id, x, y, block.offsetWidth, block.offsetHeight) })
    }

    function handlePointerUp() {
        const current = drag
        startRef.current = null
        setDrag(null)

        if (current && !current.blocked && (current.x !== node.pos_x || current.y !== node.pos_y)) {
            actions.moveNode(node.id, current.x, current.y)
        }
    }

    function handlePointerCancel() {
        startRef.current = null
        setDrag(null)
    }

    async function handleRename(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        const next = title.trim()

        if (next === '' || next === node.title) {
            setRenaming(false)
            setTitle(node.title)
            return
        }

        if (await actions.renameNode(node.id, next)) {
            setRenaming(false)
        }
    }

    async function handleDelete() {
        if (window.confirm(TextConfig.TodoDeleteNodeConfirm)) {
            await actions.deleteNode(node.id)
        }
    }

    async function handleCreate(input: Parameters<BoardActions['createTask']>[2]) {
        const done = await actions.createTask(node.id, null, input)

        if (done) {
            setForm(null)
        }

        return done
    }

    const className = drag ? (drag.blocked ? 'todo-node dragging blocked' : 'todo-node dragging') : 'todo-node'

    return (
        <section
            ref={blockRef}
            className={className}
            data-node-id={node.id}
            style={{ left: drag?.x ?? node.pos_x, top: drag?.y ?? node.pos_y, width: NODE_WIDTH }}
        >
            <div
                className="todo-node-header"
                title={TextConfig.TodoDragBlock}
                onPointerDown={handlePointerDown}
                onPointerMove={handlePointerMove}
                onPointerUp={handlePointerUp}
                onPointerCancel={handlePointerCancel}
            >
                {renaming ? (
                    <form className="todo-node-rename" onSubmit={handleRename}>
                        <input type="text" value={title} onChange={(event) => setTitle(event.target.value)} maxLength={200} aria-label={TextConfig.TodoNodeNamePlaceholder} autoFocus />
                    </form>
                ) : (
                    <h3>{node.title}</h3>
                )}
                {counts.total > 0 && <span className="todo-node-count">{counts.done}/{counts.total}</span>}
                <div className="todo-node-actions">
                    <button type="button" onClick={() => setRenaming(true)} title={TextConfig.TodoRenameNode} aria-label={TextConfig.TodoRenameNode}>
                        <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M12 20h9M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z" /></svg>
                    </button>
                    <button type="button" onClick={handleDelete} title={TextConfig.TodoDeleteNode} aria-label={TextConfig.TodoDeleteNode}>×</button>
                </div>
            </div>
            <div className="todo-node-body">
                <ul className="task-list">
                    {tree.map((item) => (
                        <TaskItem key={item.task.id} item={item} nodeId={node.id} actions={actions} form={form} onForm={setForm} />
                    ))}
                </ul>
                {form?.mode === 'create' && form.parentId === null ? (
                    <TaskForm onSubmit={handleCreate} onCancel={() => setForm(null)} />
                ) : (
                    <button type="button" className="todo-new-task" onClick={() => setForm({ mode: 'create', parentId: null })}>+ {TextConfig.TodoAddTask}</button>
                )}
            </div>
        </section>
    )
}
export default NodeBlock