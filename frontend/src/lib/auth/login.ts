import { API_CONFIG } from "../configs/api/Api.config"
import { apiRequest, toApiError } from "./api"

export interface LoginResponse {
    token_type?: string
    expires_in?: number
    message?: string
    error?: string
}

export async function login(email: string, password: string): Promise<LoginResponse> {
    const response = await apiRequest(API_CONFIG.AUTH.AUTH_LOGIN, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            email,
            password,
        }),
    })

    if (!response.ok) {
        throw await toApiError(response)
    }

    return response.json()
}