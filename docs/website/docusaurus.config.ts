import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'AI-Pack',
  tagline: 'A comprehensive AI agent workflow framework for software development',
  favicon: 'img/favicon.ico',

  // Set the production url of your site here
  // For GitHub Pages: https://YOUR_ORG.github.io
  // For custom domain: https://yourdomain.com
  url: 'https://cortexa-llc.github.io',

  // For GitHub Pages with repo: /YOUR_REPO_NAME/
  // For custom domain or org.github.io repo: /
  baseUrl: '/ai-pack/',

  organizationName: 'Cortexa-LLC', // Usually your GitHub org/user name
  projectName: 'ai-pack', // Usually your repo name

  onBrokenLinks: 'warn',
  onBrokenAnchors: 'warn',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  themes: ['@docusaurus/theme-mermaid'],

  markdown: {
    mermaid: true,
  },

  presets: [
    [
      'classic',
      {
        docs: {
          // Path to your documentation content (relative to website/)
          path: '../content',
          routeBasePath: 'docs',
          sidebarPath: './sidebars.ts',
          // Edit this page link
          editUrl: 'https://github.com/Cortexa-LLC/ai-pack/tree/main/docs/',
        },
        blog: false, // Disable blog
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: 'img/social-card.png',
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'AI-Pack',
      logo: {
        alt: 'AI-Pack Logo',
        src: 'img/logo-navbar.png',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'tutorialSidebar',
          position: 'left',
          label: 'Documentation',
        },
        {
          href: 'https://github.com/Cortexa-LLC/ai-pack',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Documentation',
          items: [
            {
              label: 'Getting Started',
              to: '/docs/',
            },
            {
              label: 'Workflows',
              to: '/docs/workflows/bugfix',
            },
            {
              label: 'Roles',
              to: '/docs/roles/archaeologist',
            },
          ],
        },
        {
          title: 'Resources',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/Cortexa-LLC/ai-pack',
            },
            {
              label: 'Clean Code Standards',
              to: '/docs/quality/clean-code/general-rules',
            },
          ],
        },
      ],
      copyright: `Copyright © 2025-${new Date().getFullYear()} Cortexa LLC. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'python', 'javascript', 'typescript', 'json', 'yaml'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
