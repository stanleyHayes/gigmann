import { RouterProvider } from 'react-router-dom'

import { AppProviders } from './app/providers'
import { router } from './app/router'
import { AuthProvider } from './auth/AuthProvider'
import { useAuth } from './auth/authContext'
import { LoginScreen } from './screens/LoginScreen'

function Gate() {
  const { isAuthenticated } = useAuth()
  return isAuthenticated ? <RouterProvider router={router} /> : <LoginScreen />
}

export function App() {
  return (
    <AppProviders>
      <AuthProvider>
        <Gate />
      </AuthProvider>
    </AppProviders>
  )
}
