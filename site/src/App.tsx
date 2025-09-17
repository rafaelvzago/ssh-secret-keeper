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
      description: "Complete backup of your ~/.ssh folder with permissions preserved. Securely stores keys, config files, and authorized_keys to HashiCorp Vault."
    },
    {
      icon: <Server className="w-8 h-8 text-amber-400" />,
      title: "Restore Anywhere",
      description: "Restore your SSH keys on any machine with Vault access. Perfect for setting up new workstations or recovering from system failures."
    },
    {
      icon: <Key className="w-8 h-8 text-cyan-400" />,
      title: "Analyze & Validate",
      description: "Check SSH key structure and permissions before backup. Detects key types, validates permissions, and categorizes your SSH setup."
    },
    {
      icon: <Zap className="w-8 h-8 text-red-400" />,
      title: "List & Manage",
      description: "View and manage your stored backups with ease. List available backups, check status, and delete old backups when needed."
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
                Securely backup your ~/.ssh folder to HashiCorp Vault and restore it anywhere.<br />
                Perfect for developers, system administrators, and DevOps teams<br />
                who need reliable SSH key management.
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
        <div className="max-w-6xl mx-auto">
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

          <div className="grid md:grid-cols-2 gap-8">
            <div>
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

                    <div className="text-gray-400"># Backup SSH keys</div>
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
                      <span className="ml-4 text-white">rafaelvzago/ssh-secret-keeper:latest backup "my-backup"</span>
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
                        Securely backup your entire SSH directory to HashiCorp Vault
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

                    <div className="text-yellow-400 font-bold">## Step 3: Backup Options</div>
                    <div className="text-gray-400"># Simple backup</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk backup</span>
                    </div>

                    <div className="text-gray-400"># Named backup</div>
                    <div>
                      <span className="text-green-400">$</span>
                      <span className="ml-2 text-white">sshsk backup "laptop-$(date +%Y%m%d)"</span>
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

                    <div className="bg-blue-900/30 border border-blue-500/30 rounded p-2 mt-4">
                      <div className="text-blue-400 text-xs">
                        ✓ Preserves exact permissions (0600/0644)<br/>
                        ✓ Includes MD5 checksums for integrity<br/>
                        ✓ Stores metadata (timestamps, file sizes)
                      </div>
                    </div>
                  </div>
                </TerminalWindow>
              )}

              {activeTab === 'restore' && (
                <TerminalWindow title="🔄 SSH Restore Guide">
                  <div className="space-y-4 text-sm">
                    <div className="bg-cyan-900/30 border border-cyan-500/30 rounded p-3">
                      <div className="text-cyan-400 font-bold mb-2">📥 Complete Restore Workflow</div>
                      <div className="text-gray-300 text-xs">
                        Restore SSH keys from Vault with flexible options
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

                    <div className="bg-green-900/30 border border-green-500/30 rounded p-2 mt-4">
                      <div className="text-green-400 text-xs">
                        ✓ Preserves original permissions exactly<br/>
                        ✓ Verifies MD5 checksums during restore<br/>
                        ✓ Interactive confirmation for overwrites
                      </div>
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
                    <div className="text-gray-400"># Example configuration:</div>
                    <pre className="text-white">
{`vault:
  address: "https://vault.company.com:8200"
  mount_path: "ssh-backups"
backup:
  ssh_dir: "~/.ssh"
  retention_count: 10
security:
  verify_integrity: true`}
                    </pre>
                  </div>
                </TerminalWindow>
              )}
            </div>

            <div className="space-y-6">
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
                    Future: Support for other secret stores
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
                  <a href="#" className="block text-cyan-400 hover:text-cyan-300 text-sm">
                    → Complete Make Target Reference
                  </a>
                  <a href="#" className="block text-cyan-400 hover:text-cyan-300 text-sm">
                    → Backup/Restore Best Practices
                  </a>
                  <a href="#" className="block text-cyan-400 hover:text-cyan-300 text-sm">
                    → Vault Security Configuration
                  </a>
                  <a href="#" className="block text-cyan-400 hover:text-cyan-300 text-sm">
                    → Container Deployment Guide
                  </a>
                  <a href="#" className="block text-cyan-400 hover:text-cyan-300 text-sm">
                    → GitHub Repository
                  </a>
                </div>
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
