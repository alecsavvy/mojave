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
  ogImageUrl: '/og.png',
  sidebar: [
    {
      text: 'One Pager',
      link: '/onepager',
    },
    {
      text: 'Primitives',
      items: [
        { text: 'DDEX', link: '/primitives/ddex' },
        { text: 'CometBFT', link: '/primitives/cometbft' },
        { text: 'OpenTDF', link: '/primitives/opentdf' },
        { text: 'BitTorrent', link: '/primitives/bittorrent' },
        { text: 'ConnectRPC', link: '/primitives/connectrpc' },
      ],
    },
    {
      text: 'Architecture',
      items: [
        { text: 'Overview', link: '/architecture/overview' },
        { text: 'Metadata', link: '/architecture/metadata' },
        { text: 'Upload', link: '/architecture/upload' },
        { text: 'File Replication', link: '/architecture/file-replication' },
      ],
    },
    {
      text: 'Economics',
      items: [
        { text: 'VOX', link: '/economics/vox' },
        { text: 'Purchases', link: '/economics/purchases' },
      ]
    },
    {
      text: 'Actors',
      items: [
        { text: 'Operators', items: [
          { text: 'Validators', link: '/actors/validators' },
          { text: 'Archivers', link: '/actors/archivers' },
          { text: 'Oracles', link: '/actors/oracles' },
        ]},
        { text: 'Producers', items: [
          { text: 'Distributors', link: '/actors/distributors' },
          { text: 'Labels', link: '/actors/labels' },
          { text: 'Artists', link: '/actors/artists' },
        ]},
        { text: 'Consumers', items: [
          { text: 'DSPs', link: '/actors/dsps' },
          { text: 'Listeners', link: '/actors/listeners' },
        ]},
      ],
    },
    {
      text: 'Glossary',
      link: '/glossary',
    },
  ],
})
