import { defineConfig } from 'vocs'

export default defineConfig({
  title: 'Docs',
  description: 'The protocol for open audio distribution.',
  rootDir: 'src',
  sidebar: [
    {
      text: 'Whitepaper',
      link: '/whitepaper',
    },
    {
      text: 'Yellowpaper',
      link: '/yellowpaper',
    },
  ],
})
