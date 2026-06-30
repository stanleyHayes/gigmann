export const OPEN_HELP_EVENT = 'gigmann:open-help'
export const REPLAY_TOUR_EVENT = 'gigmann:replay-tour'

export function dispatchOpenHelp(): void {
  window.dispatchEvent(new Event(OPEN_HELP_EVENT))
}

export function dispatchReplayTour(): void {
  window.dispatchEvent(new Event(REPLAY_TOUR_EVENT))
}
