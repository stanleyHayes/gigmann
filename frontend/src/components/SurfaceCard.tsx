import Box from '@mui/material/Box'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import type { SxProps, Theme } from '@mui/material/styles'
import type { SvgIconComponent } from '@mui/icons-material'
import type { ReactNode } from 'react'

type SurfaceCardProps = {
  title?: string
  description?: string
  icon?: SvgIconComponent
  actions?: ReactNode
  children: ReactNode
  sx?: SxProps<Theme>
}

/** SurfaceCard mirrors Aura's card sections while staying in MUI/Gigmann tokens. */
export function SurfaceCard({ title, description, icon: Icon, actions, children, sx }: SurfaceCardProps) {
  return (
    <Card variant="outlined" sx={sx}>
      {(title || description || actions) ? (
        <Box
          sx={{
            px: { xs: 2, md: 2.5 },
            pt: { xs: 2, md: 2.5 },
            pb: 1.5,
            borderBottom: 1,
            borderColor: 'divider',
          }}
        >
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} sx={{ justifyContent: 'space-between', gap: 1.5 }}>
            <Stack direction="row" spacing={1.5} sx={{ minWidth: 0, alignItems: 'flex-start' }}>
              {Icon ? (
                <Box
                  aria-hidden="true"
                  sx={{
                    display: 'grid',
                    placeItems: 'center',
                    width: 40,
                    height: 40,
                    borderRadius: 2,
                    bgcolor: 'action.hover',
                    color: 'primary.main',
                    border: 1,
                    borderColor: 'divider',
                    flex: '0 0 auto',
                  }}
                >
                  <Icon fontSize="small" />
                </Box>
              ) : null}
              <Box sx={{ minWidth: 0 }}>
                {title ? <Typography variant="h6" component="h2">{title}</Typography> : null}
                {description ? (
                  <Typography variant="body2" color="text.secondary" sx={{ mt: title ? 0.25 : 0, lineHeight: 1.65 }}>
                    {description}
                  </Typography>
                ) : null}
              </Box>
            </Stack>
            {actions ? (
              <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap', justifyContent: { xs: 'flex-start', sm: 'flex-end' } }}>
                {actions}
              </Stack>
            ) : null}
          </Stack>
        </Box>
      ) : null}
      <CardContent sx={{ p: { xs: 2, md: 2.5 }, '&:last-child': { pb: { xs: 2, md: 2.5 } } }}>
        {children}
      </CardContent>
    </Card>
  )
}
