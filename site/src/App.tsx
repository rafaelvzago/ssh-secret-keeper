import React, { useState, useEffect } from 'react';
import { Terminal, Shield, Key, Server, Copy, Check, Github, Book, Zap, Lock } from 'lucide-react';
import { config } from './config';

const TypewriterText: React.FC<{ text: string; delay?: number }> = ({ text, delay = 50 }) => {
  const [displayText, setDisplayText] = useState('');
  const [currentIndex, setCurrentIndex] = useState(0);

  useEffect(() => {
    if (currentIndex < text.length) {
      const timeout = setTimeout(() => {
        setDisplayText(prev => prev + text[currentIndex]);
        setCurrentIndex(prev => prev + 1);
      }, delay);
      return () => clearTimeout(timeout);
    }
  }, [currentIndex, text, delay]);

  return <span>{displayText}<span className="animate-pulse text-green-400">▊</span></span>;
};

const CopyButton: React.FC<{ text: string }> = ({ text }) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <button
      onClick={handleCopy}
      className="ml-3 p-2 rounded bg-gray-800 hover:bg-gray-700 transition-colors border border-gray-600"
      title="Copy to clipboard"
    >
      {copied ? <Check className="w-4 h-4 text-green-400" /> : <Copy className="w-4 h-4 text-cyan-400" />}
    </button>
  );
};

const TerminalWindow: React.FC<{ title: string; children: React.ReactNode }> = ({ title, children }) => (
  <div className="bg-gray-900 border border-gray-700 rounded-lg overflow-hidden">
    <div className="bg-gray-800 px-4 py-2 flex items-center gap-2 border-b border-gray-700">
      <div className="flex gap-1">
        <div className="w-3 h-3 rounded-full bg-red-500"></div>
        <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
        <div className="w-3 h-3 rounded-full bg-green-500"></div>
      </div>
      <span className="text-gray-300 text-sm font-mono">{title}</span>
    </div>
    <div className="p-4">
      {children}
    </div>
  </div>
);

const ASCIIBorder: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div className="font-mono text-cyan-400 text-xs">
    <div>┌─────────────────────────────────────────────────────────────────────┐</div>
    <div className="px-2 py-4 text-base text-white">
      {children}
    </div>
    <div>└─────────────────────────────────────────────────────────────────────┘</div>
  </div>
);

function App() {
  const [activeTab, setActiveTab] = useState('installation');

  const installCommand = "curl -sSL https://raw.githubusercontent.com/rafaelvzago/ssh-secret-keeper/refs/heads/main/install.sh | bash";

  const features = [
    {
      icon: <Shield className="w-8 h-8 text-green-400" />,
      title: "Backup SSH Directory",
      description: "Complete backup of your ~/.ssh folder with permissions preserved. Securely stores keys, config files, and authorized_keys to HashiCorp Vault with client-side encryption."
    },
    {
      icon: <Server className="w-8 h-8 text-amber-400" />,
      title: "Cross-Machine Restore",
      description: "NEW: Backup on laptop, restore on desktop! Universal storage enables cross-machine and cross-user restore. Perfect for developers working across multiple machines."
    },
    {
      icon: <Key className="w-8 h-8 text-cyan-400" />,
      title: "Flexible Storage Strategies",
      description: "Choose from universal (shared), user-scoped, machine-user (legacy), or custom storage strategies. Migration tools help upgrade existing installations."
    },
    {
      icon: <Zap className="w-8 h-8 text-red-400" />,
      title: "Smart Analysis & Management",
      description: "Intelligent SSH analysis with service categorization. List, analyze, and manage backups with comprehensive CLI tools and validation."
    }
  ];

  return (
    <div className="min-h-screen bg-gray-900 text-white font-mono">
      {/* Header */}
      <header className="border-b border-gray-800 bg-gray-900/95 backdrop-blur-sm sticky top-0 z-50">
        <div className="max-w-6xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Terminal className="w-8 h-8 text-green-400" />
            <div>
              <h1 className="text-xl font-bold text-green-400">SSH Secret Keeper</h1>
              <div className="text-xs text-gray-400">v{config.app.version}</div>
            </div>
          </div>
          <nav className="hidden md:flex items-center gap-6">
            <a href="#features" className="text-cyan-400 hover:text-cyan-300 transition-colors">Features</a>
            <a href="#installation" className="text-cyan-400 hover:text-cyan-300 transition-colors">Install</a>
            <a href="#docs" className="text-cyan-400 hover:text-cyan-300 transition-colors">Documentation</a>
            <a href={config.github.url} target="_blank" rel="noopener noreferrer" className="text-gray-400 hover:text-gray-300 transition-colors">
              <Github className="w-5 h-5" />
            </a>
          </nav>
        </div>
      </header>

      {/* Hero Section */}
      <section className="py-20 px-4">
        <div className="max-w-4xl mx-auto text-center">
          <ASCIIBorder>
            <div className="space-y-6">
              <h2 className="text-4xl md:text-6xl font-bold text-green-400 mb-4">
                <TypewriterText text="SSH Secret Keeper" delay={100} />
              </h2>
              <p className="text-xl text-cyan-400 leading-relaxed">
                Securely backup your ~/.ssh folder to HashiCorp Vault with <span className="text-green-400 font-semibold">cross-machine restore</span>.<br />
                Backup on laptop, restore on desktop! Perfect for developers, system administrators,<br />
                and DevOps teams who work across multiple machines and environments.
              </p>
              <div className="flex justify-center gap-4 mt-8">
                <button className="bg-green-600 hover:bg-green-500 px-6 py-3 rounded border border-green-500 transition-colors">
                  Get Started
                </button>
                <button className="border border-cyan-500 text-cyan-400 hover:bg-cyan-900/20 px-6 py-3 rounded transition-colors">
                  View Documentation
                </button>
              </div>
            </div>
          </ASCIIBorder>
        </div>
      </section>

      {/* Installation Section */}
      <section id="installation" className="py-16 px-4 bg-gray-800/30">
        <div className="max-w-4xl mx-auto">
          <h3 className="text-3xl font-bold text-center mb-12 text-green-400">
            ┌─[ Installation ]─┐
          </h3>

          <TerminalWindow title="bash">
            <div className="space-y-4">
              <div className="flex items-center">
                <span className="text-green-400">$</span>
                <code className="ml-2 text-white flex-1">{installCommand}</code>
                <CopyButton text={installCommand} />
              </div>
              <div className="text-gray-400 text-sm">
                # Quick installation script - detects your OS and installs automatically
              </div>
            </div>
          </TerminalWindow>

          <div className="mt-8 grid md:grid-cols-3 gap-6">
            <div className="bg-gray-900 border border-gray-700 rounded p-4">
              <h4 className="text-cyan-400 font-bold mb-2">Quick Install</h4>
              <p className="text-gray-300 text-sm">Single command installation script</p>
            </div>
            <div className="bg-gray-900 border border-gray-700 rounded p-4">
              <h4 className="text-cyan-400 font-bold mb-2">Container Ready</h4>
              <p className="text-gray-300 text-sm">Docker and Podman support</p>
            </div>
            <div className="bg-gray-900 border border-gray-700 rounded p-4">
              <h4 className="text-cyan-400 font-bold mb-2">Build from Source</h4>
              <p className="text-gray-300 text-sm">See documentation for details</p>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="py-16 px-4">
        <div className="max-w-6xl mx-auto">
          <h3 className="text-3xl font-bold text-center mb-12 text-green-400">
            ┌─[ Core Features ]─┐
          </h3>

          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
            {features.map((feature, index) => (
              <div key={index} className="bg-gray-800 border border-gray-700 rounded-lg p-6 hover:border-gray-600 transition-colors">
                <div className="flex items-center gap-3 mb-4">
                  {feature.icon}
                  <h4 className="text-lg font-bold text-white">{feature.title}</h4>
                </div>
                <p className="text-gray-300 text-sm leading-relaxed">
                  {feature.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Documentation Section */}
      <section id="docs" className="py-16 px-4 bg-gray-800/30">
        <div className="max-w-7xl mx-auto">
          <h3 className="text-3xl font-bold text-center mb-12 text-green-400">
            ┌─[ Documentation ]─┐
          </h3>

          <div className="flex justify-center mb-8">
            <div className="bg-gray-900 border border-gray-700 rounded-lg p-1 inline-flex">
              {[
                { id: 'installation', label: 'Installation' },
                { id: 'build', label: 'Build from Source' },
                { id: 'docker', label: 'Docker Usage' },
                { id: 'backup', label: 'Backup Guide' },
                { id: 'restore', label: 'Restore Guide' },
                { id: 'storage', label: 'Storage Strategies' },
                { id: 'commands', label: 'All Commands' },
                { id: 'config', label: 'Configuration' }
              ].map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`px-4 py-2 rounded text-sm font-medium transition-colors ${
                    activeTab === tab.id
                      ? 'bg-gray-700 text-cyan-400'
                      : 'text-gray-400 hover:text-gray-300'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </div>
          </div>

          {/* Full Width Documentation Content */}
          <div className="w-full mb-12">
              {activeTab === 'installation' && (
                <TerminalWindow title="Manual Installation">
                  <div className="space-y-4 text-sm">
                    <div className="text-yellow-400 font-bold">## Manual Binary Download</div>
                    <div className="text-gray-400"># Download for your platform</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">curl -L https://github.com/rafaelvzago/ssh-secret-keeper/releases/latest/download/ssh-secret-keeper-VERSION-linux-amd64.tar.gz -o sshsk.tar.gz</span>
                    </div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">tar -xzf sshsk.tar.gz</span>
                    </div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">chmod +x sshsk</span>
                    </div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sudo mv sshsk /usr/local/bin/</span>
                    </div>

                    <div className="text-yellow-400 font-bold">## Verify Installation</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk --version</span>
                    </div>
                    <div className="text-cyan-400">SSH Secret Keeper v{config.app.version}</div>

                    <div className="bg-blue-900/30 border border-blue-500/30 rounded p-2 mt-4">
                      <div className="text-blue-400 text-xs">
                        💡 For easier installation, use the quick install script above
                      </div>
                    </div>
                  </div>
                </TerminalWindow>
              )}

              {activeTab === 'build' && (
                <TerminalWindow title="Build from Source">
                  <div className="space-y-4 text-sm">
                    <div className="text-yellow-400 font-bold">## Install Dependencies</div>
                    <div className="text-gray-400"># Arch Linux</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sudo pacman -S go make git</span>
                    </div>

                    <div className="text-gray-400"># Ubuntu/Debian</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sudo apt-get install golang-go make git</span>
                    </div>

                    <div className="text-gray-400"># Fedora/RHEL</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sudo dnf install golang make git</span>
                    </div>

                    <div className="text-yellow-400 font-bold">## Clone and Build</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">git clone https://github.com/rafaelvzago/ssh-secret-keeper.git</span>
                    </div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">cd ssh-secret-keeper</span>
                    </div>

                    <div className="text-gray-400"># Build for current platform</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">make build</span>
                    </div>

                    <div className="text-gray-400"># Install to /usr/local/bin</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sudo make install</span>
                    </div>

                    <div className="text-gray-400"># Verify installation</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk --version</span>
                    </div>
                    <div className="text-cyan-400">✓ SSH Secret Keeper v{config.app.version}</div>
                  </div>
                </TerminalWindow>
              )}

              {activeTab === 'docker' && (
                <TerminalWindow title="Docker Usage">
                  <div className="space-y-4 text-sm">
                    <div className="text-yellow-400 font-bold">## Pull from Docker Hub</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">docker pull rafaelvzago/ssh-secret-keeper:latest</span>
                    </div>

                    <div className="text-yellow-400 font-bold">## Basic Usage</div>
                    <div className="text-gray-400"># Analyze SSH directory</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">docker run --rm -v ~/.ssh:/ssh:ro \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">-e VAULT_ADDR="https://your-vault:8200" \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">-e VAULT_TOKEN="your-token" \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">rafaelvzago/ssh-secret-keeper:latest analyze</span>
                    </div>

                    <div className="text-gray-400"># Quick backup (auto-generated name)</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">docker run --rm -v ~/.ssh:/ssh:ro \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">-e VAULT_ADDR="https://your-vault:8200" \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">-e VAULT_TOKEN="your-token" \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">rafaelvzago/ssh-secret-keeper:latest backup</span>
                    </div>

                    <div className="text-gray-400"># Named backup</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">docker run --rm -v ~/.ssh:/ssh:ro \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">-e VAULT_ADDR="https://your-vault:8200" \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">-e VAULT_TOKEN="your-token" \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">rafaelvzago/ssh-secret-keeper:latest backup "container-$(date +%Y%m%d)"</span>
                    </div>

                    <div className="text-gray-400"># Dry run mode</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">docker run --rm -v ~/.ssh:/ssh:ro \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">rafaelvzago/ssh-secret-keeper:latest backup --dry-run</span>
                    </div>

                    <div className="text-gray-400"># Restore SSH keys</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">docker run --rm -v ~/.ssh:/ssh \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">-e VAULT_ADDR="https://your-vault:8200" \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">-e VAULT_TOKEN="your-token" \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">rafaelvzago/ssh-secret-keeper:latest restore</span>
                    </div>

                    <div className="text-yellow-400 font-bold">## NEW: Storage Strategy Configuration</div>
                    <div className="text-gray-400"># Use universal storage (default - cross-machine)</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">docker run --rm -v ~/.ssh:/ssh:ro \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">-e VAULT_ADDR="https://your-vault:8200" \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">-e VAULT_TOKEN="your-token" \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">-e SSH_VAULT_STORAGE_STRATEGY="universal" \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">rafaelvzago/ssh-secret-keeper:latest backup "cross-machine-backup"</span>
                    </div>

                    <div className="text-gray-400"># Use legacy machine-user storage</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">docker run --rm -v ~/.ssh:/ssh:ro \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">-e SSH_VAULT_STORAGE_STRATEGY="machine-user" \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">rafaelvzago/ssh-secret-keeper:latest backup "isolated-backup"</span>
                    </div>

                    <div className="text-yellow-400 font-bold">## Using Podman</div>
                    <div className="text-gray-400"># Same commands, just replace 'docker' with 'podman'</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">podman run --rm -v ~/.ssh:/ssh:ro \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">-e VAULT_ADDR="https://your-vault:8200" \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">-e VAULT_TOKEN="your-token" \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">-e SSH_VAULT_STORAGE_STRATEGY="universal" \</span>
                    </div>
                    <div>
                      <span className="ml-4 text-white">rafaelvzago/ssh-secret-keeper:latest analyze</span>
                    </div>
                  </div>
                </TerminalWindow>
              )}

              {activeTab === 'backup' && (
                <TerminalWindow title="🔒 SSH Backup Guide">
                  <div className="space-y-4 text-sm">
                    <div className="bg-green-900/30 border border-green-500/30 rounded p-3">
                      <div className="text-green-400 font-bold mb-2">📋 Complete Backup Workflow</div>
                      <div className="text-gray-300 text-xs">
                        Securely backup your entire SSH directory to HashiCorp Vault with cross-machine restore capability
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Step 1: Initialize</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">export VAULT_ADDR="https://vault.company.com:8200"</span>
                    </div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">export VAULT_TOKEN="your-vault-token"</span>
                    </div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk init</span>
                    </div>

                    <div className="text-yellow-400 font-bold">## Step 2: Analyze (Optional)</div>
                    <div className="text-gray-400"># See what will be backed up</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk analyze --verbose</span>
                    </div>

                    <div className="text-yellow-400 font-bold">## Step 3: Backup Modes</div>
                    
                    <div className="bg-green-900/20 border border-green-500/30 rounded p-3">
                      <div className="text-green-400 font-bold">🚀 Quick Backup Mode</div>
                      <div className="text-gray-300 text-xs mb-2">Backup everything with auto-generated name</div>
                      <div>
                        <span className="text-green-400">$</span>
                        <span className="ml-2 text-white">sshsk backup</span>
                      </div>
                      <div className="text-gray-400 text-xs mt-1">Creates: backup-YYYYMMDD-HHMMSS</div>
                    </div>

                    <div className="bg-blue-900/20 border border-blue-500/30 rounded p-3">
                      <div className="text-blue-400 font-bold">📝 Named Backup Mode</div>
                      <div className="text-gray-300 text-xs mb-2">Backup with custom name for organization</div>
                      <div>
                        <span className="text-green-400">$</span>
                        <span className="ml-2 text-white">sshsk backup "laptop-$(date +%Y%m%d)"</span>
                      </div>
                      <div>
                        <span className="text-green-400">$</span>
                        <span className="ml-2 text-white">sshsk backup "dev-environment-keys"</span>
                      </div>
                    </div>

                    <div className="bg-purple-900/20 border border-purple-500/30 rounded p-3">
                      <div className="text-purple-400 font-bold">🎯 Interactive Selection Mode</div>
                      <div className="text-gray-300 text-xs mb-2">Choose specific files to backup</div>
                      <div>
                        <span className="text-green-400">$</span>
                        <span className="ml-2 text-white">sshsk backup --interactive "selective-backup"</span>
                      </div>
                      <div className="text-gray-400 text-xs mt-1">Prompts for each file to include/exclude</div>
                    </div>

                    <div className="bg-yellow-900/20 border border-yellow-500/30 rounded p-3">
                      <div className="text-yellow-400 font-bold">👁️ Preview Mode (Dry Run)</div>
                      <div className="text-gray-300 text-xs mb-2">See what would be backed up without doing it</div>
                      <div>
                        <span className="text-green-400">$</span>
                        <span className="ml-2 text-white">sshsk backup --dry-run</span>
                      </div>
                      <div>
                        <span className="text-green-400">$</span>
                        <span className="ml-2 text-white">sshsk backup --dry-run "test-backup"</span>
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Step 4: Advanced Options</div>
                    <div className="text-gray-400"># Custom SSH directory</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk backup --ssh-dir "/custom/ssh/path"</span>
                    </div>

                    <div className="text-gray-400"># Combine options</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk backup --interactive --dry-run "preview-selective"</span>
                    </div>

                    <div className="text-yellow-400 font-bold">## Step 5: File Filtering & Categories</div>
                    
                    <div className="bg-amber-900/20 border border-amber-500/30 rounded p-3">
                      <div className="text-amber-400 font-bold mb-2">📁 Include/Exclude Patterns</div>
                      <div className="text-gray-300 text-xs mb-2">Configure which files to backup (config.yaml)</div>
                      <div className="text-white text-xs">
                        <div className="text-cyan-400">include_patterns:</div>
                        <div className="ml-2">- "*.rsa"        # RSA private keys</div>
                        <div className="ml-2">- "*.pub"        # Public keys</div>
                        <div className="ml-2">- "config"       # SSH config file</div>
                        <div className="ml-2">- "known_hosts*" # Known hosts</div>
                        <div className="ml-2">- "*work*"       # Work-related keys</div>
                        <br/>
                        <div className="text-cyan-400">exclude_patterns:</div>
                        <div className="ml-2">- "*.tmp"        # Temporary files</div>
                        <div className="ml-2">- "*.bak"        # Backup files</div>
                        <div className="ml-2">- "*.old"        # Old files</div>
                      </div>
                    </div>

                    <div className="bg-indigo-900/20 border border-indigo-500/30 rounded p-3">
                      <div className="text-indigo-400 font-bold mb-2">🏷️ File Categories</div>
                      <div className="text-gray-300 text-xs mb-2">Organize keys by purpose for better management</div>
                      <div className="text-white text-xs">
                        <div className="text-cyan-400">categories:</div>
                        <div className="ml-2">services: ["github_rsa", "gitlab_rsa"]</div>
                        <div className="ml-2">personal: ["id_rsa", "local_rsa"]</div>
                        <div className="ml-2">work: ["*work*", "*corp*", "*office*"]</div>
                        <div className="ml-2">cloud: ["*aws*", "*gcp*", "*azure*"]</div>
                      </div>
                    </div>

                    <div className="bg-blue-900/30 border border-blue-500/30 rounded p-2 mt-4">
                      <div className="text-blue-400 text-xs">
                        ✓ Preserves exact permissions (0600/0644)<br/>
                        ✓ Includes MD5 checksums for integrity<br/>
                        ✓ Stores metadata (timestamps, file sizes)<br/>
                        ✓ <span className="text-green-400">NEW:</span> Cross-machine compatible paths
                      </div>
                    </div>
                  </div>
                </TerminalWindow>
              )}

              {activeTab === 'restore' && (
                <TerminalWindow title="🔄 SSH Restore Guide">
                  <div className="space-y-4 text-sm">
                    <div className="bg-cyan-900/30 border border-cyan-500/30 rounded p-3">
                      <div className="text-cyan-400 font-bold mb-2">📥 Cross-Machine Restore Workflow</div>
                      <div className="text-gray-300 text-xs">
                        Restore SSH keys from Vault on any machine or user account with universal storage
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Step 1: List Available Backups</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk list</span>
                    </div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk list --detailed</span>
                    </div>

                    <div className="text-yellow-400 font-bold">## Step 2: Restore Options</div>
                    <div className="text-gray-400"># Restore latest backup</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk restore</span>
                    </div>

                    <div className="text-gray-400"># Restore specific backup</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk restore "laptop-20240315"</span>
                    </div>

                    <div className="text-gray-400"># Interactive backup selection</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk restore --select</span>
                    </div>

                    <div className="text-gray-400"># Interactive file selection</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk restore --interactive</span>
                    </div>

                    <div className="text-gray-400"># Restore to custom directory</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk restore --target-dir "/tmp/ssh-restore"</span>
                    </div>

                    <div className="text-gray-400"># Restore specific files only</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk restore --files "github*,gitlab*"</span>
                    </div>

                    <div className="text-yellow-400 font-bold">## Step 3: Safety Options</div>
                    <div className="text-gray-400"># Preview restore (dry run)</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk restore --dry-run</span>
                    </div>

                    <div className="text-gray-400"># Overwrite existing files</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk restore --overwrite</span>
                    </div>

                    <div className="text-yellow-400 font-bold">## NEW: Cross-Machine Restore</div>
                    <div className="text-gray-400"># Backup on laptop</div>
                    <div>
                      <span className="text-green-400">laptop$</span>
                      <span className="ml-2 text-white">sshsk backup "my-dev-keys"</span>
                    </div>

                    <div className="text-gray-400"># Restore on desktop (different machine/user!)</div>
                    <div>
                      <span className="text-green-400">desktop$</span>
                      <span className="ml-2 text-white">sshsk restore "my-dev-keys"</span>
                    </div>

                    <div className="text-gray-400"># Works across different users too</div>
                    <div>
                      <span className="text-green-400">server$</span>
                      <span className="ml-2 text-white">sshsk restore "my-dev-keys" --target-dir ~/.ssh</span>
                    </div>

                    <div className="bg-green-900/30 border border-green-500/30 rounded p-2 mt-4">
                      <div className="text-green-400 text-xs">
                        ✓ Preserves original permissions exactly<br/>
                        ✓ Verifies MD5 checksums during restore<br/>
                        ✓ Interactive confirmation for overwrites<br/>
                        ✓ <span className="text-yellow-400">NEW:</span> Cross-machine and cross-user restore
                      </div>
                    </div>
                  </div>
                </TerminalWindow>
              )}

              {activeTab === 'storage' && (
                <TerminalWindow title="🗂️ Storage Strategies Guide">
                  <div className="space-y-4 text-sm">
                    <div className="bg-purple-900/30 border border-purple-500/30 rounded p-3">
                      <div className="text-purple-400 font-bold mb-2">🎯 Cross-Machine Restore</div>
                      <div className="text-gray-300 text-xs">
                        Choose the right storage strategy for your use case
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Available Storage Strategies</div>
                    
                    <div className="bg-green-900/20 border border-green-500/30 rounded p-3">
                      <div className="text-green-400 font-bold">✅ Universal Storage (Default)</div>
                      <div className="text-gray-300 text-xs mb-2">Path: shared/backups/backup-name</div>
                      <div className="text-white">
                        • ✅ Cross-machine restore: backup on laptop, restore on desktop<br/>
                        • ✅ Cross-user restore: backup as user1, restore as user2<br/>
                        • ✅ Team sharing: shared backup namespace<br/>
                        • ✅ Container friendly: perfect for CI/CD
                      </div>
                    </div>

                    <div className="bg-blue-900/20 border border-blue-500/30 rounded p-3">
                      <div className="text-blue-400 font-bold">👤 User-Scoped Storage</div>
                      <div className="text-gray-300 text-xs mb-2">Path: users/username/backups/backup-name</div>
                      <div className="text-white">
                        • ✅ Cross-machine restore for same user<br/>
                        • ✅ User isolation in shared Vault<br/>
                        • ⚠️ Limited to same username
                      </div>
                    </div>

                    <div className="bg-yellow-900/20 border border-yellow-500/30 rounded p-3">
                      <div className="text-yellow-400 font-bold">🔒 Machine-User Storage (Legacy)</div>
                      <div className="text-gray-300 text-xs mb-2">Path: users/hostname-username/backups/backup-name</div>
                      <div className="text-white">
                        • ✅ Maximum isolation per machine-user<br/>
                        • ❌ No cross-machine restore<br/>
                        • ❌ No cross-user restore<br/>
                        • 📦 Existing installations (until migrated)
                      </div>
                    </div>

                    <div className="bg-purple-900/20 border border-purple-500/30 rounded p-3">
                      <div className="text-purple-400 font-bold">🎨 Custom Storage</div>
                      <div className="text-gray-300 text-xs mb-2">Path: custom-prefix/backups/backup-name</div>
                      <div className="text-white">
                        • ✅ Team/project organization<br/>
                        • ✅ Flexible prefix configuration<br/>
                        • ⚙️ Requires custom_prefix setting
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Migration for Existing Users</div>
                    <div className="text-gray-400"># Check current storage strategy</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk migrate-status</span>
                    </div>

                    <div className="text-gray-400"># Migrate to universal storage (dry run)</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk migrate --from machine-user --to universal --dry-run</span>
                    </div>

                    <div className="text-gray-400"># Perform actual migration with cleanup</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk migrate --from machine-user --to universal --cleanup</span>
                    </div>

                    <div className="text-yellow-400 font-bold">## Configuration Examples</div>
                    
                    <div className="text-gray-400"># Universal storage (default)</div>
                    <div className="text-cyan-400">vault:</div>
                    <div className="text-white ml-2">storage_strategy: "universal"</div>
                    <div className="text-white ml-2">backup_namespace: "personal"  # Optional</div>

                    <div className="text-gray-400"># User-scoped storage</div>
                    <div className="text-cyan-400">vault:</div>
                    <div className="text-white ml-2">storage_strategy: "user"</div>

                    <div className="text-gray-400"># Legacy machine-user storage</div>
                    <div className="text-cyan-400">vault:</div>
                    <div className="text-white ml-2">storage_strategy: "machine-user"</div>

                    <div className="text-gray-400"># Custom team storage</div>
                    <div className="text-cyan-400">vault:</div>
                    <div className="text-white ml-2">storage_strategy: "custom"</div>
                    <div className="text-white ml-2">custom_prefix: "team-devops"</div>

                    <div className="text-yellow-400 font-bold">## Use Cases</div>
                    <div className="text-white">
                      <div className="text-green-400">📱 Personal Use:</div> Universal storage for laptop ↔ desktop<br/>
                      <div className="text-blue-400">👥 Team Environment:</div> Custom prefix for team organization<br/>
                      <div className="text-yellow-400">🏢 Shared Vault:</div> User-scoped for multi-user isolation<br/>
                      <div className="text-gray-400">🔒 Maximum Security:</div> Machine-user for strict isolation
                    </div>
                  </div>
                </TerminalWindow>
              )}

              {activeTab === 'commands' && (
                <TerminalWindow title="Complete Command Reference">
                  <div className="space-y-3 text-sm">
                    <div className="text-yellow-400 font-bold">## Core Operations</div>
                    <div className="space-y-1">
                      <div className="flex justify-between">
                        <span className="text-cyan-400">sshsk init</span>
                        <span className="text-gray-400">Initialize configuration</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-cyan-400">sshsk analyze</span>
                        <span className="text-gray-400">Analyze SSH directory</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-cyan-400">sshsk backup</span>
                        <span className="text-gray-400">Create backup</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-cyan-400">sshsk restore</span>
                        <span className="text-gray-400">Restore from backup</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-cyan-400">sshsk list</span>
                        <span className="text-gray-400">List available backups</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-cyan-400">sshsk status</span>
                        <span className="text-gray-400">Show system status</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-cyan-400">sshsk delete</span>
                        <span className="text-gray-400">Delete backup</span>
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## NEW: Storage Management</div>
                    <div className="space-y-1">
                      <div className="flex justify-between">
                        <span className="text-green-400">sshsk migrate-status</span>
                        <span className="text-gray-400">Show storage strategy info</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-green-400">sshsk migrate</span>
                        <span className="text-gray-400">Migrate between strategies</span>
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Build Commands (Make)</div>
                    <div className="space-y-1">
                      <div className="flex justify-between">
                        <span className="text-cyan-400">make build</span>
                        <span className="text-gray-400">Build for current platform</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-cyan-400">make build-all</span>
                        <span className="text-gray-400">Build all platforms + containers</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-cyan-400">make build-binaries</span>
                        <span className="text-gray-400">Build all platform binaries</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-cyan-400">make install</span>
                        <span className="text-gray-400">Install to /usr/local/bin</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-cyan-400">make container-build</span>
                        <span className="text-gray-400">Build container image</span>
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Testing Commands</div>
                    <div className="space-y-1">
                      <div className="flex justify-between">
                        <span className="text-cyan-400">make test</span>
                        <span className="text-gray-400">Run tests with coverage</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-cyan-400">make test-coverage</span>
                        <span className="text-gray-400">Generate HTML coverage report</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-cyan-400">make test-coverage-check</span>
                        <span className="text-gray-400">Verify 40%+ coverage (target 85%)</span>
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Release Commands</div>
                    <div className="space-y-1">
                      <div className="flex justify-between">
                        <span className="text-cyan-400">make release VERSION=x.y.z</span>
                        <span className="text-gray-400">Complete release workflow</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-cyan-400">make tag-release VERSION=x.y.z</span>
                        <span className="text-gray-400">Create git tag</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-cyan-400">make release-snapshot</span>
                        <span className="text-gray-400">Test release locally</span>
                      </div>
                    </div>
                  </div>
                </TerminalWindow>
              )}

              {activeTab === 'config' && (
                <TerminalWindow title="Configuration">
                  <div className="space-y-4 text-sm">
                    <div className="text-yellow-400 font-bold">## Environment Variables (Recommended)</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">export VAULT_ADDR="https://vault.company.com:8200"</span>
                    </div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">export VAULT_TOKEN="your-vault-token"</span>
                    </div>

                    <div className="text-yellow-400 font-bold">## Config File</div>
                    <div className="text-gray-400"># Config file location:</div>
                    <div className="text-cyan-400">~/.ssh-secret-keeper/config.yaml</div>
                    <div className="text-gray-400"># Universal storage configuration (default):</div>
                    <pre className="text-white">
{`vault:
  address: "https://vault.company.com:8200"
  storage_strategy: "universal"
  backup_namespace: "personal"  # Optional
backup:
  ssh_dir: "~/.ssh"
  normalize_paths: true
  cross_machine_restore: true`}
                    </pre>

                    <div className="text-gray-400"># Legacy machine-user configuration:</div>
                    <pre className="text-white">
{`vault:
  address: "https://vault.company.com:8200"
  storage_strategy: "machine-user"
backup:
  ssh_dir: "~/.ssh"
  hostname_prefix: true`}
                    </pre>

                    <div className="text-gray-400"># Complete backup configuration:</div>
                    <pre className="text-white">
{`backup:
  ssh_dir: "~/.ssh"
  hostname_prefix: false          # Use with machine-user strategy
  retention_count: 10             # Keep 10 backup versions
  normalize_paths: true           # Enable cross-user compatibility
  cross_machine_restore: true     # Enable cross-machine restore
  
  # File filtering patterns
  include_patterns:
    - "*.rsa"
    - "*.pem" 
    - "*.pub"
    - "id_rsa*"
    - "config"
    - "known_hosts*"
    - "authorized_keys"
    - "*_rsa"
    - "*.ed25519"
    - "*.ecdsa"
  
  exclude_patterns:
    - "*.tmp"
    - "*.bak"
    - "*.old"
    - "*.orig"
    - "*.swp"
  
  # File categorization for organization
  categories:
    services: ["service1_rsa", "service2_rsa"]
    personal: ["id_rsa", "local_rsa"]
    work: ["*work*", "*corp*", "*office*"]
    cloud: ["*cloud*", "*aws*", "*gcp*", "*azure*"]`}
                    </pre>

                    <div className="text-yellow-400 font-bold">## Backup Modes & Options</div>
                    <div className="text-gray-400"># Basic backup</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk backup "my-backup"</span>
                    </div>

                    <div className="text-gray-400"># Interactive file selection</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk backup --interactive</span>
                    </div>

                    <div className="text-gray-400"># Dry run (preview only)</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk backup --dry-run</span>
                    </div>

                    <div className="text-gray-400"># Custom SSH directory</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk backup --ssh-dir "/custom/ssh/path"</span>
                    </div>

                    <div className="text-gray-400"># Auto-generated backup name</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk backup  # Creates backup-YYYYMMDD-HHMMSS</span>
                    </div>

                    <div className="text-yellow-400 font-bold">## Environment Variables</div>
                    <div className="text-gray-400"># Storage strategy configuration</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">export SSH_VAULT_STORAGE_STRATEGY="universal"</span>
                    </div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">export SSH_VAULT_BACKUP_NAMESPACE="team-devops"</span>
                    </div>

                    <div className="text-gray-400"># Backup behavior overrides</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">export SSH_VAULT_BACKUP_SSH_DIR="/custom/ssh"</span>
                    </div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">export SSH_VAULT_BACKUP_RETENTION_COUNT="20"</span>
                    </div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">export SSH_VAULT_BACKUP_NORMALIZE_PATHS="true"</span>
                    </div>
                  </div>
                </TerminalWindow>
              )}
          </div>

          {/* Information and Quick Links moved below documentation */}
          <div className="grid md:grid-cols-2 gap-8 mt-12">
            <div className="bg-gray-900 border border-gray-700 rounded-lg p-6">
              <h4 className="text-amber-400 font-bold mb-3 flex items-center gap-2">
                <Lock className="w-5 h-5" />
                Information
              </h4>
              <ul className="space-y-2 text-gray-300 text-sm">
                <li className="flex items-center gap-2">
                  <span className="text-green-400">✓</span>
                  Requires HashiCorp Vault installation
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-green-400">✓</span>
                  MD5 checksums verify backup integrity
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-green-400">✓</span>
                  List and manage multiple backup versions
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-green-400">✓</span>
                  Cross-machine and cross-user restore support
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-green-400">✓</span>
                  Migration tools for existing installations
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-green-400">✓</span>
                  Open-source community project
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-green-400">✓</span>
                  Production-ready security features
                </li>
              </ul>
            </div>

            <div className="bg-gray-900 border border-gray-700 rounded-lg p-6">
              <h4 className="text-cyan-400 font-bold mb-3 flex items-center gap-2">
                <Book className="w-5 h-5" />
                Quick Links
              </h4>
              <div className="space-y-2">
                <a href="https://github.com/rafaelvzago/ssh-secret-keeper/blob/main/docs/STORAGE_STRATEGIES.md" className="block text-cyan-400 hover:text-cyan-300 text-sm">
                  → Storage Strategies Guide
                </a>
                <a href="https://github.com/rafaelvzago/ssh-secret-keeper/blob/main/docs/CONFIGURATION.md" className="block text-cyan-400 hover:text-cyan-300 text-sm">
                  → Complete Configuration Reference
                </a>
                <a href="https://github.com/rafaelvzago/ssh-secret-keeper/blob/main/Makefile" className="block text-cyan-400 hover:text-cyan-300 text-sm">
                  → Make Target Reference
                </a>
                <a href="https://github.com/rafaelvzago/ssh-secret-keeper/blob/main/docs/QUICK_START.md" className="block text-cyan-400 hover:text-cyan-300 text-sm">
                  → Quick Start Guide
                </a>
                <a href="https://github.com/rafaelvzago/ssh-secret-keeper" className="block text-cyan-400 hover:text-cyan-300 text-sm">
                  → GitHub Repository
                </a>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-gray-800 py-8 px-4">
        <div className="max-w-6xl mx-auto">
          <div className="text-center">
            <div className="font-mono text-cyan-400 text-sm mb-4">
              ┌─────────────────────────────────────────────────────────────┐<br />
              │                   SSH Secret Keeper v{config.app.version}                 │<br />
              │              Open-Source SSH Key Backup Tool               │<br />
              └─────────────────────────────────────────────────────────────┘
            </div>
            <p className="text-gray-400 text-sm">
              Built with ❤️ for the community • By Rafeal Zago • Licensed under Apache 2.0
            </p>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default App;
