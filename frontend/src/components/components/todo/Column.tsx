import { useState, type FormEvent } from 'react'
import TextConfig from "../../../lib/configs/site/text.configs"
import type { TaskCounts, TaskTreeNode } from "../../../lib/todo/tree"
import type { BoardActions, TaskFormState, TodoNode } from "../../../lib/todo/types"
import type { DragState } from "./engine/DragState"
import TaskDrag from "./engine/TaskDrag"
import TaskDropZone from "./engine/TaskDropZone"
import ColumnMenu from "./ColumnMenu"
import TaskForm from "./TaskForm"
import TaskRow from "./TaskRow"

type ColumnProps = {
    node: TodoNode
    tree: TaskTreeNode[]
    totals: TaskCounts
    filtering: boolean
    actions: BoardActions
    drag: DragState
}

// Колонка доски. Только внешний вид: за перетаскивание отвечают компоненты из ./engine,
// мы лишь получаем от них props и вешаем на нужные элементы
function Column({ node, tree, totals, filtering, actions, drag }: ColumnProps) {
    const [form, setForm] = useState<TaskFormState>(null)
    const [renaming, setRenaming] = useState(false)
    const [title, setTitle] = useState(node.title)

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

    function cancelRename() {
        setRenaming(false)
        setTitle(node.title)
    }

    async function handleDelete() {
        if (await actions.confirm(TextConfig.TodoDeleteNodeConfirm)) {
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

    return (
        <section ref={drag.innerRef} {...drag.containerProps} className={drag.dragging ? 'todo-node dragging' : 'todo-node'}>
            <div className="todo-node-header" title={TextConfig.TodoDragBlock} {...drag.handleProps}>
                {renaming ? (
                    <form className="todo-node-rename" onSubmit={handleRename} onKeyDown={(event) => { if (event.key === 'Escape') { cancelRename() } }}>
                        <input type="text" value={title} onChange={(event) => setTitle(event.target.value)} maxLength={200} aria-label={TextConfig.TodoNodeNamePlaceholder} autoFocus />
                    </form>
                ) : (
                    <h3>{node.title}</h3>
                )}
                {totals.total > 0 && <span className="todo-node-count">{totals.done}/{totals.total}</span>}
                <ColumnMenu onRename={() => setRenaming(true)} onAddTask={() => setForm({ mode: 'create', parentId: null })} onDelete={handleDelete} />
                <button type="button" className="todo-icon-button" onClick={handleDelete} title={TextConfig.TodoDeleteNode} aria-label={TextConfig.TodoDeleteNode}>×</button>
            </div>
            <TaskDropZone id={node.id}>
                {(zone) => (
                    <ul ref={zone.innerRef} {...zone.zoneProps} className="task-list task-drop">
                        {tree.map((item, index) => (
                            <TaskDrag key={item.task.id} id={item.task.id} index={index} disabled={filtering}>
                                {(card) => (
                                    <li ref={card.innerRef} {...card.containerProps} {...card.handleProps} className={card.dragging ? 'task-card dragging' : 'task-card'}>
                                        <TaskRow item={item} nodeId={node.id} actions={actions} form={form} onForm={setForm} />
                                    </li>
                                )}
                            </TaskDrag>
                        ))}
                        {zone.placeholder}
                    </ul>
                )}
            </TaskDropZone>
            {tree.length === 0 && form === null && (
                <p className="todo-node-empty">{filtering && totals.total > 0 ? TextConfig.TodoNoResults : TextConfig.TodoNoTasks}</p>
            )}
            <div className="todo-node-footer">
                {form?.mode === 'create' && form.parentId === null ? (
                    <TaskForm onSubmit={handleCreate} onCancel={() => setForm(null)} />
                ) : (
                    <button type="button" className="todo-new-task" onClick={() => setForm({ mode: 'create', parentId: null })}>+ {TextConfig.TodoAddTask}</button>
                )}
            </div>
        </section>
    )
}
export default Column