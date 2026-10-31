import { API_CONFIG } from "../configs/api/Api.config"
import { apiRequest, toApiError } from "./api"

export interface RegisterResponse {
    username?: string
    email?: string
    error?: string
}

export async function register(email: string, password: string, name: string): Promise<RegisterResponse> {
    const response = await apiRequest(API_CONFIG.AUTH.REGISTERS, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            username: name,
            email,
            password,
        }),
    })

    if (!response.ok) {
        throw await toApiError(response)
    }

    return response.json()
}