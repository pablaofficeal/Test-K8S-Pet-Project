import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { useState, useEffect, lazy, Suspense } from 'react'

const App = lazy(() => import('./pages/app'))
const Login = lazy(() => import('./pages/login'))
const Register = lazy(() => import('./pages/register'))


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
                <Route path="/" element={<App />} />
                <Route path="/Start/log/in/my/account/login.php" element={<Login />} />
                <Route path="/Start/create/my/account/register.php" element={<Register />} />
            </Routes>
            </Suspense>
        </BrowserRouter>
    )
}
export default Router