import { Draggable } from '@hello-pangea/dnd'
import type { ReactElement } from 'react'
import type { DragState } from './DragState'

type ColumnDragProps = {
    id: string
    index: number
    children: (drag: DragState) => ReactElement<HTMLElement>
}

function ColumnDrag({ id, index, children }: ColumnDragProps) {
    return (
        <Draggable draggableId={id} index={index}>
            {(provided, snapshot) => children({
                innerRef: provided.innerRef,
                containerProps: provided.draggableProps,
                handleProps: provided.dragHandleProps,
                dragging: snapshot.isDragging,
            })}
        </Draggable>
    )
}
export default ColumnDrag