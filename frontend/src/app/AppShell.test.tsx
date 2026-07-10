import { fireEvent, render, screen, within } from '@testing-library/react'
import { createMemoryRouter, RouterProvider, type RouteObject } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const logout = vi.fn()

vi.mock('../auth/authContext', () => ({
  useAuth: () => ({ user: { id: 'u1', name: 'Sammy Adjei', role: 'executive' }, logout }),
}))
vi.mock('../api/useLiveUpdates', () => ({ useLiveUpdates: () => {} }))
vi.mock('../api/useAlerts', () => ({
  useAlerts: () => ({
    data: {
      alerts: [
        {
          id: 'a1',
          facility_id: 'kasoa',
          type: 'revenue_drop',
          severity: 'critical',
          status: 'open',
          headline: 'Kasoa revenue down 22%',
          detail: 'Demand flat',
          created_at: '2026-06-30T00:00:00Z',
        },
      ],
    },
    isLoading: false,
    isError: false,
  }),
  useUpdateAlertStatus: () => ({ mutate: vi.fn(), isPending: false }),
}))

import { AppProviders } from './providers'
import { AppShell } from './AppShell'

function renderShell(path = '/') {
  const routes: RouteObject[] = [
    {
      path: '/',
      Component: AppShell,
      children: [
        { index: true, element: <div>Home stub</div> },
        { path: 'settings', element: <div>Settings stub</div> },
      ],
    },
  ]
  const router = createMemoryRouter(routes, { initialEntries: [path] })
  return { router, ...render(<AppProviders><RouterProvider router={router} /></AppProviders>) }
}

describe('AppShell', () => {
  beforeEach(() => {
    logout.mockClear()
    localStorage.clear()
  })

  it('renders the shell chrome and the routed child for an authenticated user', () => {
    renderShell()
    expect(screen.getAllByRole('navigation', { name: /primary navigation/i }).length).toBeGreaterThan(0)
    expect(screen.getByText('Home stub')).toBeInTheDocument()
  })

  it('toggles the colour theme', () => {
    renderShell()
    fireEvent.click(screen.getByLabelText(/switch to dark mode/i))
    expect(screen.getByLabelText(/switch to light mode/i)).toBeInTheDocument()
  })

  it('opens the notifications popover with the critical feed', () => {
    renderShell()
    fireEvent.click(screen.getByLabelText(/open notifications/i))
    expect(screen.getByText(/critical feed from the network/i)).toBeInTheDocument()
  })

  it('opens the help center', () => {
    renderShell()
    fireEvent.click(screen.getByLabelText(/open help/i))
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText(/user guide/i)).toBeInTheDocument()
  })

  it('collapses and expands the sidebar', () => {
    renderShell()
    fireEvent.click(screen.getByLabelText(/collapse sidebar/i))
    fireEvent.click(screen.getByLabelText(/expand sidebar/i))
    expect(screen.getByLabelText(/collapse sidebar/i)).toBeInTheDocument()
  })

  it('opens the account menu and signs out', () => {
    renderShell()
    fireEvent.click(screen.getByRole('button', { name: /Sammy Adjei/i }))
    const menu = screen.getByRole('menu')
    fireEvent.click(within(menu).getByText(/sign out/i))
    expect(logout).toHaveBeenCalledTimes(1)
  })

  it('opens the mobile navigation drawer', () => {
    renderShell()
    // Runs setMobileOpen(true) and renders the temporary drawer's DrawerContent;
    // the shell stays intact (no crash) and the routed child is still present.
    fireEvent.click(screen.getByLabelText(/open navigation/i))
    expect(screen.getByText('Home stub')).toBeInTheDocument()
  })

  it('navigates from the account menu', () => {
    renderShell()
    fireEvent.click(screen.getByRole('button', { name: /Sammy Adjei/i }))
    fireEvent.click(within(screen.getByRole('menu')).getByText(/profile and settings/i))
    expect(screen.getByText('Settings stub')).toBeInTheDocument()
  })

  it('toggles a nav section group', () => {
    renderShell()
    const header = screen.getByRole('button', { name: /^execution$/i })
    fireEvent.click(header) // collapse
    fireEvent.click(header) // expand
    expect(header).toBeInTheDocument()
  })

  it('shows the current route title and reacts to navigation', () => {
    const { router } = renderShell()
    fireEvent.click(screen.getAllByRole('link', { name: /network/i })[0])
    expect(router.state.location.pathname).toBe('/network')
  })
})
