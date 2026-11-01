export const API_CONFIG = {
    BASE: "/main/api/v1",
    AUTH: {
        AUTH_LOGIN: "/start/log/in/my/account",
        REGISTERS: "/start/create/my/account",
        REFRESH: "/start/refresh/my/token",
        LOGOUT: "/start/log/out/my/account"
    },
    VERIFY: {
        STATUS: "/verify/status",
        RESEND: "/verify/resend",
        CONFIRM: "/verify/confirm"
    },
    PROFILE: {
        INFO: "/get/my/profile/info",
        EDIT_NAME: "/edit/my/name",
        EDIT_EMAIL: "/edit/my/email",
        UPLOAD_AVATAR: "/start/upload/my/avatar",
        DELETE_AVATAR: "/delete/my/avatar",
        IMAGE: "/get/my/image"
    },
    TODO: {
        BOARDS: "/todo/get/my/boards",
        BOARD: "/todo/get/my/board",
        CREATE_BOARD: "/todo/create/my/board",
        EDIT_BOARD: "/todo/edit/my/board",
        DELETE_BOARD: "/todo/delete/my/board",
        IMPORT_BOARD: "/todo/import/my/board",
        EXPORT_BOARD: "/todo/export/my/board",
        EDIT_NODE: "/todo/edit/my/node",
        MOVE_NODE: "/todo/move/my/node",
        DELETE_NODE: "/todo/delete/my/node",
        CREATE_TASK: "/todo/create/my/node",
        EDIT_TASK: "/todo/edit/my/task",
        DELETE_TASK: "/todo/delete/my/task"
    },
}