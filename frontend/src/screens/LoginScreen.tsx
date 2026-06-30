import { useState, type FormEvent } from 'react'
import ArrowBackOutlined from '@mui/icons-material/ArrowBackOutlined'
import CheckCircleOutlined from '@mui/icons-material/CheckCircleOutlined'
import KeyOutlined from '@mui/icons-material/KeyOutlined'
import LockResetOutlined from '@mui/icons-material/LockResetOutlined'
import MonitorHeartOutlined from '@mui/icons-material/MonitorHeartOutlined'
import SecurityOutlined from '@mui/icons-material/SecurityOutlined'
import ShieldOutlined from '@mui/icons-material/ShieldOutlined'
import VisibilityOffOutlined from '@mui/icons-material/VisibilityOffOutlined'
import VisibilityOutlined from '@mui/icons-material/VisibilityOutlined'
import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Chip from '@mui/material/Chip'
import Divider from '@mui/material/Divider'
import IconButton from '@mui/material/IconButton'
import InputAdornment from '@mui/material/InputAdornment'
import MuiLink from '@mui/material/Link'
import Paper from '@mui/material/Paper'
import Stack from '@mui/material/Stack'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'

import { usePasswordResetConfirm, usePasswordResetRequest } from '../api/usePasswordReset'
import { useAuth } from '../auth/authContext'
import { ButtonLoadingDots } from '../components/ButtonLoadingDots'

type AuthMode = 'login' | 'reset-request' | 'reset-confirm'

const trustSignals = [
  { label: 'Computed figures', icon: MonitorHeartOutlined },
  { label: 'MFA ready', icon: SecurityOutlined },
  { label: 'Grounded AI', icon: CheckCircleOutlined },
]

function resetErrorMessage(error: unknown) {
  const code = (error as { error?: string } | null)?.error
  if (code === 'weak_password') {
    return 'Use at least 12 characters with a mix of letters, numbers, or symbols.'
  }
  if (code === 'invalid_reset_token') {
    return 'That reset token is invalid or has expired. Request a new one.'
  }
  return 'We could not complete the reset. Try again.'
}

/** LoginScreen gates the cockpit until the user signs in. */
export function LoginScreen() {
  const { login, loginPending, loginError, mfaRequired } = useAuth()
  const resetRequest = usePasswordResetRequest()
  const resetConfirm = usePasswordResetConfirm()
  const [mode, setMode] = useState<AuthMode>('login')
  const [email, setEmail] = useState('ceo@gigmann.health')
  const [password, setPassword] = useState('')
  const [code, setCode] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [resetEmail, setResetEmail] = useState('ceo@gigmann.health')
  const [resetToken, setResetToken] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [showNewPassword, setShowNewPassword] = useState(false)
  const [resetComplete, setResetComplete] = useState(false)

  const onSubmit = (e: FormEvent) => {
    e.preventDefault()
    setResetComplete(false)
    login(email, password, mfaRequired ? code : undefined)
  }

  const onResetRequest = (e: FormEvent) => {
    e.preventDefault()
    setResetComplete(false)
    resetRequest.mutate(resetEmail, {
      onSuccess: (result) => {
        setResetToken(result.reset_token ?? '')
        setMode('reset-confirm')
      },
    })
  }

  const onResetConfirm = (e: FormEvent) => {
    e.preventDefault()
    resetConfirm.mutate(
      { token: resetToken, password: newPassword },
      {
        onSuccess: () => {
          setEmail(resetEmail)
          setPassword('')
          setCode('')
          setNewPassword('')
          setResetComplete(true)
          setMode('login')
        },
      },
    )
  }

  const goToReset = () => {
    setResetEmail(email)
    setResetComplete(false)
    resetRequest.reset()
    resetConfirm.reset()
    setMode('reset-request')
  }

  const goToLogin = () => {
    resetRequest.reset()
    resetConfirm.reset()
    setMode('login')
  }

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'grid',
        alignItems: 'stretch',
        bgcolor: 'background.default',
        backgroundImage: (theme) =>
          `linear-gradient(135deg, ${theme.palette.primary.main}14 0%, transparent 34%), radial-gradient(circle at 84% 12%, ${theme.palette.warning.main}1f, transparent 28%), radial-gradient(circle at 12% 78%, ${theme.palette.success.main}1c, transparent 30%)`,
      }}
    >
      <Box
        sx={{
          width: '100%',
          maxWidth: 1180,
          mx: 'auto',
          px: { xs: 2, md: 4 },
          py: { xs: 3, md: 5 },
          display: 'grid',
          gridTemplateColumns: { xs: '1fr', lg: 'minmax(0, 1.05fr) minmax(440px, 0.95fr)' },
          gap: { xs: 2, md: 3 },
          alignItems: 'stretch',
        }}
      >
        <Paper
          elevation={0}
          sx={{
            display: { xs: 'none', lg: 'flex' },
            minHeight: 640,
            p: 4,
            border: '1px solid',
            borderColor: 'divider',
            borderRadius: 4,
            color: 'primary.contrastText',
            bgcolor: 'primary.main',
            overflow: 'hidden',
            position: 'relative',
          }}
        >
          <Box
            aria-hidden="true"
            sx={{
              position: 'absolute',
              inset: 'auto -18% -20% auto',
              width: 360,
              height: 360,
              borderRadius: '50%',
              bgcolor: 'rgba(255,255,255,0.09)',
            }}
          />
          <Stack spacing={4} sx={{ width: '100%', position: 'relative', zIndex: 1 }}>
            <Stack direction="row" spacing={1.5} sx={{ alignItems: 'center' }}>
              <Box
                aria-hidden="true"
                sx={{
                  display: 'grid',
                  placeItems: 'center',
                  width: 46,
                  height: 46,
                  borderRadius: 2,
                  bgcolor: 'rgba(255,255,255,0.14)',
                  border: '1px solid rgba(255,255,255,0.2)',
                }}
              >
                <ShieldOutlined />
              </Box>
              <Stack spacing={0}>
                <Typography variant="overline" sx={{ color: 'rgba(255,255,255,0.72)', fontWeight: 800, letterSpacing: 0 }}>
                  Gigmann Executive Cockpit
                </Typography>
                <Typography variant="h4" sx={{ fontWeight: 900 }}>
                  Ahenfie
                </Typography>
              </Stack>
            </Stack>

            <Stack spacing={2.5} sx={{ maxWidth: 540, mt: 'auto' }}>
              <Typography variant="h1" sx={{ fontSize: 'clamp(3rem, 5vw, 4.9rem)', lineHeight: 0.94, color: 'inherit' }}>
                Secure command starts at the door.
              </Typography>
              <Typography variant="body1" sx={{ maxWidth: 500, color: 'rgba(255,255,255,0.78)', lineHeight: 1.8 }}>
                Sign in to the daily brief, network pulse, approvals, and grounded Ask workspace. Every figure is computed before AI gets to narrate it.
              </Typography>
            </Stack>

            <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(3, minmax(0, 1fr))', gap: 1.5, mt: 'auto' }}>
              {trustSignals.map(({ label, icon: Icon }) => (
                <Box
                  key={label}
                  sx={{
                    p: 1.5,
                    borderRadius: 2,
                    border: '1px solid rgba(255,255,255,0.2)',
                    bgcolor: 'rgba(255,255,255,0.1)',
                  }}
                >
                  <Icon sx={{ fontSize: 20, mb: 0.75 }} />
                  <Typography variant="caption" sx={{ display: 'block', color: 'rgba(255,255,255,0.86)', fontWeight: 800 }}>
                    {label}
                  </Typography>
                </Box>
              ))}
            </Box>
          </Stack>
        </Paper>

        <Paper
          elevation={0}
          sx={{
            alignSelf: 'center',
            borderRadius: 4,
            border: '1px solid',
            borderColor: 'divider',
            boxShadow: '0 24px 80px rgba(15, 23, 42, 0.13)',
            overflow: 'hidden',
            bgcolor: 'background.paper',
          }}
        >
          <Box sx={{ p: { xs: 3, sm: 4 } }}>
            <Stack spacing={3}>
              <Stack spacing={1}>
                <Chip
                  icon={mode === 'login' ? <KeyOutlined /> : <LockResetOutlined />}
                  label={mode === 'login' ? 'Secure access' : 'Password recovery'}
                  color="primary"
                  variant="outlined"
                  sx={{ alignSelf: 'flex-start', fontWeight: 800 }}
                />
                <Typography variant="h1" sx={{ fontSize: { xs: '2.35rem', sm: '3rem' }, lineHeight: 1 }}>
                  {mode === 'login' ? 'Welcome back.' : mode === 'reset-request' ? 'Reset your password.' : 'Create a new password.'}
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ maxWidth: 420 }}>
                  {mode === 'login'
                    ? 'Enter your executive credentials. If MFA is enabled, you will confirm with your authenticator or recovery code.'
                    : mode === 'reset-request'
                      ? 'Enter your account email and we will prepare a short-lived reset token.'
                      : 'Paste the reset token and choose a new password. MFA remains active after reset.'}
                </Typography>
              </Stack>

              {mode === 'login' ? (
                <Stack component="form" spacing={2.25} onSubmit={onSubmit}>
                  {resetComplete ? <Alert severity="success">Password reset. Sign in with the new password.</Alert> : null}
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
                    type={showPassword ? 'text' : 'password'}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    autoComplete="current-password"
                    required
                    slotProps={{
                      htmlInput: { 'aria-label': 'Password' },
                      input: {
                        endAdornment: (
                        <InputAdornment position="end">
                          <IconButton
                            aria-label={showPassword ? 'Hide sign-in password' : 'Show sign-in password'}
                            edge="end"
                            onClick={() => setShowPassword((value) => !value)}
                          >
                            {showPassword ? <VisibilityOffOutlined /> : <VisibilityOutlined />}
                          </IconButton>
                        </InputAdornment>
                        ),
                      },
                    }}
                  />
                  <Stack direction="row" spacing={2} sx={{ alignItems: 'center', justifyContent: 'space-between' }}>
                    <Typography variant="caption" color="text.secondary">
                      Demo: ceo@gigmann.health
                    </Typography>
                    <Button variant="text" size="small" onClick={goToReset}>
                      Forgot password?
                    </Button>
                  </Stack>
                  {mfaRequired ? (
                    <TextField
                      label="Authenticator or recovery code"
                      value={code}
                      onChange={(e) => setCode(e.target.value)}
                      autoComplete="one-time-code"
                      inputMode="numeric"
                      helperText="Enter the 6-digit code from your authenticator app or a one-time recovery code."
                      autoFocus
                      slotProps={{ htmlInput: { 'aria-label': 'Authenticator or recovery code' } }}
                    />
                  ) : null}
                  {loginError ? <Alert severity="error">{loginError}</Alert> : null}
                  <Button type="submit" variant="contained" size="large" disabled={loginPending} sx={{ py: 1.35 }}>
                    {loginPending ? <ButtonLoadingDots /> : null}
                    {mfaRequired ? 'Verify and enter' : 'Sign in'}
                  </Button>
                </Stack>
              ) : null}

              {mode === 'reset-request' ? (
                <Stack component="form" spacing={2.25} onSubmit={onResetRequest}>
                  <TextField
                    label="Account email"
                    type="email"
                    value={resetEmail}
                    onChange={(e) => setResetEmail(e.target.value)}
                    autoComplete="username"
                    required
                  />
                  {resetRequest.isError ? <Alert severity="error">We could not start the reset. Try again.</Alert> : null}
                  <Button type="submit" variant="contained" size="large" disabled={resetRequest.isPending} sx={{ py: 1.35 }}>
                    {resetRequest.isPending ? <ButtonLoadingDots /> : null}
                    Send reset instructions
                  </Button>
                  <Button variant="text" startIcon={<ArrowBackOutlined />} onClick={goToLogin}>
                    Back to sign in
                  </Button>
                </Stack>
              ) : null}

              {mode === 'reset-confirm' ? (
                <Stack component="form" spacing={2.25} onSubmit={onResetConfirm}>
                  <Alert severity={resetToken ? 'info' : 'success'}>
                    {resetToken
                      ? 'Demo delivery: the reset token is filled below. In production this would arrive by email or SMS.'
                      : resetRequest.data?.message}
                  </Alert>
                  <TextField
                    label="Reset token"
                    value={resetToken}
                    onChange={(e) => setResetToken(e.target.value)}
                    autoComplete="one-time-code"
                    required
                    slotProps={{ htmlInput: { 'aria-label': 'Reset token' } }}
                  />
                  <TextField
                    label="New password"
                    type={showNewPassword ? 'text' : 'password'}
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    autoComplete="new-password"
                    helperText="Use at least 12 characters with a mix of letters, numbers, or symbols."
                    required
                    slotProps={{
                      htmlInput: { 'aria-label': 'New password' },
                      input: {
                        endAdornment: (
                        <InputAdornment position="end">
                          <IconButton
                            aria-label={showNewPassword ? 'Hide new password' : 'Show new password'}
                            edge="end"
                            onClick={() => setShowNewPassword((value) => !value)}
                          >
                            {showNewPassword ? <VisibilityOffOutlined /> : <VisibilityOutlined />}
                          </IconButton>
                        </InputAdornment>
                        ),
                      },
                    }}
                  />
                  {resetConfirm.isError ? <Alert severity="error">{resetErrorMessage(resetConfirm.error)}</Alert> : null}
                  <Button type="submit" variant="contained" size="large" disabled={resetConfirm.isPending} sx={{ py: 1.35 }}>
                    {resetConfirm.isPending ? <ButtonLoadingDots /> : null}
                    Reset password
                  </Button>
                  <Stack direction="row" spacing={1.5} sx={{ flexWrap: 'wrap', justifyContent: 'space-between', rowGap: 1 }}>
                    <Button variant="text" startIcon={<ArrowBackOutlined />} onClick={() => setMode('reset-request')}>
                      Request again
                    </Button>
                    <Button variant="text" onClick={goToLogin}>
                      Back to sign in
                    </Button>
                  </Stack>
                </Stack>
              ) : null}

              <Divider />

              <Stack spacing={1} sx={{ alignItems: 'center', textAlign: 'center' }}>
                <Typography variant="caption" color="text.secondary" sx={{ maxWidth: 360 }}>
                  Protected by short-lived sessions, MFA, and role-scoped access.
                </Typography>
                <MuiLink href="/welcome.html" variant="caption" sx={{ fontWeight: 800 }}>
                  Learn more about Ahenfie
                </MuiLink>
              </Stack>
            </Stack>
          </Box>
        </Paper>
      </Box>
    </Box>
  )
}
