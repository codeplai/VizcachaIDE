import { mount } from 'svelte'
import App from './App.svelte'
import { setupI18n } from './lib/i18n'
import './lib/theme/fonts'
import './lib/theme/tokens.css'
import './lib/theme/ansi.css'
import './lib/theme/base.css'

setupI18n('auto')

const app = mount(App, { target: document.getElementById('app')! })

export default app
