import Box from '@mui/material/Box'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import type { SvgIconComponent } from '@mui/icons-material'
import DashboardOutlined from '@mui/icons-material/DashboardOutlined'
import type { ReactNode } from 'react'

type PageHeaderProps = {
  title: string
  eyebrow?: string
  description?: string
  icon?: SvgIconComponent
  actions?: ReactNode
}

/** PageHeader gives every cockpit screen the same Aura-inspired hierarchy. */
export function PageHeader({ title, eyebrow, description, icon: Icon = DashboardOutlined, actions }: PageHeaderProps) {
  return (
    <Box
      component="header"
      sx={{
        position: 'relative',
        overflow: 'hidden',
        borderBottom: 1,
        borderColor: 'divider',
        pb: { xs: 2.5, md: 3 },
      }}
    >
      <Box
        aria-hidden="true"
        sx={{
          position: 'absolute',
          right: { xs: -36, md: 0 },
          top: '50%',
          transform: 'translateY(-50%) rotate(-8deg)',
          color: 'primary.main',
          opacity: 0.06,
          display: { xs: 'none', sm: 'block' },
        }}
      >
        <Icon sx={{ fontSize: { sm: 128, md: 168 } }} />
      </Box>
      <Stack
        direction={{ xs: 'column', md: 'row' }}
        spacing={2}
        sx={{ position: 'relative', alignItems: { xs: 'flex-start', md: 'center' }, justifyContent: 'space-between' }}
      >
        <Stack direction="row" spacing={2} sx={{ minWidth: 0, alignItems: 'flex-start' }}>
          <Box
            aria-hidden="true"
            sx={{
              display: 'grid',
              placeItems: 'center',
              width: 52,
              height: 52,
              flex: '0 0 auto',
              borderRadius: 2,
              bgcolor: 'primary.main',
              color: 'primary.contrastText',
              boxShadow: (theme) => `0 14px 32px ${theme.palette.mode === 'dark' ? 'rgba(0,0,0,.28)' : 'rgba(11,92,173,.2)'}`,
            }}
          >
            <Icon fontSize="medium" />
          </Box>
          <Box sx={{ minWidth: 0 }}>
            {eyebrow ? (
              <Typography
                variant="overline"
                sx={{ color: 'text.secondary', fontWeight: 800, letterSpacing: 0, lineHeight: 1.4 }}
              >
                {eyebrow}
              </Typography>
            ) : null}
            <Typography variant="h1" sx={{ fontSize: { xs: '2rem', md: '2.6rem' }, lineHeight: 1.05 }}>
              {title}
            </Typography>
            {description ? (
              <Typography variant="body2" color="text.secondary" sx={{ mt: 1, maxWidth: 760, lineHeight: 1.7 }}>
                {description}
              </Typography>
            ) : null}
          </Box>
        </Stack>
        {actions ? (
          <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap', justifyContent: { xs: 'flex-start', md: 'flex-end' } }}>
            {actions}
          </Stack>
        ) : null}
      </Stack>
    </Box>
  )
}
