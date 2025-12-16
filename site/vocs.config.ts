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
    // colorScheme not set - enables manual dark/light mode toggle
  },
  sidebar: [
    {
      text: 'One Pager',
      link: '/onepager',
    },
    {
      text: 'Whitepaper',
      link: '/whitepaper',
    },
    {
      text: 'Yellowpaper',
      link: '/yellowpaper',
    },
    {
      text: 'Glossary',
      link: '/glossary',
    },
  ],
})
