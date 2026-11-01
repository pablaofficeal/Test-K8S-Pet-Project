import { useRef, useState, type ChangeEvent, type FormEvent } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import Shell from '../../components/menu/shell'
import ProjectCard from '../../components/components/todo/ProjectCard'
import TextConfig from '../../lib/configs/site/text.configs'
import { userTodoBoardPath } from '../../lib/configs/site/routes'
import { getUserId } from '../../lib/auth/session'
import { useBoards } from '../../lib/todo/useBoards'
import type { TodoBoard } from '../../lib/todo/types'
import '../../assets/css/dashboard.css'
import '../../assets/css/todo.css'

const MAX_IMPORT_BYTES = 1024 * 1024

function Projects() {
    const params = useParams()
    const uuid = params.uuid ?? getUserId() ?? ''
    const navigate = useNavigate()
    const { boards, error, pending, create, importFile, remove } = useBoards()
    const fileRef = useRef<HTMLInputElement>(null)
    const [title, setTitle] = useState('')
    const [showForm, setShowForm] = useState(false)
    const [importError, setImportError] = useState('')

    async function handleCreate(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        const name = title.trim()

        if (name === '') {
            return
        }

        const board = await create(name)

        if (board) {
            navigate(userTodoBoardPath(uuid, board.id))
        }
    }

    async function handleFile(event: ChangeEvent<HTMLInputElement>) {
        const file = event.target.files?.[0]
        event.target.value = ''
        setImportError('')

        if (!file) {
            return
        }

        if (file.size > MAX_IMPORT_BYTES) {
            setImportError(TextConfig.TodoImportFailed)
            return
        }

        const board = await importFile(await file.text())

        if (board) {
            navigate(userTodoBoardPath(uuid, board.id))
        }
    }

    function handleDelete(board: TodoBoard) {
        if (window.confirm(TextConfig.TodoDeleteProjectConfirm)) {
            remove(board.id)
        }
    }

    return (
        <Shell dashboard>
            <main className="dashboard">
                <div className="projects-header">
                    <div>
                        <h1>{TextConfig.TodoProjects}</h1>
                        <p className="dashboard-muted">{TextConfig.TodoProjectsDescription}</p>
                    </div>
                    <div className="projects-actions">
                        <button type="button" className="todo-button secondary" onClick={() => fileRef.current?.click()} disabled={pending}>
                            {pending ? TextConfig.TodoImporting : TextConfig.TodoImport}
                        </button>
                        <button type="button" className="todo-button" onClick={() => setShowForm(!showForm)} disabled={pending}>{TextConfig.TodoNewProject}</button>
                    </div>
                </div>
                <input ref={fileRef} type="file" accept=".md,text/markdown,text/plain" onChange={handleFile} hidden />
                {showForm && (
                    <form className="project-form" onSubmit={handleCreate}>
                        <input
                            type="text"
                            value={title}
                            onChange={(event) => setTitle(event.target.value)}
                            placeholder={TextConfig.TodoProjectNamePlaceholder}
                            aria-label={TextConfig.TodoProjectNamePlaceholder}
                            maxLength={200}
                            disabled={pending}
                            autoFocus
                        />
                        <button type="submit" className="todo-button" disabled={pending || title.trim() === ''}>
                            {pending ? TextConfig.TodoCreating : TextConfig.TodoCreate}
                        </button>
                    </form>
                )}
                {(error || importError) && <p className="dashboard-error" role="alert">{error || importError}</p>}
                {!boards && !error && <p className="dashboard-muted">{TextConfig.DashboardLoading}</p>}
                {boards && boards.length === 0 && <p className="dashboard-muted">{TextConfig.TodoEmpty}</p>}
                {boards && boards.length > 0 && (
                    <div className="projects-grid">
                        {boards.map((board) => (
                            <ProjectCard key={board.id} board={board} uuid={uuid} disabled={pending} onDelete={handleDelete} />
                        ))}
                    </div>
                )}
            </main>
        </Shell>
    )
}
export default Projects