import { DragDropContext, Droppable, type DropResult } from '@hello-pangea/dnd'
import type { ReactNode } from 'react'

type EngineProps = {
    order: string[]
    onReorder: (order: string[]) => void
    onMoveTask: (taskId: string, nodeId: string, index: number) => void
    children: ReactNode
}

// Движок доски: только перетаскивание. Ряд колонок - один горизонтальный Droppable,
// карточки внутри колонок - Droppable'ы типа task (их даёт TaskDropZone)
function Engine({ order, onReorder, onMoveTask, children }: EngineProps) {
    function handleDragEnd(result: DropResult) {
        const { source, destination, type, draggableId } = result

        if (!destination || (source.droppableId === destination.droppableId && source.index === destination.index)) {
            return
        }

        if (type === 'column') {
            const next = [...order]
            const [moved] = next.splice(source.index, 1)
            next.splice(destination.index, 0, moved)
            onReorder(next)
            return
        }

        onMoveTask(draggableId, destination.droppableId, destination.index)
    }

    return (
        <DragDropContext onDragEnd={handleDragEnd}>
            <Droppable droppableId="board" type="column" direction="horizontal">
                {(drop) => (
                    <div ref={drop.innerRef} {...drop.droppableProps} className="todo-columns">
                        {children}
                        {drop.placeholder}
                    </div>
                )}
            </Droppable>
        </DragDropContext>
    )
}
export default Engine