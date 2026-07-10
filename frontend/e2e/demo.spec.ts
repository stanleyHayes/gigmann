import { expect, test } from '@playwright/test'

const EMAIL = 'ceo@gigmann.health'
const PASSWORD = 'ahenfie-demo'

// The §3.3 demo narrative end-to-end against the in-memory network.
test('demo narrative: login → brief → network → ask → my day → approvals', async ({ page }) => {
  await page.goto('/')

  await page.getByLabel(/email/i).fill(EMAIL)
  // Exact label: the visibility toggle is aria-label "Show sign-in password",
  // so a /password/i regex matches two elements (the input and the button).
  await page.getByLabel('Password', { exact: true }).fill(PASSWORD)
  await page.getByRole('button', { name: /sign in/i }).click()

  // The Daily Brief is the hero; the worst item (Tafo) surfaces first.
  await expect(page.getByRole('link', { name: /Today/i })).toBeVisible()
  await expect(page.getByText(/Tafo/i).first()).toBeVisible()

  await page.getByRole('link', { name: /Network/i }).click()
  await expect(page).toHaveURL(/\/network/)
  await expect(page.getByText(/Kasoa Polyclinic/i)).toBeVisible()

  await page.getByRole('link', { name: /Ask/i }).click()
  await page.getByLabel(/question/i).fill('Why is Tafo critical?')
  await page.getByRole('button', { name: /^ask$/i }).click()
  await expect(page.getByText(/.+/).first()).toBeVisible()

  await page.getByRole('link', { name: /My Day/i }).click()
  await expect(page).toHaveURL(/\/my-day/)

  await page.getByRole('link', { name: /Approvals/i }).click()
  await expect(page).toHaveURL(/\/approvals/)
})
