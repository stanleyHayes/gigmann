import { render } from '@testing-library/react'
import { axe } from 'jest-axe'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

import type { Brief } from './api/useBrief'
import { DailyBrief } from './components/DailyBrief'
import { StatusChip } from './components/StatusChip'
import { LoginScreen } from './screens/LoginScreen'
import { AskScreen } from './screens/AskScreen'
import { NetworkScreen } from './screens/NetworkScreen'
import { KpisScreen } from './screens/KpisScreen'
import { MyDayScreen } from './screens/MyDayScreen'
import { ApprovalsScreen } from './screens/ApprovalsScreen'

// LoginScreen reads the auth context; provide a stable stub.
vi.mock('./auth/authContext', () => ({
  useAuth: () => ({
    user: undefined,
    isAuthenticated: false,
    mfaRequired: false,
    login: () => {},
    logout: () => {},
    loginPending: false,
    loginError: null,
  }),
}))

const brief: Brief = {
  id: 'b1', date: '2026-06-26', prose: 'Good morning, Sammy.', model: 'local',
  items: [
    { severity: 'critical', facility_id: 'tafo-maternity', headline: 'Claims not submitted', explanation: 'demand is flat', suggested_actions: ['Why?'] },
    { severity: 'watch', facility_id: 'asokwa', headline: 'Stock low' },
  ],
}

vi.mock('./api/useAsk', () => ({
  useAsk: () => ({ mutate: () => {}, isPending: false, isError: false, data: undefined }),
}))
vi.mock('./api/usePasswordReset', () => ({
  usePasswordResetRequest: () => ({ mutate: () => {}, reset: () => {}, isPending: false, isError: false, data: undefined }),
  usePasswordResetConfirm: () => ({ mutate: () => {}, reset: () => {}, isPending: false, isError: false, error: null }),
}))
vi.mock('./api/useDrafts', () => ({
  useCreateDraft: () => ({ mutate: () => {}, isPending: false, isError: false, data: undefined }),
}))
vi.mock('./api/useFacilities', () => ({
  useFacilities: () => ({ data: [], isLoading: false, isError: false }),
}))
vi.mock('./api/useMetrics', () => ({
  useMetrics: () => ({ data: { as_of: '2026-06-09', kpis: [] }, isLoading: false, isError: false }),
}))
vi.mock('./api/useTasks', () => ({
  useTasks: () => ({ data: [], isLoading: false, isError: false }),
  useUpdateTaskStatus: () => ({ mutate: () => {}, isPending: false }),
}))
vi.mock('./api/useApprovals', () => ({
  useApprovals: () => ({ data: [], isLoading: false, isError: false }),
  useDecideApproval: () => ({ mutate: () => {}, isPending: false }),
}))

describe('accessibility (axe)', () => {
  it('the Daily Brief has no violations', async () => {
    const { container } = render(<DailyBrief brief={brief} isLoading={false} isError={false} />)
    const results = await axe(container)
    expect(results.violations).toEqual([])
  })

  it('the status chips have no violations', async () => {
    const { container } = render(
      <div>
        <StatusChip status="critical" label="Tafo" />
        <StatusChip status="watch" label="Asokwa" />
        <StatusChip status="good" label="Kasoa" />
      </div>,
    )
    expect((await axe(container)).violations).toEqual([])
  })

  it('the Ask screen has no violations', async () => {
    const { container } = render(
      <MemoryRouter>
        <AskScreen />
      </MemoryRouter>,
    )
    expect((await axe(container)).violations).toEqual([])
  })

  it('the network screen has no violations', async () => {
    const { container } = render(<MemoryRouter><NetworkScreen /></MemoryRouter>)
    expect((await axe(container)).violations).toEqual([])
  })

  it('the KPIs screen has no violations', async () => {
    const { container } = render(<MemoryRouter><KpisScreen /></MemoryRouter>)
    expect((await axe(container)).violations).toEqual([])
  })

  it('the My Day screen has no violations', async () => {
    const { container } = render(<MemoryRouter><MyDayScreen /></MemoryRouter>)
    expect((await axe(container)).violations).toEqual([])
  })

  it('the approvals screen has no violations', async () => {
    const { container } = render(<MemoryRouter><ApprovalsScreen /></MemoryRouter>)
    expect((await axe(container)).violations).toEqual([])
  })

  it('the login screen has no violations', async () => {
    const { container } = render(<LoginScreen />, { wrapper: MemoryRouter })
    expect((await axe(container)).violations).toEqual([])
  })
})
