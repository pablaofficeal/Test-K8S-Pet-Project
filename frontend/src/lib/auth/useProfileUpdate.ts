import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import TextConfig from '../configs/site/text.configs'
import { ROUTES, loginPathWithRedirect } from '../configs/site/routes'
import { describeError } from './api'
import { UnauthorizedError } from './profile'
import { clearAuthSession } from './session'

// Общая обработка сохранения поля профиля: загрузка, ошибка, "сохранено" и выход при 401
export function useProfileUpdate<T>(save: (value: string) => Promise<T>) {
    const navigate = useNavigate()
    const [pending, setPending] = useState(false)
    const [error, setError] = useState('')
    const [saved, setSaved] = useState(false)

    async function run(value: string): Promise<T | null> {
        setPending(true)
        setError('')
        setSaved(false)

        try {
            const result = await save(value)
            setSaved(true)

            return result
        } catch (err) {
            if (err instanceof UnauthorizedError) {
                clearAuthSession()
                navigate(loginPathWithRedirect(ROUTES.SETTINGS), { replace: true })
                return null
            }

            setError(describeError(err, TextConfig.SettingsSaveFailed))

            return null
        } finally {
            setPending(false)
        }
    }

    return { run, pending, error, saved, resetSaved: () => setSaved(false) }
}