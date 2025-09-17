// Development configuration for SSH Secret Keeper Site
// No versioning during development phase

export const config = {
  app: {
    name: 'SSH Secret Keeper',
    status: 'Development',
  },
  github: {
    repo: 'rafaelvzago/ssh-secret-keeper',
    url: 'https://github.com/rafaelvzago/ssh-secret-keeper',
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
