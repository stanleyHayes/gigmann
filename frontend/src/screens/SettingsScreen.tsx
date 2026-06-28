import { useEffect, useState } from 'react'
import Alert from '@mui/material/Alert'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Checkbox from '@mui/material/Checkbox'
import FormControlLabel from '@mui/material/FormControlLabel'
import FormGroup from '@mui/material/FormGroup'
import Stack from '@mui/material/Stack'
import Switch from '@mui/material/Switch'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'

import { useMfaConfirm, useMfaEnroll } from '../api/useMfa'
import { usePreferences, useSavePreferences } from '../api/usePreferences'
import { usePush } from '../api/usePush'
import { ButtonLoadingDots } from '../components/ButtonLoadingDots'
import { monoFont } from '../theme'

/** SettingsScreen lets the user opt into two-factor authentication. */
const WATCHABLE = [
  { key: 'revenue', label: 'Revenue' },
  { key: 'patients', label: 'Patients seen' },
  { key: 'occupancy', label: 'Occupancy' },
  { key: 'denial_rate', label: 'NHIS denial rate' },
]

export function SettingsScreen() {
  const enroll = useMfaEnroll()
  const confirm = useMfaConfirm()
  const [code, setCode] = useState('')
  const secret = enroll.data?.secret

  const push = usePush()
  const prefs = usePreferences()
  const savePrefs = useSavePreferences()
  const [watched, setWatched] = useState<string[]>([])
  useEffect(() => {
    if (prefs.data) {
      setWatched(prefs.data.watched_metrics)
    }
  }, [prefs.data])
  const toggleWatched = (key: string) =>
    setWatched((w) => (w.includes(key) ? w.filter((k) => k !== key) : [...w, key]))

  return (
    <Stack spacing={3}>
      <Typography variant="h1" sx={{ fontSize: { xs: '2rem', md: '2.5rem' } }}>
        Settings
      </Typography>

      <Card variant="outlined">
        <CardContent>
          <Stack spacing={2}>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              Two-factor authentication
            </Typography>
            {confirm.isSuccess ? (
              <Alert severity="success">Two-factor authentication is on for your account.</Alert>
            ) : !secret ? (
              <Stack spacing={1} sx={{ alignItems: 'flex-start' }}>
                <Typography variant="body2" color="text.secondary">
                  Protect your account with a time-based one-time code from an authenticator app.
                </Typography>
                <Button variant="contained" onClick={() => enroll.mutate()} disabled={enroll.isPending}>
                  {enroll.isPending ? <ButtonLoadingDots /> : null}
                  Set up two-factor auth
                </Button>
              </Stack>
            ) : (
              <Stack spacing={2}>
                <Typography variant="body2" color="text.secondary">
                  Add this key to your authenticator app, then enter the 6-digit code to confirm.
                </Typography>
                <Typography sx={{ fontFamily: monoFont, wordBreak: 'break-all' }}>{secret}</Typography>
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
                  onClick={() => confirm.mutate({ secret, code })}
                  disabled={confirm.isPending || code === ''}
                  sx={{ alignSelf: 'flex-start' }}
                >
                  {confirm.isPending ? <ButtonLoadingDots /> : null}
                  Confirm
                </Button>
              </Stack>
            )}
          </Stack>
        </CardContent>
      </Card>

      <Card variant="outlined">
        <CardContent>
          <Stack spacing={2}>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              What you watch
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Choose the metrics the cockpit should prioritise for you.
            </Typography>
            <FormGroup>
              {WATCHABLE.map((m) => (
                <FormControlLabel
                  key={m.key}
                  control={<Checkbox checked={watched.includes(m.key)} onChange={() => toggleWatched(m.key)} />}
                  label={m.label}
                />
              ))}
            </FormGroup>
            {savePrefs.isSuccess ? <Alert severity="success">Preferences saved.</Alert> : null}
            <Button
              variant="contained"
              onClick={() => savePrefs.mutate({ watched_metrics: watched, thresholds: prefs.data?.thresholds ?? {} })}
              disabled={savePrefs.isPending || prefs.isLoading}
              sx={{ alignSelf: 'flex-start' }}
            >
              {savePrefs.isPending ? <ButtonLoadingDots /> : null}
              Save preferences
            </Button>
          </Stack>
        </CardContent>
      </Card>

      <Card variant="outlined">
        <CardContent>
          <Stack spacing={2}>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              Critical alerts
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Quiet by default: get a push notification only for things that genuinely need you — an
              imminent stock-out, a sharp revenue drop, or an approval waiting.
            </Typography>
            {push.error ? <Alert severity="error">{push.error}</Alert> : null}
            {!push.supported ? (
              <Alert severity="info">This browser doesn&apos;t support push notifications.</Alert>
            ) : !push.available ? (
              <Alert severity="info">Push notifications aren&apos;t configured on the server yet.</Alert>
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
        </CardContent>
      </Card>
    </Stack>
  )
}
