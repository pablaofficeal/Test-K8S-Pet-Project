import { Draggable } from '@hello-pangea/dnd'
import type { ReactElement } from 'react'
import type { DragState } from './DragState'

type TaskDragProps = {
    id: string
    index: number
    disabled: boolean
    children: (drag: DragState) => ReactElement<HTMLElement>
}

function TaskDrag({ id, index, disabled, children }: TaskDragProps) {
    return (
        <Draggable draggableId={id} index={index} isDragDisabled={disabled}>
            {(provided, snapshot) => children({
                innerRef: provided.innerRef,
                containerProps: provided.draggableProps,
                handleProps: provided.dragHandleProps,
                dragging: snapshot.isDragging,
            })}
        </Draggable>
    )
}
export default TaskDrag