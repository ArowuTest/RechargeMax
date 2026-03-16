import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import './index.css'

// NOTE: Toast providers are mounted inside App.tsx (<Toaster /> + <Sonner />).
// Do NOT add a second Toaster here — it causes duplicate notifications.
createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)