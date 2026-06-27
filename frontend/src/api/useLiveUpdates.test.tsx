import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { useLiveUpdates } from './useLiveUpdates'

function Probe() {
  useLiveUpdates()
  return null
}

describe('useLiveUpdates', () => {
  it('no-ops without WebSocket/token and never throws', () => {
    const qc = new QueryClient()
    expect(() =>
      render(
        <QueryClientProvider client={qc}>
          <Probe />
        </QueryClientProvider>,
      ),
    ).not.toThrow()
  })
})
