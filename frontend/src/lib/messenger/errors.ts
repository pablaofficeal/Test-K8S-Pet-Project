import { describeError } from '../auth/api'

// Ошибка с уже локализованным текстом: показываем как есть
export class LocalizedError extends Error {}

export function messageOf(err: unknown, fallback: string): string {
    return err instanceof LocalizedError ? err.message : describeError(err, fallback)
}