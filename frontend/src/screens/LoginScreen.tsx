import { useState, type FormEvent } from 'react'
import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import MuiLink from '@mui/material/Link'
import Stack from '@mui/material/Stack'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'

import { useAuth } from '../auth/authContext'
import { ButtonLoadingDots } from '../components/ButtonLoadingDots'

/** LoginScreen gates the cockpit until the user signs in. */
export function LoginScreen() {
  const { login, loginPending, loginError, mfaRequired } = useAuth()
  const [email, setEmail] = useState('ceo@gigmann.health')
  const [password, setPassword] = useState('')
  const [code, setCode] = useState('')

  const onSubmit = (e: FormEvent) => {
    e.preventDefault()
    login(email, password, mfaRequired ? code : undefined)
  }

  return (
    <Box sx={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', p: 2 }}>
      <Card variant="outlined" sx={{ width: '100%', maxWidth: 380 }}>
        <CardContent>
          <Stack component="form" spacing={3} onSubmit={onSubmit}>
            <Stack spacing={0.5}>
              <Typography variant="h1" sx={{ fontSize: '2rem' }}>
                Ahenfie
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Gigmann Executive Cockpit
              </Typography>
              <MuiLink href="/welcome.html" variant="caption" sx={{ alignSelf: 'flex-start' }}>
                Learn more about Ahenfie →
              </MuiLink>
            </Stack>
            <TextField
              label="Email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoComplete="username"
              required
            />
            <TextField
              label="Password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="current-password"
              required
            />
            {mfaRequired ? (
              <TextField
                label="Authenticator code"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                autoComplete="one-time-code"
                inputMode="numeric"
                helperText="Enter the 6-digit code from your authenticator app."
                autoFocus
              />
            ) : null}
            {loginError ? <Alert severity="error">{loginError}</Alert> : null}
            <Button type="submit" variant="contained" size="large" disabled={loginPending}>
              {loginPending ? <ButtonLoadingDots /> : null}
              {mfaRequired ? 'Verify' : 'Sign in'}
            </Button>
          </Stack>
        </CardContent>
      </Card>
    </Box>
  )
}
