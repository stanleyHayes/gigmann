import '@fontsource-variable/fraunces'
import '@fontsource-variable/outfit'
import '@fontsource-variable/jetbrains-mono'

import React from 'react'
import ReactDOM from 'react-dom/client'

import { App } from './App'
import { ReloadPrompt } from './app/ReloadPrompt'

const rootElement = document.getElementById('root')
if (!rootElement) throw new Error('root element not found')

ReactDOM.createRoot(rootElement).render(
  <React.StrictMode>
    <App />
    <ReloadPrompt />
  </React.StrictMode>,
)
