export const API_CONFIG = {
    BASE: "http://192.168.0.30:30080/main/api/v1",
    AUTH: {
        AUTH_LOGIN: "/start/log/in/my/account",
        REGISTERS: "/start/create/my/account",
        REFRESH: "/start/refresh/my/token",
        LOGOUT: "/start/log/out/my/account"
    },
    PROFILE: {
        INFO: "/get/my/profile/info",
        EDIT_NAME: "/edit/my/name",
        EDIT_EMAIL: "/edit/my/email"
    },
    BOTS: {
        LIST: "/get/my/bots",
        GET: "/get/my/bot",
        TEMPLATES: "/get/templates",
        LIMITS: "/get/my/limits",
        USAGE: "/get/my/usage",
        CREATE: "/create/my/bot",
        EDIT: "/edit/my/bot",
        DELETE: "/delete/my/bot"
    },
}