const headerPattern = /^(?:[^\w\s]+\s)?(\w+)(?:\(([^)]*)\))?!?: (.+)$/u;

export default {
  branches: ['main'],
  plugins: [
    [
      '@semantic-release/commit-analyzer',
      {
        preset: 'conventionalcommits',
        parserOpts: { headerPattern },
      },
    ],
    [
      '@semantic-release/release-notes-generator',
      {
        preset: 'conventionalcommits',
        parserOpts: { headerPattern },
      },
    ],
    '@semantic-release/changelog',
    [
      '@semantic-release/exec',
      {
        publishCmd: 'task release:publish VERSION=${nextRelease.version}',
      },
    ],
    '@semantic-release/github',
    [
      '@semantic-release/git',
      {
        assets: ['CHANGELOG.md'],
        message: '🔖 chore(release): ${nextRelease.version} [skip ci]',
      },
    ],
  ],
};
