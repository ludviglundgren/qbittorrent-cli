// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
  // GitHub Pages project site: served from https://<user>.github.io/<repo>/
  site: 'https://ludviglundgren.github.io',
  base: '/qbittorrent-cli',
  integrations: [
    starlight({
      title: 'qbittorrent-cli',
      description:
        'A CLI to manage qBittorrent - add torrents, manage categories and tags, reannounce, import from other clients and more.',
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/ludviglundgren/qbittorrent-cli',
        },
      ],
      editLink: {
        baseUrl:
          'https://github.com/ludviglundgren/qbittorrent-cli/edit/master/docs/',
      },
      sidebar: [
        {
          label: 'Getting started',
          items: [
            { label: 'Installation', slug: 'getting-started/installation' },
            { label: 'Configuration', slug: 'getting-started/configuration' },
          ],
        },
        {
          label: 'Command reference',
          items: [{ autogenerate: { directory: 'commands' } }],
        },
      ],
    }),
  ],
});
