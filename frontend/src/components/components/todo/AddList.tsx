import { useState, type FormEvent } from 'react'
import TextConfig from "../../../lib/configs/site/text.configs"

type AddListProps = {
    onAdd: (title: string) => Promise<boolean>
}

function AddList({ onAdd }: AddListProps) {
    const [open, setOpen] = useState(false)
    const [title, setTitle] = useState('')
    const [pending, setPending] = useState(false)

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        const name = title.trim()

        if (name === '') {
            return
        }

        setPending(true)
        const done = await onAdd(name)
        setPending(false)

        if (done) {
            setTitle('')
            setOpen(false)
        }
    }

    if (!open) {
        return <button type="button" className="todo-add-list" onClick={() => setOpen(true)}>+ {TextConfig.TodoNewNode}</button>
    }

    return (
        <form className="todo-add-list-form" onSubmit={handleSubmit}>
            <input type="text" value={title} onChange={(event) => setTitle(event.target.value)} placeholder={TextConfig.TodoNodeNamePlaceholder} aria-label={TextConfig.TodoNodeNamePlaceholder} maxLength={200} disabled={pending} autoFocus />
            <div className="task-form-actions">
                <button type="submit" disabled={pending || title.trim() === ''}>{TextConfig.TodoCreate}</button>
                <button type="button" className="secondary" onClick={() => setOpen(false)} disabled={pending}>{TextConfig.TodoCancel}</button>
            </div>
        </form>
    )
}
export default AddList