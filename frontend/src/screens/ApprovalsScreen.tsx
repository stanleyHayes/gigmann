import { useState } from 'react'
import Alert from '@mui/material/Alert'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Chip from '@mui/material/Chip'
import Dialog from '@mui/material/Dialog'
import DialogActions from '@mui/material/DialogActions'
import DialogContent from '@mui/material/DialogContent'
import DialogContentText from '@mui/material/DialogContentText'
import DialogTitle from '@mui/material/DialogTitle'
import Skeleton from '@mui/material/Skeleton'
import Stack from '@mui/material/Stack'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'

import { useApprovals, useDecideApproval, type Approval, type Decision } from '../api/useApprovals'
import { ButtonLoadingDots } from '../components/ButtonLoadingDots'
import { monoFont } from '../theme'

const TYPE_LABEL: Record<Approval['type'], string> = { capital: 'Capital', hire: 'Hire', reorder: 'Reorder' }
const STATUS_COLOR: Record<Approval['status'], 'warning' | 'success' | 'error'> = {
  pending: 'warning',
  approved: 'success',
  declined: 'error',
}

function formatCedis(pesewas: number): string {
  return `GH₵ ${Math.round(pesewas / 100).toLocaleString('en-US')}`
}

/** ApprovalCard shows one approval and (when pending) the decision controls. */
export function ApprovalCard({ approval, onDecide }: { approval: Approval; onDecide: (a: Approval, d: Decision) => void }) {
  return (
    <Card variant="outlined">
      <CardContent>
        <Stack spacing={1}>
          <Stack direction="row" spacing={1} sx={{ justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              {approval.title}
            </Typography>
            <Chip size="small" label={approval.status} color={STATUS_COLOR[approval.status]} sx={{ textTransform: 'capitalize' }} />
          </Stack>
          <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap', gap: 1 }}>
            <Chip size="small" variant="outlined" label={TYPE_LABEL[approval.type]} />
            <Chip size="small" variant="outlined" label={formatCedis(approval.amount_pesewas)} sx={{ fontFamily: monoFont }} />
            <Chip size="small" variant="outlined" label={approval.facility_id} />
          </Stack>
          {approval.context ? (
            <Typography variant="body2" color="text.secondary">
              {approval.context}
            </Typography>
          ) : null}
          <Typography variant="caption" color="text.secondary">
            Requested by {approval.requested_by}
          </Typography>
          {approval.status === 'pending' ? (
            <Stack direction="row" spacing={1} sx={{ pt: 1 }}>
              <Button size="small" variant="contained" onClick={() => onDecide(approval, 'approve')}>
                Approve
              </Button>
              <Button size="small" variant="outlined" color="error" onClick={() => onDecide(approval, 'decline')}>
                Decline
              </Button>
            </Stack>
          ) : approval.decision_note ? (
            <Typography variant="body2" color="text.secondary">
              Note: {approval.decision_note}
            </Typography>
          ) : null}
        </Stack>
      </CardContent>
    </Card>
  )
}

type PendingDecision = { approval: Approval; decision: Decision }

/** ApprovalsScreen lists the executive's approval queue. A decision is only ever
 *  committed after an explicit confirmation step (never a one-click side-effect). */
export function ApprovalsScreen() {
  const { data, isLoading, isError } = useApprovals()
  const decide = useDecideApproval()
  const [pending, setPending] = useState<PendingDecision | null>(null)
  const [note, setNote] = useState('')

  const openConfirm = (approval: Approval, decision: Decision) => {
    setNote('')
    setPending({ approval, decision })
  }
  const close = () => setPending(null)
  const confirm = () => {
    if (!pending) {
      return
    }
    decide.mutate(
      { id: pending.approval.id, decision: pending.decision, note: note || undefined },
      { onSuccess: close },
    )
  }

  const verb = pending?.decision === 'approve' ? 'Approve' : 'Decline'

  return (
    <Stack spacing={3}>
      <Typography variant="h1" sx={{ fontSize: { xs: '2rem', md: '2.5rem' } }}>
        Approvals
      </Typography>
      {isLoading ? (
        <Stack spacing={2} data-testid="approvals-skeleton">
          {[0, 1, 2].map((i) => (
            <Skeleton key={i} variant="rounded" height={132} />
          ))}
        </Stack>
      ) : isError || !data ? (
        <Alert severity="error">Couldn&apos;t load approvals. Try again shortly.</Alert>
      ) : data.length === 0 ? (
        <Alert severity="info">No approvals waiting for you.</Alert>
      ) : (
        <Stack spacing={2}>
          {data.map((a) => (
            <ApprovalCard key={a.id} approval={a} onDecide={openConfirm} />
          ))}
        </Stack>
      )}

      <Dialog open={pending !== null} onClose={close} fullWidth maxWidth="xs">
        <DialogTitle>{verb} request</DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ mb: 2 }}>
            {pending ? `${verb} “${pending.approval.title}”? This records your decision.` : ''}
          </DialogContentText>
          <TextField
            label="Note (optional)"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            fullWidth
            multiline
            minRows={2}
          />
          {decide.isError ? (
            <Alert severity="error" sx={{ mt: 2 }}>
              Couldn&apos;t record your decision. Try again.
            </Alert>
          ) : null}
        </DialogContent>
        <DialogActions>
          <Button onClick={close} disabled={decide.isPending}>
            Cancel
          </Button>
          <Button
            onClick={confirm}
            variant="contained"
            color={pending?.decision === 'approve' ? 'primary' : 'error'}
            disabled={decide.isPending}
          >
            {decide.isPending ? <ButtonLoadingDots /> : null}
            Confirm {verb.toLowerCase()}
          </Button>
        </DialogActions>
      </Dialog>
    </Stack>
  )
}
