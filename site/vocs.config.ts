import { defineConfig } from 'vocs'

export default defineConfig({
  title: 'Mojave',
  description: 'The protocol for open audio distribution.',
  rootDir: 'src',
  theme: {
    accentColor: {
      light: '#d97706', // Warm amber/orange for Mojave desert theme
      dark: '#f59e0b',  // Brighter amber for dark mode visibility
    },
  },
  iconUrl: '/logo.png',
  sidebar: [
    {
      text: 'One Pager',
      link: '/onepager',
    },
    {
      text: 'Glossary',
      link: '/glossary',
    },
  ],
})
