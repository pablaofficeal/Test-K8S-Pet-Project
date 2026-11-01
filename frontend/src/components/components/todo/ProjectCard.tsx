import { Link } from 'react-router-dom'
import TextConfig from "../../../lib/configs/site/text.configs"
import { userTodoBoardPath } from "../../../lib/configs/site/routes"
import { formatDate } from "../../../lib/i18n/date"
import type { TodoBoard } from "../../../lib/todo/types"

type ProjectCardProps = {
    board: TodoBoard
    uuid: string
    disabled: boolean
    onDelete: (board: TodoBoard) => void
}

function ProjectCard({ board, uuid, disabled, onDelete }: ProjectCardProps) {
    return (
        <article className="project-card">
            <Link to={userTodoBoardPath(uuid, board.id)} className="project-card-title">{board.title}</Link>
            <p className="project-card-date">{TextConfig.TodoCreatedAt}: <time dateTime={board.created_at}>{formatDate(board.created_at)}</time></p>
            <div className="project-card-actions">
                <Link to={userTodoBoardPath(uuid, board.id)}>{TextConfig.TodoOpen}</Link>
                <button type="button" onClick={() => onDelete(board)} disabled={disabled}>{TextConfig.TodoDelete}</button>
            </div>
        </article>
    )
}
export default ProjectCard