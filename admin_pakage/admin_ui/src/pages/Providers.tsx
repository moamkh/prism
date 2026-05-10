import { useEffect, useState } from 'react';
import { providersApi } from '../api/client';
import type { Provider } from '../types';

export default function Providers() {
  const [providers, setProviders] = useState<Provider[]>([]);
  const [form, setForm] = useState({ name: '', base_url: '', api_token: '', http_proxy: '', socks5_proxy: '', enable_proxy: true, max_concurrent_requests: 100, is_active: true });
  const [editing, setEditing] = useState<string | null>(null);

  const load = () => providersApi.list().then((res) => setProviders(res.data));

  useEffect(() => {
    load();
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (editing) {
      await providersApi.update(editing, form);
      setEditing(null);
    } else {
      await providersApi.create(form);
    }
    setForm({ name: '', base_url: '', api_token: '', http_proxy: '', socks5_proxy: '', enable_proxy: true, max_concurrent_requests: 100, is_active: true });
    load();
  };

  const startEdit = (p: Provider) => {
    setEditing(p.id);
    setForm({ name: p.name, base_url: p.base_url, api_token: p.api_token, http_proxy: p.http_proxy || '', socks5_proxy: p.socks5_proxy || '', enable_proxy: p.enable_proxy, max_concurrent_requests: p.max_concurrent_requests, is_active: p.is_active });
  };

  const remove = async (id: string) => {
    await providersApi.remove(id);
    load();
  };

  return (
    <div>
      <h2 style={{ marginBottom: 20 }}>Providers</h2>
      <form onSubmit={handleSubmit} style={{ background: '#fff', padding: 20, borderRadius: 8, marginBottom: 24 }}>
        <div style={{ display: 'grid', gap: 12, gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))' }}>
          <input placeholder="Name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required style={inputStyle} />
          <input placeholder="Base URL" value={form.base_url} onChange={(e) => setForm({ ...form, base_url: e.target.value })} required style={inputStyle} />
          <input placeholder="API Token" value={form.api_token} onChange={(e) => setForm({ ...form, api_token: e.target.value })} required style={inputStyle} />
          <input placeholder="HTTP Proxy (optional)" value={form.http_proxy} onChange={(e) => setForm({ ...form, http_proxy: e.target.value })} style={inputStyle} />
          <input placeholder="SOCKS5 Proxy (optional)" value={form.socks5_proxy} onChange={(e) => setForm({ ...form, socks5_proxy: e.target.value })} style={inputStyle} />
          <input placeholder="Max Concurrent Requests" type="number" value={form.max_concurrent_requests} onChange={(e) => setForm({ ...form, max_concurrent_requests: parseInt(e.target.value) || 100 })} style={inputStyle} />
          <label style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <input type="checkbox" checked={form.enable_proxy} onChange={(e) => setForm({ ...form, enable_proxy: e.target.checked })} />
            Enable Proxy
          </label>
          <label style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <input type="checkbox" checked={form.is_active} onChange={(e) => setForm({ ...form, is_active: e.target.checked })} />
            Active
          </label>
        </div>
        <div style={{ marginTop: 12 }}>
          <button type="submit" style={btnPrimary}>{editing ? 'Update' : 'Add'} Provider</button>
          {editing && <button type="button" onClick={() => { setEditing(null); setForm({ name: '', base_url: '', api_token: '', http_proxy: '', socks5_proxy: '', enable_proxy: true, max_concurrent_requests: 100, is_active: true }); }} style={btnSecondary}>Cancel</button>}
        </div>
      </form>

      <table style={{ width: '100%', borderCollapse: 'collapse', background: '#fff', borderRadius: 8, overflow: 'hidden' }}>
        <thead>
          <tr style={{ background: '#eee' }}>
            <th style={thStyle}>Name</th>
            <th style={thStyle}>Base URL</th>
            <th style={thStyle}>HTTP Proxy</th>
            <th style={thStyle}>SOCKS5 Proxy</th>
            <th style={thStyle}>Max Concurrent</th>
            <th style={thStyle}>Proxy</th>
            <th style={thStyle}>Active</th>
            <th style={thStyle}>Actions</th>
          </tr>
        </thead>
        <tbody>
          {providers.map((p) => (
            <tr key={p.id} style={{ borderBottom: '1px solid #eee' }}>
              <td style={tdStyle}>{p.name}</td>
              <td style={tdStyle}>{p.base_url}</td>
              <td style={tdStyle}>{p.http_proxy || '-'}</td>
              <td style={tdStyle}>{p.socks5_proxy || '-'}</td>
              <td style={tdStyle}>{p.max_concurrent_requests}</td>
              <td style={tdStyle}>{p.enable_proxy ? 'Yes' : 'No'}</td>
              <td style={tdStyle}>{p.is_active ? 'Yes' : 'No'}</td>
              <td style={tdStyle}>
                <button onClick={() => startEdit(p)} style={btnSmall}>Edit</button>
                <button onClick={() => remove(p.id)} style={btnSmallDanger}>Delete</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

const inputStyle: React.CSSProperties = { padding: '8px 12px', border: '1px solid #ccc', borderRadius: 4 };
const btnPrimary: React.CSSProperties = { padding: '8px 16px', background: '#1a1a2e', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer' };
const btnSecondary: React.CSSProperties = { padding: '8px 16px', background: '#ccc', border: 'none', borderRadius: 4, cursor: 'pointer', marginLeft: 8 };
const btnSmall: React.CSSProperties = { padding: '4px 10px', background: '#1a1a2e', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer', marginRight: 4 };
const btnSmallDanger: React.CSSProperties = { padding: '4px 10px', background: '#c0392b', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer' };
const thStyle: React.CSSProperties = { padding: '12px 16px', textAlign: 'left', fontWeight: 600 };
const tdStyle: React.CSSProperties = { padding: '12px 16px' };
