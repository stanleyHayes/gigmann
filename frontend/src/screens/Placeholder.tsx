import Alert from '@mui/material/Alert'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'

/** Placeholder is a lightweight stand-in for screens planned in later stories. */
export function Placeholder({ title, note }: { title: string; note?: string }) {
  return (
    <Stack spacing={2}>
      <Typography variant="h1" sx={{ fontSize: { xs: '2rem', md: '2.5rem' } }}>
        {title}
      </Typography>
      <Alert severity="info">{note ?? 'This screen is planned in a later story.'}</Alert>
    </Stack>
  )
}
