import { useCallback, useRef } from 'react'
import { NODE_GAP, NODE_WIDTH } from "../../../lib/todo/grid"
import { buildTaskTree } from "../../../lib/todo/tree"
import type { BoardActions, TodoNode, TodoTask } from "../../../lib/todo/types"
import NodeBlock from "./NodeBlock"

type CanvasProps = {
    nodes: TodoNode[]
    tasks: TodoTask[]
    actions: BoardActions
}

function Canvas({ nodes, tasks, actions }: CanvasProps) {
    const innerRef = useRef<HTMLDivElement>(null)
    const width = Math.max(0, ...nodes.map((node) => node.pos_x)) + NODE_WIDTH + NODE_GAP * 2
    const height = Math.max(0, ...nodes.map((node) => node.pos_y)) + 640

    // Блоки стоят в общей сетке и не должны наезжать друг на друга: сверяем с реальными размерами в DOM
    const fits = useCallback((id: string, x: number, y: number, blockWidth: number, blockHeight: number) => {
        const root = innerRef.current

        if (!root) {
            return true
        }

        return Array.from(root.querySelectorAll<HTMLElement>('[data-node-id]')).every((other) => {
            if (other.dataset.nodeId === id) {
                return true
            }

            return x >= other.offsetLeft + other.offsetWidth
                || x + blockWidth <= other.offsetLeft
                || y >= other.offsetTop + other.offsetHeight
                || y + blockHeight <= other.offsetTop
        })
    }, [])

    return (
        <div className="todo-canvas">
            <div ref={innerRef} className="todo-canvas-inner" style={{ width, height }}>
                {nodes.map((node) => (
                    <NodeBlock key={node.id} node={node} tree={buildTaskTree(tasks, node.id)} actions={actions} fits={fits} />
                ))}
            </div>
        </div>
    )
}
export default Canvas