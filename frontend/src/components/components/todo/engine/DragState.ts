import type { DraggableProvidedDraggableProps, DraggableProvidedDragHandleProps } from '@hello-pangea/dnd'

// Всё, что нужно визуальному компоненту, чтобы стать перетаскиваемым: он ничего не знает о самой библиотеке
export interface DragState {
    innerRef: (element: HTMLElement | null) => void
    containerProps: DraggableProvidedDraggableProps
    handleProps: DraggableProvidedDragHandleProps | null | undefined
    dragging: boolean
}