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
    MESSENGER: {
        MY_KEY: "/messenger/put/my/key",
        GET_MY_KEY: "/messenger/get/my/key",
        KEY_BACKUP: "/messenger/put/my/key/backup",
        USER_KEY: "/messenger/get/user/key",
        SEARCH_USERS: "/messenger/search/users",
        SEND: "/messenger/send/message",
        MESSAGES: "/messenger/get/my/messages",
        CONVERSATIONS: "/messenger/get/my/conversations",
        MARK_READ: "/messenger/mark/my/messages/read",
        DELETE_MESSAGE: "/messenger/delete/my/message"
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
        REORDER_NODES: "/todo/reorder/my/board",
        MOVE_TASK: "/todo/move/my/task",
        DELETE_NODE: "/todo/delete/my/node",
        CREATE_TASK: "/todo/create/my/node",
        EDIT_TASK: "/todo/edit/my/task",
        DELETE_TASK: "/todo/delete/my/task"
    },
}