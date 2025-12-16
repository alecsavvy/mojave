import { defineConfig } from 'vocs'

export default defineConfig({
  title: 'Mojave',
  description: 'The protocol for open audio distribution.',
  rootDir: 'src',
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
