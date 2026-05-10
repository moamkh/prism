import { useEffect, useState } from 'react';
import { usageApi } from '../api/client';
import type { UsageLog } from '../types';

export default function UsageLogs() {
  const [logs, setLogs] = useState<UsageLog[]>([]);
  const [page, setPage] = useState(0);
  const limit = 50;

  const load = () => {
    usageApi.logs({ skip: page * limit, limit }).then((res) => setLogs(res.data));
  };

  useEffect(() => {
    load();
  }, [page]);

  return (
    <div>
      <h2 style={{ marginBottom: 20 }}>Usage Logs</h2>
      <table style={{ width: '100%', borderCollapse: 'collapse', background: '#fff', borderRadius: 8, overflow: 'hidden' }}>
        <thead>
          <tr style={{ background: '#eee' }}>
            <th style={thStyle}>Time</th>
            <th style={thStyle}>Path</th>
            <th style={thStyle}>Input</th>
            <th style={thStyle}>Output</th>
            <th style={thStyle}>Total</th>
            <th style={thStyle}>Latency</th>
            <th style={thStyle}>Status</th>
          </tr>
        </thead>
        <tbody>
          {logs.map((log) => (
            <tr key={log.id} style={{ borderBottom: '1px solid #eee' }}>
              <td style={tdStyle}>{new Date(log.created_at).toLocaleString()}</td>
              <td style={tdStyle}>{log.request_path}</td>
              <td style={tdStyle}>{log.input_tokens}</td>
              <td style={tdStyle}>{log.output_tokens}</td>
              <td style={tdStyle}>{log.total_tokens}</td>
              <td style={tdStyle}>{log.latency_ms}ms</td>
              <td style={tdStyle}>{log.status_code}</td>
            </tr>
          ))}
        </tbody>
      </table>
      <div style={{ marginTop: 16 }}>
        <button onClick={() => setPage((p) => Math.max(0, p - 1))} disabled={page === 0} style={btnPrimary}>Previous</button>
        <span style={{ margin: '0 12px' }}>Page {page + 1}</span>
        <button onClick={() => setPage((p) => p + 1)} disabled={logs.length < limit} style={btnPrimary}>Next</button>
      </div>
    </div>
  );
}

const btnPrimary: React.CSSProperties = { padding: '8px 16px', background: '#1a1a2e', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer' };
const thStyle: React.CSSProperties = { padding: '12px 16px', textAlign: 'left', fontWeight: 600 };
const tdStyle: React.CSSProperties = { padding: '12px 16px' };
