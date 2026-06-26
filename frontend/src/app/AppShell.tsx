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
import ChecklistOutlined from '@mui/icons-material/ChecklistOutlined'
import DarkModeOutlined from '@mui/icons-material/DarkModeOutlined'
import ForumOutlined from '@mui/icons-material/ForumOutlined'
import HubOutlined from '@mui/icons-material/HubOutlined'
import InsightsOutlined from '@mui/icons-material/InsightsOutlined'
import LightModeOutlined from '@mui/icons-material/LightModeOutlined'
import LogoutOutlined from '@mui/icons-material/LogoutOutlined'
import TaskAltOutlined from '@mui/icons-material/TaskAltOutlined'
import TodayOutlined from '@mui/icons-material/TodayOutlined'
import { NavLink, Outlet, useNavigation } from 'react-router-dom'
import type { ReactNode } from 'react'

import { useAuth } from '../auth/authContext'
import { useColorMode } from './colorMode'

type NavItem = { to: string; label: string; icon: ReactNode; end?: boolean }

const NAV: NavItem[] = [
  { to: '/', label: 'Today', icon: <TodayOutlined />, end: true },
  { to: '/network', label: 'Network', icon: <HubOutlined /> },
  { to: '/kpis', label: 'KPIs', icon: <InsightsOutlined /> },
  { to: '/ask', label: 'Ask', icon: <ForumOutlined /> },
  { to: '/my-day', label: 'My Day', icon: <ChecklistOutlined /> },
  { to: '/approvals', label: 'Approvals', icon: <TaskAltOutlined /> },
]

const DRAWER_WIDTH = 248

/** AppShell is the persistent cockpit frame: brand bar, nav rail, content outlet. */
export function AppShell() {
  const { mode, toggle } = useColorMode()
  const { user, logout } = useAuth()
  const navigation = useNavigation()

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
          <IconButton
            color="inherit"
            onClick={toggle}
            aria-label={mode === 'light' ? 'Switch to dark mode' : 'Switch to light mode'}
          >
            {mode === 'light' ? <DarkModeOutlined /> : <LightModeOutlined />}
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
          <Outlet />
        </Container>
      </Box>
    </Box>
  )
}
