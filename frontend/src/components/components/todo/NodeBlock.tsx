import { useState, type FormEvent } from 'react'
import { Draggable, Droppable, type DraggableProvided } from '@hello-pangea/dnd'
import TextConfig from "../../../lib/configs/site/text.configs"
import type { TaskCounts, TaskTreeNode } from "../../../lib/todo/tree"
import type { BoardActions, TaskFormState, TodoNode } from "../../../lib/todo/types"
import TaskForm from "./TaskForm"
import TaskItem from "./TaskItem"

type NodeBlockProps = {
    node: TodoNode
    tree: TaskTreeNode[]
    totals: TaskCounts
    filtering: boolean
    actions: BoardActions
    drag: DraggableProvided
    dragging: boolean
}

// Один список (колонка) доски: заголовок - ручка для перетаскивания списка, карточки перетаскиваются между списками.
// Счётчики берём по всем задачам списка, а карточки показываем уже с учётом поиска
function NodeBlock({ node, tree, totals, filtering, actions, drag, dragging }: NodeBlockProps) {
    const [form, setForm] = useState<TaskFormState>(null)
    const [renaming, setRenaming] = useState(false)
    const [title, setTitle] = useState(node.title)
    const percent = totals.total === 0 ? 0 : Math.round((totals.done / totals.total) * 100)

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
        <section ref={drag.innerRef} {...drag.draggableProps} className={dragging ? 'todo-node dragging' : 'todo-node'}>
            <div className="todo-node-header" title={TextConfig.TodoDragBlock} {...drag.dragHandleProps}>
                {renaming ? (
                    <form className="todo-node-rename" onSubmit={handleRename} onKeyDown={(event) => { if (event.key === 'Escape') { cancelRename() } }}>
                        <input type="text" value={title} onChange={(event) => setTitle(event.target.value)} maxLength={200} aria-label={TextConfig.TodoNodeNamePlaceholder} autoFocus />
                    </form>
                ) : (
                    <h3>{node.title}</h3>
                )}
                {totals.total > 0 && <span className="todo-node-count">{totals.done}/{totals.total}</span>}
                <div className="todo-node-actions">
                    <button type="button" onClick={() => setRenaming(true)} title={TextConfig.TodoRenameNode} aria-label={TextConfig.TodoRenameNode}>
                        <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M12 20h9M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z" /></svg>
                    </button>
                    <button type="button" onClick={handleDelete} title={TextConfig.TodoDeleteNode} aria-label={TextConfig.TodoDeleteNode}>×</button>
                </div>
            </div>
            {totals.total > 0 && (
                <div className="todo-node-progress" role="progressbar" aria-valuemin={0} aria-valuemax={100} aria-valuenow={percent}>
                    <span style={{ width: `${percent}%` }} />
                </div>
            )}
            <Droppable droppableId={node.id} type="task">
                {(drop) => (
                    <ul ref={drop.innerRef} {...drop.droppableProps} className="task-list task-drop">
                        {tree.map((item, index) => (
                            <Draggable key={item.task.id} draggableId={item.task.id} index={index} isDragDisabled={filtering}>
                                {(card, snapshot) => (
                                    <li
                                        ref={card.innerRef}
                                        {...card.draggableProps}
                                        {...card.dragHandleProps}
                                        className={`task-card priority-${item.task.priority}${snapshot.isDragging ? ' dragging' : ''}`}
                                    >
                                        <TaskItem item={item} nodeId={node.id} actions={actions} form={form} onForm={setForm} />
                                    </li>
                                )}
                            </Draggable>
                        ))}
                        {drop.placeholder}
                    </ul>
                )}
            </Droppable>
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
export default NodeBlock