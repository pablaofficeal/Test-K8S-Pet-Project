import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { useState, useEffect, lazy, Suspense } from 'react'

const App = lazy(() => import('./pages/app'))


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
            </Routes>
            </Suspense>
        </BrowserRouter>
    )
}
export default Router