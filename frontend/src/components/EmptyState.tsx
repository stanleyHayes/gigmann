import Box from '@mui/material/Box'
import Paper from '@mui/material/Paper'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import type { SvgIconComponent } from '@mui/icons-material'
import InfoOutlined from '@mui/icons-material/InfoOutlined'
import type { ReactNode } from 'react'

type EmptyStateProps = {
  title: string
  description?: string
  icon?: SvgIconComponent
  actions?: ReactNode
  compact?: boolean
}

/** EmptyState gives quiet/blank states the same structured treatment as Aura. */
export function EmptyState({ title, description, icon: Icon = InfoOutlined, actions, compact = false }: EmptyStateProps) {
  return (
    <Paper
      variant="outlined"
      role="status"
      sx={{
        position: 'relative',
        overflow: 'hidden',
        p: compact ? 2.5 : { xs: 3, md: 5 },
        textAlign: 'center',
        bgcolor: 'background.paper',
      }}
    >
      <Icon
        aria-hidden="true"
        sx={{
          position: 'absolute',
          right: -18,
          top: 18,
          fontSize: compact ? 72 : 132,
          color: 'primary.main',
          opacity: 0.06,
          transform: 'rotate(8deg)',
        }}
      />
      <Stack spacing={1.5} sx={{ position: 'relative', alignItems: 'center' }}>
        <Box
          aria-hidden="true"
          sx={{
            display: 'grid',
            placeItems: 'center',
            width: compact ? 44 : 64,
            height: compact ? 44 : 64,
            borderRadius: 2,
            bgcolor: 'action.hover',
            color: 'primary.main',
            border: 1,
            borderColor: 'divider',
          }}
        >
          <Icon fontSize={compact ? 'small' : 'medium'} />
        </Box>
        <Typography variant={compact ? 'subtitle1' : 'h6'} component={compact ? 'h3' : 'h2'}>
          {title}
        </Typography>
        {description ? (
          <Typography variant="body2" color="text.secondary" sx={{ maxWidth: 460, lineHeight: 1.7 }}>
            {description}
          </Typography>
        ) : null}
        {actions ? (
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} sx={{ pt: 1, justifyContent: 'center' }}>
            {actions}
          </Stack>
        ) : null}
      </Stack>
    </Paper>
  )
}
