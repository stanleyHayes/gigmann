// English (Ghana) message catalog. UI strings live here so they can be swapped per
// locale; `t` is the single lookup point. (No framework yet — react-i18next/Lingui
// can wrap this when a second locale is actually required.)
export const messages = {
  'nav.today': 'Today',
  'nav.network': 'Network',
  'nav.kpis': 'KPIs',
  'nav.reports': 'Reports',
  'nav.ask': 'Ask',
  'nav.myDay': 'My Day',
  'nav.delegation': 'Delegation',
  'nav.approvals': 'Approvals',
  'brief.source.claude': 'Narrated by Claude',
  'brief.source.local': 'Deterministic summary — AI narration unavailable',
} as const

export type MessageKey = keyof typeof messages

/** t looks up a UI string by key (en-GH). */
export function t(key: MessageKey): string {
  return messages[key]
}
