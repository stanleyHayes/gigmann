import { createTheme, type Theme } from '@mui/material/styles'

// Owner typography directive:
//   Fraunces — titles/display · Outfit — body · JetBrains Mono — statuses/figures.
const titleFont = '"Fraunces", Georgia, serif'
const bodyFont = '"Outfit", system-ui, -apple-system, sans-serif'
export const monoFont = '"JetBrains Mono", ui-monospace, "SFMono-Regular", monospace'

// Status palette (critical / watch / good). Per a11y rules, colour is never the
// only signal — components also render a text label.
export const statusColors = {
  good: '#1f9d55',
  watch: '#d9822b',
  critical: '#d64545',
} as const

export type ThemeMode = 'light' | 'dark'

export function buildTheme(mode: ThemeMode): Theme {
  return createTheme({
    palette: {
      mode,
      primary: { main: '#0b5cad' },
    },
    typography: {
      fontFamily: bodyFont,
      h1: { fontFamily: titleFont, fontWeight: 600 },
      h2: { fontFamily: titleFont, fontWeight: 600 },
      h3: { fontFamily: titleFont, fontWeight: 600 },
    },
    shape: { borderRadius: 12 },
  })
}
