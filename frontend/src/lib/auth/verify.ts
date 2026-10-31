import { API_CONFIG } from "../configs/api/Api.config"
import TextConfig from "../configs/site/text.configs"
import { apiFetch } from "./api"
import { UnauthorizedError } from "./profile"

export interface VerifyStatus {
    verified: boolean
    email: string
    resend_in: number
    expires_in: number
}

export class VerifyError extends Error {
    retryAfter: number

    constructor(message: string, retryAfter = 0) {
        super(message)
        this.retryAfter = retryAfter
    }
}

type VerifyBody = {
    error?: string
    attempts_left?: number
    retry_after?: number
    resend_in?: number
}

function failureMessage(status: number, body: VerifyBody) {
    if (status === 400 && typeof body.attempts_left === 'number') {
        return `${TextConfig.VerifyErrorInvalid} ${body.attempts_left}`
    }
    if (status === 400 && body.error?.includes('6 digits')) {
        return TextConfig.VerifyErrorFormat
    }
    if (status === 400) {
        return TextConfig.VerifyErrorNoCode
    }
    if (status === 429 && typeof body.retry_after === 'number') {
        return TextConfig.VerifyErrorCooldown
    }
    if (status === 429 && body.error?.includes('attempts')) {
        return TextConfig.VerifyErrorAttempts
    }
    if (status === 429) {
        return TextConfig.VerifyErrorRateLimit
    }
    if (status === 502) {
        return TextConfig.VerifyErrorSend
    }

    return TextConfig.UnexpectedError
}

async function call(path: string, init: RequestInit = {}) {
    const response = await apiFetch(path, init)

    if (response.status === 401) {
        throw new UnauthorizedError()
    }

    const body: VerifyBody = await response.json().catch(() => ({}))

    return { response, body }
}

export async function getVerifyStatus(): Promise<VerifyStatus> {
    const { response, body } = await call(API_CONFIG.VERIFY.STATUS)

    if (!response.ok) {
        throw new VerifyError(failureMessage(response.status, body))
    }

    return body as VerifyStatus
}

export async function confirmCode(code: string): Promise<void> {
    const { response, body } = await call(API_CONFIG.VERIFY.CONFIRM, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({ code }),
    })

    if (!response.ok) {
        throw new VerifyError(failureMessage(response.status, body))
    }
}

export async function resendCode(): Promise<{ verified: boolean, resendIn: number }> {
    const { response, body } = await call(API_CONFIG.VERIFY.RESEND, { method: "POST" })

    if (response.status === 409) {
        return { verified: true, resendIn: 0 }
    }

    if (!response.ok) {
        throw new VerifyError(failureMessage(response.status, body), body.retry_after ?? 0)
    }

    return { verified: false, resendIn: body.resend_in ?? 0 }
}