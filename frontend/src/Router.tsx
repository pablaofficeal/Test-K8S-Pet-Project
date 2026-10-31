import { BrowserRouter, Navigate, Routes, Route } from 'react-router-dom'
import { useState, useEffect, lazy, Suspense } from 'react'
import { ROUTES } from './lib/configs/site/routes'

const App = lazy(() => import('./pages/app'))
const Login = lazy(() => import('./pages/login'))
const Register = lazy(() => import('./pages/register'))
const Dashboard = lazy(() => import('./pages/dashboard'))
const DashboardRedirect = lazy(() => import('./pages/dashboardRedirect'))


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
                <Route path={ROUTES.DASHBOARD} element={<DashboardRedirect />} />
                <Route path="/users/:uuid/dashboard" element={<Dashboard />} />
                <Route path="*" element={<Navigate to={ROUTES.HOME} replace />} />
            </Routes>
            </Suspense>
        </BrowserRouter>
    )
}
export default Router