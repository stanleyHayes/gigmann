import { render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

vi.mock('./api/useBrief', () => ({
  useBrief: () => ({ data: undefined, isLoading: true, isError: false }),
}))

vi.mock('./api/useFacilities', () => ({
  useFacilities: () => ({ data: [], isLoading: false, isError: false }),
}))

vi.mock('./api/useApprovals', () => ({
  useApprovals: () => ({ data: [], isLoading: false, isError: false }),
  useDecideApproval: () => ({ mutate: () => {}, isPending: false }),
}))

import { AppProviders } from './app/providers'
import { routes } from './app/routes'
import { AuthProvider } from './auth/AuthProvider'

function renderAt(path: string) {
  const router = createMemoryRouter(routes, { initialEntries: [path] })
  return render(
    <AppProviders>
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>
    </AppProviders>,
  )
}

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
})
