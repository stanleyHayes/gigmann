import { alpha, createTheme, type Theme } from '@mui/material/styles'

// Owner typography directive:
//   Fraunces — titles/display · Outfit — body · JetBrains Mono — statuses/figures.
const titleFont = '"Fraunces Variable", Georgia, serif'
const bodyFont = '"Outfit Variable", system-ui, -apple-system, sans-serif'
export const monoFont = '"JetBrains Mono Variable", ui-monospace, "SFMono-Regular", monospace'

// Status palette (critical / watch / good). Per a11y rules, colour is never the
// only signal — components also render a text label.
export const statusColors = {
  good: '#157f3c',
  watch: '#a85f0a',
  critical: '#c62828',
} as const

export type ThemeMode = 'light' | 'dark'
export type ThemePreset = 'gigmann' | 'cedar' | 'gold' | 'graphite'

export const THEME_PRESETS: Record<
  ThemePreset,
  { label: string; description: string; primary: string; primaryDark: string; primaryLight: string; swatches: string[] }
> = {
  gigmann: {
    label: 'Gigmann Blue',
    description: 'The default executive cockpit palette.',
    primary: '#0b5cad',
    primaryDark: '#08437f',
    primaryLight: '#73b7f4',
    swatches: ['#08437f', '#0b5cad', '#73b7f4'],
  },
  cedar: {
    label: 'Cedar Green',
    description: 'A calmer clinical operations tint.',
    primary: '#176f56',
    primaryDark: '#0d4f3d',
    primaryLight: '#75c7ae',
    swatches: ['#0d4f3d', '#176f56', '#75c7ae'],
  },
  gold: {
    label: 'Boardroom Gold',
    description: 'A warmer finance-and-ownership accent.',
    primary: '#8a5a00',
    primaryDark: '#623f00',
    primaryLight: '#d9a441',
    swatches: ['#623f00', '#8a5a00', '#d9a441'],
  },
  graphite: {
    label: 'Graphite',
    description: 'A restrained neutral accent for low-light reviews.',
    primary: '#4f5d75',
    primaryDark: '#30394b',
    primaryLight: '#9fb0ca',
    swatches: ['#30394b', '#4f5d75', '#9fb0ca'],
  },
}

export function buildTheme(mode: ThemeMode, preset: ThemePreset = 'gigmann'): Theme {
  const isDark = mode === 'dark'
  const themePreset = THEME_PRESETS[preset]
  const primary = themePreset.primary
  const primaryDark = themePreset.primaryDark
  const primaryLight = themePreset.primaryLight
  const paper = isDark ? '#101927' : '#ffffff'
  const background = isDark ? '#08111f' : '#f6f8fb'
  const surface = isDark ? '#132033' : '#fbfdff'
  const textPrimary = isDark ? '#edf4ff' : '#172033'
  const textSecondary = isDark ? '#9fb1c8' : '#5b687a'
  const divider = isDark ? 'rgba(148, 163, 184, 0.22)' : '#dce4ef'
  const actionHover = alpha(primary, isDark ? 0.16 : 0.08)
  const actionSelected = alpha(primary, isDark ? 0.28 : 0.13)
  const actionFocus = alpha(primary, isDark ? 0.3 : 0.16)

  return createTheme({
    palette: {
      mode,
      primary: { main: primary, dark: primaryDark, light: primaryLight },
      background: {
        default: background,
        paper,
      },
      text: {
        primary: textPrimary,
        secondary: textSecondary,
      },
      divider,
      action: {
        hover: actionHover,
        selected: actionSelected,
        focus: actionFocus,
        disabledBackground: alpha(textSecondary, isDark ? 0.18 : 0.1),
      },
    },
    typography: {
      fontFamily: bodyFont,
      h1: { fontFamily: titleFont, fontWeight: 600 },
      h2: { fontFamily: titleFont, fontWeight: 600 },
      h3: { fontFamily: titleFont, fontWeight: 600 },
      h4: { fontFamily: titleFont, fontWeight: 600 },
      h5: { fontWeight: 700 },
      h6: { fontWeight: 700 },
      button: { fontWeight: 700, textTransform: 'none' },
    },
    shape: { borderRadius: 8 },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          body: {
            background:
              mode === 'dark'
                ? 'linear-gradient(180deg, #08111f 0%, #0d1625 56%, #08111f 100%)'
                : 'linear-gradient(180deg, #f6f8fb 0%, #eef4fb 58%, #f8fafc 100%)',
          },
          '::selection': {
            backgroundColor: alpha(primary, 0.18),
          },
        },
      },
      MuiAppBar: {
        styleOverrides: {
          root: {
            boxShadow: 'none',
          },
        },
      },
      MuiButton: {
        defaultProps: {
          disableElevation: true,
        },
        styleOverrides: {
          root: {
            borderRadius: 8,
            '&.MuiButton-containedPrimary': {
              backgroundColor: primary,
              '&:hover': {
                backgroundColor: primaryDark,
              },
            },
          },
        },
      },
      MuiCard: {
        styleOverrides: {
          root: {
            borderColor: divider,
            backgroundImage: 'none',
            backgroundColor: paper,
            boxShadow: `0 18px 44px ${alpha(isDark ? '#000000' : '#31415f', isDark ? 0.28 : 0.08)}`,
          },
        },
      },
      MuiPaper: {
        styleOverrides: {
          root: {
            backgroundImage: 'none',
          },
          outlined: {
            borderColor: divider,
          },
        },
      },
      MuiChip: {
        styleOverrides: {
          root: {
            fontWeight: 700,
          },
        },
      },
      MuiTextField: {
        defaultProps: {
          variant: 'outlined',
        },
      },
      MuiOutlinedInput: {
        styleOverrides: {
          root: {
            backgroundColor: surface,
          },
        },
      },
      MuiDrawer: {
        styleOverrides: {
          paper: {
            backgroundImage: 'none',
          },
        },
      },
    },
  })
}
