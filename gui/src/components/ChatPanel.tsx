import { useState, useRef, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import mermaid from 'mermaid';

interface Message {
  role: 'user' | 'assistant';
  content: string;
  replyTo?: number; // Index of the message being replied to
  id?: number; // Unique ID for this message
}

interface ChatSession {
  id: string;
  name: string;
  messages: Message[];
  createdAt: number;
  updatedAt: number;
}

interface ProjectChats {
  [projectPath: string]: ChatSession[];
}

interface CurrentChatPerProject {
  [projectPath: string]: string; // chatId
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

export default function ChatPanel() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [isStreaming, setIsStreaming] = useState(false);
  const [streamingMessage, setStreamingMessage] = useState('');
  const [selectedRole, setSelectedRole] = useState('orchestrator');
  const [mode] = useState<'chat' | 'agent'>('chat'); // Always chat mode with orchestrator
  const [chatId, setChatId] = useState<string>('');
  const [promptHistory, setPromptHistory] = useState<string[]>([]);
  const [historyIndex, setHistoryIndex] = useState(-1);
  const [tempInput, setTempInput] = useState('');
  const [showCommandMenu, setShowCommandMenu] = useState(false);
  const [commandFilter, setCommandFilter] = useState('');
  const [detectedFiles, setDetectedFiles] = useState<string[]>([]);
  const [replyingTo, setReplyingTo] = useState<number | null>(null);
  const [mentionedFiles, setMentionedFiles] = useState<string[]>([]);
  const [showFileMentions, setShowFileMentions] = useState(false);
  const [_fileMentionQuery, setFileMentionQuery] = useState('');
  const [attachedFiles, setAttachedFiles] = useState<Array<{ name: string; content: string; size: number }>>([]);
  const [attachedImages, setAttachedImages] = useState<Array<{ name: string; dataUrl: string; size: number }>>([]);
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [currentChatName, setCurrentChatName] = useState('New Chat');
  const [projectChats, setProjectChats] = useState<ChatSession[]>([]);
  const [showChatList, setShowChatList] = useState(false);
  const [suggestion, setSuggestion] = useState('');
  const [useProjectContext, setUseProjectContext] = useState(true);
  const [contextLoadedFile, setContextLoadedFile] = useState('');
  const abortControllerRef = useRef<AbortController | null>(null);
  const [orchestratorMonitoring, setOrchestratorMonitoring] = useState(false);
  const orchestratorEventSourceRef = useRef<EventSource | null>(null);
  const [modal, setModal] = useState<{
    show: boolean;
    title: string;
    message: string;
    type: 'alert' | 'confirm';
    onConfirm?: () => void;
  }>({ show: false, title: '', message: '', type: 'alert' });

  // Helper functions for modals
  const showAlert = (title: string, message: string) => {
    setModal({ show: true, title, message, type: 'alert' });
  };

  const showConfirm = (title: string, message: string, onConfirm: () => void) => {
    setModal({ show: true, title, message, type: 'confirm', onConfirm });
  };

  // Helper functions for project-based chat management
  const loadProjectChats = (projectPath: string): ChatSession[] => {
    try {
      const allChats = localStorage.getItem('ai-pack-project-chats');
      if (!allChats) return [];
      const parsed: ProjectChats = JSON.parse(allChats);
      return parsed[projectPath] || [];
    } catch (err) {
      console.error('Failed to load project chats:', err);
      return [];
    }
  };

  const saveProjectChats = (projectPath: string, chats: ChatSession[]) => {
    try {
      const allChats = localStorage.getItem('ai-pack-project-chats');
      const parsed: ProjectChats = allChats ? JSON.parse(allChats) : {};
      parsed[projectPath] = chats;
      localStorage.setItem('ai-pack-project-chats', JSON.stringify(parsed));
    } catch (err) {
      console.error('Failed to save project chats:', err);
    }
  };

  const getCurrentChatForProject = (projectPath: string): string | null => {
    try {
      const currentChats = localStorage.getItem('ai-pack-current-chat-per-project');
      if (!currentChats) return null;
      const parsed: CurrentChatPerProject = JSON.parse(currentChats);
      return parsed[projectPath] || null;
    } catch (err) {
      console.error('Failed to get current chat for project:', err);
      return null;
    }
  };

  const setCurrentChatForProject = (projectPath: string, chatId: string) => {
    try {
      const currentChats = localStorage.getItem('ai-pack-current-chat-per-project');
      const parsed: CurrentChatPerProject = currentChats ? JSON.parse(currentChats) : {};
      parsed[projectPath] = chatId;
      localStorage.setItem('ai-pack-current-chat-per-project', JSON.stringify(parsed));
    } catch (err) {
      console.error('Failed to set current chat for project:', err);
    }
  };

  const generateChatName = (firstMessage: string): string => {
    const cleaned = firstMessage.trim().replace(/\n/g, ' ');
    if (cleaned.length === 0) {
      return `New Chat ${new Date().toLocaleDateString()}`;
    }
    return cleaned.substring(0, 50) + (cleaned.length > 50 ? '...' : '');
  };

  const migrateOldChatToProject = () => {
    const oldChatId = localStorage.getItem('ai-pack-current-chat-id');
    if (!oldChatId) return;

    const oldMessages = localStorage.getItem(`ai-pack-chat-${oldChatId}`);
    if (!oldMessages) return;

    try {
      const msgs = JSON.parse(oldMessages);
      const defaultProject = '/default';
      const migratedChat: ChatSession = {
        id: oldChatId,
        name: 'Migrated Chat',
        messages: msgs,
        createdAt: Date.now(),
        updatedAt: Date.now()
      };

      saveProjectChats(defaultProject, [migratedChat]);
      setCurrentChatForProject(defaultProject, oldChatId);

      // Clean up old keys
      localStorage.removeItem('ai-pack-current-chat-id');
      localStorage.removeItem(`ai-pack-chat-${oldChatId}`);

      console.log('Successfully migrated old chat to default project');
    } catch (err) {
      console.error('Failed to migrate old chat:', err);
    }
  };

  const slashCommands = [
    { name: '/commit', description: 'Create a git commit', action: () => { setInput('Spawn an engineer to create a git commit with the recent changes'); } },
    { name: '/test', description: 'Run tests', action: () => { setInput('Spawn an engineer to run the test suite'); } },
    { name: '/review', description: 'Review code changes', action: () => { setInput('Spawn a reviewer to review the recent code changes'); } },
    { name: '/search', description: 'Search codebase', action: () => { setInput('Search the codebase for: '); } },
    { name: '/fix', description: 'Fix an issue', action: () => { setInput('Spawn an engineer to fix the following issue: '); } },
    { name: '/refactor', description: 'Refactor code', action: () => { setInput('Spawn an engineer to refactor the following code: '); } },
    { name: '/explain', description: 'Explain code', action: () => { setInput('Explain how this works: '); } },
    { name: '/docs', description: 'Generate documentation', action: () => { setInput('Spawn an engineer to generate documentation for: '); } },
    { name: '/status', description: 'Check system status', action: () => { setInput('What is the current status of all tasks and agents?'); } },
  ];

  const [projectRoot, setProjectRoot] = useState('');
  const [projectRoots, setProjectRoots] = useState<string[]>([]);
  const [showProjectDropdown, setShowProjectDropdown] = useState(false);
  const [directorySuggestions, setDirectorySuggestions] = useState<string[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const autocompleteTimerRef = useRef<number | null>(null);
  const lastProjectRef = useRef<string>('');

  // Chat management functions
  const createNewChatInProject = (project: string) => {
    if (!project) {
      showAlert('Project Required', 'Please select a project first');
      return;
    }

    const newChat: ChatSession = {
      id: Date.now().toString(),
      name: 'New Chat',
      messages: [],
      createdAt: Date.now(),
      updatedAt: Date.now()
    };

    const updatedChats = [newChat, ...projectChats];
    setProjectChats(updatedChats);
    saveProjectChats(project, updatedChats);
    switchToChat(newChat);
    setCurrentChatForProject(project, newChat.id);
  };

  const switchToChat = (chat: ChatSession) => {
    setChatId(chat.id);
    setMessages(chat.messages);
    setCurrentChatName(chat.name);
    setShowChatList(false);
  };

  const deleteChat = (chatIdToDelete: string) => {
    if (!projectRoot) return;

    const updatedChats = projectChats.filter(c => c.id !== chatIdToDelete);
    setProjectChats(updatedChats);
    saveProjectChats(projectRoot, updatedChats);

    // If deleting current chat, switch to first available
    if (chatIdToDelete === chatId && updatedChats.length > 0) {
      switchToChat(updatedChats[0]);
      setCurrentChatForProject(projectRoot, updatedChats[0].id);
    } else if (updatedChats.length === 0) {
      // Create new chat if none left
      createNewChatInProject(projectRoot);
    }
  };

  const updateChatName = (chatIdToUpdate: string, firstMessage: string) => {
    if (!projectRoot) return;

    const chatToUpdate = projectChats.find(c => c.id === chatIdToUpdate);
    if (!chatToUpdate || chatToUpdate.messages.length > 0) return; // Only update on first message

    const newName = generateChatName(firstMessage);
    const updatedChats = projectChats.map(c =>
      c.id === chatIdToUpdate ? { ...c, name: newName, updatedAt: Date.now() } : c
    );

    setProjectChats(updatedChats);
    setCurrentChatName(newName);
    saveProjectChats(projectRoot, updatedChats);
  };

  // Load project roots and chat history from localStorage on mount
  useEffect(() => {
    // Migrate old chat data if exists
    migrateOldChatToProject();

    // Load use project context preference
    const savedUseContext = localStorage.getItem('ai-pack-use-project-context');
    if (savedUseContext !== null) {
      setUseProjectContext(savedUseContext === 'true');
    }

    // Load project roots
    const savedRoots = localStorage.getItem('ai-pack-project-roots');
    if (savedRoots) {
      try {
        const roots = JSON.parse(savedRoots);
        setProjectRoots(roots);
      } catch (err) {
        console.error('Failed to load project roots:', err);
      }
    }

    // Load current project
    const currentProj = localStorage.getItem('ai-pack-current-project');
    if (currentProj) {
      setProjectRoot(currentProj);

      // Load chats for this project
      const chats = loadProjectChats(currentProj);
      setProjectChats(chats);

      // Load current chat for this project
      const currentChatId = getCurrentChatForProject(currentProj);

      if (currentChatId && chats.find(c => c.id === currentChatId)) {
        // Load existing chat
        const chat = chats.find(c => c.id === currentChatId)!;
        setChatId(chat.id);
        setMessages(chat.messages);
        setCurrentChatName(chat.name);
      } else if (chats.length > 0) {
        // Load first chat
        const chat = chats[0];
        setChatId(chat.id);
        setMessages(chat.messages);
        setCurrentChatName(chat.name);
        setCurrentChatForProject(currentProj, chat.id);
      } else {
        // Create first chat for this project
        const newChat: ChatSession = {
          id: Date.now().toString(),
          name: 'New Chat',
          messages: [],
          createdAt: Date.now(),
          updatedAt: Date.now()
        };
        setChatId(newChat.id);
        setMessages([]);
        setCurrentChatName(newChat.name);
        setProjectChats([newChat]);
        saveProjectChats(currentProj, [newChat]);
        setCurrentChatForProject(currentProj, newChat.id);
      }

      // Load prompt history for this chat
      if (currentChatId) {
        const savedHistory = localStorage.getItem(`ai-pack-prompt-history-${currentChatId}`);
        if (savedHistory) {
          try {
            const history = JSON.parse(savedHistory);
            setPromptHistory(history);
          } catch (err) {
            console.error('Failed to load prompt history:', err);
          }
        }
      }
    } else {
      // No project selected - create default empty state
      const newChatId = Date.now().toString();
      setChatId(newChatId);
      setMessages([]);
      setCurrentChatName('New Chat');
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

  // Handle ESC key to cancel streaming
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isStreaming) {
        e.preventDefault();
        if (abortControllerRef.current) {
          abortControllerRef.current.abort();
          setIsStreaming(false);
          setStreamingMessage('');
          // Add cancelled message
          setMessages(prev => [...prev, {
            role: 'assistant',
            content: '⚠️ Response cancelled by user',
          }]);
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isStreaming]);

  // Watch for project changes and load appropriate chats
  useEffect(() => {
    if (!projectRoot) return;

    // Prevent running if project hasn't actually changed
    if (lastProjectRef.current === projectRoot) return;
    lastProjectRef.current = projectRoot;

    // Save current project
    localStorage.setItem('ai-pack-current-project', projectRoot);

    // Load chats for new project
    const chats = loadProjectChats(projectRoot);
    setProjectChats(chats);

    // Load last active chat for this project
    const currentChatId = getCurrentChatForProject(projectRoot);

    if (currentChatId && chats.find(c => c.id === currentChatId)) {
      const chat = chats.find(c => c.id === currentChatId)!;
      setChatId(chat.id);
      setMessages(chat.messages);
      setCurrentChatName(chat.name);
    } else if (chats.length > 0) {
      const chat = chats[0];
      setChatId(chat.id);
      setMessages(chat.messages);
      setCurrentChatName(chat.name);
      setCurrentChatForProject(projectRoot, chat.id);
    } else {
      // Create first chat
      const newChat: ChatSession = {
        id: Date.now().toString(),
        name: 'New Chat',
        messages: [],
        createdAt: Date.now(),
        updatedAt: Date.now()
      };
      setChatId(newChat.id);
      setMessages([]);
      setCurrentChatName(newChat.name);
      setProjectChats([newChat]);
      saveProjectChats(projectRoot, [newChat]);
      setCurrentChatForProject(projectRoot, newChat.id);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [projectRoot]);

  // Connect to orchestrator SSE stream when project is selected
  useEffect(() => {
    if (!projectRoot) {
      // Disconnect if no project
      if (orchestratorEventSourceRef.current) {
        orchestratorEventSourceRef.current.close();
        orchestratorEventSourceRef.current = null;
        setOrchestratorMonitoring(false);
      }
      return;
    }

    // Connect to orchestrator SSE
    const eventSource = new EventSource(
      `/api/orchestrator/stream?project_root=${encodeURIComponent(projectRoot)}`
    );

    eventSource.onopen = () => {
      console.log('[Orchestrator] SSE connected');
      setOrchestratorMonitoring(true);
    };

    eventSource.addEventListener('connected', (e) => {
      const data = JSON.parse(e.data);
      console.log('[Orchestrator] Session connected:', data.session_id);
    });

    eventSource.addEventListener('update', (e) => {
      const update = JSON.parse(e.data);
      console.log('[Orchestrator] Update:', update);

      // Add orchestrator update as assistant message
      const orchestratorMessage: Message = {
        role: 'assistant',
        content: `🤖 ${update.message}`,
      };
      setMessages(prev => [...prev, orchestratorMessage]);
    });

    eventSource.onerror = (err) => {
      console.error('[Orchestrator] SSE error:', err);
      setOrchestratorMonitoring(false);
      // Will automatically reconnect
    };

    orchestratorEventSourceRef.current = eventSource;

    // Cleanup on unmount or project change
    return () => {
      eventSource.close();
      setOrchestratorMonitoring(false);
    };
  }, [projectRoot]);

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
    if (!chatId || !projectRoot || messages.length === 0) return;

    // Load current chats from localStorage to avoid stale state
    const currentChats = loadProjectChats(projectRoot);
    if (currentChats.length === 0) return;

    // Check if this is the first user message for auto-naming
    const shouldAutoName = messages.length === 1 && messages[0].role === 'user';
    const existingChat = currentChats.find(c => c.id === chatId);

    if (shouldAutoName && existingChat && existingChat.messages.length === 0) {
      // Auto-name on first message
      const newName = generateChatName(messages[0].content);
      const updatedChats = currentChats.map(c =>
        c.id === chatId
          ? { ...c, name: newName, messages, updatedAt: Date.now() }
          : c
      );
      saveProjectChats(projectRoot, updatedChats);
      setProjectChats(updatedChats);
      setCurrentChatName(newName);
    } else {
      // Just update messages
      const updatedChats = currentChats.map(c =>
        c.id === chatId
          ? { ...c, messages, updatedAt: Date.now() }
          : c
      );
      saveProjectChats(projectRoot, updatedChats);
      setProjectChats(updatedChats);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [messages, chatId, projectRoot]);

  // Auto-scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, streamingMessage]);


  const sendMessage = async () => {
    console.log('[ChatPanel] sendMessage called', {
      inputLength: input.length,
      isStreaming,
      projectRoot,
      chatId,
      mode,
      selectedRole,
      messagesCount: messages.length
    });

    if (!input.trim() || isStreaming) {
      console.log('[ChatPanel] sendMessage aborted - input empty or already streaming');
      return;
    }

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
      replyTo: replyingTo !== null ? replyingTo : undefined,
      id: Date.now(),
    };

    console.log('[ChatPanel] Adding user message to state');
    setMessages(prev => [...prev, userMessage]);
    setReplyingTo(null); // Clear reply state after sending
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

    const requestPayload = {
      message: messageContent,
      messages: messages,
      role: selectedRole,
      mode: mode,
      project_root: projectRoot,
      use_project_context: useProjectContext,
    };

    console.log('[ChatPanel] Preparing to fetch /api/chat', {
      url: '/api/chat',
      payload: requestPayload
    });

    // Create AbortController for cancellation
    const abortController = new AbortController();
    abortControllerRef.current = abortController;

    try {
      // Fetch with streaming response
      console.log('[ChatPanel] Calling fetch...');
      const response = await fetch('/api/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestPayload),
        signal: abortController.signal,
      });

      console.log('[ChatPanel] Fetch response received', {
        ok: response.ok,
        status: response.status,
        statusText: response.statusText,
        headers: Object.fromEntries(response.headers.entries())
      });

      if (!response.ok) {
        console.error('[ChatPanel] Response not OK', {
          status: response.status,
          statusText: response.statusText
        });
        // Try to parse JSON error message
        try {
          const errorData = await response.json();
          throw new Error(errorData.message || `HTTP ${response.status}: ${response.statusText}`);
        } catch (parseError) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
      }

      // Handle agent mode response (JSON, not streaming)
      if (mode === 'agent') {
        console.log('[ChatPanel] Agent mode - reading JSON response');
        const data = await response.json();
        console.log('[ChatPanel] Agent response data:', data);

        if (data.status === 'agent_spawned') {
          const projectDisplay = data.project_name || data.project_root || 'Unknown';
          const assistantMessage: Message = {
            role: 'assistant',
            content: `✅ Agent task spawned!\n\n**Project:** ${projectDisplay}\n**Role:** ${data.role || selectedRole}\n**Task ID:** ${data.task_id}\n**Beads ID:** ${data.beads_task_id || 'N/A'}\n\nThe ${data.role || selectedRole} agent is now working on this task in the project directory. You can track progress in the Kanban board or task logs.`,
          };
          setMessages(prev => [...prev, assistantMessage]);
          setIsStreaming(false);
          return;
        } else if (data.status === 'error') {
          throw new Error(data.message || 'Agent task failed');
        } else {
          throw new Error(`Unexpected agent response status: ${data.status}`);
        }
      }

      // Read SSE stream from response body (chat mode)
      console.log('[ChatPanel] Chat mode - reading SSE stream');
      const reader = response.body?.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      if (!reader) {
        console.error('[ChatPanel] No response body reader available');
        throw new Error('No response body');
      }

      console.log('[ChatPanel] Starting stream read loop');
      while (true) {
        const { done, value } = await reader.read();
        if (done) {
          console.log('[ChatPanel] Stream reading complete');
          break;
        }

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
              console.log('[ChatPanel] SSE event:', parsed);

              if (parsed.status === 'connected') {
                console.log('[ChatPanel] Chat stream connected');
              } else if (parsed.status === 'complete') {
                console.log('[ChatPanel] Stream complete');
                // Completion event - check this BEFORE parsed.text
                const assistantMessage: Message = {
                  role: 'assistant',
                  content: parsed.text || streamingMessage,
                };
                setMessages(prev => [...prev, assistantMessage]);
                setStreamingMessage('');
                setIsStreaming(false);
                // Set suggestion from backend if provided
                if (parsed.suggestion && input === '') {
                  setSuggestion(parsed.suggestion);
                }
                // Set context loaded info if provided
                if (parsed.context_loaded) {
                  setContextLoadedFile(parsed.context_loaded);
                }
                return;
              } else if (parsed.text) {
                // For delta events
                setStreamingMessage(prev => prev + parsed.text.replace(/\\n/g, '\n').replace(/\\"/g, '"'));
              } else if (parsed.error) {
                console.error('[ChatPanel] SSE error event:', parsed.error);
                throw new Error(parsed.error);
              }
            } catch (err) {
              console.error('[ChatPanel] Failed to parse SSE data:', err, 'Line:', line);
            }
          }
        }
      }

      console.log('[ChatPanel] Stream ended, setting isStreaming to false');
      setIsStreaming(false);
    } catch (err) {
      // Ignore abort errors (user cancelled with ESC)
      if (err instanceof Error && err.name === 'AbortError') {
        console.log('[ChatPanel] Request aborted by user');
        return;
      }

      console.error('[ChatPanel] Error in sendMessage:', err);
      console.error('[ChatPanel] Error stack:', err instanceof Error ? err.stack : 'No stack');
      setMessages(prev => [
        ...prev,
        {
          role: 'assistant',
          content: `Error: ${err instanceof Error ? err.message : 'Unknown error'}`,
        },
      ]);
      setIsStreaming(false);
    } finally {
      abortControllerRef.current = null;
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    } else if (e.key === 'ArrowRight' && input === '' && suggestion) {
      // Accept suggestion on right arrow when input is empty
      e.preventDefault();
      setInput(suggestion);
      setSuggestion('');
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

    // Clear suggestion when user starts typing
    if (value && suggestion) {
      setSuggestion('');
    }

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

    // Detect file paths (simple pattern: path/to/file.ext)
    // Exclude URLs (anything with :// or starting with //)
    const filePattern = /[a-zA-Z0-9_\-./]+\.[a-zA-Z]{2,6}/g;
    const matches = value.match(filePattern);
    if (matches) {
      setDetectedFiles(matches.filter(m =>
        m.includes('/') &&
        !m.includes('://') &&
        !m.startsWith('//')
      ));
    } else {
      setDetectedFiles([]);
    }
  };

  const removeFileMention = (filename: string) => {
    setMentionedFiles(mentionedFiles.filter(f => f !== filename));
  };

  // Generate follow-up suggestions based on response content

  const executeCommand = (command: typeof slashCommands[0]) => {
    command.action();
    setShowCommandMenu(false);
  };

  const filteredCommands = slashCommands.filter(cmd =>
    cmd.name.toLowerCase().includes(commandFilter.toLowerCase())
  );

  const handleFileSelect = async (files: FileList | null) => {
    if (!files || files.length === 0) return;

    const newFiles: Array<{ name: string; content: string; size: number }> = [];

    for (let i = 0; i < files.length; i++) {
      const file = files[i];

      // Only handle text files (< 1MB)
      if (file.size > 1024 * 1024) {
        showAlert('File Too Large', `File ${file.name} is too large (max 1MB)`);
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
        showAlert('File Read Error', `Failed to read file ${file.name}`);
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
    if (!projectRoot) {
      showAlert('Project Required', 'Please select a project first');
      return;
    }

    createNewChatInProject(projectRoot);
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

  const compressOldMessages = async () => {
    if (messages.length < 10) {
      showAlert('Cannot Compress', 'Not enough messages to compress. Need at least 10 messages.');
      return;
    }

    // Keep the last 5 messages and compress the rest
    const messagesToKeep = messages.slice(-5);
    const messagesToCompress = messages.slice(0, -5);

    // Create a simple summary
    const summary = `[Previous conversation compressed: ${messagesToCompress.length} messages removed to save context. Token estimate: ~${messagesToCompress.reduce((sum, msg) => sum + estimateTokens(msg.content), 0).toLocaleString()} tokens saved]`;

    const summaryMessage: Message = {
      role: 'assistant',
      content: summary,
      id: Date.now(),
    };

    setMessages([summaryMessage, ...messagesToKeep]);

    // Save to localStorage
    localStorage.setItem(`ai-pack-chat-${chatId}`, JSON.stringify([summaryMessage, ...messagesToKeep]));

    showAlert('Compression Complete', `Compressed ${messagesToCompress.length} old messages!`);
  };

  return (
    <div className="w-full h-full bg-gray-800 flex flex-col">
      {/* Header */}
      <div className="p-3 border-b border-gray-700 bg-gray-900">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <div
              className={`w-2 h-2 rounded-full ${orchestratorMonitoring ? 'bg-green-400 animate-pulse' : 'bg-gray-500'}`}
              title={orchestratorMonitoring ? 'Orchestrator monitoring active' : 'Orchestrator idle'}
            ></div>
            <h3 className="font-semibold text-sm">
              {projectRoot ? (
                <>
                  📁 {projectRoot.split('/').pop()} › 🤖 {currentChatName}
                </>
              ) : (
                '🤖 Orchestrator'
              )}
            </h3>
            {orchestratorMonitoring && (
              <span className="text-xs text-green-400 ml-2">monitoring</span>
            )}
          </div>
          <div className="flex gap-1">
            <button
              onClick={compressOldMessages}
              disabled={messages.length < 10}
              className={`w-8 h-6 text-xs rounded flex items-center justify-center ${
                getTotalTokens() > 150000
                  ? 'bg-yellow-700 hover:bg-yellow-600'
                  : 'bg-gray-700 hover:bg-gray-600'
              } disabled:bg-gray-800 disabled:cursor-not-allowed text-gray-300`}
              title="Compress"
            >
              🗜️
            </button>
            <button
              onClick={startNewChat}
              disabled={isStreaming}
              className="w-8 h-6 text-xs bg-gray-700 hover:bg-gray-600 disabled:bg-gray-800 disabled:cursor-not-allowed text-gray-300 rounded flex items-center justify-center"
              title="New chat"
            >
              +
            </button>
          </div>
        </div>

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
                    🗑️
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Project Context Toggle */}
        {projectRoot && (
          <div className="mt-2 flex items-center gap-2">
            <label className="flex items-center gap-2 text-xs text-gray-400 hover:text-gray-300 cursor-pointer">
              <input
                type="checkbox"
                checked={useProjectContext}
                onChange={(e) => {
                  const newValue = e.target.checked;
                  setUseProjectContext(newValue);
                  localStorage.setItem('ai-pack-use-project-context', String(newValue));
                }}
                className="w-3 h-3 rounded"
              />
              <span>Load project context</span>
            </label>
            {contextLoadedFile && (
              <span className="text-xs text-green-400" title={`Loaded ${contextLoadedFile}`}>
                💡 {contextLoadedFile}
              </span>
            )}
          </div>
        )}

        {/* Chat List Selector */}
        {projectRoot && projectChats.length > 0 && (
          <div className="relative mt-2">
            <div className="flex gap-1">
              <div className="flex-1 px-3 py-1.5 bg-gray-700 text-white text-xs rounded flex items-center">
                <span className="truncate">💬 {currentChatName}</span>
              </div>
              <button
                onClick={() => setShowChatList(!showChatList)}
                className="px-2 py-1.5 bg-gray-700 hover:bg-gray-600 text-white text-xs rounded"
                title="Show saved chats"
              >
                ▼
              </button>
            </div>

            {showChatList && (
              <div className="absolute top-full left-0 right-0 mt-1 bg-gray-700 rounded shadow-lg border border-gray-600 z-10 max-h-60 overflow-y-auto">
                {projectChats.map((chat) => (
                  <div
                    key={chat.id}
                    className={`flex items-center justify-between px-3 py-2 hover:bg-gray-600 group ${
                      chat.id === chatId ? 'bg-gray-600' : ''
                    }`}
                  >
                    <button
                      onClick={() => {
                        switchToChat(chat);
                        setCurrentChatForProject(projectRoot, chat.id);
                        setShowChatList(false);
                      }}
                      className="flex-1 text-left text-xs text-white truncate"
                      title={chat.name}
                    >
                      {chat.id === chatId && '✓ '}
                      {chat.name}
                      <span className="text-gray-400 ml-2">
                        ({chat.messages.length} msgs)
                      </span>
                    </button>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        showConfirm(
                          'Delete Chat',
                          `Delete chat "${chat.name}"?`,
                          () => deleteChat(chat.id)
                        );
                      }}
                      className="ml-2 px-1.5 py-0.5 text-xs text-red-400 hover:text-red-300 opacity-0 group-hover:opacity-100 transition-opacity"
                      title="Delete chat"
                    >
                      🗑️
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.length === 0 && (
          <div className="text-center text-gray-500 mt-8">
            <p className="text-lg mb-2">🤖 Orchestrator Active</p>
            <p className="text-sm mb-1">I monitor your agent system and coordinate task execution.</p>
            <p className="text-xs text-gray-600">Ask me to spawn agents, check task status, or discuss the project.</p>
          </div>
        )}

        {messages.map((msg, idx) => (
          <div
            key={idx}
            data-msg-id={idx}
            className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
          >
            <div
              className={`max-w-[85%] rounded-lg px-4 py-2 relative group ${
                msg.role === 'user'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-700 text-gray-100'
              }`}
            >
              {/* Reply-to indicator */}
              {msg.replyTo !== undefined && messages[msg.replyTo] && (
                <div className="mb-2 pb-2 border-b border-gray-600 text-xs opacity-70">
                  <div className="flex items-center gap-1">
                    <span>↩️ Replying to:</span>
                    <button
                      onClick={() => {
                        const replyEl = document.querySelector(`[data-msg-id="${msg.replyTo}"]`);
                        replyEl?.scrollIntoView({ behavior: 'smooth', block: 'center' });
                      }}
                      className="hover:underline truncate max-w-xs"
                    >
                      {messages[msg.replyTo].content.substring(0, 50)}...
                    </button>
                  </div>
                </div>
              )}
              {msg.role === 'assistant' ? (
                <>
                  {/* Reply button */}
                  <button
                    onClick={() => {
                      setReplyingTo(idx);
                      document.querySelector('textarea')?.focus();
                    }}
                    className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity bg-gray-800 hover:bg-gray-600 text-white text-xs w-6 h-6 flex items-center justify-center rounded"
                    title="Reply to this message"
                  >
                    ↩️
                  </button>
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
                </>
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
        {replyingTo !== null && messages[replyingTo] && (
          <div className="mb-2 p-2 bg-blue-900/30 border-l-4 border-blue-500 rounded text-sm">
            <div className="flex items-start justify-between gap-2">
              <div className="flex-1">
                <div className="text-blue-300 text-xs mb-1">↩️ Replying to:</div>
                <div className="text-gray-300 text-xs truncate">
                  {messages[replyingTo].content.substring(0, 100)}...
                </div>
              </div>
              <button
                onClick={() => setReplyingTo(null)}
                className="text-gray-400 hover:text-white text-xs"
                title="Cancel reply"
              >
                ✕
              </button>
            </div>
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
                  🗑️
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
                  🗑️
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
                  🗑️
                </button>
                <div className="text-xs text-gray-400 mt-1 text-center">{Math.round(img.size / 1024)}KB</div>
              </div>
            ))}
          </div>
        )}
        <div className="flex justify-end items-center mb-2">
          <button
            onClick={() => fileInputRef.current?.click()}
            className="text-xs px-2 py-1 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors"
            title="Attach files (or drag & drop)"
          >
            📎 Attach Files
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
        {detectedFiles.length > 0 && (
          <div className="mb-2 p-2 bg-blue-900/30 border border-blue-600 rounded text-xs text-blue-300">
            💡 Detected file paths: {detectedFiles.join(', ')} - Consider attaching these files using the 📎 button
          </div>
        )}
        <div className="flex gap-2 items-stretch">
          <div className="flex-1 relative flex">
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
            {/* Suggestion overlay */}
            {!input && suggestion && (
              <div className="absolute top-0 left-0 right-0 px-3 py-2 text-sm text-gray-500 pointer-events-none whitespace-pre-wrap">
                {suggestion}
              </div>
            )}
            <textarea
              value={input}
              onChange={(e) => handleInputChange(e.target.value)}
              onKeyDown={handleKeyPress}
              placeholder={suggestion ? '' : 'Ask me anything... (type / for commands)'}
              disabled={isStreaming}
              className="w-full h-full bg-gray-800 text-white rounded-lg px-3 py-2 text-sm resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 relative z-10"
              rows={2}
              style={{ background: 'transparent' }}
            />
            <div className="absolute inset-0 bg-gray-800 rounded-lg -z-10" />
          </div>
          <button
            onClick={sendMessage}
            disabled={!input.trim() || isStreaming}
            className="px-4 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white rounded-lg transition-colors text-sm font-medium"
          >
            {isStreaming ? '...' : 'Send'}
          </button>
        </div>
        <div className="mt-2">
          {/* Context Window Progress Bar */}
          <div className="mb-2">
            <div className="flex items-center justify-between text-xs mb-1">
              <span className="text-gray-500">Context Window</span>
              <span className={
                getTotalTokens() > 180000 ? 'text-red-400 font-semibold' :
                getTotalTokens() > 150000 ? 'text-yellow-400 font-semibold' :
                'text-gray-400'
              }>
                {getTotalTokens().toLocaleString()} / 200,000 tokens ({Math.round((getTotalTokens() / 200000) * 100)}%)
              </span>
            </div>
            <div className="w-full bg-gray-700 rounded-full h-2 overflow-hidden">
              <div
                className={`h-full transition-all duration-300 ${
                  getTotalTokens() > 180000 ? 'bg-red-500' :
                  getTotalTokens() > 150000 ? 'bg-yellow-500' :
                  'bg-green-500'
                }`}
                style={{ width: `${Math.min((getTotalTokens() / 200000) * 100, 100)}%` }}
              />
            </div>
          </div>
          <p className="text-xs text-gray-500">
            {isStreaming ? (
              <span className="text-yellow-400 font-semibold">Press ESC to cancel</span>
            ) : (
              'Press Enter to send, Shift+Enter for new line, ↑↓ for history'
            )}
          </p>
        </div>
      </div>

      {/* Custom Modal for Alerts and Confirms */}
      {modal.show && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-lg shadow-xl border border-gray-600 p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-white mb-3">{modal.title}</h3>
            <p className="text-gray-300 mb-6 whitespace-pre-wrap">{modal.message}</p>
            <div className="flex justify-end gap-3">
              {modal.type === 'confirm' && (
                <button
                  onClick={() => setModal({ ...modal, show: false })}
                  className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded transition-colors"
                >
                  Cancel
                </button>
              )}
              <button
                onClick={() => {
                  if (modal.type === 'confirm' && modal.onConfirm) {
                    modal.onConfirm();
                  }
                  setModal({ ...modal, show: false });
                }}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded transition-colors"
              >
                OK
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
