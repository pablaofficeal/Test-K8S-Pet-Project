import { useEffect, useRef, useState } from 'react'
import TextConfig from "../../../lib/configs/site/text.configs"

type ColumnMenuProps = {
    onRename: () => void
    onAddTask: () => void
    onDelete: () => void
}

function ColumnMenu({ onRename, onAddTask, onDelete }: ColumnMenuProps) {
    const [open, setOpen] = useState(false)
    const ref = useRef<HTMLDivElement>(null)

    useEffect(() => {
        if (!open) {
            return
        }

        function handleDown(event: MouseEvent) {
            if (ref.current && !ref.current.contains(event.target as Node)) {
                setOpen(false)
            }
        }

        function handleKey(event: KeyboardEvent) {
            if (event.key === 'Escape') {
                setOpen(false)
            }
        }

        document.addEventListener('mousedown', handleDown)
        document.addEventListener('keydown', handleKey)

        return () => {
            document.removeEventListener('mousedown', handleDown)
            document.removeEventListener('keydown', handleKey)
        }
    }, [open])

    function choose(action: () => void) {
        setOpen(false)
        action()
    }

    return (
        <div className="todo-menu" ref={ref}>
            <button type="button" className="todo-icon-button" onClick={() => setOpen(!open)} aria-haspopup="menu" aria-expanded={open} aria-label={TextConfig.TodoMenu} title={TextConfig.TodoMenu}>
                <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor" aria-hidden="true"><circle cx="5" cy="12" r="1.8" /><circle cx="12" cy="12" r="1.8" /><circle cx="19" cy="12" r="1.8" /></svg>
            </button>
            {open && (
                <div className="todo-menu-list" role="menu">
                    <button type="button" role="menuitem" onClick={() => choose(onRename)}>{TextConfig.TodoRenameNode}</button>
                    <button type="button" role="menuitem" onClick={() => choose(onAddTask)}>{TextConfig.TodoAddTask}</button>
                    <button type="button" role="menuitem" className="danger" onClick={() => choose(onDelete)}>{TextConfig.TodoDeleteNode}</button>
                </div>
            )}
        </div>
    )
}
export default ColumnMenu