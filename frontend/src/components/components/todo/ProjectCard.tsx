import { Link } from 'react-router-dom'
import TextConfig from "../../../lib/configs/site/text.configs"
import { userTodoBoardPath } from "../../../lib/configs/site/routes"
import { formatDate } from "../../../lib/i18n/date"
import type { BoardSummary } from "../../../lib/todo/types"

type ProjectCardProps = {
    board: BoardSummary
    uuid: string
    disabled: boolean
    onDelete: (board: BoardSummary) => void
}

function ProjectCard({ board, uuid, disabled, onDelete }: ProjectCardProps) {
    const percent = board.tasks_total === 0 ? 0 : Math.round((board.tasks_done / board.tasks_total) * 100)

    return (
        <article className="project-card">
            <Link to={userTodoBoardPath(uuid, board.id)} className="project-card-title">{board.title}</Link>
            <p className="project-card-stats">
                {board.nodes_count} {TextConfig.TodoBlocksLabel} · {board.tasks_done}/{board.tasks_total} {TextConfig.TodoTasksLabel}
            </p>
            <div className="project-card-progress" role="progressbar" aria-valuemin={0} aria-valuemax={100} aria-valuenow={percent}>
                <span style={{ width: `${percent}%` }} />
            </div>
            <p className="project-card-date">{TextConfig.ProfileUpdated}: <time dateTime={board.updated_at}>{formatDate(board.updated_at)}</time></p>
            <div className="project-card-actions">
                <Link to={userTodoBoardPath(uuid, board.id)}>{TextConfig.TodoOpen}</Link>
                <button type="button" onClick={() => onDelete(board)} disabled={disabled}>{TextConfig.TodoDelete}</button>
            </div>
        </article>
    )
}
export default ProjectCard