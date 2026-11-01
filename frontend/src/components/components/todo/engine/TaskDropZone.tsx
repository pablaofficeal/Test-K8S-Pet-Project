import { Droppable, type DroppableProvidedProps } from '@hello-pangea/dnd'
import type { ReactElement, ReactNode } from 'react'

export interface DropZoneState {
    innerRef: (element: HTMLElement | null) => void
    zoneProps: DroppableProvidedProps
    placeholder: ReactNode
}

type TaskDropZoneProps = {
    id: string
    children: (zone: DropZoneState) => ReactElement<HTMLElement>
}

function TaskDropZone({ id, children }: TaskDropZoneProps) {
    return (
        <Droppable droppableId={id} type="task">
            {(provided) => children({
                innerRef: provided.innerRef,
                zoneProps: provided.droppableProps,
                placeholder: provided.placeholder,
            })}
        </Droppable>
    )
}
export default TaskDropZone