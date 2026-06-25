import Button from '@mui/material/Button'
import Snackbar from '@mui/material/Snackbar'
import { useRegisterSW } from 'virtual:pwa-register/react'

/** ReloadPrompt surfaces offline-readiness and lets the user opt into updates. */
export function ReloadPrompt() {
  const {
    offlineReady: [offlineReady, setOfflineReady],
    needRefresh: [needRefresh, setNeedRefresh],
    updateServiceWorker,
  } = useRegisterSW()

  const open = offlineReady || needRefresh
  const close = () => {
    setOfflineReady(false)
    setNeedRefresh(false)
  }

  return (
    <Snackbar
      open={open}
      onClose={close}
      message={needRefresh ? 'A new version is available.' : 'Ready to work offline.'}
      action={
        needRefresh ? (
          <Button color="secondary" size="small" onClick={() => void updateServiceWorker(true)}>
            Reload
          </Button>
        ) : (
          <Button color="inherit" size="small" onClick={close}>
            Dismiss
          </Button>
        )
      }
    />
  )
}
