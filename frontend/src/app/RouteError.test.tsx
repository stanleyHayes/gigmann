import { render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { RouteError } from './RouteError'

afterEach(() => vi.restoreAllMocks())

describe('RouteError', () => {
  it('renders a friendly error when a route throws', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {}) // silence React Router's error log
    function Boom(): React.ReactNode {
      throw new Error('kaboom')
    }
    const router = createMemoryRouter([{ path: '/', Component: Boom, ErrorBoundary: RouteError }])
    render(<RouterProvider router={router} />)

    expect(await screen.findByText(/something went wrong/i)).toBeInTheDocument()
    expect(screen.getByText('kaboom')).toBeInTheDocument()
  })
})
