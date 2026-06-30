import { useState, type FormEvent } from 'react'
import { useLocation } from 'react-router-dom'
import Alert from '@mui/material/Alert'
import Button from '@mui/material/Button'
import Chip from '@mui/material/Chip'
import FormControl from '@mui/material/FormControl'
import InputLabel from '@mui/material/InputLabel'
import MenuItem from '@mui/material/MenuItem'
import Select from '@mui/material/Select'
import Snackbar from '@mui/material/Snackbar'
import Stack from '@mui/material/Stack'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'
import ContentCopyOutlined from '@mui/icons-material/ContentCopyOutlined'
import ForumOutlined from '@mui/icons-material/ForumOutlined'
import PsychologyOutlined from '@mui/icons-material/PsychologyOutlined'
import SendOutlined from '@mui/icons-material/SendOutlined'

import { useAsk } from '../api/useAsk'
import { useCreateDraft, type DraftRequest } from '../api/useDrafts'
import { useFacilities } from '../api/useFacilities'
import { ButtonLoadingDots } from '../components/ButtonLoadingDots'
import { PageHeader } from '../components/PageHeader'
import { SurfaceCard } from '../components/SurfaceCard'
import { answerToText } from './exportBrief'

const SUGGESTIONS = [
  'Which facility needs me most today, and why?',
  'What is driving the NHIS denials?',
  'Where am I about to run out of stock?',
]

/** AskScreen answers natural-language questions, grounded in today's figures. */
export function AskScreen() {
  const ask = useAsk()
  const draft = useCreateDraft()
  const { data: facilities = [] } = useFacilities()
  const location = useLocation()
  const prefill = (location.state as { question?: string } | null)?.question ?? ''
  const [question, setQuestion] = useState(prefill)
  const [draftKind, setDraftKind] = useState<DraftRequest['kind']>('message')
  const [draftFacility, setDraftFacility] = useState('none')
  const [draftInstruction, setDraftInstruction] = useState('Draft a concise update for the facility manager with the next action and deadline.')
  const [draftCopied, setDraftCopied] = useState(false)

  const submit = (q: string) => {
    const trimmed = q.trim()
    if (trimmed) {
      ask.mutate(trimmed)
    }
  }
  const onSubmit = (e: FormEvent) => {
    e.preventDefault()
    submit(question)
  }

  const createDraft = () => {
    const instruction = draftInstruction.trim()
    if (!instruction) {
      return
    }
    draft.mutate({
      kind: draftKind,
      facility_id: draftFacility === 'none' ? undefined : draftFacility,
      instruction,
    })
  }

  const copyDraft = async () => {
    if (draft.data?.draft) {
      await navigator.clipboard?.writeText(draft.data.draft)
      setDraftCopied(true)
    }
  }

  return (
    <Stack spacing={3}>
      <PageHeader
        title="Ask"
        eyebrow="Grounded answers"
        description="Ask about the network. Answers cite today&apos;s computed figures and facilities."
        icon={ForumOutlined}
      />

      <Stack
        component="form"
        direction={{ xs: 'column', sm: 'row' }}
        spacing={1}
        onSubmit={onSubmit}
        sx={{
          p: 1,
          border: 1,
          borderColor: 'divider',
          borderRadius: 2,
          bgcolor: 'background.paper',
        }}
      >
        <TextField
          fullWidth
          placeholder="e.g. Why is Tafo critical?"
          value={question}
          onChange={(e) => setQuestion(e.target.value)}
          slotProps={{ htmlInput: { 'aria-label': 'Question' } }}
        />
        <Button type="submit" variant="contained" disabled={ask.isPending || !question.trim()}>
          {ask.isPending ? <ButtonLoadingDots /> : null}
          Ask
        </Button>
      </Stack>

      <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap', gap: 1 }}>
        {SUGGESTIONS.map((s) => (
          <Chip
            key={s}
            label={s}
            variant="outlined"
            onClick={() => {
              setQuestion(s)
              submit(s)
            }}
          />
        ))}
      </Stack>

      <SurfaceCard
        title="Draft utility"
        description="Generate an unsent message or summary using the same guarded cockpit context."
        icon={SendOutlined}
      >
        <Stack spacing={2}>
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5}>
            <FormControl sx={{ minWidth: { xs: '100%', md: 160 } }}>
              <InputLabel id="draft-kind-label">Draft type</InputLabel>
              <Select
                labelId="draft-kind-label"
                label="Draft type"
                value={draftKind}
                onChange={(e) => setDraftKind(e.target.value as DraftRequest['kind'])}
              >
                <MenuItem value="message">Message</MenuItem>
                <MenuItem value="summary">Summary</MenuItem>
              </Select>
            </FormControl>
            <FormControl sx={{ minWidth: { xs: '100%', md: 240 } }}>
              <InputLabel id="draft-facility-label">Facility</InputLabel>
              <Select
                labelId="draft-facility-label"
                label="Facility"
                value={draftFacility}
                onChange={(e) => setDraftFacility(e.target.value)}
              >
                <MenuItem value="none">Network-wide</MenuItem>
                {facilities.map((facility) => (
                  <MenuItem key={facility.id} value={facility.id}>
                    {facility.name}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Stack>
          <TextField
            label="Instruction"
            value={draftInstruction}
            onChange={(e) => setDraftInstruction(e.target.value)}
            multiline
            minRows={3}
            fullWidth
          />
          <Button
            variant="contained"
            startIcon={draft.isPending ? undefined : <SendOutlined />}
            onClick={createDraft}
            disabled={draft.isPending || !draftInstruction.trim()}
            sx={{ alignSelf: 'flex-start' }}
          >
            {draft.isPending ? <ButtonLoadingDots /> : null}
            Generate draft
          </Button>
          {draft.isError ? <Alert severity="error">Couldn&apos;t generate the draft. Try again shortly.</Alert> : null}
        </Stack>
      </SurfaceCard>

      {draft.data ? (
        <SurfaceCard
          title="Generated draft"
          description="Draft-only output. Review before sending."
          icon={SendOutlined}
          actions={
            <Button size="small" startIcon={<ContentCopyOutlined />} onClick={() => void copyDraft()}>
              Copy draft
            </Button>
          }
        >
          <Typography variant="body1" sx={{ whiteSpace: 'pre-wrap', lineHeight: 1.8 }}>
            {draft.data.draft}
          </Typography>
        </SurfaceCard>
      ) : null}

      {ask.isError ? <Alert severity="error">Couldn&apos;t get an answer. Try again shortly.</Alert> : null}

      {ask.data ? (
        <SurfaceCard
          title="Grounded answer"
          description="Generated from supplied network context and returned citations."
          icon={PsychologyOutlined}
        >
            <Typography variant="body1" sx={{ mt: 1, whiteSpace: 'pre-wrap', lineHeight: 1.8 }}>
              {ask.data.text}
            </Typography>
            {ask.data.citations && ask.data.citations.length > 0 ? (
              <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap', gap: 1, mt: 2 }}>
                {[...new Set(ask.data.citations)].map((c) => (
                  <Chip key={c} size="small" label={c} />
                ))}
              </Stack>
            ) : null}
            <Stack direction="row" sx={{ mt: 2, justifyContent: 'flex-end' }}>
              <Button
                size="small"
                startIcon={<ContentCopyOutlined fontSize="small" />}
                onClick={() => void navigator.clipboard?.writeText(answerToText(ask.data))}
              >
                Copy answer
              </Button>
            </Stack>
        </SurfaceCard>
      ) : null}
      <Snackbar
        open={draftCopied}
        autoHideDuration={2500}
        onClose={() => setDraftCopied(false)}
        message="Draft copied"
      />
    </Stack>
  )
}
