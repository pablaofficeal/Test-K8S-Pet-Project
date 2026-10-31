import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import Router from './Router'
import './assets/css/main.css'


if (typeof window !== 'undefined') {
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <Router />
    </StrictMode>,
  )
}