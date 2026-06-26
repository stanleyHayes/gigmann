import { useState } from 'react'
import Alert from '@mui/material/Alert'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Stack from '@mui/material/Stack'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'

import { useMfaConfirm, useMfaEnroll } from '../api/useMfa'
import { ButtonLoadingDots } from '../components/ButtonLoadingDots'
import { monoFont } from '../theme'

/** SettingsScreen lets the user opt into two-factor authentication. */
export function SettingsScreen() {
  const enroll = useMfaEnroll()
  const confirm = useMfaConfirm()
  const [code, setCode] = useState('')
  const secret = enroll.data?.secret

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
    </Stack>
  )
}
