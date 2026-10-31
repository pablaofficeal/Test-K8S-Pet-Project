type Level = 'debug' | 'info' | 'warn' | 'error'
type Context = Record<string, unknown>

const ORDER: Record<Level, number> = { debug: 0, info: 1, warn: 2, error: 3 }
const MIN_LEVEL: Level = import.meta.env.DEV ? 'debug' : 'info'

function write(level: Level, message: string, context?: Context) {
    if (ORDER[level] < ORDER[MIN_LEVEL]) {
        return
    }

    const line = `${new Date().toISOString()} [${level.toUpperCase()}] ${message}`

    if (context) {
        console[level](line, context)
        return
    }

    console[level](line)
}

export const logger = {
    debug: (message: string, context?: Context) => write('debug', message, context),
    info: (message: string, context?: Context) => write('info', message, context),
    warn: (message: string, context?: Context) => write('warn', message, context),
    error: (message: string, context?: Context) => write('error', message, context),
}

export function initGlobalErrorLogging() {
    window.addEventListener('error', (event) => {
        logger.error('uncaught error', {
            message: event.message,
            source: event.filename,
            line: event.lineno,
        })
    })

    window.addEventListener('unhandledrejection', (event) => {
        const reason = event.reason as { message?: string, traceId?: string | null } | undefined

        logger.error('unhandled rejection', {
            message: reason?.message ?? String(event.reason),
            traceId: reason?.traceId ?? null,
        })
    })
}