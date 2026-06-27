// Locale-aware formatters (en-GH). Centralising these keeps the app i18n-ready:
// a new locale is a new constant + catalog, not a hunt through components.
export const LOCALE = 'en-GH'

export const fmt = {
  number: (n: number) => new Intl.NumberFormat(LOCALE).format(n),
  /** Format Ghana cedis from a whole-cedi amount (figures arrive pre-formatted from the API; this is for client-side math). */
  cedis: (cedis: number) =>
    new Intl.NumberFormat(LOCALE, { style: 'currency', currency: 'GHS', currencyDisplay: 'symbol' }).format(cedis),
  date: (d: string | Date) => new Intl.DateTimeFormat(LOCALE, { dateStyle: 'medium' }).format(new Date(d)),
  dateTime: (d: string | Date) =>
    new Intl.DateTimeFormat(LOCALE, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(d)),
}
