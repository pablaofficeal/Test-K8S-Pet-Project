import { API_CONFIG } from "../configs/api/Api.config"
import { apiFetch, toApiError } from "./api"

export interface ProfileResponse {
    uuid: string
    username: string
    email: string
    email_verified: boolean
    avatar_key: string | null
    created_at: string
    updated_at: string
}

export class UnauthorizedError extends Error {}

export async function getProfile(): Promise<ProfileResponse> {
    const response = await apiFetch(API_CONFIG.PROFILE.INFO)

    if (response.status === 401) {
        throw new UnauthorizedError()
    }

    if (!response.ok) {
        throw await toApiError(response)
    }

    return response.json()
}

async function patchProfile<T>(path: string, body: Record<string, string>): Promise<T> {
    const response = await apiFetch(path, {
        method: "PATCH",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
    })

    if (response.status === 401) {
        throw new UnauthorizedError()
    }

    if (!response.ok) {
        throw await toApiError(response)
    }

    return response.json()
}

export function updateName(username: string) {
    return patchProfile<{ username: string }>(API_CONFIG.PROFILE.EDIT_NAME, { username })
}

export function updateEmail(email: string) {
    return patchProfile<{ email: string }>(API_CONFIG.PROFILE.EDIT_EMAIL, { email })
}

export function avatarUrl(key: string) {
    return `${API_CONFIG.BASE}${API_CONFIG.PROFILE.IMAGE}/${encodeURIComponent(key)}`
}

export async function uploadAvatar(file: File): Promise<{ avatar_key: string }> {
    const form = new FormData()
    form.append("file", file)

    const response = await apiFetch(API_CONFIG.PROFILE.UPLOAD_AVATAR, {
        method: "POST",
        body: form,
    })

    if (response.status === 401) {
        throw new UnauthorizedError()
    }

    if (!response.ok) {
        throw await toApiError(response)
    }

    return response.json()
}

export async function deleteAvatar(): Promise<{ deleted: true }> {
    const response = await apiFetch(API_CONFIG.PROFILE.DELETE_AVATAR, { method: "DELETE" })

    if (response.status === 401) {
        throw new UnauthorizedError()
    }

    if (!response.ok) {
        throw await toApiError(response)
    }

    return { deleted: true }
}