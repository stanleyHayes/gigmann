import { useState, type FormEvent } from 'react'
import Alert from '@mui/material/Alert'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Chip from '@mui/material/Chip'
import Stack from '@mui/material/Stack'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'

import { useAsk } from '../api/useAsk'
import { ButtonLoadingDots } from '../components/ButtonLoadingDots'

const SUGGESTIONS = [
  'Which facility needs me most today, and why?',
  'What is driving the NHIS denials?',
  'Where am I about to run out of stock?',
]

/** AskScreen answers natural-language questions, grounded in today's figures. */
export function AskScreen() {
  const ask = useAsk()
  const [question, setQuestion] = useState('')

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

  return (
    <Stack spacing={3}>
      <Typography variant="h1" sx={{ fontSize: { xs: '2rem', md: '2.5rem' } }}>
        Ask
      </Typography>
      <Typography variant="body2" color="text.secondary">
        Ask anything about the network — answers are grounded in today&apos;s computed figures.
      </Typography>

      <Stack component="form" direction="row" spacing={1} onSubmit={onSubmit}>
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

      {ask.isError ? <Alert severity="error">Couldn&apos;t get an answer. Try again shortly.</Alert> : null}

      {ask.data ? (
        <Card variant="outlined">
          <CardContent>
            <Typography variant="body1" sx={{ whiteSpace: 'pre-wrap' }}>
              {ask.data.text}
            </Typography>
            {ask.data.citations && ask.data.citations.length > 0 ? (
              <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap', gap: 1, mt: 2 }}>
                {ask.data.citations.map((c) => (
                  <Chip key={c} size="small" label={c} />
                ))}
              </Stack>
            ) : null}
          </CardContent>
        </Card>
      ) : null}
    </Stack>
  )
}
