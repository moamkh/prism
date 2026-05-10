import { useEffect, useRef, useState } from 'react';

const PROXY_BASE = (import.meta as any).env?.VITE_PROXY_BASE_URL || 'http://localhost:8080';
const STORAGE_KEY = 'playground_token';

interface Message {
  role: 'system' | 'user' | 'assistant';
  content: string;
}

interface ModelItem {
  id: string;
  object: string;
  created: number;
  owned_by: string;
}

export default function Playground() {
  const [token, setToken] = useState(() => localStorage.getItem(STORAGE_KEY) || '');
  const [models, setModels] = useState<ModelItem[]>([]);
  const [selectedModel, setSelectedModel] = useState('');
  const [systemMessage, setSystemMessage] = useState('You are a helpful assistant.');
  const [temperature, setTemperature] = useState(0.7);
  const [maxTokens, setMaxTokens] = useState(1024);
  const [topP, setTopP] = useState(1);
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [streamEnabled, setStreamEnabled] = useState(true);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, token);
  }, [token]);

  useEffect(() => {
    if (bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, loading]);

  const fetchModels = async () => {
    setError('');
    try {
      const res = await fetch(`${PROXY_BASE}/v1/models`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error || `HTTP ${res.status}`);
      }
      const data = await res.json();
      const list: ModelItem[] = data.data || [];
      setModels(list);
      if (list.length > 0 && !selectedModel) {
        setSelectedModel(list[0].id);
      }
    } catch (e: any) {
      setError(e.message || 'Failed to load models');
    }
  };

  const sendMessage = async () => {
    if (!input.trim() || !selectedModel) return;
    setError('');

    const userMsg: Message = { role: 'user', content: input.trim() };
    const newMessages = [...messages, userMsg];
    if (systemMessage.trim() && !messages.some((m) => m.role === 'system')) {
      newMessages.unshift({ role: 'system', content: systemMessage.trim() });
    }
    setMessages(newMessages);
    setInput('');
    setLoading(true);

    try {
      const res = await fetch(`${PROXY_BASE}/v1/chat/completions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          model: selectedModel,
          messages: newMessages,
          temperature,
          max_tokens: maxTokens,
          top_p: topP,
          stream: streamEnabled,
        }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error || `HTTP ${res.status}`);
      }

      if (streamEnabled) {
        const reader = res.body?.getReader();
        const decoder = new TextDecoder();
        let assistantContent = '';
        setMessages((prev) => [...prev, { role: 'assistant', content: '' }]);

        if (reader) {
          while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            const chunk = decoder.decode(value, { stream: true });
            const lines = chunk.split('\n');
            for (const line of lines) {
              const trimmed = line.trim();
              if (!trimmed.startsWith('data:')) continue;
              const jsonStr = trimmed.slice(5).trim();
              if (jsonStr === '[DONE]') continue;
              try {
                const parsed = JSON.parse(jsonStr);
                const delta = parsed.choices?.[0]?.delta?.content;
                if (delta) {
                  assistantContent += delta;
                  setMessages((prev) => {
                    const copy = [...prev];
                    const last = copy[copy.length - 1];
                    if (last && last.role === 'assistant') {
                      last.content = assistantContent;
                    }
                    return copy;
                  });
                }
              } catch {
                // ignore malformed SSE lines
              }
            }
          }
        }
      } else {
        const data = await res.json();
        const content = data.choices?.[0]?.message?.content || '';
        setMessages((prev) => [...prev, { role: 'assistant', content: content }]);
      }
    } catch (e: any) {
      setError(e.message || 'Request failed');
    } finally {
      setLoading(false);
    }
  };

  const clearChat = () => setMessages([]);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: 'calc(100vh - 48px)' }}>
      <h2 style={{ marginBottom: 12 }}>Playground</h2>

      {/* Configuration bar */}
      <div style={{ background: '#fff', padding: 16, borderRadius: 8, marginBottom: 12, display: 'flex', flexWrap: 'wrap', gap: 12, alignItems: 'flex-end' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 4, minWidth: 260, flex: 1 }}>
          <label style={{ fontSize: 12, fontWeight: 600, color: '#555' }}>Token</label>
          <input
            type="password"
            placeholder="Paste your token here (rpm_...)"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            style={{ ...inputStyle, width: '100%' }}
          />
        </div>
        <button onClick={fetchModels} disabled={!token} style={{ ...btnPrimary, opacity: token ? 1 : 0.5 }}>
          Load Models
        </button>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 4, minWidth: 200 }}>
          <label style={{ fontSize: 12, fontWeight: 600, color: '#555' }}>Model</label>
          <select value={selectedModel} onChange={(e) => setSelectedModel(e.target.value)} style={inputStyle}>
            {models.length === 0 && <option value="">No models loaded</option>}
            {models.map((m) => (
              <option key={m.id} value={m.id}>{m.id}</option>
            ))}
          </select>
        </div>
        <label style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 13 }}>
          <input type="checkbox" checked={streamEnabled} onChange={(e) => setStreamEnabled(e.target.checked)} />
          Stream
        </label>
      </div>

      {/* Parameters */}
      <div style={{ background: '#fff', padding: '12px 16px', borderRadius: 8, marginBottom: 12, display: 'flex', flexWrap: 'wrap', gap: 16, alignItems: 'center' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
          <label style={{ fontSize: 12, fontWeight: 600, color: '#555' }}>System</label>
          <input value={systemMessage} onChange={(e) => setSystemMessage(e.target.value)} style={{ ...inputStyle, width: 280 }} />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
          <label style={{ fontSize: 12, fontWeight: 600, color: '#555' }}>Temperature ({temperature})</label>
          <input type="range" min={0} max={2} step={0.1} value={temperature} onChange={(e) => setTemperature(parseFloat(e.target.value))} style={{ width: 140 }} />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
          <label style={{ fontSize: 12, fontWeight: 600, color: '#555' }}>Max Tokens</label>
          <input type="number" min={1} max={8192} value={maxTokens} onChange={(e) => setMaxTokens(parseInt(e.target.value) || 0)} style={{ ...inputStyle, width: 100 }} />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
          <label style={{ fontSize: 12, fontWeight: 600, color: '#555' }}>Top P ({topP})</label>
          <input type="range" min={0} max={1} step={0.05} value={topP} onChange={(e) => setTopP(parseFloat(e.target.value))} style={{ width: 120 }} />
        </div>
        <button onClick={clearChat} style={{ ...btnPrimary, background: '#666', marginLeft: 'auto' }}>Clear Chat</button>
      </div>

      {error && (
        <div style={{ background: '#f8d7da', color: '#721c24', padding: 10, borderRadius: 6, marginBottom: 12, fontSize: 14 }}>
          {error}
        </div>
      )}

      {/* Chat area */}
      <div style={{ flex: 1, background: '#f5f5f5', borderRadius: 8, padding: 16, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: 12, marginBottom: 12 }}>
        {messages.length === 0 && (
          <div style={{ color: '#888', textAlign: 'center', marginTop: 40 }}>Start a conversation…</div>
        )}
        {messages.map((msg, i) => (
          <div
            key={i}
            style={{
              alignSelf: msg.role === 'user' ? 'flex-end' : 'flex-start',
              background: msg.role === 'user' ? '#1a1a2e' : '#fff',
              color: msg.role === 'user' ? '#fff' : '#333',
              padding: '10px 14px',
              borderRadius: 12,
              maxWidth: '80%',
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-word',
              fontSize: 14,
              lineHeight: 1.5,
              boxShadow: msg.role === 'assistant' ? '0 1px 3px rgba(0,0,0,0.08)' : 'none',
            }}
          >
            {msg.role === 'system' && (
              <div style={{ fontSize: 10, textTransform: 'uppercase', letterSpacing: 1, opacity: 0.6, marginBottom: 4 }}>System</div>
            )}
            {msg.content}
          </div>
        ))}
        {loading && streamEnabled && messages[messages.length - 1]?.role === 'assistant' ? null : loading ? (
          <div style={{ alignSelf: 'flex-start', color: '#888', fontSize: 13 }}>Thinking…</div>
        ) : null}
        <div ref={bottomRef} />
      </div>

      {/* Input area */}
      <div style={{ display: 'flex', gap: 8 }}>
        <textarea
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault();
              sendMessage();
            }
          }}
          placeholder="Type a message… (Shift+Enter for new line)"
          rows={2}
          style={{ flex: 1, ...inputStyle, resize: 'vertical' }}
        />
        <button onClick={sendMessage} disabled={loading || !input.trim() || !selectedModel} style={{ ...btnPrimary, alignSelf: 'flex-end', opacity: loading || !input.trim() || !selectedModel ? 0.5 : 1 }}>
          Send
        </button>
      </div>
    </div>
  );
}

const inputStyle: React.CSSProperties = {
  padding: '8px 12px',
  border: '1px solid #ccc',
  borderRadius: 4,
  fontFamily: 'inherit',
  fontSize: 14,
};

const btnPrimary: React.CSSProperties = {
  padding: '8px 16px',
  background: '#1a1a2e',
  color: '#fff',
  border: 'none',
  borderRadius: 4,
  cursor: 'pointer',
  fontSize: 14,
};
