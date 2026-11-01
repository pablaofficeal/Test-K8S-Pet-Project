import { useState, type FormEvent } from 'react'
import { Draggable, Droppable, type DraggableProvided } from '@hello-pangea/dnd'
import TextConfig from "../../../lib/configs/site/text.configs"
import { countTasks, type TaskTreeNode } from "../../../lib/todo/tree"
import type { BoardActions, TaskFormState, TodoNode } from "../../../lib/todo/types"
import TaskForm from "./TaskForm"
import TaskItem from "./TaskItem"

type NodeBlockProps = {
    node: TodoNode
    tree: TaskTreeNode[]
    actions: BoardActions
    drag: DraggableProvided
    dragging: boolean
}

// Один список (колонка) доски: заголовок - ручка для перетаскивания списка, карточки перетаскиваются между списками
function NodeBlock({ node, tree, actions, drag, dragging }: NodeBlockProps) {
    const [form, setForm] = useState<TaskFormState>(null)
    const [renaming, setRenaming] = useState(false)
    const [title, setTitle] = useState(node.title)
    const counts = countTasks(tree)

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

    return (
        <section ref={drag.innerRef} {...drag.draggableProps} className={dragging ? 'todo-node dragging' : 'todo-node'}>
            <div className="todo-node-header" title={TextConfig.TodoDragBlock} {...drag.dragHandleProps}>
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
            <Droppable droppableId={node.id} type="task">
                {(drop) => (
                    <ul ref={drop.innerRef} {...drop.droppableProps} className="task-list task-drop">
                        {tree.map((item, index) => (
                            <Draggable key={item.task.id} draggableId={item.task.id} index={index}>
                                {(card, snapshot) => (
                                    <li ref={card.innerRef} {...card.draggableProps} {...card.dragHandleProps} className={snapshot.isDragging ? 'task-card dragging' : 'task-card'}>
                                        <TaskItem item={item} nodeId={node.id} actions={actions} form={form} onForm={setForm} />
                                    </li>
                                )}
                            </Draggable>
                        ))}
                        {drop.placeholder}
                    </ul>
                )}
            </Droppable>
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