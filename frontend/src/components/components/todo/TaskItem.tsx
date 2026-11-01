import { useState } from 'react'
import TextConfig from "../../../lib/configs/site/text.configs"
import { formatDeadline } from "../../../lib/todo/date"
import { countTasks, type TaskTreeNode } from "../../../lib/todo/tree"
import type { BoardActions, TaskFormState } from "../../../lib/todo/types"
import TaskCheck from "./TaskCheck"
import TaskForm from "./TaskForm"

type TaskItemProps = {
    item: TaskTreeNode
    nodeId: string
    actions: BoardActions
    form: TaskFormState
    onForm: (form: TaskFormState) => void
}

function TaskItem({ item, nodeId, actions, form, onForm }: TaskItemProps) {
    const { task, children } = item
    const [open, setOpen] = useState(false)
    const counts = countTasks(children)
    const editing = form?.mode === 'edit' && form.taskId === task.id
    const addingChild = form?.mode === 'create' && form.parentId === task.id
    const showChildren = open || addingChild

    async function handleToggle() {
        await actions.editTask(task.id, { status: task.status === 'done' ? 'todo' : 'done' })
    }

    async function handleDelete() {
        if (window.confirm(TextConfig.TodoDeleteTaskConfirm)) {
            await actions.deleteTask(task.id)
        }
    }

    async function handleSave(input: Parameters<BoardActions['editTask']>[1]) {
        const done = await actions.editTask(task.id, input)

        if (done) {
            onForm(null)
        }

        return done
    }

    async function handleCreate(input: Parameters<BoardActions['createTask']>[2]) {
        const done = await actions.createTask(nodeId, task.id, input)

        if (done) {
            onForm(null)
        }

        return done
    }

    if (editing) {
        return (
            <li className="task-item">
                <TaskForm initial={task} onSubmit={(input) => handleSave(input)} onCancel={() => onForm(null)} />
            </li>
        )
    }

    return (
        <li className="task-item">
            <div className="task-row">
                {children.length > 0 ? (
                    <button type="button" className={open ? 'task-caret open' : 'task-caret'} onClick={() => setOpen(!open)} aria-expanded={open} aria-label={TextConfig.TodoToggleSubtasks} title={TextConfig.TodoToggleSubtasks}>
                        <svg viewBox="0 0 24 24" width="12" height="12" fill="currentColor" aria-hidden="true"><path d="M8 5l10 7-10 7z" /></svg>
                    </button>
                ) : (
                    <span className="task-caret-space" aria-hidden="true" />
                )}
                <TaskCheck status={task.status} onToggle={handleToggle} />
                <div className="task-main">
                    <span className={task.status === 'done' ? 'task-title done' : 'task-title'}>{task.title}</span>
                    {task.description !== '' && <p className="task-description">{task.description}</p>}
                    <div className="task-meta">
                        {counts.total > 0 && <span className="task-counter">{counts.done}/{counts.total}</span>}
                        {task.priority !== 'normal' && (
                            <span className={`task-chip priority-${task.priority}`}>
                                {task.priority === 'high' ? TextConfig.TodoPriorityHigh : TextConfig.TodoPriorityLow}
                            </span>
                        )}
                        {task.tags.map((tag) => <span key={tag} className="task-chip tag">{tag}</span>)}
                        {task.deadline && <time className="task-chip deadline" dateTime={task.deadline}>{formatDeadline(task.deadline)}</time>}
                    </div>
                </div>
                <div className="task-actions">
                    <button type="button" onClick={() => { setOpen(true); onForm({ mode: 'create', parentId: task.id }) }} title={TextConfig.TodoAddSubtask} aria-label={TextConfig.TodoAddSubtask}>+</button>
                    <button type="button" onClick={() => onForm({ mode: 'edit', taskId: task.id })} title={TextConfig.TodoEditTask} aria-label={TextConfig.TodoEditTask}>
                        <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M12 20h9M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z" /></svg>
                    </button>
                    <button type="button" onClick={handleDelete} title={TextConfig.TodoDeleteTask} aria-label={TextConfig.TodoDeleteTask}>×</button>
                </div>
            </div>
            {showChildren && (
                <ul className="task-list task-children">
                    {children.map((child) => (
                        <TaskItem key={child.task.id} item={child} nodeId={nodeId} actions={actions} form={form} onForm={onForm} />
                    ))}
                    {addingChild && (
                        <li className="task-item">
                            <TaskForm onSubmit={handleCreate} onCancel={() => onForm(null)} />
                        </li>
                    )}
                </ul>
            )}
        </li>
    )
}
export default TaskItem