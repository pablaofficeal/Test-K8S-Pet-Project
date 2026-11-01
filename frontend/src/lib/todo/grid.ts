export const GRID_SIZE = 20
export const NODE_WIDTH = 300
export const NODE_GAP = 40

export function snapToGrid(value: number) {
    return Math.max(0, Math.round(value / GRID_SIZE) * GRID_SIZE)
}

// Новый блок ставим в ряд правее самого правого, чтобы он не наехал на существующие
export function nextNodePosition(positionsX: number[]) {
    if (positionsX.length === 0) {
        return { x: 0, y: 0 }
    }

    return { x: snapToGrid(Math.max(...positionsX) + NODE_WIDTH + NODE_GAP), y: 0 }
}