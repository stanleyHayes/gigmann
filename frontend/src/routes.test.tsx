import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { createMemoryRouter, RouterProvider, type RouteObject } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

vi.mock('./api/useBrief', () => ({
  useBrief: () => ({ data: undefined, isLoading: true, isError: false }),
}))

vi.mock('./api/useAlerts', () => ({
  useAlerts: () => ({ data: { alerts: [] }, isLoading: false, isError: false }),
  useUpdateAlertStatus: () => ({ mutate: () => {}, isPending: false }),
}))

vi.mock('./api/useTasks', () => ({
  useCreateTask: () => ({ mutate: () => {}, isPending: false, isError: false, reset: () => {} }),
}))

vi.mock('./api/useFacilities', () => ({
  useFacilities: () => ({ data: [], isLoading: false, isError: false }),
}))

vi.mock('./api/useApprovals', () => ({
  useApprovals: () => ({ data: [], isLoading: false, isError: false }),
  useDecideApproval: () => ({ mutate: () => {}, isPending: false }),
}))

import { AppProviders } from './app/providers'
import { AppShell } from './app/AppShell'
import { Placeholder } from './screens/Placeholder'
import { RouteError } from './app/RouteError'
import { AuthProvider } from './auth/AuthProvider'
import { HomeScreen } from './screens/HomeScreen'
import { NetworkScreen } from './screens/NetworkScreen'
import { ApprovalsScreen } from './screens/ApprovalsScreen'

// Eager test routes avoid dynamic imports, which hang in isolated Vitest runs
// because jsdom cannot resolve the lazy chunks before other tests have loaded
// the screen modules via static imports.
const testRoutes: RouteObject[] = [
  {
    path: '/',
    Component: AppShell,
    ErrorBoundary: RouteError,
    children: [
      { index: true, Component: HomeScreen },
      { path: 'network', Component: NetworkScreen },
      { path: 'approvals', Component: ApprovalsScreen },
      { path: 'settings', element: <Placeholder title="Settings" note="Cockpit settings." /> },
      { path: '*', element: <Placeholder title="Not found" note="That page does not exist." /> },
    ],
  },
]

function renderAt(path: string) {
  const router = createMemoryRouter(testRoutes, { initialEntries: [path] })
  const view = render(
    <AppProviders>
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>
    </AppProviders>,
  )
  return { router, ...view }
}

describe('route chunks', () => {
  it('every lazy route resolves its screen component', async () => {
    const { routes } = await import('./app/routes')
    const children = routes[0].children ?? []
    const lazy = children.filter((c) => c.lazy && typeof c.lazy === 'object' && 'Component' in c.lazy)
    expect(lazy.length).toBeGreaterThan(5)
    for (const child of lazy) {
      const load = (child.lazy as unknown as { Component: () => Promise<Record<string, unknown>> }).Component
      expect(await load()).toBeTruthy()
    }
  })
})

describe('routing', () => {
  it('renders the brief on the index route', async () => {
    renderAt('/')
    expect(await screen.findByRole('heading', { name: /The Brief/i })).toBeInTheDocument()
    expect(await screen.findByTestId('brief-skeleton')).toBeInTheDocument()
  })

  it('renders the network screen', async () => {
    renderAt('/network')
    expect(await screen.findByRole('heading', { name: /^Network$/i })).toBeInTheDocument()
  })

  it('renders the approvals screen', async () => {
    renderAt('/approvals')
    expect(await screen.findByRole('heading', { name: /Approvals/i })).toBeInTheDocument()
  })

  it('renders a not-found placeholder for unknown paths', async () => {
    renderAt('/does-not-exist')
    expect(await screen.findByRole('heading', { name: /Not found/i })).toBeInTheDocument()
  })

  it('keeps clicked sidebar routes on the selected destination', async () => {
    const { router } = renderAt('/')

    fireEvent.click((await screen.findAllByRole('link', { name: /Network/i }))[0])
    await waitFor(() => expect(router.state.location.pathname).toBe('/network'))
    expect(await screen.findByRole('heading', { name: /^Network$/i })).toBeInTheDocument()

    fireEvent.click((await screen.findAllByRole('link', { name: /Approvals/i }))[0])
    await waitFor(() => expect(router.state.location.pathname).toBe('/approvals'))
    expect(await screen.findByRole('heading', { name: /Approvals/i })).toBeInTheDocument()

    fireEvent.click((await screen.findAllByRole('link', { name: /Settings/i }))[0])
    await waitFor(() => expect(router.state.location.pathname).toBe('/settings'))
    expect(await screen.findByRole('heading', { name: /Settings/i })).toBeInTheDocument()
  })
})
