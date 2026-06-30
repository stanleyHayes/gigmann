import { useEffect, useMemo, useState } from 'react'
import FormControl from '@mui/material/FormControl'
import InputLabel from '@mui/material/InputLabel'
import MenuItem from '@mui/material/MenuItem'
import Pagination from '@mui/material/Pagination'
import Select from '@mui/material/Select'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'

type UsePaginationOptions = {
  initialPageSize: number
  resetKey?: string
}

export function usePagination<T>(items: readonly T[], { initialPageSize, resetKey }: UsePaginationOptions) {
  const [rawPage, setRawPage] = useState(1)
  const [pageSize, setPageSizeState] = useState(initialPageSize)
  const pageCount = Math.max(1, Math.ceil(items.length / pageSize))
  const page = Math.min(rawPage, pageCount)

  useEffect(() => {
    if (rawPage !== page) {
      setRawPage(page)
    }
  }, [page, rawPage])

  useEffect(() => {
    setRawPage(1)
  }, [resetKey])

  const pageItems = useMemo(() => {
    const start = (page - 1) * pageSize
    return items.slice(start, start + pageSize)
  }, [items, page, pageSize])

  return {
    page,
    pageCount,
    pageItems,
    pageSize,
    total: items.length,
    setPage: setRawPage,
    setPageSize: (next: number) => {
      setPageSizeState(next)
      setRawPage(1)
    },
  }
}

type PaginationControlsProps = {
  id: string
  itemLabel: string
  page: number
  pageCount: number
  pageSize: number
  pageSizeOptions: number[]
  total: number
  onPageChange: (page: number) => void
  onPageSizeChange: (pageSize: number) => void
}

/** PaginationControls keeps growable cockpit lists scannable without changing their card layout. */
export function PaginationControls({
  id,
  itemLabel,
  page,
  pageCount,
  pageSize,
  pageSizeOptions,
  total,
  onPageChange,
  onPageSizeChange,
}: PaginationControlsProps) {
  const smallestPage = Math.min(...pageSizeOptions)
  if (total <= smallestPage) {
    return null
  }

  const start = (page - 1) * pageSize + 1
  const end = Math.min(total, page * pageSize)

  return (
    <Stack
      direction={{ xs: 'column', sm: 'row' }}
      spacing={1.5}
      aria-label={`${itemLabel} pagination`}
      data-testid={`${id}-pagination`}
      sx={{
        alignItems: { xs: 'stretch', sm: 'center' },
        justifyContent: 'space-between',
        pt: 0.5,
      }}
    >
      <Typography variant="body2" color="text.secondary" sx={{ fontWeight: 700 }}>
        {start}-{end} of {total} {itemLabel}
      </Typography>
      <Stack direction="row" spacing={1} sx={{ alignItems: 'center', justifyContent: { xs: 'space-between', sm: 'flex-end' } }}>
        <Pagination
          count={pageCount}
          page={page}
          onChange={(_, next) => onPageChange(next)}
          size="small"
          shape="rounded"
          siblingCount={0}
          boundaryCount={1}
        />
        <FormControl size="small" sx={{ minWidth: 116 }}>
          <InputLabel id={`${id}-page-size-label`}>Per page</InputLabel>
          <Select
            labelId={`${id}-page-size-label`}
            label="Per page"
            value={pageSize}
            onChange={(event) => onPageSizeChange(Number(event.target.value))}
          >
            {pageSizeOptions.map((option) => (
              <MenuItem key={option} value={option}>
                {option}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      </Stack>
    </Stack>
  )
}
