import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

vi.mock('qrcode', () => ({
  default: {
    toDataURL: vi.fn(),
  },
}))

import QRCode from 'qrcode'
import { MfaQrCode } from './MfaQrCode'

const toDataURL = QRCode.toDataURL as ReturnType<typeof vi.fn>

describe('MfaQrCode', () => {
  it('renders a generated QR code image', async () => {
    toDataURL.mockResolvedValue('data:image/png;base64,qr')
    render(<MfaQrCode uri="otpauth://totp/test" />)
    const img = await screen.findByRole('img', { name: /qr code/i })
    expect(img).toHaveAttribute('src', 'data:image/png;base64,qr')
  })

  it('shows a fallback message when generation fails', async () => {
    toDataURL.mockRejectedValue(new Error('fail'))
    render(<MfaQrCode uri="otpauth://totp/test" />)
    expect(await screen.findByText(/Couldn.t generate the QR code/i)).toBeInTheDocument()
  })
})
