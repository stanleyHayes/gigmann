import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import { useEffect, useState } from 'react'
import QRCode from 'qrcode'

type MfaQrCodeProps = {
  /** otpauth:// URI to encode in the QR code. */
  uri: string
  /** Rendered size in pixels. */
  size?: number
}

/**
 * MfaQrCode renders a scannable QR code for an authenticator app from an
 * otpauth:// URI. It uses an image data URL so the component does not rely on a
 * browser canvas implementation or raw SVG injection.
 */
export function MfaQrCode({ uri, size = 200 }: MfaQrCodeProps) {
  const [dataUrl, setDataUrl] = useState<string | null>(null)
  const [error, setError] = useState(false)

  useEffect(() => {
    let cancelled = false
    setDataUrl(null)
    setError(false)
    QRCode.toDataURL(uri, {
      type: 'image/png',
      width: size,
      margin: 1,
      color: { dark: '#111827', light: '#ffffff' },
      errorCorrectionLevel: 'M',
    })
      .then((value) => {
        if (!cancelled) setDataUrl(value)
      })
      .catch(() => {
        if (!cancelled) setError(true)
      })
    return () => {
      cancelled = true
    }
  }, [uri, size])

  if (error) {
    return <Alert severity="error">Couldn&apos;t generate the QR code. You can still enter the key manually.</Alert>
  }

  if (!dataUrl) {
    return (
      <Box
        role="status"
        aria-label="Generating QR code"
        sx={{
          width: size,
          height: size,
          bgcolor: 'action.hover',
          borderRadius: 1,
        }}
      />
    )
  }

  return (
    <Box
      component="img"
      src={dataUrl}
      alt="QR code for authenticator app setup"
      sx={{
        width: size,
        height: size,
        display: 'block',
        border: '1px solid',
        borderColor: 'divider',
        borderRadius: 1,
      }}
    />
  )
}
