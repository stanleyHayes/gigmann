import AppBar from '@mui/material/AppBar'
import Avatar from '@mui/material/Avatar'
import Badge from '@mui/material/Badge'
import BottomNavigation from '@mui/material/BottomNavigation'
import BottomNavigationAction from '@mui/material/BottomNavigationAction'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Collapse from '@mui/material/Collapse'
import Container from '@mui/material/Container'
import Divider from '@mui/material/Divider'
import Dialog from '@mui/material/Dialog'
import DialogActions from '@mui/material/DialogActions'
import DialogContent from '@mui/material/DialogContent'
import DialogTitle from '@mui/material/DialogTitle'
import Drawer from '@mui/material/Drawer'
import IconButton from '@mui/material/IconButton'
import LinearProgress from '@mui/material/LinearProgress'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemButton from '@mui/material/ListItemButton'
import ListItemIcon from '@mui/material/ListItemIcon'
import ListItemText from '@mui/material/ListItemText'
import Menu from '@mui/material/Menu'
import MenuItem from '@mui/material/MenuItem'
import Popover from '@mui/material/Popover'
import Skeleton from '@mui/material/Skeleton'
import Stack from '@mui/material/Stack'
import Step from '@mui/material/Step'
import StepLabel from '@mui/material/StepLabel'
import Stepper from '@mui/material/Stepper'
import Toolbar from '@mui/material/Toolbar'
import Tooltip from '@mui/material/Tooltip'
import Typography from '@mui/material/Typography'
import AssignmentIndOutlined from '@mui/icons-material/AssignmentIndOutlined'
import ChecklistOutlined from '@mui/icons-material/ChecklistOutlined'
import ChevronLeftOutlined from '@mui/icons-material/ChevronLeftOutlined'
import ChevronRightOutlined from '@mui/icons-material/ChevronRightOutlined'
import DarkModeOutlined from '@mui/icons-material/DarkModeOutlined'
import DoneAllOutlined from '@mui/icons-material/DoneAllOutlined'
import ExpandMoreOutlined from '@mui/icons-material/ExpandMoreOutlined'
import ForumOutlined from '@mui/icons-material/ForumOutlined'
import HelpOutlineOutlined from '@mui/icons-material/HelpOutlineOutlined'
import HubOutlined from '@mui/icons-material/HubOutlined'
import InsightsOutlined from '@mui/icons-material/InsightsOutlined'
import SummarizeOutlined from '@mui/icons-material/SummarizeOutlined'
import LightModeOutlined from '@mui/icons-material/LightModeOutlined'
import LogoutOutlined from '@mui/icons-material/LogoutOutlined'
import MenuBookOutlined from '@mui/icons-material/MenuBookOutlined'
import MenuOutlined from '@mui/icons-material/MenuOutlined'
import NotificationsActiveOutlined from '@mui/icons-material/NotificationsActiveOutlined'
import PersonOutlineOutlined from '@mui/icons-material/PersonOutlineOutlined'
import PlayCircleOutlineOutlined from '@mui/icons-material/PlayCircleOutlineOutlined'
import RocketLaunchOutlined from '@mui/icons-material/RocketLaunchOutlined'
import SettingsOutlined from '@mui/icons-material/SettingsOutlined'
import TaskAltOutlined from '@mui/icons-material/TaskAltOutlined'
import TodayOutlined from '@mui/icons-material/TodayOutlined'
import VisibilityOffOutlined from '@mui/icons-material/VisibilityOffOutlined'
import { NavLink, Outlet, useLocation, useNavigate, useNavigation } from 'react-router-dom'
import { motion, useReducedMotion } from 'framer-motion'
import { flushSync } from 'react-dom'
import { useCallback, useEffect, useMemo, useState, type MouseEvent, type ReactNode } from 'react'

import { useAlerts, useUpdateAlertStatus, type AlertItem } from '../api/useAlerts'
import { FacilitySearch } from '../components/FacilitySearch'
import { StatusChip, type FacilityStatus } from '../components/StatusChip'
import { useLiveUpdates } from '../api/useLiveUpdates'
import { useAuth } from '../auth/authContext'
import { useColorMode } from './colorMode'
import { OPEN_HELP_EVENT, REPLAY_TOUR_EVENT } from './helpEvents'
import { t } from '../i18n/messages'

type NavItem = { to: string; label: string; icon: ReactNode; end?: boolean }
type NavSection = { heading: string; items: NavItem[] }

const EXPANDED_DRAWER_WIDTH = 296
const COLLAPSED_DRAWER_WIDTH = 84
const SIDEBAR_KEY = 'gigmann-sidebar-collapsed'
const NAV_GROUPS_KEY = 'gigmann-nav-groups'

const NAV_SECTIONS: NavSection[] = [
  {
    heading: 'Command',
    items: [
      { to: '/', label: t('nav.today'), icon: <TodayOutlined />, end: true },
      { to: '/network', label: t('nav.network'), icon: <HubOutlined /> },
      { to: '/kpis', label: t('nav.kpis'), icon: <InsightsOutlined /> },
      { to: '/reports', label: t('nav.reports'), icon: <SummarizeOutlined /> },
      { to: '/ask', label: t('nav.ask'), icon: <ForumOutlined /> },
    ],
  },
  {
    heading: 'Execution',
    items: [
      { to: '/my-day', label: t('nav.myDay'), icon: <ChecklistOutlined /> },
      { to: '/delegation', label: t('nav.delegation'), icon: <AssignmentIndOutlined /> },
      { to: '/approvals', label: t('nav.approvals'), icon: <TaskAltOutlined /> },
    ],
  },
]

const SETTINGS_ITEM: NavItem = { to: '/settings', label: 'Settings', icon: <SettingsOutlined /> }
const NAV_ITEMS = [...NAV_SECTIONS.flatMap((section) => section.items), SETTINGS_ITEM]
const MOBILE_ITEMS: NavItem[] = [
  { to: '/', label: t('nav.today'), icon: <TodayOutlined />, end: true },
  { to: '/network', label: t('nav.network'), icon: <HubOutlined /> },
  { to: '/ask', label: t('nav.ask'), icon: <ForumOutlined /> },
  { to: '/my-day', label: t('nav.myDay'), icon: <ChecklistOutlined /> },
  { to: '/approvals', label: t('nav.approvals'), icon: <TaskAltOutlined /> },
]

const TOUR_STEPS = [
  {
    label: 'Today',
    route: '/',
    title: 'Start with the Brief',
    body: 'This is the morning command surface: narrated summary, worst-first issues, exports, and the attention feed.',
  },
  {
    label: 'Network',
    route: '/network',
    title: 'Scan the whole network',
    body: 'Use search, status, region, and sort controls to find the facilities that need a decision first.',
  },
  {
    label: 'Ask',
    route: '/ask',
    title: 'Ask and draft',
    body: 'Ask grounded questions, then generate unsent manager messages or summaries from the same computed context.',
  },
  {
    label: 'My Day',
    route: '/my-day',
    title: 'Turn signals into work',
    body: 'Brief items and alerts can become tasks here, so follow-through stays tied to the source issue.',
  },
  {
    label: 'Settings',
    route: '/settings',
    title: 'Tune the cockpit',
    body: 'Manage MFA, watched metrics, thresholds, critical alerts, theme presets, and the guide controls.',
  },
] as const

const HELP_LINKS = [
  { title: 'Daily command routine', body: 'Read the Brief, clear the attention feed, then move follow-ups into My Day.', route: '/' },
  { title: 'Investigate a facility', body: 'Open Network, filter to critical or watch, then drill into the facility card.', route: '/network' },
  { title: 'Generate a manager draft', body: 'Use Ask or Facility Detail to generate a draft, then copy it after review.', route: '/ask' },
  { title: 'Change account controls', body: 'Open Settings for MFA, watched metrics, alert devices, and appearance.', route: '/settings' },
] as const

function readSidebarCollapsed(): boolean {
  if (typeof window === 'undefined') {
    return false
  }
  return window.localStorage.getItem(SIDEBAR_KEY) === 'true'
}

function readNavGroups(): Record<string, boolean> {
  if (typeof window === 'undefined') {
    return {}
  }
  try {
    return JSON.parse(window.localStorage.getItem(NAV_GROUPS_KEY) ?? '{}') as Record<string, boolean>
  } catch {
    return {}
  }
}

function writeNavGroups(groups: Record<string, boolean>) {
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(NAV_GROUPS_KEY, JSON.stringify(groups))
  }
}

function itemSx(collapsed: boolean) {
  return {
    mx: collapsed ? 1 : 1.25,
    my: 0.25,
    minHeight: 44,
    justifyContent: collapsed ? 'center' : 'flex-start',
    borderRadius: 2,
    color: 'text.secondary',
    '& .MuiListItemIcon-root': {
      minWidth: collapsed ? 0 : 38,
      color: 'inherit',
    },
    '&.active': {
      bgcolor: 'action.selected',
      color: 'primary.main',
      boxShadow: collapsed ? 'inset 0 -3px 0 currentColor' : 'inset 3px 0 0 currentColor',
    },
    '&:hover': {
      bgcolor: 'action.hover',
      color: 'text.primary',
    },
  } as const
}

function NavButton({ item, collapsed, onNavigate }: { item: NavItem; collapsed: boolean; onNavigate?: () => void }) {
  const button = (
    <ListItemButton component={NavLink} to={item.to} end={item.end} onClick={onNavigate} sx={itemSx(collapsed)}>
      <ListItemIcon>{item.icon}</ListItemIcon>
      {!collapsed ? (
        <ListItemText
          primary={
            <Typography component="span" variant="body2" sx={{ fontSize: 14, fontWeight: 700 }}>
              {item.label}
            </Typography>
          }
        />
      ) : null}
    </ListItemButton>
  )

  return (
    <ListItem disablePadding>
      {collapsed ? (
        <Tooltip title={item.label} placement="right">
          {button}
        </Tooltip>
      ) : (
        button
      )}
    </ListItem>
  )
}

function NavList({ collapsed, onNavigate }: { collapsed: boolean; onNavigate?: () => void }) {
  const [groups, setGroups] = useState<Record<string, boolean>>(() => readNavGroups())

  const toggleGroup = (heading: string) => {
    setGroups((current) => {
      const next = { ...current, [heading]: !(current[heading] ?? true) }
      writeNavGroups(next)
      return next
    })
  }

  if (collapsed) {
    return (
      <Box component="nav" aria-label="Primary navigation">
        <List dense>
          {NAV_SECTIONS.flatMap((section) => section.items).map((item) => (
            <NavButton key={item.to} item={item} collapsed onNavigate={onNavigate} />
          ))}
        </List>
      </Box>
    )
  }

  return (
    <Box component="nav" aria-label="Primary navigation">
      {NAV_SECTIONS.map((section) => {
        const open = groups[section.heading] ?? true
        return (
          <List key={section.heading} dense disablePadding sx={{ mb: 1 }}>
            <Button
              fullWidth
              color="inherit"
              onClick={() => toggleGroup(section.heading)}
              aria-expanded={open}
              endIcon={
                <ExpandMoreOutlined
                  fontSize="small"
                  sx={{ transform: open ? 'rotate(0deg)' : 'rotate(-90deg)', transition: 'transform .16s ease' }}
                />
              }
              sx={{
                justifyContent: 'space-between',
                px: 2.5,
                py: 0.75,
                color: 'text.secondary',
                fontSize: 11,
                fontWeight: 800,
                letterSpacing: 0,
                textTransform: 'uppercase',
              }}
            >
              {section.heading}
            </Button>
            <Collapse in={open} timeout="auto" unmountOnExit>
              {section.items.map((item) => (
                <NavButton key={item.to} item={item} collapsed={false} onNavigate={onNavigate} />
              ))}
            </Collapse>
          </List>
        )
      })}
    </Box>
  )
}

function DrawerContent({
  collapsed,
  onToggleCollapse,
  onNavigate,
}: {
  collapsed: boolean
  onToggleCollapse?: () => void
  onNavigate?: () => void
}) {
  return (
    <Stack sx={{ height: '100%' }}>
      <Stack spacing={2} sx={{ p: collapsed ? 1.5 : 2.5, borderBottom: 1, borderColor: 'divider' }}>
        <Stack
          direction="row"
          spacing={1.5}
          sx={{ alignItems: 'center', justifyContent: collapsed ? 'center' : 'flex-start' }}
        >
          <Box
            aria-hidden="true"
            sx={{
              display: 'grid',
              width: 44,
              height: 44,
              placeItems: 'center',
              borderRadius: 2,
              bgcolor: 'primary.main',
              color: 'primary.contrastText',
              fontFamily: '"Fraunces Variable", serif',
              fontSize: 24,
              fontWeight: 700,
              flex: '0 0 auto',
            }}
          >
            A
          </Box>
          {!collapsed ? (
            <Box sx={{ minWidth: 0 }}>
              <Typography variant="h6" sx={{ lineHeight: 1.1 }}>
                Ahenfie
              </Typography>
              <Typography variant="caption" color="text.secondary">
                Gigmann Executive Cockpit
              </Typography>
            </Box>
          ) : null}
        </Stack>
        {!collapsed ? (
          <Box
            sx={{
              border: 1,
              borderColor: 'divider',
              borderRadius: 2,
              bgcolor: 'background.default',
              px: 1.5,
              py: 1.25,
            }}
          >
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', fontWeight: 700 }}>
              Network posture
            </Typography>
            <Typography variant="body2" sx={{ fontWeight: 700 }}>
              Computed figures only
            </Typography>
          </Box>
        ) : null}
      </Stack>

      <Box sx={{ flex: 1, overflowY: 'auto', py: 2 }}>
        <NavList collapsed={collapsed} onNavigate={onNavigate} />
      </Box>

      <Box sx={{ p: 1.25, borderTop: 1, borderColor: 'divider' }}>
        <List dense>
          <NavButton item={SETTINGS_ITEM} collapsed={collapsed} onNavigate={onNavigate} />
        </List>
        {onToggleCollapse ? (
          <>
            <Divider sx={{ my: 1 }} />
            <Tooltip title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'} placement="right">
              <IconButton
                onClick={onToggleCollapse}
                aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
                sx={{ width: '100%', borderRadius: 2 }}
              >
                {collapsed ? <ChevronRightOutlined /> : <ChevronLeftOutlined />}
              </IconButton>
            </Tooltip>
          </>
        ) : null}
      </Box>
    </Stack>
  )
}

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

function AccountMenu({
  userName,
  role,
  onLogout,
  onOpenHelp,
  onReplayTour,
}: {
  userName: string
  role: string
  onLogout: () => void
  onOpenHelp: () => void
  onReplayTour: () => void
}) {
  const navigate = useNavigate()
  const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null)
  const open = Boolean(anchorEl)
  const close = () => setAnchorEl(null)
  const go = (path: string) => {
    navigate(path)
    close()
  }

  return (
    <>
      <Button
        color="inherit"
        onClick={(e) => setAnchorEl(e.currentTarget)}
        aria-controls={open ? 'account-menu' : undefined}
        aria-haspopup="menu"
        aria-expanded={open}
        endIcon={<ExpandMoreOutlined fontSize="small" />}
        sx={{
          minWidth: 0,
          px: { xs: 0.75, sm: 1 },
          py: 0.5,
          border: 1,
          borderColor: 'divider',
          bgcolor: 'background.paper',
        }}
      >
        <Avatar sx={{ width: 30, height: 30, bgcolor: 'primary.main', color: 'primary.contrastText', fontSize: 13 }}>
          {initials(userName)}
        </Avatar>
        <Box sx={{ display: { xs: 'none', lg: 'block' }, ml: 1, textAlign: 'left' }}>
          <Typography component="span" variant="body2" sx={{ display: 'block', fontWeight: 800, lineHeight: 1.2 }}>
            {userName}
          </Typography>
          <Typography component="span" variant="caption" color="text.secondary" sx={{ display: 'block', lineHeight: 1.2 }}>
            {role}
          </Typography>
        </Box>
      </Button>
      <Menu
        id="account-menu"
        anchorEl={anchorEl}
        open={open}
        onClose={close}
        slotProps={{ paper: { sx: { width: 320, maxWidth: 'calc(100vw - 32px)', mt: 1 } } }}
      >
        <Box sx={{ px: 2, py: 1.5 }}>
          <Typography variant="subtitle2" sx={{ fontWeight: 800 }}>
            {userName}
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ textTransform: 'capitalize' }}>
            {role}
          </Typography>
        </Box>
        <Divider />
        <MenuItem onClick={() => go('/settings')}>
          <ListItemIcon>
            <PersonOutlineOutlined fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Profile and settings" secondary="Security, preferences, device alerts" />
        </MenuItem>
        <MenuItem onClick={() => go('/my-day')}>
          <ListItemIcon>
            <ChecklistOutlined fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="My Day" secondary="Tasks created from briefs and alerts" />
        </MenuItem>
        <MenuItem onClick={() => go('/ask')}>
          <ListItemIcon>
            <ForumOutlined fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Ask the cockpit" secondary="Ground a question in current figures" />
        </MenuItem>
        <MenuItem
          onClick={() => {
            close()
            onOpenHelp()
          }}
        >
          <ListItemIcon>
            <MenuBookOutlined fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="User guide" secondary="Workflow reference and help" />
        </MenuItem>
        <MenuItem
          onClick={() => {
            close()
            onReplayTour()
          }}
        >
          <ListItemIcon>
            <PlayCircleOutlineOutlined fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Replay tour" secondary="Step through the cockpit again" />
        </MenuItem>
        <Divider />
        <MenuItem
          onClick={() => {
            close()
            onLogout()
          }}
        >
          <ListItemIcon>
            <LogoutOutlined fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Sign out" />
        </MenuItem>
      </Menu>
    </>
  )
}

function HelpCenter({
  open,
  onClose,
  onReplayTour,
}: {
  open: boolean
  onClose: () => void
  onReplayTour: () => void
}) {
  const navigate = useNavigate()
  const go = (route: string) => {
    navigate(route)
    onClose()
  }

  return (
    <Drawer
      anchor="right"
      open={open}
      onClose={onClose}
      sx={{ '& .MuiDrawer-paper': { width: { xs: '100%', sm: 420 }, boxSizing: 'border-box' } }}
    >
      <Stack sx={{ height: '100%' }}>
        <Box sx={{ p: 2.5, borderBottom: 1, borderColor: 'divider' }}>
          <Stack direction="row" spacing={1.5} sx={{ alignItems: 'center' }}>
            <Box
              aria-hidden="true"
              sx={{
                display: 'grid',
                placeItems: 'center',
                width: 42,
                height: 42,
                borderRadius: 2,
                bgcolor: 'action.hover',
                color: 'primary.main',
              }}
            >
              <MenuBookOutlined />
            </Box>
            <Box sx={{ minWidth: 0 }}>
              <Typography variant="h6">User guide</Typography>
              <Typography variant="body2" color="text.secondary">
                A quick cockpit reference.
              </Typography>
            </Box>
          </Stack>
        </Box>
        <Stack spacing={1.5} sx={{ flex: 1, overflowY: 'auto', p: 2.5 }}>
          <Button
            variant="contained"
            startIcon={<RocketLaunchOutlined />}
            onClick={() => {
              onClose()
              onReplayTour()
            }}
            sx={{ alignSelf: 'flex-start' }}
          >
            Show me around
          </Button>
          {HELP_LINKS.map((item) => (
            <Card key={item.title} variant="outlined">
              <CardContent>
                <Stack spacing={1}>
                  <Typography variant="subtitle2" sx={{ fontWeight: 800 }}>
                    {item.title}
                  </Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ lineHeight: 1.6 }}>
                    {item.body}
                  </Typography>
                  <Button size="small" onClick={() => go(item.route)} sx={{ alignSelf: 'flex-start' }}>
                    Open
                  </Button>
                </Stack>
              </CardContent>
            </Card>
          ))}
        </Stack>
      </Stack>
    </Drawer>
  )
}

function GuidedTour({
  open,
  step,
  onClose,
  onStep,
}: {
  open: boolean
  step: number
  onClose: () => void
  onStep: (step: number) => void
}) {
  const current = TOUR_STEPS[step] ?? TOUR_STEPS[0]
  const last = step === TOUR_STEPS.length - 1

  return (
    <Dialog open={open} onClose={onClose} aria-label="Gigmann cockpit tour" maxWidth="sm" fullWidth>
      <DialogTitle sx={{ pb: 1 }}>
        <Stack spacing={1}>
          <Typography variant="overline" color="text.secondary" sx={{ fontWeight: 800, letterSpacing: 0 }}>
            Step {step + 1} of {TOUR_STEPS.length}
          </Typography>
          <Typography variant="h5">{current.title}</Typography>
        </Stack>
      </DialogTitle>
      <DialogContent>
        <Stepper activeStep={step} alternativeLabel sx={{ mb: 3, display: { xs: 'none', sm: 'flex' } }}>
          {TOUR_STEPS.map((item) => (
            <Step key={item.label}>
              <StepLabel>{item.label}</StepLabel>
            </Step>
          ))}
        </Stepper>
        <Typography variant="body1" color="text.secondary" sx={{ lineHeight: 1.8 }}>
          {current.body}
        </Typography>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2.5 }}>
        <Button color="inherit" onClick={onClose}>
          Skip
        </Button>
        <Button disabled={step === 0} onClick={() => onStep(step - 1)}>
          Back
        </Button>
        <Button variant="contained" onClick={() => (last ? onClose() : onStep(step + 1))}>
          {last ? 'Finish' : 'Next'}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

function NotificationsBell() {
  const navigate = useNavigate()
  const { data, isLoading, isError } = useAlerts(8)
  const updateStatus = useUpdateAlertStatus()
  const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null)
  const open = Boolean(anchorEl)
  const alerts = data?.alerts ?? []
  const unread = alerts.filter((a) => a.status === 'open').length
  const close = () => setAnchorEl(null)

  const update = (alert: AlertItem, status: 'dismissed' | 'resolved') => {
    updateStatus.mutate({ id: alert.id, status })
  }

  return (
    <>
      <Tooltip title="Notifications">
        <IconButton color="inherit" onClick={(e) => setAnchorEl(e.currentTarget)} aria-label="Open notifications">
          <Badge color="error" badgeContent={unread} max={9}>
            <NotificationsActiveOutlined />
          </Badge>
        </IconButton>
      </Tooltip>
      <Popover
        open={open}
        anchorEl={anchorEl}
        onClose={close}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
        transformOrigin={{ vertical: 'top', horizontal: 'right' }}
        slotProps={{ paper: { sx: { width: 420, maxWidth: 'calc(100vw - 24px)', mt: 1 } } }}
      >
        <Box sx={{ p: 2 }}>
          <Stack direction="row" spacing={1} sx={{ justifyContent: 'space-between', alignItems: 'center', mb: 1.5 }}>
            <Box>
              <Typography variant="h6" sx={{ fontSize: 18 }}>
                Notifications
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Critical feed from the network.
              </Typography>
            </Box>
            <Button
              size="small"
              onClick={() => {
                navigate('/')
                close()
              }}
            >
              Open feed
            </Button>
          </Stack>
          <Stack spacing={1}>
            {isLoading ? (
              [0, 1, 2].map((i) => <Skeleton key={i} variant="rounded" height={76} data-testid="notification-skeleton" />)
            ) : isError ? (
              <Typography variant="body2" color="error.main">
                Couldn&apos;t load notifications.
              </Typography>
            ) : alerts.length === 0 ? (
              <Typography variant="body2" color="text.secondary" sx={{ py: 2 }}>
                No open alerts right now.
              </Typography>
            ) : (
              alerts.map((alert) => (
                <Card key={alert.id} variant="outlined">
                  <CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 } }}>
                    <Stack spacing={1}>
                      <Stack direction="row" spacing={1} sx={{ justifyContent: 'space-between', alignItems: 'flex-start' }}>
                        <Typography variant="body2" sx={{ fontWeight: 800, lineHeight: 1.4 }}>
                          {alert.title}
                        </Typography>
                        <StatusChip status={alert.severity as FacilityStatus} />
                      </Stack>
                      {alert.detail ? (
                        <Typography variant="caption" color="text.secondary" sx={{ lineHeight: 1.5 }}>
                          {alert.detail}
                        </Typography>
                      ) : null}
                      <Stack direction="row" spacing={1}>
                        <Button
                          size="small"
                          startIcon={<DoneAllOutlined fontSize="small" />}
                          onClick={() => update(alert, 'resolved')}
                          disabled={updateStatus.isPending}
                        >
                          Resolve
                        </Button>
                        <Button
                          size="small"
                          color="inherit"
                          startIcon={<VisibilityOffOutlined fontSize="small" />}
                          onClick={() => update(alert, 'dismissed')}
                          disabled={updateStatus.isPending}
                        >
                          Dismiss
                        </Button>
                      </Stack>
                    </Stack>
                  </CardContent>
                </Card>
              ))
            )}
          </Stack>
        </Box>
      </Popover>
    </>
  )
}

function currentRouteTitle(pathname: string): string {
  if (pathname.startsWith('/facilities/')) {
    return 'Facility detail'
  }
  const exact = NAV_ITEMS.find((item) => (item.end ? pathname === item.to : pathname === item.to || pathname.startsWith(`${item.to}/`)))
  return exact?.label ?? 'Executive cockpit'
}

function mobileValue(pathname: string): string | false {
  if (pathname.startsWith('/facilities/')) {
    return '/network'
  }
  const match = MOBILE_ITEMS.find((item) => (item.end ? pathname === item.to : pathname === item.to || pathname.startsWith(`${item.to}/`)))
  return match?.to ?? false
}

/** AppShell is the persistent cockpit frame: brand bar, nav rail, content outlet. */
export function AppShell() {
  const { mode, toggle } = useColorMode()
  const { user, logout } = useAuth()
  const navigation = useNavigation()
  const location = useLocation()
  const navigate = useNavigate()
  const reduceMotion = useReducedMotion()
  const [mobileOpen, setMobileOpen] = useState(false)
  const [helpOpen, setHelpOpen] = useState(false)
  const [tourOpen, setTourOpen] = useState(false)
  const [tourStep, setTourStep] = useState(0)
  const [desktopCollapsed, setDesktopCollapsed] = useState(() => readSidebarCollapsed())
  const drawerWidth = desktopCollapsed ? COLLAPSED_DRAWER_WIDTH : EXPANDED_DRAWER_WIDTH
  const routeTitle = useMemo(() => currentRouteTitle(location.pathname), [location.pathname])
  useLiveUpdates()

  const startTour = useCallback(() => {
    setHelpOpen(false)
    setTourStep(0)
    setTourOpen(true)
    navigate(TOUR_STEPS[0].route)
  }, [navigate])

  const closeTour = () => {
    setTourOpen(false)
    if (typeof window !== 'undefined') {
      window.localStorage.setItem('gigmann-tour-complete', 'true')
    }
  }

  const goToTourStep = useCallback((nextStep: number) => {
    const bounded = Math.min(Math.max(nextStep, 0), TOUR_STEPS.length - 1)
    setTourStep(bounded)
    navigate(TOUR_STEPS[bounded].route)
  }, [navigate])

  useEffect(() => {
    const openHelp = () => setHelpOpen(true)
    const replayTour = () => startTour()
    window.addEventListener(OPEN_HELP_EVENT, openHelp)
    window.addEventListener(REPLAY_TOUR_EVENT, replayTour)
    return () => {
      window.removeEventListener(OPEN_HELP_EVENT, openHelp)
      window.removeEventListener(REPLAY_TOUR_EVENT, replayTour)
    }
  }, [startTour])

  const toggleDesktopCollapsed = () => {
    setDesktopCollapsed((current) => {
      const next = !current
      if (typeof window !== 'undefined') {
        window.localStorage.setItem(SIDEBAR_KEY, String(next))
      }
      return next
    })
  }

  // Theme toggle with a circular clip-path reveal (View Transitions API), with a
  // graceful fallback where it is unsupported or reduced motion is requested.
  const toggleTheme = (e: MouseEvent) => {
    const doc = document as Document & {
      startViewTransition?: (cb: () => void) => { ready: Promise<void> }
    }
    if (reduceMotion || typeof doc.startViewTransition !== 'function') {
      toggle()
      return
    }
    const { clientX: x, clientY: y } = e
    const transition = doc.startViewTransition(() => flushSync(toggle))
    void transition.ready.then(() => {
      const radius = Math.hypot(Math.max(x, window.innerWidth - x), Math.max(y, window.innerHeight - y))
      document.documentElement.animate(
        { clipPath: [`circle(0px at ${x}px ${y}px)`, `circle(${radius}px at ${x}px ${y}px)`] },
        { duration: 400, easing: 'ease-out', pseudoElement: '::view-transition-new(root)' },
      )
    })
  }

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      <Box
        component="a"
        href="#main-content"
        sx={{
          position: 'absolute',
          left: 12,
          top: 12,
          zIndex: (t) => t.zIndex.drawer + 3,
          px: 2,
          py: 1,
          borderRadius: 1,
          bgcolor: 'primary.main',
          color: 'primary.contrastText',
          fontWeight: 600,
          textDecoration: 'none',
          transform: 'translateY(-150%)',
          transition: 'transform .15s ease',
          '&:focus, &:focus-visible': { transform: 'translateY(0)' },
          '@media (prefers-reduced-motion: reduce)': { transition: 'none' },
        }}
      >
        Skip to main content
      </Box>
      {navigation.state === 'loading' ? (
        <LinearProgress
          aria-label="Loading"
          sx={{ position: 'fixed', top: 0, left: 0, right: 0, zIndex: (t) => t.zIndex.drawer + 2 }}
        />
      ) : null}
      <AppBar
        position="fixed"
        color="transparent"
        sx={{
          zIndex: (t) => t.zIndex.drawer + 1,
          width: { md: `calc(100% - ${drawerWidth}px)` },
          ml: { md: `${drawerWidth}px` },
          borderBottom: 1,
          borderColor: 'divider',
          backdropFilter: 'blur(18px)',
          bgcolor: (theme) => (theme.palette.mode === 'dark' ? 'rgba(8, 17, 31, 0.82)' : 'rgba(246, 248, 251, 0.82)'),
          color: 'text.primary',
          transition: 'width .18s ease, margin-left .18s ease',
        }}
      >
        <Toolbar sx={{ minHeight: 72, gap: 1.25 }}>
          <IconButton
            color="inherit"
            edge="start"
            onClick={() => setMobileOpen(true)}
            aria-label="Open navigation"
            sx={{ display: { md: 'none' } }}
          >
            <MenuOutlined />
          </IconButton>
          <Box sx={{ flexGrow: 1, minWidth: 0 }}>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', fontWeight: 800 }}>
              Executive cockpit
            </Typography>
            <Typography variant="body2" sx={{ fontWeight: 800, display: { xs: 'none', sm: 'block' } }}>
              {routeTitle}
            </Typography>
          </Box>
          <FacilitySearch />
          <Tooltip title="Help">
            <IconButton color="inherit" onClick={() => setHelpOpen(true)} aria-label="Open help">
              <HelpOutlineOutlined />
            </IconButton>
          </Tooltip>
          <NotificationsBell />
          <Tooltip title={mode === 'light' ? 'Switch to dark mode' : 'Switch to light mode'}>
            <IconButton
              color="inherit"
              onClick={toggleTheme}
              aria-label={mode === 'light' ? 'Switch to dark mode' : 'Switch to light mode'}
            >
              {mode === 'light' ? <DarkModeOutlined /> : <LightModeOutlined />}
            </IconButton>
          </Tooltip>
          {user ? (
            <AccountMenu
              userName={user.name}
              role={user.role}
              onLogout={logout}
              onOpenHelp={() => setHelpOpen(true)}
              onReplayTour={startTour}
            />
          ) : null}
        </Toolbar>
      </AppBar>

      <Drawer
        variant="temporary"
        open={mobileOpen}
        onClose={() => setMobileOpen(false)}
        ModalProps={{ keepMounted: true }}
        sx={{
          display: { xs: 'block', md: 'none' },
          '& .MuiDrawer-paper': { width: EXPANDED_DRAWER_WIDTH, boxSizing: 'border-box' },
        }}
      >
        <DrawerContent collapsed={false} onNavigate={() => setMobileOpen(false)} />
      </Drawer>

      <Drawer
        variant="permanent"
        sx={{
          display: { xs: 'none', md: 'block' },
          width: drawerWidth,
          flexShrink: 0,
          transition: 'width .18s ease',
          '& .MuiDrawer-paper': {
            width: drawerWidth,
            boxSizing: 'border-box',
            borderRight: 1,
            borderColor: 'divider',
            overflowX: 'hidden',
            transition: 'width .18s ease',
          },
        }}
        open
      >
        <DrawerContent collapsed={desktopCollapsed} onToggleCollapse={toggleDesktopCollapsed} />
      </Drawer>

      <Box
        component="main"
        id="main-content"
        tabIndex={-1}
        sx={{ flexGrow: 1, minWidth: 0, p: { xs: 2, md: 3.5 }, pb: { xs: 10, md: 3.5 }, '&:focus': { outline: 'none' } }}
      >
        <Toolbar sx={{ minHeight: 72 }} />
        <Container maxWidth="xl" disableGutters>
          <motion.div
            key={location.pathname}
            initial={reduceMotion ? false : { opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.22, ease: 'easeOut' }}
          >
            <Outlet />
          </motion.div>
        </Container>
      </Box>

      <BottomNavigation
        value={mobileValue(location.pathname)}
        onChange={(_, value: unknown) => {
          if (typeof value === 'string' && value !== location.pathname) {
            navigate(value)
          }
        }}
        showLabels
        sx={{
          display: { xs: 'flex', md: 'none' },
          position: 'fixed',
          left: 0,
          right: 0,
          bottom: 0,
          zIndex: (t) => t.zIndex.appBar,
          borderTop: 1,
          borderColor: 'divider',
          bgcolor: 'background.paper',
        }}
      >
        {MOBILE_ITEMS.map((item) => (
          <BottomNavigationAction key={item.to} value={item.to} label={item.label} icon={item.icon} />
        ))}
      </BottomNavigation>

      <HelpCenter open={helpOpen} onClose={() => setHelpOpen(false)} onReplayTour={startTour} />
      <GuidedTour open={tourOpen} step={tourStep} onClose={closeTour} onStep={goToTourStep} />
    </Box>
  )
}
