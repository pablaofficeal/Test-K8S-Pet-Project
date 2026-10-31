import { API_CONFIG } from "../configs/api/Api.config"
import TextConfig from "../configs/site/text.configs"
import { logger } from "../log/logger"

const TRACE_HEADER = "X-Request-ID"

let refreshing: Promise<boolean> | null = null

const sentTraceIds = new WeakMap<Response, string>()

export class ApiError extends Error {
    status: number
    traceId: string | null
    unexpected: boolean

    constructor(message: string, status: number, traceId: string | null, unexpected: boolean) {
        super(message)
        this.status = status
        this.traceId = traceId
        this.unexpected = unexpected
    }
}

function newTraceId() {
    const bytes = crypto.getRandomValues(new Uint8Array(16))

    return Array.from(bytes, (byte) => byte.toString(16).padStart(2, "0")).join("")
}

export async function apiRequest(path: string, init: RequestInit = {}) {
    const traceId = newTraceId()

    try {
        const response = await fetch(`${API_CONFIG.BASE}${path}`, {
            ...init,
            credentials: "include",
            headers: {
                "X-Requested-With": "XMLHttpRequest",
                [TRACE_HEADER]: traceId,
                ...init.headers,
            },
        })
        sentTraceIds.set(response, traceId)

        return response
    } catch (error) {
        logger.error("network error", {
            path,
            method: init.method ?? "GET",
            traceId,
            message: error instanceof Error ? error.message : String(error),
        })
        throw new ApiError(TextConfig.UnexpectedError, 0, traceId, true)
    }
}

export async function toApiError(response: Response) {
    const data = await response.json().catch(() => ({}))
    const traceId = response.headers.get(TRACE_HEADER) ?? sentTraceIds.get(response) ?? null
    const unexpected = response.status >= 500 || !data.error

    logger[unexpected ? "error" : "warn"]("api request failed", {
        url: response.url,
        status: response.status,
        traceId,
        message: data.error ?? null,
    })

    return new ApiError(unexpected ? TextConfig.UnexpectedError : data.error, response.status, traceId, unexpected)
}

export function describeError(error: unknown, fallback: string) {
    if (error instanceof ApiError) {
        return error.unexpected && error.traceId ? `${error.message} (${TextConfig.TraceId}: ${error.traceId})` : error.message
    }

    logger.error("unexpected error", { message: error instanceof Error ? error.message : String(error) })

    return fallback
}

// refresh-токен ротируется, поэтому параллельные обновления схлопываем в один запрос
function refreshSession() {
    refreshing ??= apiRequest(API_CONFIG.AUTH.REFRESH, { method: "POST" })
        .then((response) => response.ok)
        .catch(() => false)
        .finally(() => {
            refreshing = null
        })

    return refreshing
}

export async function apiFetch(path: string, init: RequestInit = {}) {
    const response = await apiRequest(path, init)

    if (response.status !== 401 || !(await refreshSession())) {
        return response
    }

    return apiRequest(path, init)
}