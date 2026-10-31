import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import TextConfig from '../configs/site/text.configs'
import { loginPathWithRedirect } from '../configs/site/routes'
import { describeError } from './api'
import { getProfile, UnauthorizedError, type ProfileResponse } from './profile'
import { clearAuthSession } from './session'

// Грузит профиль для страницы вида /users/:uuid/...
// pathFor - адрес этой страницы для настоящего uuid, returnTo - куда вернуться после входа
export function useProfile(pathFor: (uuid: string) => string, returnTo: string) {
    const { uuid } = useParams()
    const navigate = useNavigate()
    const [profile, setProfile] = useState<ProfileResponse | null>(null)
    const [error, setError] = useState('')

    useEffect(() => {
        let cancelled = false

        getProfile().then(
            (data) => {
                if (cancelled) {
                    return
                }

                setProfile(data)

                // uuid в адресе и в куке - только для красоты, настоящий берём из ответа API
                if (data.uuid !== uuid) {
                    navigate(pathFor(data.uuid), { replace: true })
                }
            },
            (err) => {
                if (cancelled) {
                    return
                }

                if (err instanceof UnauthorizedError) {
                    clearAuthSession()
                    navigate(loginPathWithRedirect(returnTo), { replace: true })
                    return
                }

                setError(describeError(err, TextConfig.DashboardLoadFailed))
            },
        )

        return () => {
            cancelled = true
        }
    }, [uuid, navigate, pathFor, returnTo])

    return { profile, setProfile, error }
}