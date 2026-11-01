import { BrowserRouter, Navigate, Routes, Route } from 'react-router-dom'
import { useState, useEffect, lazy, Suspense } from 'react'
import { ROUTES, userDashboardPath, userProfilePath, userSettingsPath, userTodoPath } from './lib/configs/site/routes'

const App = lazy(() => import('./pages/app'))
const Login = lazy(() => import('./pages/auth/login'))
const Register = lazy(() => import('./pages/auth/register'))
const Verify = lazy(() => import('./pages/auth/verify'))
const Dashboard = lazy(() => import('./pages/users/dashboard'))
const Profile = lazy(() => import('./pages/users/profile'))
const Settings = lazy(() => import('./pages/users/settings'))
const TodoProjects = lazy(() => import('./pages/todo/projects'))
const TodoBoard = lazy(() => import('./pages/todo/board'))
const UserRedirect = lazy(() => import('./pages/users/userRedirect'))


function Router() {
    const [mounted, setMounted] = useState(false)

    useEffect(() => {
        setMounted(true)
    }, [])

    if (!mounted) {
        return null
    }

    return (
        <BrowserRouter>
            <Suspense fallback={null}>
            <Routes>
                <Route path={ROUTES.HOME} element={<App />} />
                <Route path={ROUTES.LOGIN} element={<Login />} />
                <Route path={ROUTES.REGISTER} element={<Register />} />
                <Route path={ROUTES.VERIFY} element={<Verify />} />
                <Route path={ROUTES.DASHBOARD} element={<UserRedirect from={ROUTES.DASHBOARD} to={userDashboardPath} />} />
                <Route path={ROUTES.PROFILE} element={<UserRedirect from={ROUTES.PROFILE} to={userProfilePath} />} />
                <Route path={ROUTES.SETTINGS} element={<UserRedirect from={ROUTES.SETTINGS} to={userSettingsPath} />} />
                <Route path={ROUTES.TODO} element={<UserRedirect from={ROUTES.TODO} to={userTodoPath} />} />
                <Route path="/users/:uuid/dashboard" element={<Dashboard />} />
                <Route path="/users/:uuid/profile" element={<Profile />} />
                <Route path="/users/:uuid/settings" element={<Settings />} />
                <Route path="/users/:uuid/todo" element={<TodoProjects />} />
                <Route path="/users/:uuid/todo/:boardId" element={<TodoBoard />} />
                <Route path="*" element={<Navigate to={ROUTES.HOME} replace />} />
            </Routes>
            </Suspense>
        </BrowserRouter>
    )
}
export default Router