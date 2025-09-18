import React, { useState, useEffect } from 'react';
import { Terminal, Shield, Key, Server, Copy, Check, Github, Book, Lock, RefreshCw } from 'lucide-react';
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
      icon: <RefreshCw className="w-8 h-8 text-purple-400" />,
      title: "Self-Updating",
      description: "Built-in update command with automatic backup and rollback. Update to latest version with a single command. Checksum verification ensures integrity and safety."
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
            <a href="#motivation" className="text-cyan-400 hover:text-cyan-300 transition-colors">Motivation</a>
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
                Securely backup your ~/.ssh folder to HashiCorp Vault with <span className="text-green-400 font-semibold">cross-machine restore</span>. Backup on laptop, restore on desktop! <br />Perfect for developers, system administrators and DevOps teams who work across multiple machines and environments.
              </p>
              <div className="flex justify-center gap-4 mt-8">
                <a href="#installation" className="bg-green-600 hover:bg-green-500 px-6 py-3 rounded border border-green-500 transition-colors">
                  Get Started
                </a>
                <a href="#docs" className="border border-cyan-500 text-cyan-400 hover:bg-cyan-900/20 px-6 py-3 rounded transition-colors">
                  View Documentation
                </a>
              </div>
            </div>
          </ASCIIBorder>
        </div>
      </section>

      {/* Motivation Section */}
      <section id="motivation" className="py-16 px-4">
        <div className="max-w-4xl mx-auto">
          <h3 className="text-3xl font-bold text-center mb-12 text-green-400">
            ┌─[ Why SSH Secret Keeper? ]─┐
          </h3>

          <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-8">
            <div className="text-gray-300 space-y-6 leading-relaxed">
              <p className="text-lg">
                As a tech guy for more than 25 years, I've faced the same recurring problem every time I switch laptops or try out a new Linux distro like the amazing <a href="https://omarchy.org/" target="_blank" rel="noopener noreferrer" className="text-cyan-400 hover:text-cyan-300 transition-colors underline">Omarchy</a>: how do I securely handle my SSH keys?
              </p>

              <p>
                I know the common workarounds—flash drives, NFS servers, even scp—but I was tempted by the usual over-engineering approach. Why not deploy a Kubernetes cluster and a HashiCorp Vault solution just for my keys? Well, I did it.
              </p>

              <p>
                The brand new problem I created for myself was even more complex: <span className="text-amber-400 font-semibold">how do I handle the backup and restore process for all of that?</span> This cycle of creating complexity to solve a simple problem is exactly why I built SSH Secret Keeper.
              </p>

              <div className="bg-green-900/20 border border-green-500/30 rounded p-4 my-6">
                <p className="text-green-400 font-semibold mb-2">The Solution</p>
                <p className="text-sm">
                  It's my straightforward, open-source solution to securely back up your <code className="bg-gray-800 px-2 py-1 rounded text-cyan-400">~/.ssh</code> folder to an existing HashiCorp Vault instance and restore it anywhere. You can back up on your laptop and restore on your desktop, with the flexibility to create multiple backups and choose if they are open to any machine or tied to a specific user and hostname.
                </p>
              </div>

              <p>
                This tool simply relies on your Vault's own proven security; I barely touch it. I'm sharing this in the same spirit as other great community projects.
              </p>

              <div className="bg-blue-900/20 border border-blue-500/30 rounded p-4 mt-6">
                <div className="flex items-center gap-2 mb-2">
                  <span className="text-blue-400 font-semibold">Connect with me:</span>
                </div>
                <div className="flex flex-wrap gap-4 text-sm">
                  <a href="https://rafaelvzago.com" target="_blank" rel="noopener noreferrer" className="text-cyan-400 hover:text-cyan-300 transition-colors">
                    🌐 rafaelvzago.com
                  </a>
                  <a href="https://github.com/rafaelvzago/ssh-secret-keeper" target="_blank" rel="noopener noreferrer" className="text-cyan-400 hover:text-cyan-300 transition-colors">
                    📦 GitHub Project
                  </a>
                </div>
              </div>
            </div>
          </div>
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

              <div className="mt-4 pt-4 border-t border-gray-700">
                <div className="text-yellow-400 text-sm font-bold mb-2">Already installed? Update with:</div>
                <div className="flex items-center">
                  <span className="text-green-400">$</span>
                  <code className="ml-2 text-white flex-1">sudo sshsk update</code>
                  <CopyButton text="sudo sshsk update" />
                </div>
                <div className="text-gray-400 text-sm mt-1">
                  # Built-in self-updating - no need to re-run the installer
                </div>
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
                { id: 'update', label: 'Update' },
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
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">curl -L https://github.com/rafaelvzago/ssh-secret-keeper/releases/latest/download/ssh-secret-keeper-VERSION-linux-amd64.tar.gz -o sshsk.tar.gz</code>
                      <CopyButton text="curl -L https://github.com/rafaelvzago/ssh-secret-keeper/releases/latest/download/ssh-secret-keeper-VERSION-linux-amd64.tar.gz -o sshsk.tar.gz" />
                    </div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">tar -xzf sshsk.tar.gz</code>
                      <CopyButton text="tar -xzf sshsk.tar.gz" />
                    </div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">chmod +x sshsk</code>
                      <CopyButton text="chmod +x sshsk" />
                    </div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sudo mv sshsk /usr/local/bin/</code>
                      <CopyButton text="sudo mv sshsk /usr/local/bin/" />
                    </div>

                    <div className="text-yellow-400 font-bold">## Verify Installation</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk --version</code>
                      <CopyButton text="sshsk --version" />
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

              {activeTab === 'update' && (
                <TerminalWindow title="🔄 Update Management">
                  <div className="space-y-4 text-sm">
                    <div className="bg-purple-900/30 border border-purple-500/30 rounded p-3">
                      <div className="text-purple-400 font-bold mb-2">🚀 Self-Updating Feature</div>
                      <div className="text-gray-300 text-xs">
                        Keep SSH Secret Keeper up-to-date with built-in update commands
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Quick Update Commands</div>

                    <div className="text-gray-400"># Check for updates without installing</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk update --check</code>
                      <CopyButton text="sshsk update --check" />
                    </div>

                    <div className="text-gray-400"># Update to the latest stable version</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sudo sshsk update</code>
                      <CopyButton text="sudo sshsk update" />
                    </div>

                    <div className="text-gray-400"># Update to a specific version</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sudo sshsk update --version v1.0.5</code>
                      <CopyButton text="sudo sshsk update --version v1.0.5" />
                    </div>

                    <div className="text-gray-400"># Include pre-release versions</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sudo sshsk update --pre-release</code>
                      <CopyButton text="sudo sshsk update --pre-release" />
                    </div>

                    <div className="text-gray-400"># Force reinstall even if already on latest</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sudo sshsk update --force</code>
                      <CopyButton text="sudo sshsk update --force" />
                    </div>

                    <div className="text-yellow-400 font-bold">## Update Process</div>
                    <div className="bg-gray-800 rounded p-3">
                      <div className="text-gray-300 text-xs space-y-1">
                        <div className="text-green-400">The update process includes:</div>
                        <div>1. 🔍 Check for the latest release on GitHub</div>
                        <div>2. 📥 Download appropriate binary for your platform</div>
                        <div>3. ✅ Verify download integrity (checksum)</div>
                        <div>4. 💾 Create backup of current binary</div>
                        <div>5. 🧪 Verify new binary works correctly</div>
                        <div>6. 🔄 Replace old binary with new version</div>
                        <div>7. ⚡ Automatic rollback on failure</div>
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Safety Features</div>
                    <div className="bg-green-900/20 border border-green-500/30 rounded p-3">
                      <div className="text-gray-300 text-xs space-y-1">
                        <div><span className="text-green-400">✓</span> Automatic backup before update</div>
                        <div><span className="text-green-400">✓</span> Checksum verification for integrity</div>
                        <div><span className="text-green-400">✓</span> Binary verification before replacement</div>
                        <div><span className="text-green-400">✓</span> Automatic rollback on failure</div>
                        <div><span className="text-green-400">✓</span> Preserves file permissions</div>
                        <div><span className="text-green-400">✓</span> Safe handling of running processes</div>
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Update Options</div>
                    <div className="bg-gray-800 rounded p-3">
                      <div className="text-gray-300 text-xs space-y-1">
                        <div><span className="text-cyan-400">--check</span> - Check for updates without installing</div>
                        <div><span className="text-cyan-400">--version VERSION</span> - Update to specific version</div>
                        <div><span className="text-cyan-400">--pre-release</span> - Include pre-release versions</div>
                        <div><span className="text-cyan-400">--force</span> - Force update even if on latest</div>
                        <div><span className="text-cyan-400">--no-backup</span> - Skip backup creation (not recommended)</div>
                        <div><span className="text-cyan-400">--skip-checksum</span> - Skip checksum verification (not recommended)</div>
                        <div><span className="text-cyan-400">--skip-verify</span> - Skip binary verification</div>
                        <div><span className="text-cyan-400">-y, --yes</span> - Skip confirmation prompt</div>
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Common Scenarios</div>

                    <div className="text-gray-400"># Check current version</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk --version</code>
                      <CopyButton text="sshsk --version" />
                    </div>

                    <div className="text-gray-400"># Check and update if available</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk update --check && sudo sshsk update</code>
                      <CopyButton text="sshsk update --check && sudo sshsk update" />
                    </div>

                    <div className="text-gray-400"># Auto-update without prompts (for scripts)</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sudo sshsk update --yes</code>
                      <CopyButton text="sudo sshsk update --yes" />
                    </div>

                    <div className="bg-blue-900/30 border border-blue-500/30 rounded p-2 mt-4">
                      <div className="text-blue-400 text-xs">
                        💡 Updates require sudo for system-wide installations (/usr/local/bin)<br/>
                        💡 The installer script always installs the latest version<br/>
                        💡 Docker users should pull new images instead: docker pull rafaelvzago/ssh-secret-keeper:latest
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
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sudo pacman -S go make git</code>
                      <CopyButton text="sudo pacman -S go make git" />
                    </div>

                    <div className="text-gray-400"># Ubuntu/Debian</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sudo apt-get install golang-go make git</code>
                      <CopyButton text="sudo apt-get install golang-go make git" />
                    </div>

                    <div className="text-gray-400"># Fedora/RHEL</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sudo dnf install golang make git</code>
                      <CopyButton text="sudo dnf install golang make git" />
                    </div>

                    <div className="text-yellow-400 font-bold">## Clone and Build</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">git clone https://github.com/rafaelvzago/ssh-secret-keeper.git</code>
                      <CopyButton text="git clone https://github.com/rafaelvzago/ssh-secret-keeper.git" />
                    </div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">cd ssh-secret-keeper</code>
                      <CopyButton text="cd ssh-secret-keeper" />
                    </div>

                    <div className="text-gray-400"># Build for current platform</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">make build</code>
                      <CopyButton text="make build" />
                    </div>

                    <div className="text-gray-400"># Install to /usr/local/bin</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sudo make install</code>
                      <CopyButton text="sudo make install" />
                    </div>

                    <div className="text-gray-400"># Verify installation</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk --version</code>
                      <CopyButton text="sshsk --version" />
                    </div>
                    <div className="text-cyan-400">✓ SSH Secret Keeper v{config.app.version}</div>
                  </div>
                </TerminalWindow>
              )}

              {activeTab === 'docker' && (
                <TerminalWindow title="Docker Usage">
                  <div className="space-y-4 text-sm">
                    <div className="text-yellow-400 font-bold">## Pull from Docker Hub</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">docker pull rafaelvzago/ssh-secret-keeper:latest</code>
                      <CopyButton text="docker pull rafaelvzago/ssh-secret-keeper:latest" />
                    </div>

                    <div className="text-yellow-400 font-bold">## Basic Usage</div>
                    <div className="text-gray-400"># Analyze SSH directory</div>
                    <div className="relative">
                      <div className="flex items-start">
                        <span className="text-green-400 mt-1">$</span>
                        <div className="ml-2 text-white flex-1">
                          <div>docker run --rm -v ~/.ssh:/ssh:ro \</div>
                          <div className="ml-2">-e VAULT_ADDR="https://your-vault:8200" \</div>
                          <div className="ml-2">-e VAULT_TOKEN="your-token" \</div>
                          <div className="ml-2">rafaelvzago/ssh-secret-keeper:latest analyze</div>
                        </div>
                        <CopyButton text='docker run --rm -v ~/.ssh:/ssh:ro -e VAULT_ADDR="https://your-vault:8200" -e VAULT_TOKEN="your-token" rafaelvzago/ssh-secret-keeper:latest analyze' />
                      </div>
                    </div>

                    <div className="text-gray-400"># Quick backup (auto-generated name)</div>
                    <div className="relative">
                      <div className="flex items-start">
                        <span className="text-green-400 mt-1">$</span>
                        <div className="ml-2 text-white flex-1">
                          <div>docker run --rm -v ~/.ssh:/ssh:ro \</div>
                          <div className="ml-2">-e VAULT_ADDR="https://your-vault:8200" \</div>
                          <div className="ml-2">-e VAULT_TOKEN="your-token" \</div>
                          <div className="ml-2">rafaelvzago/ssh-secret-keeper:latest backup</div>
                        </div>
                        <CopyButton text='docker run --rm -v ~/.ssh:/ssh:ro -e VAULT_ADDR="https://your-vault:8200" -e VAULT_TOKEN="your-token" rafaelvzago/ssh-secret-keeper:latest backup' />
                      </div>
                    </div>

                    <div className="text-gray-400"># Named backup</div>
                    <div className="relative">
                      <div className="flex items-start">
                        <span className="text-green-400 mt-1">$</span>
                        <div className="ml-2 text-white flex-1">
                          <div>docker run --rm -v ~/.ssh:/ssh:ro \</div>
                          <div className="ml-2">-e VAULT_ADDR="https://your-vault:8200" \</div>
                          <div className="ml-2">-e VAULT_TOKEN="your-token" \</div>
                          <div className="ml-2">rafaelvzago/ssh-secret-keeper:latest backup "container-$(date +%Y%m%d)"</div>
                        </div>
                        <CopyButton text='docker run --rm -v ~/.ssh:/ssh:ro -e VAULT_ADDR="https://your-vault:8200" -e VAULT_TOKEN="your-token" rafaelvzago/ssh-secret-keeper:latest backup "container-$(date +%Y%m%d)"' />
                      </div>
                    </div>

                    <div className="text-gray-400"># Dry run mode</div>
                    <div className="relative">
                      <div className="flex items-start">
                        <span className="text-green-400 mt-1">$</span>
                        <div className="ml-2 text-white flex-1">
                          <div>docker run --rm -v ~/.ssh:/ssh:ro \</div>
                          <div className="ml-2">rafaelvzago/ssh-secret-keeper:latest backup --dry-run</div>
                        </div>
                        <CopyButton text='docker run --rm -v ~/.ssh:/ssh:ro rafaelvzago/ssh-secret-keeper:latest backup --dry-run' />
                      </div>
                    </div>

                    <div className="text-gray-400"># Restore SSH keys</div>
                    <div className="relative">
                      <div className="flex items-start">
                        <span className="text-green-400 mt-1">$</span>
                        <div className="ml-2 text-white flex-1">
                          <div>docker run --rm -v ~/.ssh:/ssh \</div>
                          <div className="ml-2">-e VAULT_ADDR="https://your-vault:8200" \</div>
                          <div className="ml-2">-e VAULT_TOKEN="your-token" \</div>
                          <div className="ml-2">rafaelvzago/ssh-secret-keeper:latest restore</div>
                        </div>
                        <CopyButton text='docker run --rm -v ~/.ssh:/ssh -e VAULT_ADDR="https://your-vault:8200" -e VAULT_TOKEN="your-token" rafaelvzago/ssh-secret-keeper:latest restore' />
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## NEW: Storage Strategy Configuration</div>
                    <div className="text-gray-400"># Use universal storage (default - cross-machine)</div>
                    <div className="relative">
                      <div className="flex items-start">
                        <span className="text-green-400 mt-1">$</span>
                        <div className="ml-2 text-white flex-1">
                          <div>docker run --rm -v ~/.ssh:/ssh:ro \</div>
                          <div className="ml-2">-e VAULT_ADDR="https://your-vault:8200" \</div>
                          <div className="ml-2">-e VAULT_TOKEN="your-token" \</div>
                          <div className="ml-2">-e SSHSK_VAULT_STORAGE_STRATEGY="universal" \</div>
                          <div className="ml-2">rafaelvzago/ssh-secret-keeper:latest backup "cross-machine-backup"</div>
                        </div>
                        <CopyButton text='docker run --rm -v ~/.ssh:/ssh:ro -e VAULT_ADDR="https://your-vault:8200" -e VAULT_TOKEN="your-token" -e SSHSK_VAULT_STORAGE_STRATEGY="universal" rafaelvzago/ssh-secret-keeper:latest backup "cross-machine-backup"' />
                      </div>
                    </div>

                    <div className="text-gray-400"># Use legacy machine-user storage</div>
                    <div className="relative">
                      <div className="flex items-start">
                        <span className="text-green-400 mt-1">$</span>
                        <div className="ml-2 text-white flex-1">
                          <div>docker run --rm -v ~/.ssh:/ssh:ro \</div>
                          <div className="ml-2">-e SSHSK_VAULT_STORAGE_STRATEGY="machine-user" \</div>
                          <div className="ml-2">rafaelvzago/ssh-secret-keeper:latest backup "isolated-backup"</div>
                        </div>
                        <CopyButton text='docker run --rm -v ~/.ssh:/ssh:ro -e SSHSK_VAULT_STORAGE_STRATEGY="machine-user" rafaelvzago/ssh-secret-keeper:latest backup "isolated-backup"' />
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Using Podman</div>
                    <div className="text-gray-400"># Same commands, just replace 'docker' with 'podman'</div>
                    <div className="relative">
                      <div className="flex items-start">
                        <span className="text-green-400 mt-1">$</span>
                        <div className="ml-2 text-white flex-1">
                          <div>podman run --rm -v ~/.ssh:/ssh:ro \</div>
                          <div className="ml-2">-e VAULT_ADDR="https://your-vault:8200" \</div>
                          <div className="ml-2">-e VAULT_TOKEN="your-token" \</div>
                          <div className="ml-2">-e SSHSK_VAULT_STORAGE_STRATEGY="universal" \</div>
                          <div className="ml-2">rafaelvzago/ssh-secret-keeper:latest analyze</div>
                        </div>
                        <CopyButton text='podman run --rm -v ~/.ssh:/ssh:ro -e VAULT_ADDR="https://your-vault:8200" -e VAULT_TOKEN="your-token" -e SSHSK_VAULT_STORAGE_STRATEGY="universal" rafaelvzago/ssh-secret-keeper:latest analyze' />
                      </div>
                    </div>
                  </div>
                </TerminalWindow>
              )}

              {activeTab === 'backup' && (
                <TerminalWindow title="🔒 SSH Backup Guide">
                  <div className="space-y-4 text-sm">
                    <div className="bg-green-900/30 border border-green-500/30 rounded p-3">
                      <div className="text-green-400 font-bold mb-2">📋 Simple Backup Workflow</div>
                      <div className="text-gray-300 text-xs">
                        Backup your SSH directory using environment variables - no configuration files needed
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Step 1: Set Environment Variables</div>
                    <div className="text-gray-400"># Configure Vault connection</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export VAULT_ADDR="https://vault.company.com:8200"</code>
                      <CopyButton text='export VAULT_ADDR="https://vault.company.com:8200"' />
                    </div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export VAULT_TOKEN="your-vault-token"</code>
                      <CopyButton text='export VAULT_TOKEN="your-vault-token"' />
                    </div>

                    <div className="text-gray-400"># Configure storage strategy (optional - defaults to "universal")</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_STORAGE_STRATEGY="universal"</code>
                      <CopyButton text='export SSHSK_VAULT_STORAGE_STRATEGY="universal"' />
                    </div>
                    <div className="text-gray-400 text-xs mt-1">Available options: "universal", "user", "machine-user", "custom"</div>

                    <div className="text-yellow-400 font-bold">## Step 2: Initialize (No Config Files Created)</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk init</code>
                      <CopyButton text="sshsk init" />
                    </div>
                    <div className="text-gray-400 text-xs">Uses environment variables only - no local files created</div>

                    <div className="text-yellow-400 font-bold">## Step 3: Backup Your SSH Keys</div>

                    <div className="bg-green-900/20 border border-green-500/30 rounded p-3">
                      <div className="text-green-400 font-bold">🚀 Quick Backup (Auto-named)</div>
                      <div className="flex items-center">
                        <span className="text-green-400">$</span>
                        <code className="ml-2 text-white flex-1">sshsk backup</code>
                        <CopyButton text="sshsk backup" />
                      </div>
                      <div className="text-gray-400 text-xs mt-1">Creates: backup-YYYYMMDD-HHMMSS</div>
                    </div>

                    <div className="bg-cyan-900/20 border border-cyan-500/30 rounded p-3 mt-3">
                      <div className="text-cyan-400 font-bold">⚡ Complete Backup Example</div>
                      <div className="text-gray-400 text-xs mb-2">Full workflow in 5 commands:</div>
                      <div className="space-y-1 text-xs">
                        <div><span className="text-green-400">$</span> export VAULT_ADDR="https://vault.company.com:8200"</div>
                        <div><span className="text-green-400">$</span> export VAULT_TOKEN="your-vault-token"</div>
                        <div><span className="text-green-400">$</span> export SSHSK_VAULT_STORAGE_STRATEGY="universal"</div>
                        <div><span className="text-green-400">$</span> sshsk init</div>
                        <div><span className="text-green-400">$</span> sshsk backup</div>
                      </div>
                    </div>

                    <div className="bg-blue-900/20 border border-blue-500/30 rounded p-3">
                      <div className="text-blue-400 font-bold">📝 Named Backup</div>
                      <div className="flex items-center">
                        <span className="text-green-400">$</span>
                        <code className="ml-2 text-white flex-1">sshsk backup "my-laptop-keys"</code>
                        <CopyButton text='sshsk backup "my-laptop-keys"' />
                      </div>
                      <div className="flex items-center mt-1">
                        <span className="text-green-400">$</span>
                        <code className="ml-2 text-white flex-1">sshsk backup "dev-$(date +%Y%m%d)"</code>
                        <CopyButton text='sshsk backup "dev-$(date +%Y%m%d)"' />
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Step 4: Backup Options</div>
                    <div className="text-gray-400"># Preview what will be backed up</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk backup --dry-run</code>
                      <CopyButton text="sshsk backup --dry-run" />
                    </div>

                    <div className="text-gray-400"># Choose specific files interactively</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk backup --interactive "selective-backup"</code>
                      <CopyButton text='sshsk backup --interactive "selective-backup"' />
                    </div>

                    <div className="text-gray-400"># Custom SSH directory</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk backup --ssh-dir "/custom/path" "custom-backup"</code>
                      <CopyButton text='sshsk backup --ssh-dir "/custom/path" "custom-backup"' />
                    </div>

                    <div className="text-yellow-400 font-bold">## Storage Strategy Examples</div>
                    <div className="text-gray-400"># Universal storage (default - cross-machine restore)</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_STORAGE_STRATEGY="universal"</code>
                      <CopyButton text='export SSHSK_VAULT_STORAGE_STRATEGY="universal"' />
                    </div>

                    <div className="text-gray-400"># User-scoped storage</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_STORAGE_STRATEGY="user"</code>
                      <CopyButton text='export SSHSK_VAULT_STORAGE_STRATEGY="user"' />
                    </div>

                    <div className="text-gray-400"># Custom team storage</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_STORAGE_STRATEGY="custom"</code>
                      <CopyButton text='export SSHSK_VAULT_STORAGE_STRATEGY="custom"' />
                    </div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_CUSTOM_PREFIX="team-devops"</code>
                      <CopyButton text='export SSHSK_VAULT_CUSTOM_PREFIX="team-devops"' />
                    </div>

                    <div className="bg-blue-900/30 border border-blue-500/30 rounded p-2 mt-4">
                      <div className="text-blue-400 text-xs">
                        ✓ No configuration files needed - uses environment variables only<br/>
                        ✓ Cross-machine restore with universal storage (default)<br/>
                        ✓ Preserves exact permissions (0600/0644)<br/>
                        ✓ Includes MD5 checksums for integrity
                      </div>
                    </div>
                  </div>
                </TerminalWindow>
              )}

              {activeTab === 'restore' && (
                <TerminalWindow title="🔄 SSH Restore Guide">
                  <div className="space-y-4 text-sm">
                    <div className="bg-cyan-900/30 border border-cyan-500/30 rounded p-3">
                      <div className="text-cyan-400 font-bold mb-2">📥 Simple Restore Workflow</div>
                      <div className="text-gray-300 text-xs">
                        Restore your SSH keys from Vault using environment variables
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Step 1: Set Environment Variables</div>
                    <div className="text-gray-400"># Same as backup - configure Vault connection</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export VAULT_ADDR="https://vault.company.com:8200"</code>
                      <CopyButton text='export VAULT_ADDR="https://vault.company.com:8200"' />
                    </div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export VAULT_TOKEN="your-vault-token"</code>
                      <CopyButton text='export VAULT_TOKEN="your-vault-token"' />
                    </div>

                    <div className="text-gray-400"># Configure storage strategy (must match backup strategy)</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_STORAGE_STRATEGY="universal"</code>
                      <CopyButton text='export SSHSK_VAULT_STORAGE_STRATEGY="universal"' />
                    </div>
                    <div className="text-gray-400 text-xs mt-1">Available options: "universal", "user", "machine-user", "custom"</div>

                    <div className="text-yellow-400 font-bold">## Step 2: List Available Backups</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk list</code>
                      <CopyButton text="sshsk list" />
                    </div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk list --detailed</code>
                      <CopyButton text="sshsk list --detailed" />
                    </div>

                    <div className="text-yellow-400 font-bold">## Step 3: Restore Your Backup</div>

                    <div className="bg-green-900/20 border border-green-500/30 rounded p-3">
                      <div className="text-green-400 font-bold">🚀 Quick Restore (Latest Backup)</div>
                      <div className="flex items-center">
                        <span className="text-green-400">$</span>
                        <code className="ml-2 text-white flex-1">sshsk restore</code>
                        <CopyButton text="sshsk restore" />
                      </div>
                      <div className="text-gray-400 text-xs mt-1">Restores the most recent backup to ~/.ssh</div>
                    </div>

                    <div className="bg-cyan-900/20 border border-cyan-500/30 rounded p-3 mt-3">
                      <div className="text-cyan-400 font-bold">⚡ Complete Restore Example</div>
                      <div className="text-gray-400 text-xs mb-2">Full workflow in 4 commands:</div>
                      <div className="space-y-1 text-xs">
                        <div><span className="text-green-400">$</span> export VAULT_ADDR="https://vault.company.com:8200"</div>
                        <div><span className="text-green-400">$</span> export VAULT_TOKEN="your-vault-token"</div>
                        <div><span className="text-green-400">$</span> export SSHSK_VAULT_STORAGE_STRATEGY="universal"</div>
                        <div><span className="text-green-400">$</span> sshsk restore</div>
                      </div>
                    </div>

                    <div className="bg-blue-900/20 border border-blue-500/30 rounded p-3">
                      <div className="text-blue-400 font-bold">📝 Named Backup Restore</div>
                      <div className="flex items-center">
                        <span className="text-green-400">$</span>
                        <code className="ml-2 text-white flex-1">sshsk restore "my-laptop-keys"</code>
                        <CopyButton text='sshsk restore "my-laptop-keys"' />
                      </div>
                      <div className="flex items-center mt-1">
                        <span className="text-green-400">$</span>
                        <code className="ml-2 text-white flex-1">sshsk restore "dev-20240315"</code>
                        <CopyButton text='sshsk restore "dev-20240315"' />
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Step 4: Restore Options</div>
                    <div className="text-gray-400"># Preview what will be restored (safe)</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk restore --dry-run</code>
                      <CopyButton text="sshsk restore --dry-run" />
                    </div>

                    <div className="text-gray-400"># Restore to custom directory</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk restore --target-dir "/tmp/ssh-backup"</code>
                      <CopyButton text='sshsk restore --target-dir "/tmp/ssh-backup"' />
                    </div>

                    <div className="text-gray-400"># Choose backup interactively</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk restore --select</code>
                      <CopyButton text="sshsk restore --select" />
                    </div>

                    <div className="text-gray-400"># Restore only specific files</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk restore --files "github*,gitlab*"</code>
                      <CopyButton text='sshsk restore --files "github*,gitlab*"' />
                    </div>

                    <div className="text-yellow-400 font-bold">## Cross-Machine Restore Example</div>
                    <div className="text-gray-400"># Backup on laptop</div>
                    <div className="flex items-center">
                      <span className="text-green-400">laptop$</span>
                      <code className="ml-2 text-white flex-1">sshsk backup "my-dev-keys"</code>
                      <CopyButton text='sshsk backup "my-dev-keys"' />
                    </div>

                    <div className="text-gray-400"># Restore on desktop (different machine)</div>
                    <div className="flex items-center">
                      <span className="text-green-400">desktop$</span>
                      <code className="ml-2 text-white flex-1">sshsk restore "my-dev-keys"</code>
                      <CopyButton text='sshsk restore "my-dev-keys"' />
                    </div>

                    <div className="bg-green-900/30 border border-green-500/30 rounded p-2 mt-4">
                      <div className="text-green-400 text-xs">
                        ✓ Works with universal storage (default) for cross-machine restore<br/>
                        ✓ Preserves original permissions exactly (0600/0644)<br/>
                        ✓ Verifies MD5 checksums during restore<br/>
                        ✓ Interactive confirmation prevents accidental overwrites
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
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk migrate-status</code>
                      <CopyButton text="sshsk migrate-status" />
                    </div>

                    <div className="text-gray-400"># Migrate to universal storage (dry run)</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk migrate --from machine-user --to universal --dry-run</code>
                      <CopyButton text="sshsk migrate --from machine-user --to universal --dry-run" />
                    </div>

                    <div className="text-gray-400"># Perform actual migration with cleanup</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">sshsk migrate --from machine-user --to universal --cleanup</code>
                      <CopyButton text="sshsk migrate --from machine-user --to universal --cleanup" />
                    </div>

                    <div className="text-yellow-400 font-bold">## Environment Variable Configuration</div>

                    <div className="text-gray-400"># Universal storage (default)</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_STORAGE_STRATEGY="universal"</code>
                      <CopyButton text='export SSHSK_VAULT_STORAGE_STRATEGY="universal"' />
                    </div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_BACKUP_NAMESPACE="personal"</code>
                      <CopyButton text='export SSHSK_VAULT_BACKUP_NAMESPACE="personal"' />
                    </div>

                    <div className="text-gray-400"># User-scoped storage</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_STORAGE_STRATEGY="user"</code>
                      <CopyButton text='export SSHSK_VAULT_STORAGE_STRATEGY="user"' />
                    </div>

                    <div className="text-gray-400"># Legacy machine-user storage</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_STORAGE_STRATEGY="machine-user"</code>
                      <CopyButton text='export SSHSK_VAULT_STORAGE_STRATEGY="machine-user"' />
                    </div>

                    <div className="text-gray-400"># Custom team storage</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_STORAGE_STRATEGY="custom"</code>
                      <CopyButton text='export SSHSK_VAULT_STORAGE_STRATEGY="custom"' />
                    </div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_CUSTOM_PREFIX="team-devops"</code>
                      <CopyButton text='export SSHSK_VAULT_CUSTOM_PREFIX="team-devops"' />
                    </div>

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
                <TerminalWindow title="SSH Secret Keeper Commands">
                  <div className="space-y-4 text-sm">
                    <div className="bg-cyan-900/30 border border-cyan-500/30 rounded p-3">
                      <div className="text-cyan-400 font-bold mb-2">🛠️ Available Commands & Options</div>
                      <div className="text-gray-300 text-xs">
                        Complete reference for all sshsk commands and their options
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Core Operations</div>
                    <div className="space-y-2">
                      <div className="bg-gray-800 rounded p-2">
                        <div className="flex justify-between items-center">
                          <span className="text-cyan-400 font-mono">sshsk init</span>
                          <span className="text-gray-400 text-xs">Initialize configuration</span>
                        </div>
                        <div className="text-gray-300 text-xs mt-1 ml-4">
                          Options: --vault-addr, --token
                        </div>
                      </div>
                      <div className="bg-gray-800 rounded p-2">
                        <div className="flex justify-between items-center">
                          <span className="text-cyan-400 font-mono">sshsk analyze</span>
                          <span className="text-gray-400 text-xs">Analyze SSH directory</span>
                        </div>
                        <div className="text-gray-300 text-xs mt-1 ml-4">
                          Options: --verbose, --ssh-dir
                        </div>
                      </div>
                      <div className="bg-gray-800 rounded p-2">
                        <div className="flex justify-between items-center">
                          <span className="text-cyan-400 font-mono">sshsk backup</span>
                          <span className="text-gray-400 text-xs">Create backup</span>
                        </div>
                        <div className="text-gray-300 text-xs mt-1 ml-4">
                          Options: --dry-run, --interactive, --ssh-dir<br/>
                          Usage: sshsk backup [backup-name]
                        </div>
                      </div>
                      <div className="bg-gray-800 rounded p-2">
                        <div className="flex justify-between items-center">
                          <span className="text-cyan-400 font-mono">sshsk restore</span>
                          <span className="text-gray-400 text-xs">Restore from backup</span>
                        </div>
                        <div className="text-gray-300 text-xs mt-1 ml-4">
                          Options: --dry-run, --select, --interactive, --target-dir, --files, --overwrite<br/>
                          Usage: sshsk restore [backup-name]
                        </div>
                      </div>
                      <div className="bg-gray-800 rounded p-2">
                        <div className="flex justify-between items-center">
                          <span className="text-cyan-400 font-mono">sshsk list</span>
                          <span className="text-gray-400 text-xs">List available backups</span>
                        </div>
                        <div className="text-gray-300 text-xs mt-1 ml-4">
                          Options: --detailed
                        </div>
                      </div>
                      <div className="bg-gray-800 rounded p-2">
                        <div className="flex justify-between items-center">
                          <span className="text-cyan-400 font-mono">sshsk status</span>
                          <span className="text-gray-400 text-xs">Show system status</span>
                        </div>
                        <div className="text-gray-300 text-xs mt-1 ml-4">
                          Options: --checksums, --vault, --ssh
                        </div>
                      </div>
                      <div className="bg-gray-800 rounded p-2">
                        <div className="flex justify-between items-center">
                          <span className="text-cyan-400 font-mono">sshsk delete</span>
                          <span className="text-gray-400 text-xs">Delete backup</span>
                        </div>
                        <div className="text-gray-300 text-xs mt-1 ml-4">
                          Options: --force, --interactive<br/>
                          Usage: sshsk delete [backup-name]
                        </div>
                      </div>
                      <div className="bg-gray-800 rounded p-2">
                        <div className="flex justify-between items-center">
                          <span className="text-cyan-400 font-mono">sshsk update</span>
                          <span className="text-gray-400 text-xs">Update to latest version</span>
                        </div>
                        <div className="text-gray-300 text-xs mt-1 ml-4">
                          Options: --check, --version, --pre-release, --force, --no-backup<br/>
                          Usage: sshsk update [--version v1.0.5]
                        </div>
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Storage Management</div>
                    <div className="space-y-2">
                      <div className="bg-gray-800 rounded p-2">
                        <div className="flex justify-between items-center">
                          <span className="text-green-400 font-mono">sshsk migrate-status</span>
                          <span className="text-gray-400 text-xs">Show storage strategy info</span>
                        </div>
                        <div className="text-gray-300 text-xs mt-1 ml-4">
                          Shows current strategy and migration options
                        </div>
                      </div>
                      <div className="bg-gray-800 rounded p-2">
                        <div className="flex justify-between items-center">
                          <span className="text-green-400 font-mono">sshsk migrate</span>
                          <span className="text-gray-400 text-xs">Migrate between strategies</span>
                        </div>
                        <div className="text-gray-300 text-xs mt-1 ml-4">
                          Options: --from, --to, --dry-run, --cleanup<br/>
                          Example: sshsk migrate --from machine-user --to universal
                        </div>
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Global Options</div>
                    <div className="bg-gray-800 rounded p-3">
                      <div className="text-gray-300 text-xs space-y-1">
                        <div><span className="text-cyan-400">--config</span> - Specify config file path</div>
                        <div><span className="text-cyan-400">--verbose</span> - Enable verbose logging</div>
                        <div><span className="text-cyan-400">--quiet</span> - Suppress output except errors</div>
                        <div><span className="text-cyan-400">--help</span> - Show help for any command</div>
                        <div><span className="text-cyan-400">--version</span> - Show version information</div>
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Command Examples</div>
                    <div className="bg-gray-800 rounded p-3">
                      <div className="text-gray-300 text-xs space-y-1">
                        <div><span className="text-green-400">$</span> sshsk backup --dry-run "test-backup"</div>
                        <div><span className="text-green-400">$</span> sshsk restore --select --interactive</div>
                        <div><span className="text-green-400">$</span> sshsk list --detailed</div>
                        <div><span className="text-green-400">$</span> sshsk status --checksums</div>
                        <div><span className="text-green-400">$</span> sshsk delete "old-backup" --force</div>
                      </div>
                    </div>

                    <div className="bg-blue-900/30 border border-blue-500/30 rounded p-2 mt-4">
                      <div className="text-blue-400 text-xs">
                        💡 Use --help with any command for detailed usage information<br/>
                        💡 All commands work with environment variables (VAULT_ADDR, VAULT_TOKEN)<br/>
                        💡 Use --dry-run for safe testing of backup and restore operations
                      </div>
                    </div>
                  </div>
                </TerminalWindow>
              )}

              {activeTab === 'config' && (
                <TerminalWindow title="Configuration">
                  <div className="space-y-4 text-sm">
                    <div className="bg-blue-900/30 border border-blue-500/30 rounded p-3">
                      <div className="text-blue-400 font-bold mb-2">⚙️ Environment Variable Configuration</div>
                      <div className="text-gray-300 text-xs">
                        Configure SSH Secret Keeper using environment variables - no config files needed
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Required Environment Variables</div>
                    <div className="text-gray-400"># Essential Vault connection settings</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export VAULT_ADDR="https://vault.company.com:8200"</code>
                      <CopyButton text='export VAULT_ADDR="https://vault.company.com:8200"' />
                    </div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export VAULT_TOKEN="your-vault-token"</code>
                      <CopyButton text='export VAULT_TOKEN="your-vault-token"' />
                    </div>

                    <div className="text-yellow-400 font-bold">## Storage Strategy Configuration</div>
                    <div className="text-gray-400"># Universal storage (default - cross-machine restore)</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_STORAGE_STRATEGY="universal"</code>
                      <CopyButton text='export SSHSK_VAULT_STORAGE_STRATEGY="universal"' />
                    </div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_BACKUP_NAMESPACE="personal"</code>
                      <CopyButton text='export SSHSK_VAULT_BACKUP_NAMESPACE="personal"' />
                    </div>

                    <div className="text-gray-400"># User-scoped storage</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_STORAGE_STRATEGY="user"</code>
                      <CopyButton text='export SSHSK_VAULT_STORAGE_STRATEGY="user"' />
                    </div>

                    <div className="text-gray-400"># Custom team storage</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_STORAGE_STRATEGY="custom"</code>
                      <CopyButton text='export SSHSK_VAULT_STORAGE_STRATEGY="custom"' />
                    </div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_CUSTOM_PREFIX="team-devops"</code>
                      <CopyButton text='export SSHSK_VAULT_CUSTOM_PREFIX="team-devops"' />
                    </div>

                    <div className="text-yellow-400 font-bold">## Optional Configuration</div>
                    <div className="text-gray-400"># Backup behavior</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_BACKUP_SSH_DIR="~/.ssh"</code>
                      <CopyButton text='export SSHSK_BACKUP_SSH_DIR="~/.ssh"' />
                    </div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_BACKUP_RETENTION_COUNT="10"</code>
                      <CopyButton text='export SSHSK_BACKUP_RETENTION_COUNT="10"' />
                    </div>

                    <div className="text-gray-400"># Vault settings</div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_MOUNT_PATH="ssh-backups"</code>
                      <CopyButton text='export SSHSK_VAULT_MOUNT_PATH="ssh-backups"' />
                    </div>
                    <div className="flex items-center">
                      <span className="text-green-400">$</span>
                      <code className="ml-2 text-white flex-1">export SSHSK_VAULT_NAMESPACE="team-namespace"</code>
                      <CopyButton text='export SSHSK_VAULT_NAMESPACE="team-namespace"' />
                    </div>

                    <div className="text-yellow-400 font-bold">## Complete Setup Example</div>
                    <div className="bg-gray-800 rounded p-3">
                      <div className="text-gray-300 text-xs space-y-1">
                        <div className="text-cyan-400"># Complete environment setup for universal storage</div>
                        <div><span className="text-green-400">$</span> export VAULT_ADDR="https://vault.company.com:8200"</div>
                        <div><span className="text-green-400">$</span> export VAULT_TOKEN="hvs.your-vault-token"</div>
                        <div><span className="text-green-400">$</span> export SSHSK_VAULT_STORAGE_STRATEGY="universal"</div>
                        <div><span className="text-green-400">$</span> export SSHSK_VAULT_BACKUP_NAMESPACE="personal"</div>
                        <div><span className="text-green-400">$</span> sshsk init</div>
                        <div><span className="text-green-400">$</span> sshsk backup "my-laptop-keys"</div>
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Environment Variable Priority</div>
                    <div className="bg-gray-800 rounded p-3">
                      <div className="text-gray-300 text-xs space-y-1">
                        <div className="text-yellow-400">Configuration priority (highest to lowest):</div>
                        <div>1. Command line flags</div>
                        <div>2. Environment variables</div>
                        <div>3. Config file (~/.ssh-secret-keeper/config.yaml)</div>
                        <div>4. Default values</div>
                      </div>
                    </div>

                    <div className="text-yellow-400 font-bold">## Configuration Tips</div>
                    <div className="bg-gray-800 rounded p-3">
                      <div className="text-gray-300 text-xs space-y-1">
                        <div>💡 Use environment variables for containers and CI/CD</div>
                        <div>💡 VAULT_TOKEN takes priority over token files</div>
                        <div>💡 Universal storage enables cross-machine restore</div>
                        <div>💡 Set variables in ~/.bashrc or ~/.zshrc for persistence</div>
                        <div>💡 Use --dry-run to test configuration changes safely</div>
                      </div>
                    </div>

                    <div className="bg-green-900/30 border border-green-500/30 rounded p-2 mt-4">
                      <div className="text-green-400 text-xs">
                        ✓ Environment variables work without any config files<br/>
                        ✓ Perfect for containers, CI/CD, and automation<br/>
                        ✓ No secrets stored on disk when using VAULT_TOKEN<br/>
                        ✓ Easy to switch between different Vault instances
                      </div>
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
                  Built-in self-updating with automatic backup
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
                <a href="https://github.com/rafaelvzago/ssh-secret-keeper/blob/main/docs/UPDATE.md" className="block text-cyan-400 hover:text-cyan-300 text-sm">
                  → Update & Self-Updating Guide
                </a>
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
              Built with ❤️ for the community • By Rafael Zago • Licensed under Apache 2.0
            </p>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default App;
