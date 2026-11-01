import { useEffect, useRef } from 'react'
import TextConfig from "../../../lib/configs/site/text.configs"

type ConfirmDialogProps = {
    message: string | null
    onAnswer: (ok: boolean) => void
}

function ConfirmDialog({ message, onAnswer }: ConfirmDialogProps) {
    const ref = useRef<HTMLDialogElement>(null)

    useEffect(() => {
        const dialog = ref.current

        if (!dialog) {
            return
        }

        if (message !== null && !dialog.open) {
            dialog.showModal()
        }

        if (message === null && dialog.open) {
            dialog.close()
        }
    }, [message])

    return (
        <dialog
            ref={ref}
            className="todo-confirm"
            onCancel={(event) => { event.preventDefault(); onAnswer(false) }}
            onClick={(event) => { if (event.target === event.currentTarget) { onAnswer(false) } }}
        >
            <p>{message}</p>
            <div className="todo-confirm-actions">
                <button type="button" className="todo-button danger" onClick={() => onAnswer(true)}>{TextConfig.TodoDelete}</button>
                <button type="button" className="todo-button secondary" onClick={() => onAnswer(false)} autoFocus>{TextConfig.TodoCancel}</button>
            </div>
        </dialog>
    )
}
export default ConfirmDialog