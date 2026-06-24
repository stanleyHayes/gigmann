import Box from '@mui/material/Box'
import { keyframes } from '@mui/system'

// Owner directive: button loading state uses animated dots (not a spinner).
const bounce = keyframes`
  0%, 80%, 100% { opacity: 0.2; transform: translateY(0); }
  40% { opacity: 1; transform: translateY(-3px); }
`

export function ButtonLoadingDots({ size = 6 }: { size?: number }) {
  return (
    <Box
      component="span"
      role="status"
      aria-label="loading"
      sx={{ display: 'inline-flex', gap: 0.5, mr: 1, alignItems: 'center' }}
    >
      {[0, 1, 2].map((i) => (
        <Box
          key={i}
          component="span"
          sx={{
            width: size,
            height: size,
            borderRadius: '50%',
            bgcolor: 'currentColor',
            animation: `${bounce} 1.2s ${i * 0.16}s infinite ease-in-out`,
            '@media (prefers-reduced-motion: reduce)': { animation: 'none', opacity: 0.6 },
          }}
        />
      ))}
    </Box>
  )
}
