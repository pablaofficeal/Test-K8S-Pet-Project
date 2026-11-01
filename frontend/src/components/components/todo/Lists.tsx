import { DragDropContext, Draggable, Droppable, type DropResult } from '@hello-pangea/dnd'
import TextConfig from "../../../lib/configs/site/text.configs"
import { buildTaskTree, countTasks, filterTasks } from "../../../lib/todo/tree"
import type { BoardActions, TodoNode, TodoTask } from "../../../lib/todo/types"
import AddList from "./AddList"
import NodeBlock from "./NodeBlock"

type ListsProps = {
    nodes: TodoNode[]
    tasks: TodoTask[]
    query: string
    hideDone: boolean
    actions: BoardActions
    onAddNode: (title: string) => Promise<boolean>
}

// Доска как в Taskcafe: списки стоят в ряд, их порядок меняется перетаскиванием заголовка,
// карточки переносятся между списками. Индекс назначения совпадает с позицией среди карточек верхнего уровня,
// поэтому пока включён фильтр, перетаскивание карточек выключено (индексы отфильтрованного списка не совпали бы с серверными)
function Lists({ nodes, tasks, query, hideDone, actions, onAddNode }: ListsProps) {
    const filtering = query.trim() !== '' || hideDone
    const visible = filterTasks(tasks, query, hideDone)

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
                                        <NodeBlock
                                            node={node}
                                            tree={buildTaskTree(visible, node.id)}
                                            totals={countTasks(buildTaskTree(tasks, node.id))}
                                            filtering={filtering}
                                            actions={actions}
                                            drag={drag}
                                            dragging={snapshot.isDragging}
                                        />
                                    )}
                                </Draggable>
                            ))}
                            {drop.placeholder}
                        </div>
                    )}
                </Droppable>
                {nodes.length === 0 && <p className="todo-board-empty">{TextConfig.TodoBoardEmpty}</p>}
                <AddList onAdd={onAddNode} />
            </div>
        </DragDropContext>
    )
}
export default Lists