import { useEffect, useState } from 'react'
import Alert from '@mui/material/Alert'
import Avatar from '@mui/material/Avatar'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Checkbox from '@mui/material/Checkbox'
import ContentCopyOutlined from '@mui/icons-material/ContentCopyOutlined'
import Divider from '@mui/material/Divider'
import FormControlLabel from '@mui/material/FormControlLabel'
import FormGroup from '@mui/material/FormGroup'
import Grid from '@mui/material/Grid'
import HelpOutlineOutlined from '@mui/icons-material/HelpOutlineOutlined'
import MenuBookOutlined from '@mui/icons-material/MenuBookOutlined'
import PaletteOutlined from '@mui/icons-material/PaletteOutlined'
import PersonOutlineOutlined from '@mui/icons-material/PersonOutlineOutlined'
import RocketLaunchOutlined from '@mui/icons-material/RocketLaunchOutlined'
import Stack from '@mui/material/Stack'
import Switch from '@mui/material/Switch'
import Tab from '@mui/material/Tab'
import Tabs from '@mui/material/Tabs'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'
import NotificationsActiveOutlined from '@mui/icons-material/NotificationsActiveOutlined'
import NotificationsOffOutlined from '@mui/icons-material/NotificationsOffOutlined'
import SecurityOutlined from '@mui/icons-material/SecurityOutlined'
import SettingsOutlined from '@mui/icons-material/SettingsOutlined'
import TuneOutlined from '@mui/icons-material/TuneOutlined'

import { useMfaConfirm, useMfaDisable, useMfaEnroll } from '../api/useMfa'
import { usePreferences, useSavePreferences } from '../api/usePreferences'
import { usePush } from '../api/usePush'
import { useAuth } from '../auth/authContext'
import { useColorMode } from '../app/colorMode'
import { dispatchOpenHelp, dispatchReplayTour } from '../app/helpEvents'
import { ButtonLoadingDots } from '../components/ButtonLoadingDots'
import { EmptyState } from '../components/EmptyState'
import { MfaQrCode } from '../components/MfaQrCode'
import { PageHeader } from '../components/PageHeader'
import { SurfaceCard } from '../components/SurfaceCard'
import { monoFont, THEME_PRESETS, type ThemePreset } from '../theme'

type SettingsTab = 'profile' | 'security' | 'preferences' | 'notifications' | 'appearance' | 'guide'

const WATCHABLE = [
  { key: 'revenue', label: 'Revenue' },
  { key: 'patients', label: 'Patients seen' },
  { key: 'occupancy', label: 'Occupancy' },
  { key: 'denial_rate', label: 'NHIS denial rate' },
]

const THRESHOLDS = [
  { key: 'denial_rate', label: 'NHIS denial rate threshold', helper: 'Use a decimal ratio, e.g. 0.12 for 12%.' },
  { key: 'occupancy', label: 'Occupancy threshold', helper: 'Use a decimal ratio, e.g. 0.85 for 85%.' },
  { key: 'stockout_days', label: 'Stockout warning days', helper: 'Number of days of stock remaining.' },
]

function initials(name: string | undefined): string {
  if (!name) {
    return 'A'
  }
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join('')
}

function setThresholdValue(current: Record<string, number>, key: string, raw: string): Record<string, number> {
  const next = { ...current }
  if (raw.trim() === '') {
    delete next[key]
    return next
  }
  const parsed = Number(raw)
  if (Number.isFinite(parsed)) {
    next[key] = parsed
  }
  return next
}

/** SettingsScreen is the Aura-style account centre for profile, MFA, preferences, alerts, and appearance. */
export function SettingsScreen() {
  const { user } = useAuth()
  const { mode, preset, setPreset, toggle } = useColorMode()
  const enroll = useMfaEnroll()
  const confirm = useMfaConfirm()
  const disable = useMfaDisable()
  const [tab, setTab] = useState<SettingsTab>('security')
  const [code, setCode] = useState('')
  const [disableCode, setDisableCode] = useState('')
  const [copyStatus, setCopyStatus] = useState<'idle' | 'copied' | 'error'>('idle')
  const secret = enroll.data?.secret
  const enrollmentSecret = secret ?? ''
  const otpauthUri = enroll.data?.otpauth_uri
  const recoveryCodes = confirm.data?.recovery_codes ?? []
  const mfaEnabled = (Boolean(user?.mfa_enabled) || confirm.isSuccess) && !disable.isSuccess
  const enrollmentStarted = Boolean(secret) && !confirm.isSuccess && !disable.isSuccess

  const push = usePush()
  const prefs = usePreferences()
  const savePrefs = useSavePreferences()
  const [watched, setWatched] = useState<string[]>([])
  const [thresholds, setThresholds] = useState<Record<string, number>>({})

  useEffect(() => {
    if (prefs.data) {
      setWatched(prefs.data.watched_metrics)
      setThresholds(prefs.data.thresholds)
    }
  }, [prefs.data])

  const toggleWatched = (key: string) =>
    setWatched((w) => (w.includes(key) ? w.filter((k) => k !== key) : [...w, key]))

  const copyRecoveryCodes = async () => {
    if (recoveryCodes.length === 0) {
      return
    }
    try {
      await navigator.clipboard.writeText(recoveryCodes.join('\n'))
      setCopyStatus('copied')
    } catch {
      setCopyStatus('error')
    }
  }

  const startEnrollment = () => {
    confirm.reset()
    disable.reset()
    setCode('')
    setDisableCode('')
    enroll.mutate()
  }

  return (
    <Stack spacing={3}>
      <PageHeader
        title="Settings"
        eyebrow="Account controls"
        description="Profile, security, watched metrics, notifications, and cockpit appearance."
        icon={SettingsOutlined}
      />

      <SurfaceCard title="Executive profile" icon={PersonOutlineOutlined}>
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} sx={{ alignItems: { xs: 'flex-start', sm: 'center' } }}>
          <Avatar sx={{ width: 64, height: 64, bgcolor: 'primary.main', color: 'primary.contrastText', fontWeight: 800 }}>
            {initials(user?.name)}
          </Avatar>
          <Box sx={{ minWidth: 0 }}>
            <Typography variant="h5">{user?.name ?? 'Executive user'}</Typography>
            <Typography variant="body2" color="text.secondary" sx={{ textTransform: 'capitalize' }}>
              {user?.role ?? 'signed-in account'} · {mfaEnabled ? 'MFA enabled' : 'MFA not enabled'}
            </Typography>
          </Box>
        </Stack>
      </SurfaceCard>

      <Tabs
        value={tab}
        onChange={(_, value: SettingsTab) => setTab(value)}
        variant="scrollable"
        allowScrollButtonsMobile
        aria-label="Settings sections"
      >
        <Tab value="profile" label="Profile" icon={<PersonOutlineOutlined />} iconPosition="start" />
        <Tab value="security" label="Security" icon={<SecurityOutlined />} iconPosition="start" />
        <Tab value="preferences" label="Preferences" icon={<TuneOutlined />} iconPosition="start" />
        <Tab value="notifications" label="Notifications" icon={<NotificationsActiveOutlined />} iconPosition="start" />
        <Tab
          value="appearance"
          label="Appearance"
          icon={<PaletteOutlined />}
          iconPosition="start"
        />
        <Tab value="guide" label="Guide" icon={<HelpOutlineOutlined />} iconPosition="start" />
      </Tabs>

      {tab === 'profile' ? (
        <SurfaceCard
          title="Profile details"
          description="The profile summary keeps the cockpit personalised without exposing editable identity controls that are owned by auth."
          icon={PersonOutlineOutlined}
        >
          <Grid container spacing={2}>
            {[
              ['Name', user?.name ?? 'Not available'],
              ['Role', user?.role ?? 'Not available'],
              ['Account ID', user?.id ?? 'Not available'],
              ['Security posture', mfaEnabled ? 'Two-factor authentication on' : 'Two-factor authentication off'],
            ].map(([label, value]) => (
              <Grid key={label} size={{ xs: 12, md: 6 }}>
                <Box sx={{ border: 1, borderColor: 'divider', borderRadius: 1.5, p: 1.5, bgcolor: 'background.default' }}>
                  <Typography variant="caption" color="text.secondary" sx={{ display: 'block', fontWeight: 800 }}>
                    {label}
                  </Typography>
                  <Typography variant="body2" sx={{ mt: 0.5, fontWeight: 800, overflowWrap: 'anywhere' }}>
                    {value}
                  </Typography>
                </Box>
              </Grid>
            ))}
          </Grid>
        </SurfaceCard>
      ) : null}

      {tab === 'security' ? (
        <SurfaceCard
          title="Two-factor authentication"
          description="Use an authenticator app and recovery codes to protect executive access."
          icon={SecurityOutlined}
        >
          <Stack spacing={2}>
            {mfaEnabled ? (
              <Stack spacing={2}>
                <Alert severity="success">Two-factor authentication is on for your account.</Alert>
                {recoveryCodes.length > 0 ? (
                  <Stack spacing={1}>
                    <Typography variant="body2" color="text.secondary">
                      Save these one-time recovery codes before leaving this screen.
                    </Typography>
                    <Stack
                      component="ul"
                      spacing={0.75}
                      sx={{
                        m: 0,
                        p: 2,
                        listStyle: 'none',
                        border: '1px solid',
                        borderColor: 'divider',
                        borderRadius: 1,
                        bgcolor: 'action.hover',
                      }}
                    >
                      {recoveryCodes.map((item) => (
                        <Typography key={item} component="li" sx={{ fontFamily: monoFont }}>
                          {item}
                        </Typography>
                      ))}
                    </Stack>
                    <Button
                      variant="outlined"
                      startIcon={<ContentCopyOutlined fontSize="small" />}
                      onClick={() => void copyRecoveryCodes()}
                      sx={{ alignSelf: 'flex-start' }}
                    >
                      Copy recovery codes
                    </Button>
                    {copyStatus === 'copied' ? (
                      <Typography variant="caption" color="success.main">
                        Copied to clipboard.
                      </Typography>
                    ) : null}
                    {copyStatus === 'error' ? (
                      <Alert severity="warning">
                        Couldn&apos;t copy automatically. Select and copy the codes manually.
                      </Alert>
                    ) : null}
                  </Stack>
                ) : null}
                <Stack spacing={1.5} sx={{ alignItems: 'flex-start', maxWidth: 360 }}>
                  <TextField
                    label="Code to disable"
                    value={disableCode}
                    onChange={(e) => setDisableCode(e.target.value)}
                    autoComplete="one-time-code"
                    helperText="Use an authenticator code or unused recovery code."
                  />
                  {disable.isError ? <Alert severity="error">That code didn&apos;t match. Try again.</Alert> : null}
                  <Button
                    variant="outlined"
                    color="error"
                    onClick={() => disable.mutate({ code: disableCode })}
                    disabled={disable.isPending || disableCode === ''}
                    sx={{ alignSelf: 'flex-start' }}
                  >
                    {disable.isPending ? <ButtonLoadingDots /> : null}
                    Disable two-factor auth
                  </Button>
                </Stack>
              </Stack>
            ) : !enrollmentStarted ? (
              <Stack spacing={1} sx={{ alignItems: 'flex-start' }}>
                {disable.isSuccess ? <Alert severity="success">Two-factor authentication is off.</Alert> : null}
                <Typography variant="body2" color="text.secondary">
                  Protect your account with a time-based one-time code from an authenticator app.
                </Typography>
                <Button variant="contained" onClick={startEnrollment} disabled={enroll.isPending}>
                  {enroll.isPending ? <ButtonLoadingDots /> : null}
                  Set up two-factor auth
                </Button>
              </Stack>
            ) : (
              <Stack spacing={2}>
                <Typography variant="body2" color="text.secondary">
                  Scan the QR code with your authenticator app, or enter the key manually, then type the
                  6-digit code to confirm.
                </Typography>
                {otpauthUri ? <MfaQrCode uri={otpauthUri} size={184} /> : null}
                <Typography sx={{ fontFamily: monoFont, wordBreak: 'break-all' }}>{enrollmentSecret}</Typography>
                <TextField
                  label="Authenticator code"
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                  inputMode="numeric"
                  autoComplete="one-time-code"
                  sx={{ maxWidth: 240 }}
                />
                {confirm.isError ? <Alert severity="error">That code didn&apos;t match. Try again.</Alert> : null}
                <Button
                  variant="contained"
                  onClick={() => confirm.mutate({ secret: enrollmentSecret, code })}
                  disabled={confirm.isPending || code === ''}
                  sx={{ alignSelf: 'flex-start' }}
                >
                  {confirm.isPending ? <ButtonLoadingDots /> : null}
                  Confirm
                </Button>
              </Stack>
            )}
          </Stack>
        </SurfaceCard>
      ) : null}

      {tab === 'preferences' ? (
        <SurfaceCard
          title="What you watch"
          description="Choose the figures and alert thresholds the cockpit should prioritise in your daily readout."
          icon={TuneOutlined}
        >
          <Stack spacing={2}>
            <FormGroup>
              {WATCHABLE.map((m) => (
                <FormControlLabel
                  key={m.key}
                  control={<Checkbox checked={watched.includes(m.key)} onChange={() => toggleWatched(m.key)} />}
                  label={m.label}
                />
              ))}
            </FormGroup>
            <Divider />
            <Grid container spacing={2}>
              {THRESHOLDS.map((item) => (
                <Grid key={item.key} size={{ xs: 12, md: 4 }}>
                  <TextField
                    fullWidth
                    type="number"
                    label={item.label}
                    value={thresholds[item.key] ?? ''}
                    onChange={(e) => setThresholds((current) => setThresholdValue(current, item.key, e.target.value))}
                    helperText={item.helper}
                    slotProps={{ htmlInput: { step: '0.01' } }}
                  />
                </Grid>
              ))}
            </Grid>
            {savePrefs.isSuccess ? <Alert severity="success">Preferences saved.</Alert> : null}
            <Button
              variant="contained"
              onClick={() => savePrefs.mutate({ watched_metrics: watched, thresholds })}
              disabled={savePrefs.isPending || prefs.isLoading}
              sx={{ alignSelf: 'flex-start' }}
            >
              {savePrefs.isPending ? <ButtonLoadingDots /> : null}
              Save preferences
            </Button>
          </Stack>
        </SurfaceCard>
      ) : null}

      {tab === 'notifications' ? (
        <SurfaceCard
          title="Critical alerts"
          description="Quiet by default: only urgent network events should reach this device."
          icon={NotificationsActiveOutlined}
        >
          <Stack spacing={2}>
            {push.error ? <Alert severity="error">{push.error}</Alert> : null}
            {!push.supported ? (
              <EmptyState
                compact
                icon={NotificationsOffOutlined}
                title="Push is not supported"
                description="This browser cannot receive critical device notifications."
              />
            ) : !push.available ? (
              <EmptyState
                compact
                icon={NotificationsOffOutlined}
                title="Push is not configured"
                description="Server push keys are not available yet."
              />
            ) : (
              <FormControlLabel
                control={
                  <Switch
                    checked={push.enabled}
                    disabled={push.busy}
                    onChange={() => (push.enabled ? void push.disable() : void push.enable())}
                  />
                }
                label="Send critical push notifications to this device"
              />
            )}
          </Stack>
        </SurfaceCard>
      ) : null}

      {tab === 'appearance' ? (
        <SurfaceCard
          title="Appearance"
          description="Match the cockpit to the working environment with light/dark mode and a focused accent preset."
          icon={PaletteOutlined}
        >
          <Stack spacing={2.5}>
            <FormControlLabel
              control={<Switch checked={mode === 'dark'} onChange={toggle} />}
              label={mode === 'dark' ? 'Use dark mode' : 'Use light mode'}
            />
            <Box role="radiogroup" aria-label="Theme preset">
              <Typography variant="overline" color="text.secondary" sx={{ display: 'block', fontWeight: 800, letterSpacing: 0 }}>
                Theme preset
              </Typography>
              <Grid container spacing={2} sx={{ mt: 0.5 }}>
                {(Object.entries(THEME_PRESETS) as [ThemePreset, (typeof THEME_PRESETS)[ThemePreset]][]).map(([key, item]) => {
                  const selected = key === preset
                  return (
                    <Grid key={key} size={{ xs: 12, sm: 6, lg: 3 }}>
                      <Box
                        component="label"
                        sx={{
                          display: 'block',
                          height: '100%',
                          cursor: 'pointer',
                          border: 1,
                          borderColor: selected ? 'primary.main' : 'divider',
                          borderRadius: 2,
                          bgcolor: 'background.default',
                          p: 1.5,
                          boxShadow: selected ? (theme) => `0 0 0 3px ${theme.palette.action.selected}` : 'none',
                        }}
                      >
                        <input
                          type="radio"
                          name="theme-preset"
                          value={key}
                          checked={selected}
                          onChange={() => setPreset(key)}
                          style={{ position: 'absolute', opacity: 0, pointerEvents: 'none' }}
                        />
                        <Stack spacing={1}>
                          <Stack direction="row" spacing={1} sx={{ alignItems: 'center', justifyContent: 'space-between' }}>
                            <Typography variant="subtitle2" sx={{ fontWeight: 800 }}>
                              {item.label}
                            </Typography>
                            {selected ? (
                              <Typography variant="caption" color="primary.main" sx={{ fontWeight: 800 }}>
                                Active
                              </Typography>
                            ) : null}
                          </Stack>
                          <Typography variant="body2" color="text.secondary" sx={{ lineHeight: 1.6 }}>
                            {item.description}
                          </Typography>
                          <Stack direction="row" sx={{ overflow: 'hidden', borderRadius: 1.25, border: 1, borderColor: 'divider' }}>
                            {item.swatches.map((swatch) => (
                              <Box key={swatch} aria-hidden="true" sx={{ height: 34, flex: 1, bgcolor: swatch }} />
                            ))}
                          </Stack>
                        </Stack>
                      </Box>
                    </Grid>
                  )
                })}
              </Grid>
            </Box>
          </Stack>
        </SurfaceCard>
      ) : null}

      {tab === 'guide' ? (
        <SurfaceCard
          title="Guide and tour"
          description="Open the cockpit reference or replay the guided walkthrough from anywhere."
          icon={MenuBookOutlined}
        >
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
            <Button variant="contained" startIcon={<RocketLaunchOutlined />} onClick={dispatchReplayTour}>
              Show me around
            </Button>
            <Button variant="outlined" startIcon={<MenuBookOutlined />} onClick={dispatchOpenHelp}>
              Open user guide
            </Button>
          </Stack>
        </SurfaceCard>
      ) : null}
    </Stack>
  )
}
