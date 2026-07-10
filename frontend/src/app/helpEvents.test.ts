import { describe, expect, it, vi } from 'vitest'

import { OPEN_HELP_EVENT, REPLAY_TOUR_EVENT, dispatchOpenHelp, dispatchReplayTour } from './helpEvents'

describe('helpEvents', () => {
  it('dispatches the open-help window event', () => {
    const handler = vi.fn()
    window.addEventListener(OPEN_HELP_EVENT, handler)
    dispatchOpenHelp()
    expect(handler).toHaveBeenCalledTimes(1)
    window.removeEventListener(OPEN_HELP_EVENT, handler)
  })

  it('dispatches the replay-tour window event', () => {
    const handler = vi.fn()
    window.addEventListener(REPLAY_TOUR_EVENT, handler)
    dispatchReplayTour()
    expect(handler).toHaveBeenCalledTimes(1)
    window.removeEventListener(REPLAY_TOUR_EVENT, handler)
  })
})
