import { render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

vi.mock('./api/useBrief', () => ({
  useBrief: () => ({ data: undefined, isLoading: true, isError: false }),
}))

import { AppProviders } from './app/providers'
import { routes } from './app/routes'

function renderAt(path: string) {
  const router = createMemoryRouter(routes, { initialEntries: [path] })
  return render(
    <AppProviders>
      <RouterProvider router={router} />
    </AppProviders>,
  )
}

describe('routing', () => {
  it('renders the brief on the index route', () => {
    renderAt('/')
    expect(screen.getByRole('heading', { name: /The Brief/i })).toBeInTheDocument()
    expect(screen.getByTestId('brief-skeleton')).toBeInTheDocument()
  })

  it('renders the network placeholder', () => {
    renderAt('/network')
    expect(screen.getByRole('heading', { name: /^Network$/i })).toBeInTheDocument()
  })

  it('renders the approvals placeholder', () => {
    renderAt('/approvals')
    expect(screen.getByRole('heading', { name: /Approvals/i })).toBeInTheDocument()
  })

  it('renders a not-found placeholder for unknown paths', () => {
    renderAt('/does-not-exist')
    expect(screen.getByRole('heading', { name: /Not found/i })).toBeInTheDocument()
  })
})
