import { useState, type FormEvent } from 'react'
import TextConfig from "../../../lib/configs/site/text.configs"
import type { TaskInput, TaskPriority, TaskStatus, TodoTask } from "../../../lib/todo/types"

type TaskFormProps = {
    initial?: TodoTask
    onSubmit: (input: TaskInput) => Promise<boolean>
    onCancel: () => void
}

function TaskForm({ initial, onSubmit, onCancel }: TaskFormProps) {
    const [title, setTitle] = useState(initial?.title ?? '')
    const [description, setDescription] = useState(initial?.description ?? '')
    const [status, setStatus] = useState<TaskStatus>(initial?.status ?? 'todo')
    const [priority, setPriority] = useState<TaskPriority>(initial?.priority ?? 'normal')
    const [tags, setTags] = useState(initial?.tags.join(', ') ?? '')
    const [deadline, setDeadline] = useState(initial?.deadline?.slice(0, 10) ?? '')
    const [pending, setPending] = useState(false)

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        const name = title.trim()

        if (name === '') {
            return
        }

        setPending(true)
        await onSubmit({
            title: name,
            description: description.trim(),
            status,
            priority,
            tags: tags.split(',').map((tag) => tag.trim()).filter((tag) => tag !== ''),
            deadline,
        })
        setPending(false)
    }

    return (
        <form className="task-form" onSubmit={handleSubmit}>
            <input
                type="text"
                value={title}
                onChange={(event) => setTitle(event.target.value)}
                placeholder={TextConfig.TodoTaskTitlePlaceholder}
                aria-label={TextConfig.TodoTaskTitle}
                maxLength={500}
                disabled={pending}
                autoFocus
                required
            />
            <textarea
                value={description}
                onChange={(event) => setDescription(event.target.value)}
                placeholder={TextConfig.TodoTaskDescription}
                aria-label={TextConfig.TodoTaskDescription}
                rows={2}
                maxLength={10000}
                disabled={pending}
            />
            <div className="task-form-row">
                <label>
                    <span>{TextConfig.TodoTaskStatus}</span>
                    <select value={status} onChange={(event) => setStatus(event.target.value as TaskStatus)} disabled={pending}>
                        <option value="todo">{TextConfig.TodoStatusTodo}</option>
                        <option value="in_progress">{TextConfig.TodoStatusInProgress}</option>
                        <option value="done">{TextConfig.TodoStatusDone}</option>
                    </select>
                </label>
                <label>
                    <span>{TextConfig.TodoTaskPriority}</span>
                    <select value={priority} onChange={(event) => setPriority(event.target.value as TaskPriority)} disabled={pending}>
                        <option value="low">{TextConfig.TodoPriorityLow}</option>
                        <option value="normal">{TextConfig.TodoPriorityNormal}</option>
                        <option value="high">{TextConfig.TodoPriorityHigh}</option>
                    </select>
                </label>
            </div>
            <div className="task-form-row">
                <label>
                    <span>{TextConfig.TodoTaskTags}</span>
                    <input type="text" value={tags} onChange={(event) => setTags(event.target.value)} placeholder={TextConfig.TodoTaskTagsPlaceholder} disabled={pending} />
                </label>
                <label>
                    <span>{TextConfig.TodoTaskDeadline}</span>
                    <input type="date" value={deadline} onChange={(event) => setDeadline(event.target.value)} disabled={pending} />
                </label>
            </div>
            <div className="task-form-actions">
                <button type="submit" disabled={pending || title.trim() === ''}>
                    {pending ? TextConfig.SettingsSaving : TextConfig.SettingsSave}
                </button>
                <button type="button" className="secondary" onClick={onCancel} disabled={pending}>{TextConfig.TodoCancel}</button>
            </div>
        </form>
    )
}
export default TaskForm