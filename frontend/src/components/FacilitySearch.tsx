import Box from '@mui/material/Box'
import CircularProgress from '@mui/material/CircularProgress'
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
  const { data: matches = [], isFetching } = useFacilitySearch(debounced)

  const close = () => {
    setOpen(false)
    setInput('')
  }
  const go = (facilityId: string) => {
    close()
    navigate(`/facilities/${facilityId}`)
  }

  const longEnough = input.trim().length >= MIN_QUERY

  return (
    <>
      <IconButton color="inherit" onClick={() => setOpen(true)} aria-label="Search facilities">
        <SearchOutlined />
      </IconButton>

      <Dialog open={open} onClose={close} fullWidth maxWidth="sm" aria-labelledby="facility-search-title">
        <DialogTitle id="facility-search-title" sx={{ pb: 1, fontFamily: '"Fraunces Variable", serif' }}>
          Find a facility
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
                endAdornment: isFetching ? <CircularProgress size={18} aria-label="Searching" /> : null,
              },
            }}
          />

          {longEnough && matches.length > 0 ? (
            <List aria-label="Facility matches" sx={{ mt: 1 }}>
              {matches.map((m) => (
                <ListItemButton key={m.facilityId} onClick={() => go(m.facilityId)}>
                  <ListItemText primary={m.name} secondary={`${Math.round(m.score * 100)}% match`} />
                </ListItemButton>
              ))}
            </List>
          ) : (
            <Box sx={{ mt: 2 }}>
              <Typography variant="body2" color="text.secondary">
                {longEnough && !isFetching
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
