import type { RouteObject } from 'react-router-dom'

import { HomeScreen } from '../screens/HomeScreen'
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
      { path: 'kpis', element: <Placeholder title="Executive KPIs" note="Planned in GEC-59." /> },
      { path: 'ask', element: <Placeholder title="Ask" note="Natural-language query and generated docs — planned in GEC-60." /> },
      { path: 'my-day', element: <Placeholder title="My Day" note="Planned in GEC-61." /> },
      { path: 'approvals', element: <Placeholder title="Approvals" note="Planned in GEC-62." /> },
      { path: '*', element: <Placeholder title="Not found" note="That page does not exist." /> },
    ],
  },
]
