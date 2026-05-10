import { useEffect, useState } from 'react';
import { tokensApi, modelsApi } from '../api/client';
import type { Token, Model, TokenModelUsage } from '../types';

interface ModelPermission {
  model_id: string;
  max_input_tokens: string;
  max_output_tokens: string;
}

export default function Tokens() {
  const [tokens, setTokens] = useState<Token[]>([]);
  const [models, setModels] = useState<Model[]>([]);
  const [form, setForm] = useState({
    name: '',
    max_input_tokens: '',
    max_output_tokens: '',
    requests_per_minute: '',
    is_active: true,
    model_permissions: [] as ModelPermission[],
  });
  const [plainKey, setPlainKey] = useState<string | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [dialogToken, setDialogToken] = useState<Token | null>(null);
  const [usageData, setUsageData] = useState<TokenModelUsage[]>([]);
  const [usageLoading, setUsageLoading] = useState(false);
  const [editToken, setEditToken] = useState<Token | null>(null);
  const [editForm, setEditForm] = useState({
    name: '',
    max_input_tokens: '',
    max_output_tokens: '',
    requests_per_minute: '',
    is_active: true,
    model_permissions: [] as ModelPermission[],
  });

  const load = () => {
    tokensApi.list().then((res) => setTokens(res.data));
    modelsApi.list().then((res) => setModels(res.data));
  };

  useEffect(() => {
    load();
  }, []);

  const toggleModel = (id: string, isEdit: boolean) => {
    const setter = isEdit ? setEditForm : setForm;
    setter((prev) => {
      const exists = prev.model_permissions.find((p) => p.model_id === id);
      if (exists) {
        return {
          ...prev,
          model_permissions: prev.model_permissions.filter((p) => p.model_id !== id),
        };
      }
      return {
        ...prev,
        model_permissions: [...prev.model_permissions, { model_id: id, max_input_tokens: '', max_output_tokens: '' }],
      };
    });
  };

  const updatePermField = (modelId: string, field: keyof ModelPermission, value: string, isEdit: boolean) => {
    const setter = isEdit ? setEditForm : setForm;
    setter((prev) => ({
      ...prev,
      model_permissions: prev.model_permissions.map((p) =>
        p.model_id === modelId ? { ...p, [field]: value } : p
      ),
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      name: form.name,
      max_input_tokens: form.max_input_tokens ? parseInt(form.max_input_tokens) : null,
      max_output_tokens: form.max_output_tokens ? parseInt(form.max_output_tokens) : null,
      requests_per_minute: form.requests_per_minute ? parseInt(form.requests_per_minute) : null,
      is_active: form.is_active,
      model_permissions: form.model_permissions.map((p) => ({
        model_id: p.model_id,
        max_input_tokens: p.max_input_tokens ? parseInt(p.max_input_tokens) : null,
        max_output_tokens: p.max_output_tokens ? parseInt(p.max_output_tokens) : null,
      })),
    };
    const res = await tokensApi.create(payload);
    setPlainKey(res.data.plain_key);
    setForm({ name: '', max_input_tokens: '', max_output_tokens: '', requests_per_minute: '', is_active: true, model_permissions: [] });
    load();
  };

  const remove = async (id: string) => {
    await tokensApi.remove(id);
    load();
  };

  const openEdit = (token: Token) => {
    setEditToken(token);
    setEditForm({
      name: token.name,
      max_input_tokens: token.max_input_tokens?.toString() || '',
      max_output_tokens: token.max_output_tokens?.toString() || '',
      requests_per_minute: token.requests_per_minute?.toString() || '',
      is_active: token.is_active,
      model_permissions: token.model_permissions?.map((p) => ({
        model_id: p.model_id,
        max_input_tokens: p.max_input_tokens?.toString() || '',
        max_output_tokens: p.max_output_tokens?.toString() || '',
      })) || [],
    });
  };

  const closeEdit = () => {
    setEditToken(null);
  };

  const handleEditSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editToken) return;
    const payload = {
      name: editForm.name,
      max_input_tokens: editForm.max_input_tokens ? parseInt(editForm.max_input_tokens) : null,
      max_output_tokens: editForm.max_output_tokens ? parseInt(editForm.max_output_tokens) : null,
      requests_per_minute: editForm.requests_per_minute ? parseInt(editForm.requests_per_minute) : null,
      is_active: editForm.is_active,
      model_permissions: editForm.model_permissions.map((p) => ({
        model_id: p.model_id,
        max_input_tokens: p.max_input_tokens ? parseInt(p.max_input_tokens) : null,
        max_output_tokens: p.max_output_tokens ? parseInt(p.max_output_tokens) : null,
      })),
    };
    await tokensApi.update(editToken.id, payload);
    setEditToken(null);
    load();
  };

  const handleRegenerate = async () => {
    if (!editToken) return;
    const res = await tokensApi.regenerate(editToken.id);
    setPlainKey(res.data.plain_key);
    load();
  };

  const openUsageDialog = async (token: Token) => {
    setDialogToken(token);
    setDialogOpen(true);
    setUsageLoading(true);
    try {
      const res = await tokensApi.usage(token.id);
      setUsageData(res.data);
    } catch {
      setUsageData([]);
    }
    setUsageLoading(false);
  };

  const closeDialog = () => {
    setDialogOpen(false);
    setDialogToken(null);
    setUsageData([]);
  };

  const renderModelTable = (
    permList: ModelPermission[],
    onToggle: (id: string) => void,
    onUpdate: (id: string, field: keyof ModelPermission, value: string) => void
  ) => (
    <table style={{ width: '100%', borderCollapse: 'collapse', background: '#f8f8f8', borderRadius: 8, overflow: 'hidden', marginTop: 8 }}>
      <thead>
        <tr style={{ background: '#eee' }}>
          <th style={thStyleSmall}>Model ID</th>
          <th style={thStyleSmall}>Display ID</th>
          <th style={thStyleSmall}>Allow</th>
          <th style={thStyleSmall}>Max Input Tokens</th>
          <th style={thStyleSmall}>Max Output Tokens</th>
        </tr>
      </thead>
      <tbody>
        {models.map((m) => {
          const perm = permList.find((p) => p.model_id === m.id);
          const checked = !!perm;
          return (
            <tr key={m.id} style={{ borderBottom: '1px solid #eee' }}>
              <td style={tdStyleSmall}><strong>{m.model_id}</strong></td>
              <td style={tdStyleSmall}>{m.display_model_id || '-'}</td>
              <td style={tdStyleSmall}>
                <input type="checkbox" checked={checked} onChange={() => onToggle(m.id)} />
              </td>
              <td style={tdStyleSmall}>
                <input
                  placeholder="Max input"
                  type="number"
                  value={perm?.max_input_tokens || ''}
                  onChange={(e) => onUpdate(m.id, 'max_input_tokens', e.target.value)}
                  style={{ ...inputStyle, width: 120 }}
                  disabled={!checked}
                />
              </td>
              <td style={tdStyleSmall}>
                <input
                  placeholder="Max output"
                  type="number"
                  value={perm?.max_output_tokens || ''}
                  onChange={(e) => onUpdate(m.id, 'max_output_tokens', e.target.value)}
                  style={{ ...inputStyle, width: 120 }}
                  disabled={!checked}
                />
              </td>
            </tr>
          );
        })}
      </tbody>
    </table>
  );

  return (
    <div>
      <h2 style={{ marginBottom: 20 }}>Tokens</h2>
      {plainKey && (
        <div style={{ background: '#d4edda', padding: 12, borderRadius: 6, marginBottom: 16 }}>
          <strong>New token created:</strong> <code>{plainKey}</code>
          <div style={{ fontSize: 12, color: '#155724', marginTop: 4 }}>Copy this now — it will not be shown again.</div>
          <button onClick={() => setPlainKey(null)} style={{ marginTop: 8, ...btnSmall }}>Dismiss</button>
        </div>
      )}
      <form onSubmit={handleSubmit} style={{ background: '#fff', padding: 20, borderRadius: 8, marginBottom: 24 }}>
        <div style={{ display: 'grid', gap: 12, gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', marginBottom: 12 }}>
          <input placeholder="Name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required style={inputStyle} />
          <input placeholder="Max Input Tokens" type="number" value={form.max_input_tokens} onChange={(e) => setForm({ ...form, max_input_tokens: e.target.value })} style={inputStyle} />
          <input placeholder="Max Output Tokens" type="number" value={form.max_output_tokens} onChange={(e) => setForm({ ...form, max_output_tokens: e.target.value })} style={inputStyle} />
          <input placeholder="Req/Min" type="number" value={form.requests_per_minute} onChange={(e) => setForm({ ...form, requests_per_minute: e.target.value })} style={inputStyle} />
          <label style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <input type="checkbox" checked={form.is_active} onChange={(e) => setForm({ ...form, is_active: e.target.checked })} />
            Active
          </label>
        </div>
        <div style={{ marginBottom: 12 }}>
          <strong>Allowed Models & Per-Model Limits:</strong>
          {renderModelTable(
            form.model_permissions,
            (id) => toggleModel(id, false),
            (id, field, value) => updatePermField(id, field, value, false)
          )}
        </div>
        <button type="submit" style={btnPrimary}>Create Token</button>
      </form>

      <table style={{ width: '100%', borderCollapse: 'collapse', background: '#fff', borderRadius: 8, overflow: 'hidden' }}>
        <thead>
          <tr style={{ background: '#eee' }}>
            <th style={thStyle}>Name</th>
            <th style={thStyle}>Allowed Models</th>
            <th style={thStyle}>RPM</th>
            <th style={thStyle}>Active</th>
            <th style={thStyle}>Actions</th>
          </tr>
        </thead>
        <tbody>
          {tokens.map((t) => (
            <tr key={t.id} style={{ borderBottom: '1px solid #eee' }}>
              <td style={tdStyle}>{t.name}</td>
              <td style={tdStyle}>
                {t.model_permissions?.map((p) => {
                  const model = models.find((m) => m.id === p.model_id);
                  const actual = model?.model_id || p.model_id;
                  const display = model?.display_model_id;
                  const limits = [];
                  if (p.max_input_tokens) limits.push(`in:${p.max_input_tokens}`);
                  if (p.max_output_tokens) limits.push(`out:${p.max_output_tokens}`);
                  const label = display ? `${actual} → ${display}` : actual;
                  return `${label}${limits.length ? ' (' + limits.join(', ') + ')' : ''}`;
                }).join(', ') || 'All'}
              </td>
              <td style={tdStyle}>{t.requests_per_minute ?? 'Default'}</td>
              <td style={tdStyle}>{t.is_active ? 'Yes' : 'No'}</td>
              <td style={tdStyle}>
                <button onClick={() => openUsageDialog(t)} style={btnSmall}>Usage</button>
                <button onClick={() => openEdit(t)} style={btnSmall}>Edit</button>
                <button onClick={() => remove(t.id)} style={btnSmallDanger}>Delete</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {/* Edit Modal */}
      {editToken && (
        <div style={overlayStyle} onClick={closeEdit}>
          <div style={dialogStyle} onClick={(e) => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
              <h3 style={{ margin: 0 }}>Edit Token: {editToken.name}</h3>
              <button onClick={closeEdit} style={btnClose}>×</button>
            </div>
            <form onSubmit={handleEditSubmit}>
              <div style={{ display: 'grid', gap: 12, gridTemplateColumns: 'repeat(auto-fill, minmax(180px, 1fr))', marginBottom: 12 }}>
                <input placeholder="Name" value={editForm.name} onChange={(e) => setEditForm({ ...editForm, name: e.target.value })} required style={inputStyle} />
                <input placeholder="Max Input Tokens" type="number" value={editForm.max_input_tokens} onChange={(e) => setEditForm({ ...editForm, max_input_tokens: e.target.value })} style={inputStyle} />
                <input placeholder="Max Output Tokens" type="number" value={editForm.max_output_tokens} onChange={(e) => setEditForm({ ...editForm, max_output_tokens: e.target.value })} style={inputStyle} />
                <input placeholder="Req/Min" type="number" value={editForm.requests_per_minute} onChange={(e) => setEditForm({ ...editForm, requests_per_minute: e.target.value })} style={inputStyle} />
                <label style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <input type="checkbox" checked={editForm.is_active} onChange={(e) => setEditForm({ ...editForm, is_active: e.target.checked })} />
                  Active
                </label>
              </div>
              <div style={{ marginBottom: 12 }}>
                <strong>Allowed Models & Per-Model Limits:</strong>
                {renderModelTable(
                  editForm.model_permissions,
                  (id) => toggleModel(id, true),
                  (id, field, value) => updatePermField(id, field, value, true)
                )}
              </div>
              <div style={{ display: 'flex', gap: 8, justifyContent: 'space-between', alignItems: 'center' }}>
                <button type="button" onClick={handleRegenerate} style={{ ...btnPrimary, background: '#e67e22' }}>Regenerate Key</button>
                <div style={{ display: 'flex', gap: 8 }}>
                  <button type="button" onClick={closeEdit} style={{ ...btnPrimary, background: '#666' }}>Cancel</button>
                  <button type="submit" style={btnPrimary}>Save</button>
                </div>
              </div>
            </form>
          </div>
        </div>
      )}

      {dialogOpen && dialogToken && (
        <div style={overlayStyle} onClick={closeDialog}>
          <div style={dialogStyle} onClick={(e) => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
              <h3 style={{ margin: 0 }}>Token Usage: {dialogToken.name}</h3>
              <button onClick={closeDialog} style={btnClose}>×</button>
            </div>
            {usageLoading ? (
              <div>Loading…</div>
            ) : usageData.length === 0 ? (
              <div style={{ color: '#666' }}>No usage data available for this token.</div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                {usageData.map((u) => (
                  <div key={u.model_id}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                      <strong>{u.model_name}</strong>
                      <span style={{ fontSize: 13, color: '#555' }}>
                        {u.current_usage.toLocaleString()} / {u.max_tokens.toLocaleString()} tokens ({u.percentage}%)
                      </span>
                    </div>
                    <div style={{ background: '#e0e0e0', borderRadius: 4, height: 20, overflow: 'hidden' }}>
                      <div
                        style={{
                          width: `${Math.min(u.percentage, 100)}%`,
                          height: '100%',
                          background: u.percentage >= 90 ? '#c0392b' : u.percentage >= 70 ? '#e67e22' : '#27ae60',
                          borderRadius: 4,
                          transition: 'width 0.4s ease',
                        }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

const inputStyle: React.CSSProperties = { padding: '8px 12px', border: '1px solid #ccc', borderRadius: 4 };
const btnPrimary: React.CSSProperties = { padding: '8px 16px', background: '#1a1a2e', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer' };
const btnSmall: React.CSSProperties = { padding: '4px 10px', background: '#1a1a2e', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer', marginRight: 4 };
const btnSmallDanger: React.CSSProperties = { padding: '4px 10px', background: '#c0392b', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer' };
const btnClose: React.CSSProperties = { background: 'none', border: 'none', fontSize: 24, cursor: 'pointer', color: '#666' };
const thStyle: React.CSSProperties = { padding: '12px 16px', textAlign: 'left', fontWeight: 600 };
const tdStyle: React.CSSProperties = { padding: '12px 16px' };
const thStyleSmall: React.CSSProperties = { padding: '8px 12px', textAlign: 'left', fontWeight: 600, fontSize: 13 };
const tdStyleSmall: React.CSSProperties = { padding: '8px 12px', fontSize: 13 };
const overlayStyle: React.CSSProperties = {
  position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
  background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000
};
const dialogStyle: React.CSSProperties = {
  background: '#fff', padding: 24, borderRadius: 8, minWidth: 420, maxWidth: 720, width: '90%', maxHeight: '85vh', overflow: 'auto'
};
