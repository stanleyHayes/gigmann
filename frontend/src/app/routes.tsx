import type { RouteObject } from 'react-router-dom'

import { ApprovalsScreen } from '../screens/ApprovalsScreen'
import { HomeScreen } from '../screens/HomeScreen'
import { KpisScreen } from '../screens/KpisScreen'
import { NetworkScreen } from '../screens/NetworkScreen'
import { Placeholder } from '../screens/Placeholder'
import { AppShell } from './AppShell'

export const routes: RouteObject[] = [
  {
    path: '/',
    Component: AppShell,
    children: [
      { index: true, Component: HomeScreen },
      { path: 'network', Component: NetworkScreen },
      { path: 'kpis', Component: KpisScreen },
      { path: 'ask', element: <Placeholder title="Ask" note="Natural-language query and generated docs — planned in GEC-60." /> },
      { path: 'my-day', element: <Placeholder title="My Day" note="Planned in GEC-61." /> },
      { path: 'approvals', Component: ApprovalsScreen },
      { path: '*', element: <Placeholder title="Not found" note="That page does not exist." /> },
    ],
  },
]
