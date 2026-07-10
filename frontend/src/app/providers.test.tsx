import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'

import { AppProviders } from './providers'
import { useColorMode } from './colorMode'

function Probe() {
  const { mode, preset, setMode, setPreset, toggle } = useColorMode()
  return (
    <div>
      <span data-testid="mode">{mode}</span>
      <span data-testid="preset">{preset}</span>
      <button onClick={() => setMode('dark')}>set-dark</button>
      <button onClick={toggle}>toggle</button>
      <button onClick={() => setPreset('cedar')}>set-preset</button>
    </div>
  )
}

describe('AppProviders colour mode', () => {
  beforeEach(() => localStorage.clear())

  it('exposes setMode, toggle, and setPreset (persisting each)', () => {
    render(
      <AppProviders>
        <Probe />
      </AppProviders>,
    )

    fireEvent.click(screen.getByText('set-dark'))
    expect(screen.getByTestId('mode')).toHaveTextContent('dark')

    fireEvent.click(screen.getByText('toggle'))
    expect(screen.getByTestId('mode')).toHaveTextContent('light')

    fireEvent.click(screen.getByText('set-preset'))
    expect(screen.getByTestId('preset')).toHaveTextContent('cedar')
  })
})
