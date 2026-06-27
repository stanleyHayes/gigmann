import type { RouteObject } from 'react-router-dom'

import { Placeholder } from '../screens/Placeholder'
import { AppShell } from './AppShell'
import { RouteError } from './RouteError'

// The AppShell layout stays eager; each real screen is a lazily-loaded chunk
// (dynamic import auto-splits in Vite). Charts/MUI-X load only when visited.
export const routes: RouteObject[] = [
  {
    path: '/',
    Component: AppShell,
    ErrorBoundary: RouteError,
    children: [
      { index: true, lazy: { Component: async () => (await import('../screens/HomeScreen')).HomeScreen } },
      { path: 'network', lazy: { Component: async () => (await import('../screens/NetworkScreen')).NetworkScreen } },
      { path: 'facilities/:facilityId', lazy: { Component: async () => (await import('../screens/FacilityDetailScreen')).FacilityDetailScreen } },
      { path: 'kpis', lazy: { Component: async () => (await import('../screens/KpisScreen')).KpisScreen } },
      { path: 'reports', lazy: { Component: async () => (await import('../screens/ReportsScreen')).ReportsScreen } },
      { path: 'ask', lazy: { Component: async () => (await import('../screens/AskScreen')).AskScreen } },
      { path: 'my-day', lazy: { Component: async () => (await import('../screens/MyDayScreen')).MyDayScreen } },
      { path: 'approvals', lazy: { Component: async () => (await import('../screens/ApprovalsScreen')).ApprovalsScreen } },
      { path: 'settings', lazy: { Component: async () => (await import('../screens/SettingsScreen')).SettingsScreen } },
      { path: '*', element: <Placeholder title="Not found" note="That page does not exist." /> },
    ],
  },
]
