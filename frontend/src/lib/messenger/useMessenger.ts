import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import TextConfig from '../configs/site/text.configs'
import { ROUTES, loginPathWithRedirect } from '../configs/site/routes'
import { describeError } from '../auth/api'
import { UnauthorizedError } from '../auth/profile'
import { clearAuthSession, getUserId } from '../auth/session'
import { getMyKey, putKeyBackup, putMyKey } from './api'
import { unwrapKey, wrapKey } from './backup'
import { clearCustomDictionaries, activeAlphabets, saveCustomDictionaries } from './dictionaries'
import { loadPrivateKey, generateKeyPair } from './crypto/keys.ts'
import { saveStoredKey, loadStoredKey } from './keystore'
import type { DualAlphabets } from './crypto/alphabet.ts'
import type { PrivateKeyDoc, PublicKeyDoc } from './crypto/types.ts'

export type SetupReason = 'new' | 'restore' | 'mismatch'

export type MessengerState =
    | { phase: 'loading' }
    | { phase: 'setup'; reason: SetupReason; hasBackup: boolean; serverKeyId: string | null }
    | { phase: 'ready'; privateDoc: PrivateKeyDoc; publicDoc: PublicKeyDoc }

export function publicDocOf(doc: PrivateKeyDoc): PublicKeyDoc {
    return { version: doc.version, algorithm: doc.algorithm, key_id: doc.key_id, public_key: doc.public_key }
}

// Ключи, словари и статус шифрования. Приватный ключ живёт только в этом браузере (MVP),
// на сервер уходит публичный ключ и, по желанию, копия приватного, зашифрованная парольной фразой
export function useMessenger() {
    const navigate = useNavigate()
    const uuid = getUserId() ?? ''
    const [state, setState] = useState<MessengerState>({ phase: 'loading' })
    const [alphabets, setAlphabets] = useState<DualAlphabets | null>(null)
    const [customDictionaries, setCustomDictionaries] = useState(false)
    const [error, setError] = useState('')
    const [notice, setNotice] = useState('')
    const [pending, setPending] = useState(false)

    const fail = useCallback((err: unknown, fallback: string) => {
        if (err instanceof UnauthorizedError) {
            clearAuthSession()
            navigate(loginPathWithRedirect(ROUTES.MESSENGER), { replace: true })
            return
        }

        setError(describeError(err, fallback))
    }, [navigate])

    const refresh = useCallback(async () => {
        try {
            const active = await activeAlphabets()
            setAlphabets(active.alphabets)
            setCustomDictionaries(active.custom)

            const stored = loadStoredKey(uuid)
            let local: PrivateKeyDoc | null = null

            if (stored) {
                try {
                    await loadPrivateKey(stored)
                    local = stored
                } catch {
                    local = null
                }
            }

            const server = await getMyKey()

            if (local && !server) {
                await putMyKey(publicDocOf(local))
            }

            if (local && (!server || server.key_id === local.key_id)) {
                setState({ phase: 'ready', privateDoc: local, publicDoc: publicDocOf(local) })
                return
            }

            setState({
                phase: 'setup',
                reason: local ? 'mismatch' : server ? 'restore' : 'new',
                hasBackup: server?.backup != null,
                serverKeyId: server?.key_id ?? null,
            })
        } catch (err) {
            fail(err, TextConfig.MsgLoadFailed)
        }
    }, [uuid, fail])

    useEffect(() => {
        refresh()
    }, [refresh])

    async function run(action: () => Promise<void>) {
        setPending(true)
        setError('')
        setNotice('')

        try {
            await action()
        } catch (err) {
            fail(err, TextConfig.MsgLoadFailed)
        } finally {
            setPending(false)
        }
    }

    async function adopt(doc: PrivateKeyDoc) {
        await loadPrivateKey(doc)
        saveStoredKey(uuid, doc)
        await putMyKey(publicDocOf(doc))
        setState({ phase: 'ready', privateDoc: doc, publicDoc: publicDocOf(doc) })
    }

    return {
        state,
        alphabets,
        customDictionaries,
        error,
        notice,
        pending,
        generate: () => run(async () => adopt((await generateKeyPair()).privateDoc)),
        importKey: (text: string) => run(async () => {
            let doc: PrivateKeyDoc

            try {
                doc = JSON.parse(text) as PrivateKeyDoc
                await loadPrivateKey(doc)
            } catch {
                throw new Error(TextConfig.MsgImportFailed)
            }

            await adopt(doc)
        }),
        restoreBackup: (passphrase: string) => run(async () => {
            const server = await getMyKey()

            if (!server?.backup) {
                throw new Error(TextConfig.MsgWrongPassphrase)
            }

            await adopt(await unwrapKey(server.backup, passphrase))
        }),
        useLocalKey: () => run(async () => {
            const local = loadStoredKey(uuid)

            if (local) {
                await adopt(local)
            }
        }),
        saveBackup: (passphrase: string) => run(async () => {
            if (state.phase !== 'ready') {
                return
            }

            await putKeyBackup(await wrapKey(state.privateDoc, passphrase))
            setNotice(TextConfig.MsgBackupSaved)
        }),
        applyDictionaries: (firstText: string, secondText: string) => run(async () => {
            await saveCustomDictionaries(firstText, secondText)
            await refresh()
        }),
        resetDictionaries: () => run(async () => {
            clearCustomDictionaries()
            await refresh()
        }),
    }
}