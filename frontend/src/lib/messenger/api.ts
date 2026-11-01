import { API_CONFIG } from "../configs/api/Api.config"
import { apiFetch, toApiError } from "../auth/api"
import { UnauthorizedError } from "../auth/profile"
import type { Envelope, PublicKeyDoc } from "./crypto/types.ts"

const MESSENGER = API_CONFIG.MESSENGER

export interface ServerKey {
    key_id: string
    public_key: PublicKeyDoc
    backup: string | null
}

export interface UserHit {
    uuid: string
    username: string
    key_id: string | null
}

export interface Conversation {
    peer_uuid: string
    username: string
    key_id: string | null
    unread: number
    last_id: number
    last_at: string
}

export interface StoredMessage {
    id: number
    sender_uuid: string
    recipient_uuid: string
    envelope: Envelope
    read_at: string | null
    created_at: string
}

async function request(path: string, init: RequestInit = {}, allowMissing = false) {
    const response = await apiFetch(path, init)

    if (response.status === 401) {
        throw new UnauthorizedError()
    }

    if (allowMissing && response.status === 404) {
        return null
    }

    if (!response.ok) {
        throw await toApiError(response)
    }

    return response
}

function jsonInit(method: string, body: unknown): RequestInit {
    return {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
    }
}

export async function getMyKey(): Promise<ServerKey | null> {
    const response = await request(MESSENGER.GET_MY_KEY, {}, true)

    return response ? response.json() : null
}

export async function putMyKey(doc: PublicKeyDoc) {
    await request(MESSENGER.MY_KEY, jsonInit("PUT", doc))
}

export async function putKeyBackup(backup: string) {
    await request(MESSENGER.KEY_BACKUP, jsonInit("PUT", { backup }))
}

export async function getUserKey(uuid: string): Promise<PublicKeyDoc | null> {
    const response = await request(`${MESSENGER.USER_KEY}/${uuid}`, {}, true)

    return response ? (await response.json()).public_key : null
}

export async function searchUsers(query: string): Promise<UserHit[]> {
    const response = await request(`${MESSENGER.SEARCH_USERS}?q=${encodeURIComponent(query)}`)

    return (await response?.json()).users
}

export async function sendMessage(recipientUuid: string, envelope: Envelope) {
    await request(MESSENGER.SEND, jsonInit("POST", { recipient_uuid: recipientUuid, envelope }))
}

export async function getMessages(peerUuid: string, after?: number): Promise<StoredMessage[]> {
    const cursor = after === undefined ? "" : `&after=${after}`
    const response = await request(`${MESSENGER.MESSAGES}?peer=${peerUuid}${cursor}`)

    return (await response?.json()).messages
}

export async function getConversations(): Promise<Conversation[]> {
    const response = await request(MESSENGER.CONVERSATIONS)

    return (await response?.json()).conversations
}

export async function markRead(peerUuid: string) {
    await request(MESSENGER.MARK_READ, jsonInit("PATCH", { peer: peerUuid }))
}

export async function deleteMessage(id: number) {
    await request(`${MESSENGER.DELETE_MESSAGE}/${id}`, { method: "DELETE" })
}