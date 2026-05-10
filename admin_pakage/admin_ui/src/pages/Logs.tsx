import { useEffect, useState, useRef } from 'react';

const PROXY_BASE = (import.meta as any).env?.VITE_PROXY_BASE_URL || 'http://localhost:8080';

export default function Logs() {
  const [lines, setLines] = useState<string[]>([]);
  const [filter, setFilter] = useState('');
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [loading, setLoading] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const load = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${PROXY_BASE}/api/logs`);
      const data = await res.json();
      setLines(data.lines || []);
    } catch (e) {
      console.error('Failed to load proxy logs', e);
    } finally {
      setLoading(false);
    }
  };

  const clear = async () => {
    try {
      await fetch(`${PROXY_BASE}/api/logs`, { method: 'DELETE' });
      setLines([]);
    } catch (e) {
      console.error('Failed to clear logs', e);
    }
  };

  useEffect(() => {
    load();
    if (!autoRefresh) return;
    const id = setInterval(load, 3000);
    return () => clearInterval(id);
  }, [autoRefresh]);

  useEffect(() => {
    if (containerRef.current && !filter) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [lines, filter]);

  const filtered = filter
    ? lines.filter((l) => l.toLowerCase().includes(filter.toLowerCase()))
    : lines;

  const getLevel = (line: string) => {
    const lower = line.toLowerCase();
    if (lower.includes('error') || lower.includes('fatal')) return 'error';
    if (lower.includes('warn')) return 'warn';
    if (lower.includes('info') || lower.includes('starting')) return 'info';
    return '';
  };

  const levelColor: Record<string, string> = {
    error: '#e74c3c',
    warn: '#f39c12',
    info: '#3498db',
  };

  return (
    <div>
      <h2 style={{ marginBottom: 16 }}>Proxy Logs</h2>
      <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 16, flexWrap: 'wrap' }}>
        <button onClick={load} style={btnPrimary} disabled={loading}>
          {loading ? 'Refreshing...' : 'Refresh'}
        </button>
        <button onClick={clear} style={btnDanger}>Clear</button>
        <label style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 14 }}>
          <input type="checkbox" checked={autoRefresh} onChange={(e) => setAutoRefresh(e.target.checked)} />
          Auto-refresh (3s)
        </label>
        <input
          type="text"
          placeholder="Filter logs..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          style={{ ...inputStyle, minWidth: 200 }}
        />
        <span style={{ fontSize: 13, color: '#666' }}>{filtered.length} lines</span>
      </div>

      <div
        ref={containerRef}
        style={{
          background: '#0f0f1a',
          color: '#eee',
          borderRadius: 8,
          padding: 16,
          fontFamily: 'monospace',
          fontSize: 13,
          lineHeight: 1.5,
          maxHeight: '70vh',
          overflowY: 'auto',
        }}
      >
        {filtered.length === 0 ? (
          <div style={{ color: '#666', textAlign: 'center', padding: 40 }}>No logs</div>
        ) : (
          filtered.map((line, i) => {
            const level = getLevel(line);
            return (
              <div
                key={i}
                style={{
                  padding: '2px 0',
                  borderBottom: '1px solid #1a1a2e',
                  color: levelColor[level] || '#eee',
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                }}
              >
                {line}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}

const btnPrimary: React.CSSProperties = {
  padding: '8px 16px',
  background: '#1a1a2e',
  color: '#fff',
  border: 'none',
  borderRadius: 4,
  cursor: 'pointer',
};

const btnDanger: React.CSSProperties = {
  padding: '8px 16px',
  background: '#c0392b',
  color: '#fff',
  border: 'none',
  borderRadius: 4,
  cursor: 'pointer',
};

const inputStyle: React.CSSProperties = {
  padding: '8px 12px',
  border: '1px solid #ccc',
  borderRadius: 4,
};
