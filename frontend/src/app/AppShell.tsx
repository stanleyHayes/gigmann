import AppBar from '@mui/material/AppBar'
import Box from '@mui/material/Box'
import Container from '@mui/material/Container'
import Drawer from '@mui/material/Drawer'
import IconButton from '@mui/material/IconButton'
import LinearProgress from '@mui/material/LinearProgress'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemButton from '@mui/material/ListItemButton'
import ListItemIcon from '@mui/material/ListItemIcon'
import ListItemText from '@mui/material/ListItemText'
import Toolbar from '@mui/material/Toolbar'
import Typography from '@mui/material/Typography'
import AssignmentIndOutlined from '@mui/icons-material/AssignmentIndOutlined'
import ChecklistOutlined from '@mui/icons-material/ChecklistOutlined'
import DarkModeOutlined from '@mui/icons-material/DarkModeOutlined'
import ForumOutlined from '@mui/icons-material/ForumOutlined'
import HubOutlined from '@mui/icons-material/HubOutlined'
import InsightsOutlined from '@mui/icons-material/InsightsOutlined'
import SummarizeOutlined from '@mui/icons-material/SummarizeOutlined'
import LightModeOutlined from '@mui/icons-material/LightModeOutlined'
import LogoutOutlined from '@mui/icons-material/LogoutOutlined'
import SettingsOutlined from '@mui/icons-material/SettingsOutlined'
import TaskAltOutlined from '@mui/icons-material/TaskAltOutlined'
import TodayOutlined from '@mui/icons-material/TodayOutlined'
import { NavLink, Outlet, useLocation, useNavigation } from 'react-router-dom'
import { motion, useReducedMotion } from 'framer-motion'
import { flushSync } from 'react-dom'
import type { MouseEvent, ReactNode } from 'react'

import { FacilitySearch } from '../components/FacilitySearch'
import { useAuth } from '../auth/authContext'
import { useColorMode } from './colorMode'
import { t } from '../i18n/messages'

type NavItem = { to: string; label: string; icon: ReactNode; end?: boolean }

const NAV: NavItem[] = [
  { to: '/', label: t('nav.today'), icon: <TodayOutlined />, end: true },
  { to: '/network', label: t('nav.network'), icon: <HubOutlined /> },
  { to: '/kpis', label: t('nav.kpis'), icon: <InsightsOutlined /> },
  { to: '/reports', label: t('nav.reports'), icon: <SummarizeOutlined /> },
  { to: '/ask', label: t('nav.ask'), icon: <ForumOutlined /> },
  { to: '/my-day', label: t('nav.myDay'), icon: <ChecklistOutlined /> },
  { to: '/delegation', label: t('nav.delegation'), icon: <AssignmentIndOutlined /> },
  { to: '/approvals', label: t('nav.approvals'), icon: <TaskAltOutlined /> },
]

const DRAWER_WIDTH = 248

/** AppShell is the persistent cockpit frame: brand bar, nav rail, content outlet. */
export function AppShell() {
  const { mode, toggle } = useColorMode()
  const { user, logout } = useAuth()
  const navigation = useNavigation()
  const location = useLocation()
  const reduceMotion = useReducedMotion()

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
      {navigation.state === 'loading' ? (
        <LinearProgress
          aria-label="Loading"
          sx={{ position: 'fixed', top: 0, left: 0, right: 0, zIndex: (t) => t.zIndex.drawer + 2 }}
        />
      ) : null}
      <AppBar position="fixed" sx={{ zIndex: (t) => t.zIndex.drawer + 1 }}>
        <Toolbar>
          <Typography variant="h6" component="span" sx={{ flexGrow: 1, fontFamily: '"Fraunces Variable", serif' }}>
            Ahenfie
          </Typography>
          {user ? (
            <Typography variant="body2" sx={{ mr: 1, display: { xs: 'none', sm: 'block' } }}>
              {user.name}
            </Typography>
          ) : null}
          <FacilitySearch />
          <IconButton
            color="inherit"
            onClick={toggleTheme}
            aria-label={mode === 'light' ? 'Switch to dark mode' : 'Switch to light mode'}
          >
            {mode === 'light' ? <DarkModeOutlined /> : <LightModeOutlined />}
          </IconButton>
          <IconButton color="inherit" component={NavLink} to="/settings" aria-label="Settings">
            <SettingsOutlined />
          </IconButton>
          <IconButton color="inherit" onClick={logout} aria-label="Sign out">
            <LogoutOutlined />
          </IconButton>
        </Toolbar>
      </AppBar>

      <Drawer
        variant="permanent"
        sx={{
          width: DRAWER_WIDTH,
          flexShrink: 0,
          '& .MuiDrawer-paper': { width: DRAWER_WIDTH, boxSizing: 'border-box' },
        }}
      >
        <Toolbar />
        <Box component="nav" aria-label="Primary navigation">
          <List>
            {NAV.map((item) => (
              <ListItem key={item.to} disablePadding>
                <ListItemButton
                  component={NavLink}
                  to={item.to}
                  end={item.end}
                  sx={{
                    '&.active': {
                      bgcolor: 'action.selected',
                      borderRight: 3,
                      borderColor: 'primary.main',
                    },
                  }}
                >
                  <ListItemIcon>{item.icon}</ListItemIcon>
                  <ListItemText primary={item.label} />
                </ListItemButton>
              </ListItem>
            ))}
          </List>
        </Box>
      </Drawer>

      <Box component="main" sx={{ flexGrow: 1, p: { xs: 2, md: 4 } }}>
        <Toolbar />
        <Container maxWidth="md" disableGutters>
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
    </Box>
  )
}
