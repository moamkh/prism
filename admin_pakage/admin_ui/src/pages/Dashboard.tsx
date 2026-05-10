import { useEffect, useState } from 'react';
import { dashboardApi } from '../api/client';
import type { DashboardStats } from '../types';

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    dashboardApi.stats().then((res) => {
      setStats(res.data);
      setLoading(false);
    });
  }, []);

  if (loading || !stats) return <div>Loading...</div>;

  return (
    <div>
      <h2 style={{ marginBottom: 20 }}>Dashboard</h2>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: 16, marginBottom: 24 }}>
        <Card title="Total Requests" value={stats.total_requests} />
        <Card title="Input Tokens" value={stats.total_input_tokens} />
        <Card title="Output Tokens" value={stats.total_output_tokens} />
        <Card title="Active Providers" value={stats.active_providers} />
        <Card title="Active Tokens" value={stats.active_tokens} />
      </div>

      <h3 style={{ marginBottom: 12 }}>Top Models (7 days)</h3>
      <table style={{ width: '100%', borderCollapse: 'collapse', background: '#fff', borderRadius: 8, overflow: 'hidden' }}>
        <thead>
          <tr style={{ background: '#eee' }}>
            <th style={thStyle}>Model</th>
            <th style={thStyle}>Provider</th>
            <th style={thStyle}>Requests</th>
            <th style={thStyle}>Input Tokens</th>
            <th style={thStyle}>Output Tokens</th>
            <th style={thStyle}>Total Tokens</th>
          </tr>
        </thead>
        <tbody>
          {stats.top_models.map((m, idx) => (
            <tr key={`${m.model_id || 'unknown'}-${idx}`} style={{ borderBottom: '1px solid #eee' }}>
              <td style={tdStyle}>{m.model_name || m.model_id || 'Unknown'}</td>
              <td style={tdStyle}>{m.provider_name || '—'}</td>
              <td style={tdStyle}>{m.total_requests}</td>
              <td style={tdStyle}>{m.total_input_tokens}</td>
              <td style={tdStyle}>{m.total_output_tokens}</td>
              <td style={tdStyle}>{m.total_tokens}</td>
            </tr>
          ))}
        </tbody>
      </table>

      <h3 style={{ margin: '24px 0 12px' }}>Recent Logs</h3>
      <table style={{ width: '100%', borderCollapse: 'collapse', background: '#fff', borderRadius: 8, overflow: 'hidden' }}>
        <thead>
          <tr style={{ background: '#eee' }}>
            <th style={thStyle}>Time</th>
            <th style={thStyle}>Path</th>
            <th style={thStyle}>Input</th>
            <th style={thStyle}>Output</th>
            <th style={thStyle}>Latency</th>
            <th style={thStyle}>Status</th>
          </tr>
        </thead>
        <tbody>
          {stats.recent_logs.map((log) => (
            <tr key={log.id} style={{ borderBottom: '1px solid #eee' }}>
              <td style={tdStyle}>{new Date(log.created_at).toLocaleString()}</td>
              <td style={tdStyle}>{log.request_path}</td>
              <td style={tdStyle}>{log.input_tokens}</td>
              <td style={tdStyle}>{log.output_tokens}</td>
              <td style={tdStyle}>{log.latency_ms}ms</td>
              <td style={tdStyle}>{log.status_code}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function Card({ title, value }: { title: string; value: number }) {
  return (
    <div style={{ background: '#fff', padding: 20, borderRadius: 8, boxShadow: '0 1px 3px rgba(0,0,0,0.1)' }}>
      <div style={{ fontSize: 12, color: '#666', textTransform: 'uppercase', letterSpacing: 1 }}>{title}</div>
      <div style={{ fontSize: 28, fontWeight: 700, marginTop: 8 }}>{value.toLocaleString()}</div>
    </div>
  );
}

const thStyle: React.CSSProperties = { padding: '12px 16px', textAlign: 'left', fontWeight: 600 };
const tdStyle: React.CSSProperties = { padding: '12px 16px' };
