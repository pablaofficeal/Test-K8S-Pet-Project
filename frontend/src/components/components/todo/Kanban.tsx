import TextConfig from "../../../lib/configs/site/text.configs"
import { buildTaskTree, countTasks, filterTasks } from "../../../lib/todo/tree"
import type { BoardActions, TodoNode, TodoTask } from "../../../lib/todo/types"
import ColumnDrag from "./engine/ColumnDrag"
import Engine from "./engine/Engine"
import AddList from "./AddList"
import Column from "./Column"

type KanbanProps = {
    nodes: TodoNode[]
    tasks: TodoTask[]
    query: string
    hideDone: boolean
    actions: BoardActions
    onAddNode: (title: string) => Promise<boolean>
}

// Канбан-доска: рамка со скроллом, колонки в ряд и композер "добавить блок".
// Пока включён фильтр, перетаскивание карточек выключено: индексы отфильтрованного списка не совпали бы с серверными
function Kanban({ nodes, tasks, query, hideDone, actions, onAddNode }: KanbanProps) {
    const filtering = query.trim() !== '' || hideDone
    const visible = filterTasks(tasks, query, hideDone)

    return (
        <div className="todo-board">
            <Engine order={nodes.map((node) => node.id)} onReorder={actions.reorderNodes} onMoveTask={actions.moveTask}>
                {nodes.map((node, index) => (
                    <ColumnDrag key={node.id} id={node.id} index={index}>
                        {(drag) => (
                            <Column
                                node={node}
                                tree={buildTaskTree(visible, node.id)}
                                totals={countTasks(buildTaskTree(tasks, node.id))}
                                filtering={filtering}
                                actions={actions}
                                drag={drag}
                            />
                        )}
                    </ColumnDrag>
                ))}
            </Engine>
            {nodes.length === 0 && <p className="todo-board-empty">{TextConfig.TodoBoardEmpty}</p>}
            <AddList onAdd={onAddNode} />
        </div>
    )
}
export default Kanban