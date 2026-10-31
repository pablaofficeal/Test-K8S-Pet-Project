import { API_CONFIG } from "../configs/api/Api.config"
import { apiRequest } from "./api"

export async function logout() {
    try {
        await apiRequest(API_CONFIG.AUTH.LOGOUT, { method: "POST" })
    } catch {
    }
}