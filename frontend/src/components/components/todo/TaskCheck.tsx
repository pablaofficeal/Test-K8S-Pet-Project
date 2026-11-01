import TextConfig from "../../../lib/configs/site/text.configs"
import type { TaskStatus } from "../../../lib/todo/types"

type TaskCheckProps = {
    status: TaskStatus
    onToggle: () => void
}

function TaskCheck({ status, onToggle }: TaskCheckProps) {
    const label = status === 'done' ? TextConfig.TodoMarkTodo : TextConfig.TodoMarkDone

    return (
        <button type="button" className={`task-check ${status}`} onClick={onToggle} aria-label={label} title={label}>
            <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                {status === 'done' && <path d="M5 12.5l4.5 4.5L19 7.5" />}
                {status === 'in_progress' && <path d="M12 5v14" />}
            </svg>
        </button>
    )
}
export default TaskCheck