import Box from '@mui/material/Box'
import Dialog from '@mui/material/Dialog'
import DialogContent from '@mui/material/DialogContent'
import DialogTitle from '@mui/material/DialogTitle'
import IconButton from '@mui/material/IconButton'
import List from '@mui/material/List'
import ListItemButton from '@mui/material/ListItemButton'
import ListItemText from '@mui/material/ListItemText'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'
import SearchOutlined from '@mui/icons-material/SearchOutlined'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { useFacilitySearch } from '../api/useFacilitySearch'
import { ButtonLoadingDots } from './ButtonLoadingDots'

const MIN_QUERY = 2

function useDebounced<T>(value: T, delay: number): T {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(id)
  }, [value, delay])
  return debounced
}

/**
 * FacilitySearch is a command-palette-style quick find in the app bar: type a
 * name or a natural-language phrase ("how is the Kasoa polyclinic doing") and
 * jump to the matched facility. Backed by the vector-search endpoint (GEC-13).
 */
export function FacilitySearch() {
  const [open, setOpen] = useState(false)
  const [input, setInput] = useState('')
  const debounced = useDebounced(input, 300)
  const navigate = useNavigate()
  const { data: matches = [], isFetching, isError } = useFacilitySearch(debounced)

  const close = () => {
    setOpen(false)
    setInput('')
  }
  const go = (facilityId: string) => {
    navigate(`/facilities/${facilityId}`)
    close()
  }

  const longEnough = input.trim().length >= MIN_QUERY

  return (
    <>
      <IconButton color="inherit" onClick={() => setOpen(true)} aria-label="Search facilities">
        <SearchOutlined />
      </IconButton>

      <Dialog
        open={open}
        onClose={close}
        fullWidth
        maxWidth="sm"
        aria-labelledby="facility-search-title"
        slotProps={{
          paper: {
            sx: { borderRadius: 2, border: 1, borderColor: 'divider', overflow: 'hidden' },
          },
        }}
      >
        <DialogTitle id="facility-search-title" sx={{ pb: 1 }}>
          <Typography variant="overline" color="text.secondary" sx={{ display: 'block', fontWeight: 800, letterSpacing: 0 }}>
            Quick search
          </Typography>
          <Typography variant="h5">Find a facility</Typography>
        </DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            fullWidth
            variant="outlined"
            placeholder="e.g. how is the Kasoa polyclinic doing"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && matches[0]) {
                go(matches[0].facilityId)
              }
            }}
            slotProps={{
              htmlInput: { 'aria-label': 'Search facilities' },
              input: {
                endAdornment: isFetching ? (
                  <Box sx={{ color: 'primary.main', display: 'inline-flex', pl: 1 }}>
                    <ButtonLoadingDots size={4} />
                  </Box>
                ) : null,
              },
            }}
          />

          {longEnough && matches.length > 0 ? (
            <List aria-label="Facility matches" sx={{ mt: 1.5, display: 'grid', gap: 0.75 }}>
              {matches.map((m) => (
                <ListItemButton
                  key={m.facilityId}
                  onClick={() => go(m.facilityId)}
                  sx={{ border: 1, borderColor: 'divider', borderRadius: 2 }}
                >
                  <ListItemText primary={m.name} secondary={`${Math.round(m.score * 100)}% match`} />
                </ListItemButton>
              ))}
            </List>
          ) : (
            <Box sx={{ mt: 2 }}>
              <Typography variant="body2" color={isError ? 'error' : 'text.secondary'}>
                {isError
                  ? 'Search is unavailable right now. Try again shortly.'
                  : longEnough && !isFetching
                    ? `No facilities match “${input.trim()}”.`
                    : 'Type a name or a natural-language phrase.'}
              </Typography>
            </Box>
          )}
        </DialogContent>
      </Dialog>
    </>
  )
}
