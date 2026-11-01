import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import Shell from '../../components/menu/shell'
import ConfirmDialog from '../../components/components/todo/ConfirmDialog'
import Lists from '../../components/components/todo/Lists'
import TextConfig from '../../lib/configs/site/text.configs'
import { userTodoPath } from '../../lib/configs/site/routes'
import { getUserId } from '../../lib/auth/session'
import { useBoard } from '../../lib/todo/useBoard'
import '../../assets/css/dashboard.css'
import '../../assets/css/todo.css'

function Board() {
    const params = useParams()
    const uuid = params.uuid ?? getUserId() ?? ''
    const boardId = params.boardId ?? ''
    const navigate = useNavigate()
    const { view, error, actions, addNode, renameBoard, removeBoard, exportMarkdown, confirmMessage, answerConfirm } = useBoard(boardId)
    const [editing, setEditing] = useState(false)
    const [title, setTitle] = useState('')
    const [query, setQuery] = useState('')
    const [hideDone, setHideDone] = useState(false)

    async function handleRename(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        const next = title.trim()

        if (next === '' || next === view?.board.title || await renameBoard(next)) {
            setEditing(false)
        }
    }

    async function handleExport() {
        const markdown = await exportMarkdown()

        if (markdown === null || !view) {
            return
        }

        const url = URL.createObjectURL(new Blob([markdown], { type: 'text/markdown' }))
        const link = document.createElement('a')
        link.href = url
        link.download = `${view.board.title.replace(/[\\/:*?"<>|]+/g, '_')}.md`
        link.click()
        URL.revokeObjectURL(url)
    }

    async function handleDelete() {
        if (await actions.confirm(TextConfig.TodoDeleteProjectConfirm) && await removeBoard()) {
            navigate(userTodoPath(uuid), { replace: true })
        }
    }

    return (
        <Shell dashboard>
            <main className="dashboard todo-page">
                {error && <p className="dashboard-error" role="alert">{error}</p>}
                {!view && !error && <p className="dashboard-muted">{TextConfig.DashboardLoading}</p>}
                {view && (
                    <>
                        <div className="todo-board-header">
                            <Link to={userTodoPath(uuid)} className="todo-back">&larr; {TextConfig.TodoTitle}</Link>
                            {editing ? (
                                <form className="todo-board-rename" onSubmit={handleRename}>
                                    <input type="text" value={title} onChange={(event) => setTitle(event.target.value)} maxLength={200} aria-label={TextConfig.TodoProjectNamePlaceholder} autoFocus />
                                    <button type="submit" className="todo-button">{TextConfig.SettingsSave}</button>
                                    <button type="button" className="todo-button secondary" onClick={() => setEditing(false)}>{TextConfig.TodoCancel}</button>
                                </form>
                            ) : (
                                <h1 className="todo-board-title">{view.board.title}</h1>
                            )}
                            <div className="projects-actions">
                                <button type="button" className="todo-button secondary" onClick={() => { setTitle(view.board.title); setEditing(true) }}>{TextConfig.TodoEdit}</button>
                                <button type="button" className="todo-button secondary" onClick={handleExport}>{TextConfig.TodoExport}</button>
                                <button type="button" className="todo-button danger" onClick={handleDelete}>{TextConfig.TodoDelete}</button>
                            </div>
                        </div>
                        <div className="todo-toolbar">
                            <input
                                type="search"
                                className="todo-search"
                                value={query}
                                onChange={(event) => setQuery(event.target.value)}
                                placeholder={TextConfig.TodoSearchTasks}
                                aria-label={TextConfig.TodoSearchTasks}
                            />
                            <label className="todo-check-option">
                                <input type="checkbox" checked={hideDone} onChange={(event) => setHideDone(event.target.checked)} />
                                <span>{TextConfig.TodoHideDone}</span>
                            </label>
                            {(query.trim() !== '' || hideDone) && <span className="todo-filter-note">{TextConfig.TodoFilterActive}</span>}
                        </div>
                        <Lists nodes={view.nodes} tasks={view.tasks} query={query} hideDone={hideDone} actions={actions} onAddNode={addNode} />
                    </>
                )}
                <ConfirmDialog message={confirmMessage} onAnswer={answerConfirm} />
            </main>
        </Shell>
    )
}
export default Board