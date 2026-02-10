import { useState, useRef, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import mermaid from 'mermaid';

interface Message {
  role: 'user' | 'assistant';
  content: string;
}

// Initialize mermaid
mermaid.initialize({
  startOnLoad: true,
  theme: 'dark',
  securityLevel: 'loose',
});

// Mermaid diagram component
function MermaidDiagram({ chart }: { chart: string }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (ref.current) {
      try {
        mermaid.contentLoaded();
      } catch (err) {
        console.error('Mermaid rendering error:', err);
      }
    }
  }, [chart]);

  return (
    <div className="mermaid bg-white p-4 rounded my-2" ref={ref}>
      {chart}
    </div>
  );
}

const ROLES = [
  { value: '', label: 'No Role (General)' },
  { value: 'orchestrator', label: 'Orchestrator' },
  { value: 'engineer', label: 'Engineer' },
  { value: 'architect', label: 'Architect' },
  { value: 'designer', label: 'Designer' },
  { value: 'tester', label: 'Tester' },
  { value: 'reviewer', label: 'Reviewer' },
  { value: 'inspector', label: 'Inspector' },
  { value: 'product-manager', label: 'Product Manager' },
  { value: 'archaeologist', label: 'Archaeologist' },
  { value: 'spelunker', label: 'Spelunker' },
  { value: 'strategist', label: 'Strategist' },
];

export default function ChatPanel() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [isStreaming, setIsStreaming] = useState(false);
  const [streamingMessage, setStreamingMessage] = useState('');
  const [selectedRole, setSelectedRole] = useState('');
  const [mode, setMode] = useState<'chat' | 'agent'>('chat');
  const [chatId, setChatId] = useState<string>('');
  const [promptHistory, setPromptHistory] = useState<string[]>([]);
  const [historyIndex, setHistoryIndex] = useState(-1);
  const [tempInput, setTempInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [showSearch, setShowSearch] = useState(false);
  const [showCommandMenu, setShowCommandMenu] = useState(false);
  const [commandFilter, setCommandFilter] = useState('');
  const [showExportMenu, setShowExportMenu] = useState(false);
  const [mentionedFiles, setMentionedFiles] = useState<string[]>([]);
  const [showFileMentions, setShowFileMentions] = useState(false);
  const [_fileMentionQuery, setFileMentionQuery] = useState('');
  const [attachedFiles, setAttachedFiles] = useState<Array<{ name: string; content: string; size: number }>>([]);
  const [attachedImages, setAttachedImages] = useState<Array<{ name: string; dataUrl: string; size: number }>>([]);
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const slashCommands = [
    { name: '/commit', description: 'Create a git commit', action: () => { setInput('Create a git commit with the recent changes'); setMode('agent'); setSelectedRole('engineer'); } },
    { name: '/test', description: 'Run tests', action: () => { setInput('Run the test suite'); setMode('agent'); setSelectedRole('engineer'); } },
    { name: '/review', description: 'Review code changes', action: () => { setInput('Review the recent code changes'); setMode('agent'); setSelectedRole('reviewer'); } },
    { name: '/search', description: 'Search codebase', action: async () => {
      const query = prompt('Enter search query:');
      if (!query) return;
      await performCodebaseSearch(query);
    } },
    { name: '/fix', description: 'Fix an issue', action: () => { setInput('Fix the following issue: '); setMode('agent'); setSelectedRole('engineer'); } },
    { name: '/refactor', description: 'Refactor code', action: () => { setInput('Refactor the following code: '); setMode('agent'); setSelectedRole('engineer'); } },
    { name: '/explain', description: 'Explain code', action: () => { setInput('Explain how this works: '); } },
    { name: '/docs', description: 'Generate documentation', action: () => { setInput('Generate documentation for: '); setMode('agent'); setSelectedRole('engineer'); } },
  ];

  // Auto-switch mode when role changes
  const handleRoleChange = (role: string) => {
    setSelectedRole(role);
    // Auto-select agent mode for specific roles, chat for general
    if (role === '') {
      setMode('chat');
    } else {
      setMode('agent');
    }
  };
  const [projectRoot, setProjectRoot] = useState('');
  const [projectRoots, setProjectRoots] = useState<string[]>([]);
  const [showProjectDropdown, setShowProjectDropdown] = useState(false);
  const [directorySuggestions, setDirectorySuggestions] = useState<string[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const autocompleteTimerRef = useRef<number | null>(null);

  // Load project roots and chat history from localStorage on mount
  useEffect(() => {
    // Load project roots
    const savedRoots = localStorage.getItem('ai-pack-project-roots');
    if (savedRoots) {
      try {
        const roots = JSON.parse(savedRoots);
        setProjectRoots(roots);
        if (roots.length > 0) {
          setProjectRoot(roots[0]); // Default to most recent
        }
      } catch (err) {
        console.error('Failed to load project roots:', err);
      }
    }

    // Load or create chat session
    const savedChatId = localStorage.getItem('ai-pack-current-chat-id');
    if (savedChatId) {
      setChatId(savedChatId);
      const savedMessages = localStorage.getItem(`ai-pack-chat-${savedChatId}`);
      if (savedMessages) {
        try {
          const msgs = JSON.parse(savedMessages);
          setMessages(msgs);
        } catch (err) {
          console.error('Failed to load chat history:', err);
        }
      }

      // Load prompt history
      const savedHistory = localStorage.getItem(`ai-pack-prompt-history-${savedChatId}`);
      if (savedHistory) {
        try {
          const history = JSON.parse(savedHistory);
          setPromptHistory(history);
        } catch (err) {
          console.error('Failed to load prompt history:', err);
        }
      }
    } else {
      // Create new chat session
      const newChatId = Date.now().toString();
      setChatId(newChatId);
      localStorage.setItem('ai-pack-current-chat-id', newChatId);
    }
  }, []);

  // Handle paste events for images
  useEffect(() => {
    const handlePaste = (e: ClipboardEvent) => {
      const items = e.clipboardData?.items;
      if (!items) return;

      for (let i = 0; i < items.length; i++) {
        const item = items[i];
        if (item.type.indexOf('image') !== -1) {
          e.preventDefault();
          const blob = item.getAsFile();
          if (!blob) continue;

          // Convert to base64
          const reader = new FileReader();
          reader.onload = () => {
            const dataUrl = reader.result as string;
            setAttachedImages(prev => [...prev, {
              name: `pasted-image-${Date.now()}.png`,
              dataUrl,
              size: blob.size
            }]);
          };
          reader.readAsDataURL(blob);
        }
      }
    };

    window.addEventListener('paste', handlePaste);
    return () => window.removeEventListener('paste', handlePaste);
  }, []);

  // Save project root to history
  const saveProjectRoot = (root: string) => {
    if (!root.trim()) return;

    const updated = [root, ...projectRoots.filter(r => r !== root)].slice(0, 10); // Keep last 10
    setProjectRoots(updated);
    localStorage.setItem('ai-pack-project-roots', JSON.stringify(updated));
  };

  // Remove project root from history
  const removeProjectRoot = (root: string) => {
    const updated = projectRoots.filter(r => r !== root);
    setProjectRoots(updated);
    localStorage.setItem('ai-pack-project-roots', JSON.stringify(updated));
    if (projectRoot === root) {
      setProjectRoot('');
    }
  };

  // Fetch directory suggestions from server
  const fetchDirectorySuggestions = async (path: string) => {
    if (!path.trim()) {
      setDirectorySuggestions([]);
      return;
    }

    try {
      const response = await fetch(`/api/browse-directories?path=${encodeURIComponent(path)}`);
      if (response.ok) {
        const data = await response.json();
        setDirectorySuggestions(data.directories || []);
        setShowSuggestions(data.directories && data.directories.length > 0);
      }
    } catch (err) {
      console.error('Failed to fetch directory suggestions:', err);
    }
  };

  // Handle input change with debounced autocomplete
  const handleProjectRootChange = (value: string) => {
    setProjectRoot(value);

    // Clear existing timer
    if (autocompleteTimerRef.current) {
      clearTimeout(autocompleteTimerRef.current);
    }

    // If path ends with /, fetch immediately (user wants to see subdirectories)
    if (value.endsWith('/')) {
      fetchDirectorySuggestions(value);
    } else {
      // Debounce autocomplete requests for normal typing
      autocompleteTimerRef.current = setTimeout(() => {
        fetchDirectorySuggestions(value);
      }, 300);
    }
  };

  // Save chat messages to localStorage whenever they change
  useEffect(() => {
    if (chatId && messages.length > 0) {
      localStorage.setItem(`ai-pack-chat-${chatId}`, JSON.stringify(messages));
    }
  }, [messages, chatId]);

  // Auto-scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, streamingMessage]);

  const sendMessage = async () => {
    if (!input.trim() || isStreaming) return;

    // Build message content with attached files and images
    let messageContent = input;
    if (attachedFiles.length > 0) {
      messageContent += '\n\n---\n\n**Attached Files:**\n\n';
      attachedFiles.forEach(file => {
        messageContent += `\n### ${file.name}\n\`\`\`\n${file.content}\n\`\`\`\n`;
      });
    }
    if (attachedImages.length > 0) {
      messageContent += '\n\n**Attached Images:**\n\n';
      attachedImages.forEach((img) => {
        messageContent += `![${img.name}](${img.dataUrl})\n\n`;
      });
    }

    const userMessage: Message = {
      role: 'user',
      content: messageContent,
    };

    setMessages(prev => [...prev, userMessage]);
    const currentInput = input;

    // Add to prompt history
    setPromptHistory(prev => {
      const updated = [currentInput, ...prev.filter(p => p !== currentInput)].slice(0, 50); // Keep last 50
      localStorage.setItem(`ai-pack-prompt-history-${chatId}`, JSON.stringify(updated));
      return updated;
    });

    setInput('');
    setHistoryIndex(-1);
    setTempInput('');
    setAttachedFiles([]);
    setAttachedImages([]);
    setIsStreaming(true);
    setStreamingMessage('');

    try {
      // Fetch with streaming response
      const response = await fetch('/api/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          message: messageContent,
          messages: messages,
          role: selectedRole,
          mode: mode,
          project_root: projectRoot,
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      // Handle agent mode response (JSON, not streaming)
      if (mode === 'agent') {
        const data = await response.json();
        if (data.status === 'agent_spawned') {
          const assistantMessage: Message = {
            role: 'assistant',
            content: `✅ Agent task spawned!\n\n**Task ID:** ${data.task_id}\n\n${data.message}\n\nYou can track this task in the Kanban board.`,
          };
          setMessages(prev => [...prev, assistantMessage]);
          setIsStreaming(false);
          return;
        }
      }

      // Read SSE stream from response body (chat mode)
      const reader = response.body?.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      if (!reader) {
        throw new Error('No response body');
      }

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (line.startsWith('event:')) {
            // Skip event type line
            continue;
          }

          if (line.startsWith('data:')) {
            const data = line.substring(5).trim();
            try {
              const parsed = JSON.parse(data);

              if (parsed.status === 'connected') {
                console.log('Chat stream connected');
              } else if (parsed.status === 'complete') {
                // Completion event - check this BEFORE parsed.text
                const assistantMessage: Message = {
                  role: 'assistant',
                  content: parsed.text || streamingMessage,
                };
                setMessages(prev => [...prev, assistantMessage]);
                setStreamingMessage('');
                setIsStreaming(false);
                return;
              } else if (parsed.text) {
                // For delta events
                setStreamingMessage(prev => prev + parsed.text.replace(/\\n/g, '\n').replace(/\\"/g, '"'));
              } else if (parsed.error) {
                throw new Error(parsed.error);
              }
            } catch (err) {
              console.error('Failed to parse SSE data:', err);
            }
          }
        }
      }

      setIsStreaming(false);
    } catch (err) {
      console.error('Failed to send message:', err);
      setMessages(prev => [
        ...prev,
        {
          role: 'assistant',
          content: `Error: ${err instanceof Error ? err.message : 'Unknown error'}`,
        },
      ]);
      setIsStreaming(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (promptHistory.length === 0) return;

      // Save current input if starting to navigate history
      if (historyIndex === -1) {
        setTempInput(input);
      }

      const newIndex = Math.min(historyIndex + 1, promptHistory.length - 1);
      setHistoryIndex(newIndex);
      setInput(promptHistory[newIndex]);
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (historyIndex === -1) return;

      const newIndex = historyIndex - 1;
      if (newIndex === -1) {
        // Restore the temporary input
        setInput(tempInput);
        setHistoryIndex(-1);
        setTempInput('');
      } else {
        setHistoryIndex(newIndex);
        setInput(promptHistory[newIndex]);
      }
    }
  };

  const rerunPrompt = (prompt: string) => {
    setInput(prompt);
    setHistoryIndex(-1);
    setTempInput('');
  };

  const handleInputChange = (value: string) => {
    setInput(value);

    // Detect slash commands
    if (value.startsWith('/') && !value.includes(' ')) {
      setShowCommandMenu(true);
      setCommandFilter(value.slice(1));
      setShowFileMentions(false);
    } else {
      setShowCommandMenu(false);

      // Detect @ mentions
      const atIndex = value.lastIndexOf('@');
      if (atIndex !== -1 && atIndex === value.lastIndexOf('@')) {
        const afterAt = value.slice(atIndex + 1);
        if (!afterAt.includes(' ')) {
          setShowFileMentions(true);
          setFileMentionQuery(afterAt);
        } else {
          setShowFileMentions(false);
        }
      } else {
        setShowFileMentions(false);
      }
    }
  };

  const removeFileMention = (filename: string) => {
    setMentionedFiles(mentionedFiles.filter(f => f !== filename));
  };

  const executeCommand = (command: typeof slashCommands[0]) => {
    command.action();
    setShowCommandMenu(false);
  };

  const filteredCommands = slashCommands.filter(cmd =>
    cmd.name.toLowerCase().includes(commandFilter.toLowerCase())
  );

  const insertCodeBlock = (language: string = '') => {
    const codeBlock = `\`\`\`${language}\n\n\`\`\``;
    setInput(input + (input ? '\n\n' : '') + codeBlock);
  };

  const handleFileSelect = async (files: FileList | null) => {
    if (!files || files.length === 0) return;

    const newFiles: Array<{ name: string; content: string; size: number }> = [];

    for (let i = 0; i < files.length; i++) {
      const file = files[i];

      // Only handle text files (< 1MB)
      if (file.size > 1024 * 1024) {
        alert(`File ${file.name} is too large (max 1MB)`);
        continue;
      }

      try {
        const content = await file.text();
        newFiles.push({
          name: file.name,
          content,
          size: file.size
        });
      } catch (err) {
        console.error(`Failed to read file ${file.name}:`, err);
        alert(`Failed to read file ${file.name}`);
      }
    }

    setAttachedFiles([...attachedFiles, ...newFiles]);
  };

  const removeAttachedFile = (index: number) => {
    setAttachedFiles(attachedFiles.filter((_, i) => i !== index));
  };

  const removeAttachedImage = (index: number) => {
    setAttachedImages(attachedImages.filter((_, i) => i !== index));
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    handleFileSelect(e.dataTransfer.files);
  };

  const performCodebaseSearch = async (query: string) => {
    if (!query.trim()) return;

    // Add user message showing search query
    const searchMessage: Message = {
      role: 'user',
      content: `🔍 Searching codebase for: **${query}**`,
    };
    setMessages(prev => [...prev, searchMessage]);

    try {
      const response = await fetch('/api/search', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          query,
          project_root: projectRoot || '',
          max_results: 50,
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const data = await response.json();

      // Format results as markdown
      let resultsContent = `### Search Results for "${query}"\n\n`;

      if (data.count === 0) {
        resultsContent += `No matches found.`;
      } else {
        resultsContent += `Found ${data.count} match${data.count !== 1 ? 'es' : ''}\n\n`;

        data.results.forEach((result: any, idx: number) => {
          resultsContent += `**${idx + 1}. ${result.file}:${result.line}**\n\`\`\`\n${result.content}\n\`\`\`\n\n`;
        });
      }

      const resultsMessage: Message = {
        role: 'assistant',
        content: resultsContent,
      };
      setMessages(prev => [...prev, resultsMessage]);
    } catch (err) {
      console.error('Search failed:', err);
      const errorMessage: Message = {
        role: 'assistant',
        content: `❌ Search failed: ${err instanceof Error ? err.message : 'Unknown error'}`,
      };
      setMessages(prev => [...prev, errorMessage]);
    }
  };

  const startNewChat = () => {
    if (isStreaming) return; // Don't start new chat while streaming

    // Create new chat session
    const newChatId = Date.now().toString();
    setChatId(newChatId);
    localStorage.setItem('ai-pack-current-chat-id', newChatId);

    // Clear messages
    setMessages([]);
    setInput('');
    setStreamingMessage('');
  };

  const estimateTokens = (text: string): number => {
    // Rough estimate: ~4 characters per token for English text
    return Math.ceil(text.length / 4);
  };

  const getTotalTokens = (): number => {
    let total = 0;
    messages.forEach(msg => {
      total += estimateTokens(msg.content);
    });
    if (streamingMessage) {
      total += estimateTokens(streamingMessage);
    }
    if (input) {
      total += estimateTokens(input);
    }
    return total;
  };

  const filteredMessages = searchQuery
    ? messages.filter(msg =>
        msg.content.toLowerCase().includes(searchQuery.toLowerCase())
      )
    : messages;

  const exportToMarkdown = () => {
    const timestamp = new Date().toISOString().split('T')[0];
    const filename = `chat-export-${timestamp}.md`;

    let markdown = `# Claude Assistant Chat Export\n\n`;
    markdown += `**Date:** ${new Date().toLocaleString()}\n`;
    markdown += `**Role:** ${selectedRole || 'General'}\n`;
    markdown += `**Mode:** ${mode}\n`;
    if (projectRoot) {
      markdown += `**Project:** ${projectRoot}\n`;
    }
    markdown += `\n---\n\n`;

    messages.forEach((msg, idx) => {
      markdown += `## ${msg.role === 'user' ? '👤 User' : '🤖 Assistant'}\n\n`;
      markdown += `${msg.content}\n\n`;
      if (idx < messages.length - 1) {
        markdown += `---\n\n`;
      }
    });

    const blob = new Blob([markdown], { type: 'text/markdown' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
  };

  const exportToJSON = () => {
    const timestamp = new Date().toISOString().split('T')[0];
    const filename = `chat-export-${timestamp}.json`;

    const exportData = {
      metadata: {
        exportDate: new Date().toISOString(),
        chatId,
        role: selectedRole || 'General',
        mode,
        projectRoot: projectRoot || null,
        messageCount: messages.length,
        totalTokens: getTotalTokens(),
      },
      messages: messages.map((msg, idx) => ({
        index: idx,
        role: msg.role,
        content: msg.content,
        timestamp: new Date().toISOString(), // Would need to track actual timestamps
      })),
      attachments: {
        files: attachedFiles.map(f => ({ name: f.name, size: f.size })),
        images: attachedImages.map(i => ({ name: i.name, size: i.size })),
      },
    };

    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="w-full h-full bg-gray-800 flex flex-col">
      {/* Header */}
      <div className="p-3 border-b border-gray-700 bg-gray-900">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 rounded-full bg-green-400"></div>
            <h3 className="font-semibold">Claude Assistant</h3>
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => setShowSearch(!showSearch)}
              disabled={messages.length === 0}
              className="px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 disabled:bg-gray-800 disabled:cursor-not-allowed text-gray-300 rounded transition-colors"
              title="Search chat history"
            >
              🔍 Search
            </button>
            <div className="relative">
              <button
                onClick={() => setShowExportMenu(!showExportMenu)}
                disabled={messages.length === 0}
                className="px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 disabled:bg-gray-800 disabled:cursor-not-allowed text-gray-300 rounded transition-colors"
                title="Export chat"
              >
                📥 Export ▼
              </button>
              {showExportMenu && messages.length > 0 && (
                <div className="absolute top-full right-0 mt-1 bg-gray-800 border border-gray-700 rounded-lg shadow-lg z-10 min-w-32">
                  <button
                    onClick={() => { exportToMarkdown(); setShowExportMenu(false); }}
                    className="w-full text-left px-3 py-2 text-xs hover:bg-gray-700 transition-colors text-white rounded-t-lg"
                  >
                    📄 Markdown
                  </button>
                  <button
                    onClick={() => { exportToJSON(); setShowExportMenu(false); }}
                    className="w-full text-left px-3 py-2 text-xs hover:bg-gray-700 transition-colors text-white rounded-b-lg"
                  >
                    📋 JSON
                  </button>
                </div>
              )}
            </div>
            <button
              onClick={startNewChat}
              disabled={isStreaming}
              className="px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 disabled:bg-gray-800 disabled:cursor-not-allowed text-gray-300 rounded transition-colors"
              title="Start new chat"
            >
              + New Chat
            </button>
          </div>
        </div>

        {/* Search Bar */}
        {showSearch && (
          <div className="mb-2 flex items-center gap-2">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search messages..."
              className="flex-1 bg-gray-800 text-white rounded px-2 py-1 text-xs focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                className="text-xs text-gray-400 hover:text-white"
              >
                ✕
              </button>
            )}
            <span className="text-xs text-gray-500">
              {filteredMessages.length}/{messages.length}
            </span>
          </div>
        )}

        {/* Mode Toggle */}
        <div className="flex gap-2 mb-2">
          <button
            onClick={() => setMode('chat')}
            className={`flex-1 px-3 py-1.5 text-xs rounded transition-colors ${
              mode === 'chat'
                ? 'bg-blue-600 text-white'
                : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
            }`}
          >
            💬 Chat
          </button>
          <button
            onClick={() => setMode('agent')}
            className={`flex-1 px-3 py-1.5 text-xs rounded transition-colors ${
              mode === 'agent'
                ? 'bg-purple-600 text-white'
                : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
            }`}
          >
            🤖 Agent
          </button>
        </div>

        {/* Role Selector */}
        <select
          value={selectedRole}
          onChange={(e) => handleRoleChange(e.target.value)}
          className="w-full px-3 py-1.5 bg-gray-700 text-white text-xs rounded focus:outline-none focus:ring-2 focus:ring-blue-500 mb-2"
        >
          {ROLES.map(role => (
            <option key={role.value} value={role.value}>
              {role.label}
            </option>
          ))}
        </select>

        {/* Project Selector */}
        <div className="relative">
          <div className="flex gap-1">
            <input
              type="text"
              value={projectRoot}
              onChange={(e) => handleProjectRootChange(e.target.value)}
              onFocus={() => {
                if (projectRoot) {
                  fetchDirectorySuggestions(projectRoot);
                }
              }}
              onBlur={() => {
                // Delay to allow clicking on dropdown items
                setTimeout(() => {
                  setShowProjectDropdown(false);
                  setShowSuggestions(false);
                }, 200);
                if (projectRoot.trim()) {
                  saveProjectRoot(projectRoot);
                }
              }}
              placeholder="Project root (e.g., ~/Projects/myapp)"
              className="flex-1 px-3 py-1.5 bg-gray-700 text-white text-xs rounded focus:outline-none focus:ring-2 focus:ring-blue-500 placeholder-gray-500"
            />
            {projectRoots.length > 0 && (
              <button
                onClick={() => setShowProjectDropdown(!showProjectDropdown)}
                className="px-2 py-1.5 bg-gray-700 hover:bg-gray-600 text-white text-xs rounded"
                title="Show saved projects"
              >
                ▼
              </button>
            )}
          </div>

          {/* Filesystem autocomplete suggestions */}
          {showSuggestions && directorySuggestions.length > 0 && (
            <div className="absolute top-full left-0 right-0 mt-1 bg-gray-700 rounded shadow-lg border border-gray-600 z-20 max-h-40 overflow-y-auto">
              {directorySuggestions.map((dir, idx) => (
                <button
                  key={idx}
                  onClick={() => {
                    setProjectRoot(dir);
                    setShowSuggestions(false);
                    saveProjectRoot(dir);
                  }}
                  className="w-full text-left px-3 py-2 hover:bg-gray-600 text-xs text-white truncate block"
                  title={dir}
                >
                  📁 {dir}
                </button>
              ))}
            </div>
          )}

          {/* Dropdown with saved projects */}
          {showProjectDropdown && projectRoots.length > 0 && (
            <div className="absolute top-full left-0 right-0 mt-1 bg-gray-700 rounded shadow-lg border border-gray-600 z-10 max-h-40 overflow-y-auto">
              {projectRoots.map((root, idx) => (
                <div
                  key={idx}
                  className="flex items-center justify-between px-3 py-2 hover:bg-gray-600 group"
                >
                  <button
                    onClick={() => {
                      setProjectRoot(root);
                      setShowProjectDropdown(false);
                    }}
                    className="flex-1 text-left text-xs text-white truncate"
                    title={root}
                  >
                    {root}
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      removeProjectRoot(root);
                    }}
                    className="ml-2 px-1.5 py-0.5 text-xs text-red-400 hover:text-red-300 opacity-0 group-hover:opacity-100 transition-opacity"
                    title="Remove"
                  >
                    ✕
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.length === 0 && (
          <div className="text-center text-gray-500 mt-8">
            <p className="text-lg mb-2">👋 Hi! I'm Claude.</p>
            <p className="text-sm">Ask me anything about your agents, code, or anything else!</p>
          </div>
        )}

        {filteredMessages.map((msg, idx) => (
          <div
            key={idx}
            className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
          >
            <div
              className={`max-w-[85%] rounded-lg px-4 py-2 ${
                msg.role === 'user'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-700 text-gray-100'
              }`}
            >
              {msg.role === 'assistant' ? (
                <div className="prose prose-invert prose-sm max-w-none">
                  <ReactMarkdown
                    components={{
                      code({ inline, className, children, ...props }: any) {
                        const match = /language-(\w+)/.exec(className || '');
                        const language = match ? match[1] : '';

                        // Handle mermaid diagrams
                        if (!inline && language === 'mermaid') {
                          return <MermaidDiagram chart={String(children)} />;
                        }

                        // Handle regular code blocks
                        return !inline && match ? (
                          <SyntaxHighlighter
                            style={vscDarkPlus as any}
                            language={language}
                            PreTag="div"
                            {...props}
                          >
                            {String(children).replace(/\n$/, '')}
                          </SyntaxHighlighter>
                        ) : (
                          <code className={className} {...props}>
                            {children}
                          </code>
                        );
                      },
                    }}
                  >
                    {msg.content}
                  </ReactMarkdown>
                </div>
              ) : (
                <div className="flex items-start gap-2">
                  <p className="whitespace-pre-wrap flex-1">{msg.content}</p>
                  <button
                    onClick={() => rerunPrompt(msg.content)}
                    className="flex-shrink-0 text-xs opacity-60 hover:opacity-100 transition-opacity"
                    title="Rerun this prompt"
                  >
                    🔄
                  </button>
                </div>
              )}
            </div>
          </div>
        ))}

        {/* Streaming message */}
        {isStreaming && streamingMessage && (
          <div className="flex justify-start">
            <div className="max-w-[85%] rounded-lg px-4 py-2 bg-gray-700 text-gray-100">
              <div className="prose prose-invert prose-sm max-w-none">
                <ReactMarkdown>{streamingMessage}</ReactMarkdown>
              </div>
            </div>
          </div>
        )}

        {/* Typing indicator */}
        {isStreaming && !streamingMessage && (
          <div className="flex justify-start">
            <div className="bg-gray-700 rounded-lg px-4 py-3">
              <div className="flex gap-1">
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.4s' }}></div>
              </div>
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div
        className={`p-3 border-t border-gray-700 bg-gray-900 ${isDragging ? 'border-blue-500 border-2 bg-gray-800' : ''}`}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        {isDragging && (
          <div className="mb-2 p-4 border-2 border-dashed border-blue-500 rounded-lg text-center text-blue-400 text-sm">
            Drop files here to attach...
          </div>
        )}
        {mentionedFiles.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-2">
            {mentionedFiles.map((file, idx) => (
              <span key={idx} className="inline-flex items-center gap-1 px-2 py-1 bg-blue-600 text-white text-xs rounded">
                📄 {file}
                <button
                  onClick={() => removeFileMention(file)}
                  className="hover:text-gray-300"
                >
                  ✕
                </button>
              </span>
            ))}
          </div>
        )}
        {attachedFiles.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-2">
            {attachedFiles.map((file, idx) => (
              <span key={idx} className="inline-flex items-center gap-1 px-2 py-1 bg-green-600 text-white text-xs rounded">
                📎 {file.name} <span className="text-green-200">({Math.round(file.size / 1024)}KB)</span>
                <button
                  onClick={() => removeAttachedFile(idx)}
                  className="hover:text-gray-300"
                >
                  ✕
                </button>
              </span>
            ))}
          </div>
        )}
        {attachedImages.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-2">
            {attachedImages.map((img, idx) => (
              <div key={idx} className="relative inline-block group">
                <img
                  src={img.dataUrl}
                  alt={img.name}
                  className="h-20 w-auto rounded border-2 border-purple-600"
                />
                <button
                  onClick={() => removeAttachedImage(idx)}
                  className="absolute -top-2 -right-2 bg-red-600 text-white rounded-full w-5 h-5 flex items-center justify-center text-xs hover:bg-red-700 opacity-0 group-hover:opacity-100 transition-opacity"
                  title="Remove image"
                >
                  ✕
                </button>
                <div className="text-xs text-gray-400 mt-1 text-center">{Math.round(img.size / 1024)}KB</div>
              </div>
            ))}
          </div>
        )}
        <div className="flex gap-1 mb-2">
          <button
            onClick={() => insertCodeBlock()}
            className="text-xs px-2 py-1 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
            title="Insert code block"
          >
            &lt;/&gt;
          </button>
          <button
            onClick={() => insertCodeBlock('javascript')}
            className="text-xs px-2 py-1 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
            title="Insert JavaScript code block"
          >
            JS
          </button>
          <button
            onClick={() => insertCodeBlock('python')}
            className="text-xs px-2 py-1 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
            title="Insert Python code block"
          >
            PY
          </button>
          <button
            onClick={() => insertCodeBlock('go')}
            className="text-xs px-2 py-1 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
            title="Insert Go code block"
          >
            GO
          </button>
          <button
            onClick={() => fileInputRef.current?.click()}
            className="text-xs px-2 py-1 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
            title="Attach files (or drag & drop)"
          >
            📎
          </button>
          <input
            ref={fileInputRef}
            type="file"
            multiple
            onChange={(e) => handleFileSelect(e.target.files)}
            className="hidden"
            accept=".txt,.md,.json,.js,.ts,.tsx,.jsx,.go,.py,.java,.c,.cpp,.h,.hpp,.css,.html,.xml,.yaml,.yml,.toml,.ini,.sh,.bash"
          />
        </div>
        <div className="flex gap-2">
          <div className="flex-1 relative">
            {showCommandMenu && filteredCommands.length > 0 && (
              <div className="absolute bottom-full left-0 mb-1 w-full bg-gray-800 border border-gray-700 rounded-lg shadow-lg max-h-48 overflow-y-auto z-10">
                {filteredCommands.map((cmd, idx) => (
                  <button
                    key={idx}
                    onClick={() => executeCommand(cmd)}
                    className="w-full text-left px-3 py-2 hover:bg-gray-700 transition-colors"
                  >
                    <div className="text-sm text-white">{cmd.name}</div>
                    <div className="text-xs text-gray-400">{cmd.description}</div>
                  </button>
                ))}
              </div>
            )}
            {showFileMentions && (
              <div className="absolute bottom-full left-0 mb-1 w-full bg-gray-800 border border-gray-700 rounded-lg shadow-lg max-h-48 overflow-y-auto z-10">
                <div className="px-3 py-2 text-xs text-gray-400">
                  Type filename to mention (e.g., @src/App.tsx)
                </div>
                <div className="px-3 py-2 text-xs text-gray-500">
                  Tip: Use full paths for better context
                </div>
              </div>
            )}
            <textarea
              value={input}
              onChange={(e) => handleInputChange(e.target.value)}
              onKeyPress={handleKeyPress}
              placeholder="Ask me anything... (type / for commands)"
              disabled={isStreaming}
              className="w-full bg-gray-800 text-white rounded-lg px-3 py-2 text-sm resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
              rows={2}
            />
          </div>
          <button
            onClick={sendMessage}
            disabled={!input.trim() || isStreaming}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white rounded-lg transition-colors text-sm font-medium"
          >
            {isStreaming ? '...' : 'Send'}
          </button>
        </div>
        <div className="flex items-center justify-between mt-2">
          <p className="text-xs text-gray-500">Press Enter to send, Shift+Enter for new line, ↑↓ for history</p>
          <p className="text-xs text-gray-500">
            <span className={getTotalTokens() > 180000 ? 'text-red-400' : getTotalTokens() > 150000 ? 'text-yellow-400' : ''}>
              ~{getTotalTokens().toLocaleString()} tokens
            </span>
          </p>
        </div>
      </div>
    </div>
  );
}
