import { defineConfig } from 'vocs'

export default defineConfig({
  title: 'Sonata',
  theme: {
    accentColor: {
      light: '#7a8f73', // sage/olive
      dark: '#d9e2c7',  // light sage
    },

    variables: {
      color: {
        background: {
          light: '#ffffff',
          dark: '#2b2b2b'
        },
        background2: {
          light: '#f7f7f7',
          dark: '#1e1e1e'
        },
        background3: {
          light: '#e9dcc5',   // soft beige
          dark: '#3f4f3a'     // deep forest accent
        },
        border: {
          light: '#e2e2e2',
          dark: '#3f3f3f'
        },
        text: {
          light: '#3f4f3a',   // readable forest green
          dark: '#d9e2c7'
        },
        textAccent: {
          light: '#7a8f73',   // muted olive
          dark: '#d9e2c7'
        },
        backgroundAccent: {
          light: '#d9e2c7',
          dark: '#3f4f3a'
        },
        backgroundAccentHover: {
          light: '#cbdab5',
          dark: '#4d6046'
        },
        backgroundAccentText: {
          light: '#3f4f3a',
          dark: '#d9e2c7'
        },
        link: {
          light: '#6d553e',   // soft brown
          dark: '#e9dcc5'
        },
        linkHover: {
          light: '#3f4f3a',
          dark: '#d9e2c7'
        },
        heading: {
          light: '#6d553e',   // warm brown headings
          dark: '#e9dcc5'
        },
      },
    },
  },
  sidebar: [
    {
      text: 'Mission',
      items: [
        { text: 'Introduction', link: '/mission/introduction' },
        { text: 'The Problem', link: '/mission/the-problem' },
        { text: 'Why Sonata', link: '/mission/why-sonata' },
        { text: 'Philosophy', link: '/mission/philosophy' },
      ],
    },
    {
      text: 'Value',
      items: [
        { text: 'Overview', link: '/value/overview' },
        { text: 'USDC Purchases', link: '/value/usdc-purchases' },
        { text: 'Portable Wallets', link: '/value/portable-wallets' },
        { text: 'DDEX Native', link: '/value/ddex-native' },
        { text: 'Encrypted Files', link: '/value/encrypted-files' },
        { text: 'Ownership & Portability', link: '/value/ownership-and-portability' },
        { text: 'Artist Benefits', link: '/value/artist-benefits' },
        { text: 'Listener Benefits', link: '/value/listener-benefits' },
      ],
    },
    {
      text: 'Architecture',
      items: [
        { text: 'Overview', link: '/architecture/overview' },
        { text: 'Components', link: '/architecture/components' },
        { text: 'Data Flow', link: '/architecture/data-flow' },
        { text: 'Accounts & Auth', link: '/architecture/accounts-and-auth' },
        { text: 'Transactions & Attestations', link: '/architecture/transactions-and-attestations' },
        { text: 'File Lifecycle', link: '/architecture/file-lifecycle' },
        { text: 'Chain Overview', link: '/architecture/chain-overview' },
      ],
    },
    {
      text: 'Internals',
      items: [
        {
          text: 'Chain',
          items: [
            { text: 'CometBFT Overview', link: '/internals/chain/cometbft-overview' },
            { text: 'Notes Module', link: '/internals/chain/notes-module' },
            { text: 'Compositions', link: '/internals/chain/compositions' },
            { text: 'State Machine', link: '/internals/chain/state-machine' },
            { text: 'RPC & SDK', link: '/internals/chain/rpc-and-sdk' },
          ],
        },
        {
          text: 'Storage',
          items: [
            { text: 'Overview', link: '/internals/storage/overview' },
            { text: 'File Normalization', link: '/internals/storage/file-normalization' },
            { text: 'Watermarking', link: '/internals/storage/watermarking' },
            { text: 'Replication', link: '/internals/storage/replication' },
            { text: 'Rendezvous Hashing', link: '/internals/storage/rendezvous-hashing' },
            { text: 'BitTorrent', link: '/internals/storage/bittorrent' },
            { text: 'GoCloud Backends', link: '/internals/storage/gocloud-backends' },
          ],
        },
        {
          text: 'DDEX',
          items: [
            { text: 'DDEX Overview', link: '/internals/ddex/ddex-overview' },
            { text: 'Ingestion Pipeline', link: '/internals/ddex/ingestion-pipeline' },
            { text: 'Metadata Enrichment', link: '/internals/ddex/metadata-enrichment' },
            { text: 'Compositions for Access', link: '/internals/ddex/compositions-for-access' },
            { text: 'Attestations', link: '/internals/ddex/attestations' },
          ],
        },
        {
          text: 'Crypto',
          items: [
            { text: 'OpenTDF Overview', link: '/internals/crypto/opentdf-overview' },
            { text: 'Encryption at Rest', link: '/internals/crypto/encryption-at-rest' },
            { text: 'FROST Keys', link: '/internals/crypto/frost-keys' },
            { text: 'Rotation Schemes', link: '/internals/crypto/rotation-schemes' },
            { text: 'Decryption Events', link: '/internals/crypto/decryption-events' },
          ],
        },
        {
          text: 'Protocol',
          items: [
            { text: 'Overview', link: '/internals/protocol/overview' },
            { text: 'Protobuf Schema', link: '/internals/protocol/protobuf-schema' },
            { text: 'ConnectRPC', link: '/internals/protocol/connectrpc' },
            { text: 'Buf Tooling', link: '/internals/protocol/buf-tooling' },
            { text: 'Code Generation', link: '/internals/protocol/code-generation' },
          ],
        },
      ],
    },
    {
      text: 'Run',
      items: [
        {
          text: 'Validators',
          items: [
            { text: 'Requirements', link: '/run/validators/requirements' },
            { text: 'Setup', link: '/run/validators/setup' },
            { text: 'Running', link: '/run/validators/running' },
            { text: 'Monitoring', link: '/run/validators/monitoring' },
            { text: 'Upgrades', link: '/run/validators/upgrades' },
          ],
        },
        {
          text: 'Client',
          items: [
            { text: 'Embedded Player', link: '/run/client/embedded-player' },
            { text: 'Phantom Login', link: '/run/client/phantom-login' },
            { text: 'Note Balances', link: '/run/client/note-balances' },
            { text: 'Purchasing', link: '/run/client/purchasing' },
          ],
        },
      ],
    },
    {
      text: 'Economics',
      items: [
        { text: 'Fundamentals', link: '/economics/fundamentals' },
        { text: 'Fees', link: '/economics/fees' },
        { text: 'Inflationary Notes', link: '/economics/inflationary-notes' },
        { text: 'Validator Rewards', link: '/economics/validator-rewards' },
        { text: 'Artist Revenue Splits', link: '/economics/artist-revenue-splits' },
        { text: 'Sustainability', link: '/economics/sustainability' },
        { text: 'USDC Bridging', link: '/economics/usdc-bridging' },
      ],
    },
    {
      text: 'Appendix',
      items: [
        { text: 'Glossary', link: '/appendix/glossary' },
        { text: 'FAQ', link: '/appendix/faq' },
        { text: 'Whitepaper Draft', link: '/appendix/whitepaper-draft' },
      ],
    },
  ],
})
