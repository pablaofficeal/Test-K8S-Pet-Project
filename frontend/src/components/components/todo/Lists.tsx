import { DragDropContext, Draggable, Droppable, type DropResult } from '@hello-pangea/dnd'
import { buildTaskTree } from "../../../lib/todo/tree"
import type { BoardActions, TodoNode, TodoTask } from "../../../lib/todo/types"
import AddList from "./AddList"
import NodeBlock from "./NodeBlock"

type ListsProps = {
    nodes: TodoNode[]
    tasks: TodoTask[]
    actions: BoardActions
    onAddNode: (title: string) => Promise<boolean>
}

// Доска как в Taskcafe: списки стоят в ряд, их порядок меняется перетаскиванием заголовка,
// карточки переносятся между списками. Индекс назначения совпадает с позицией среди карточек верхнего уровня
function Lists({ nodes, tasks, actions, onAddNode }: ListsProps) {
    function handleDragEnd(result: DropResult) {
        const { source, destination, type, draggableId } = result

        if (!destination || (source.droppableId === destination.droppableId && source.index === destination.index)) {
            return
        }

        if (type === 'column') {
            const order = nodes.map((node) => node.id)
            const [moved] = order.splice(source.index, 1)
            order.splice(destination.index, 0, moved)
            actions.reorderNodes(order)
            return
        }

        actions.moveTask(draggableId, destination.droppableId, destination.index)
    }

    return (
        <DragDropContext onDragEnd={handleDragEnd}>
            <div className="todo-board">
                <Droppable droppableId="board" type="column" direction="horizontal">
                    {(drop) => (
                        <div ref={drop.innerRef} {...drop.droppableProps} className="todo-columns">
                            {nodes.map((node, index) => (
                                <Draggable key={node.id} draggableId={node.id} index={index}>
                                    {(drag, snapshot) => (
                                        <NodeBlock node={node} tree={buildTaskTree(tasks, node.id)} actions={actions} drag={drag} dragging={snapshot.isDragging} />
                                    )}
                                </Draggable>
                            ))}
                            {drop.placeholder}
                        </div>
                    )}
                </Droppable>
                <AddList onAdd={onAddNode} />
            </div>
        </DragDropContext>
    )
}
export default Lists