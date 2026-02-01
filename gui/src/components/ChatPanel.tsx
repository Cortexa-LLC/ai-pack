import { useState, useRef, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

interface Message {
  role: 'user' | 'assistant';
  content: string;
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
  { value: 'cartographer', label: 'Cartographer' },
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
    } else {
      // Create new chat session
      const newChatId = Date.now().toString();
      setChatId(newChatId);
      localStorage.setItem('ai-pack-current-chat-id', newChatId);
    }
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

    const userMessage: Message = {
      role: 'user',
      content: input,
    };

    setMessages(prev => [...prev, userMessage]);
    const currentInput = input;
    setInput('');
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
          message: currentInput,
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

  return (
    <div className="w-full h-full bg-gray-800 flex flex-col">
      {/* Header */}
      <div className="p-3 border-b border-gray-700 bg-gray-900">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 rounded-full bg-green-400"></div>
            <h3 className="font-semibold">Claude Assistant</h3>
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

        {messages.map((msg, idx) => (
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
                        return !inline && match ? (
                          <SyntaxHighlighter
                            style={vscDarkPlus as any}
                            language={match[1]}
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
                <p className="whitespace-pre-wrap">{msg.content}</p>
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
      <div className="p-3 border-t border-gray-700 bg-gray-900">
        <div className="flex gap-2">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder="Ask me anything..."
            disabled={isStreaming}
            className="flex-1 bg-gray-800 text-white rounded-lg px-3 py-2 text-sm resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
            rows={2}
          />
          <button
            onClick={sendMessage}
            disabled={!input.trim() || isStreaming}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white rounded-lg transition-colors text-sm font-medium"
          >
            {isStreaming ? '...' : 'Send'}
          </button>
        </div>
        <p className="text-xs text-gray-500 mt-2">Press Enter to send, Shift+Enter for new line</p>
      </div>
    </div>
  );
}
