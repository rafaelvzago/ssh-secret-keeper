// Development configuration for SSH Vault Keeper Site
// No versioning during development phase

export const config = {
  app: {
    name: 'SSH Vault Keeper',
    status: 'Development',
  },
  github: {
    repo: 'rzago/ssh-vault-keeper',
    url: 'https://github.com/rzago/ssh-vault-keeper',
  },
  install: {
    // Placeholder URLs for development - will be updated when ready
    scriptUrl: '#',
    downloadBase: '#',
  },
} as const;

// Development helpers
export const isDevelopment = true;
export const showPlaceholders = true;
