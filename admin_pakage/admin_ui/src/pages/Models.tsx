import { useEffect, useState } from 'react';
import { modelsApi, providersApi } from '../api/client';
import type { Model, Provider } from '../types';

interface FetchedModel {
  id: string;
  object: string;
  created: number;
  owned_by: string;
  [key: string]: unknown;
}

export default function ModelsPage() {
  const [models, setModels] = useState<Model[]>([]);
  const [providers, setProviders] = useState<Provider[]>([]);
  const [selectedProvider, setSelectedProvider] = useState('');
  const [fetchedModels, setFetchedModels] = useState<FetchedModel[]>([]);
  const [selectedModelId, setSelectedModelId] = useState('');
  const [fetching, setFetching] = useState(false);
  const [detailModel, setDetailModel] = useState<FetchedModel | null>(null);
  const [form, setForm] = useState({ provider_id: '', model_id: '', display_model_id: '', is_active: true });

  const load = () => {
    modelsApi.list().then((res) => setModels(res.data));
    providersApi.list().then((res) => setProviders(res.data));
  };

  useEffect(() => {
    load();
  }, []);

  const handleFetch = async () => {
    if (!selectedProvider) return;
    setFetching(true);
    try {
      const res = await modelsApi.fetchAvailable(selectedProvider);
      console.log('Fetch models response:', res.data);
      const modelList = res.data?.models || [];
      setFetchedModels(modelList);
      setSelectedModelId('');
      if (modelList.length === 0) {
        alert('Provider returned 0 models');
      }
    } catch (e: any) {
      console.error('Fetch models error:', e);
      const msg = e?.response?.data?.detail || e?.message || 'Unknown error';
      alert('Failed to fetch models: ' + msg);
    } finally {
      setFetching(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await modelsApi.create(form);
    setForm({ provider_id: '', model_id: '', display_model_id: '', is_active: true });
    setSelectedProvider('');
    setFetchedModels([]);
    setSelectedModelId('');
    load();
  };

  const toggle = async (m: Model) => {
    await modelsApi.update(m.id, { is_active: !m.is_active });
    load();
  };

  const remove = async (id: string) => {
    await modelsApi.remove(id);
    load();
  };

  const showDetails = async (m: Model) => {
    const provider = providers.find((p) => p.id === m.provider_id);
    if (!provider) return;
    try {
      const res = await modelsApi.fetchAvailable(provider.id);
      const all = res.data.models || [];
      const found = all.find((x: FetchedModel) => x.id === m.model_id);
      setDetailModel(found || null);
    } catch {
      setDetailModel(null);
    }
  };

  return (
    <div>
      <h2 style={{ marginBottom: 20 }}>Models</h2>

      <form onSubmit={handleSubmit} style={{ background: '#fff', padding: 20, borderRadius: 8, marginBottom: 24 }}>
        <div style={{ display: 'grid', gap: 12, gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))' }}>
          <select
            value={selectedProvider}
            onChange={(e) => {
              setSelectedProvider(e.target.value);
              setForm({ ...form, provider_id: e.target.value });
              setFetchedModels([]);
              setSelectedModelId('');
            }}
            required
            style={inputStyle}
          >
            <option value="">Select Provider</option>
            {providers.map((p) => (
              <option key={p.id} value={p.id}>{p.name}</option>
            ))}
          </select>

          <div style={{ display: 'flex', gap: 8 }}>
            <button type="button" onClick={handleFetch} disabled={fetching || !selectedProvider} style={btnPrimary}>
              {fetching ? 'Fetching...' : 'Fetch Models'}
            </button>
          </div>

          {fetchedModels.length > 0 && (
            <select
              value={selectedModelId}
              onChange={(e) => {
                setSelectedModelId(e.target.value);
                setForm({ ...form, model_id: e.target.value });
              }}
              required
              style={inputStyle}
            >
              <option value="">Select Model ID</option>
              {fetchedModels.map((m) => (
                <option key={m.id} value={m.id}>{m.id}</option>
              ))}
            </select>
          )}

          <input
            type="text"
            placeholder="Display Model ID (optional)"
            value={form.display_model_id}
            onChange={(e) => setForm({ ...form, display_model_id: e.target.value })}
            style={inputStyle}
          />

          <label style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <input type="checkbox" checked={form.is_active} onChange={(e) => setForm({ ...form, is_active: e.target.checked })} />
            Active
          </label>
        </div>
        <div style={{ marginTop: 12 }}>
          <button type="submit" style={btnPrimary}>Add Model</button>
        </div>
      </form>

      {detailModel && (
        <div style={{ background: '#fff', padding: 20, borderRadius: 8, marginBottom: 24 }}>
          <h3 style={{ marginBottom: 12 }}>Model Details</h3>
          <pre style={{ background: '#f5f5f5', padding: 12, borderRadius: 4, overflow: 'auto' }}>
            {JSON.stringify(detailModel, null, 2)}
          </pre>
          <button onClick={() => setDetailModel(null)} style={{ ...btnSecondary, marginTop: 12 }}>Close</button>
        </div>
      )}

      <table style={{ width: '100%', borderCollapse: 'collapse', background: '#fff', borderRadius: 8, overflow: 'hidden' }}>
        <thead>
          <tr style={{ background: '#eee' }}>
            <th style={thStyle}>Model ID</th>
            <th style={thStyle}>Display ID</th>
            <th style={thStyle}>Provider</th>
            <th style={thStyle}>Active</th>
            <th style={thStyle}>Actions</th>
          </tr>
        </thead>
        <tbody>
          {models.map((m) => (
            <tr key={m.id} style={{ borderBottom: '1px solid #eee', cursor: 'pointer' }} onClick={() => showDetails(m)}>
              <td style={tdStyle}>{m.model_id}</td>
              <td style={tdStyle}>{m.display_model_id || '-'}</td>
              <td style={tdStyle}>{providers.find((p) => p.id === m.provider_id)?.name || m.provider_id}</td>
              <td style={tdStyle}>{m.is_active ? 'Yes' : 'No'}</td>
              <td style={tdStyle} onClick={(e) => e.stopPropagation()}>
                <button onClick={() => toggle(m)} style={btnSmall}>{m.is_active ? 'Deactivate' : 'Activate'}</button>
                <button onClick={() => remove(m.id)} style={btnSmallDanger}>Delete</button>
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
const btnSecondary: React.CSSProperties = { padding: '8px 16px', background: '#ccc', border: 'none', borderRadius: 4, cursor: 'pointer' };
const btnSmall: React.CSSProperties = { padding: '4px 10px', background: '#1a1a2e', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer', marginRight: 4 };
const btnSmallDanger: React.CSSProperties = { padding: '4px 10px', background: '#c0392b', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer' };
const thStyle: React.CSSProperties = { padding: '12px 16px', textAlign: 'left', fontWeight: 600 };
const tdStyle: React.CSSProperties = { padding: '12px 16px' };
