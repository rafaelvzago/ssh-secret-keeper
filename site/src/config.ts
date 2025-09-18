// Configuration for SSH Secret Keeper Site
// Open-source community project

export const config = {
  app: {
    name: 'SSH Secret Keeper',
    version: '1.0.1',
    status: 'Open Source',
  },
  github: {
    repo: 'rafaelvzago/ssh-secret-keeper',
    url: 'https://github.com/rafaelvzago/ssh-secret-keeper',
  },
  install: {
    scriptUrl: 'https://raw.githubusercontent.com/rafaelvzago/ssh-secret-keeper/refs/heads/main/install.sh',
    rawScriptUrl: 'https://raw.githubusercontent.com/rafaelvzago/ssh-secret-keeper/refs/heads/main/install.sh',
    downloadBase: 'https://github.com/rafaelvzago/ssh-secret-keeper/releases/latest/download/',
  },
  docker: {
    repository: 'rafaelvzago/ssh-secret-keeper',
    registry: 'docker.io',
  },
} as const;

// Development helpers
export const isDevelopment = false;
export const showPlaceholders = false;
