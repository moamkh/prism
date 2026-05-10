import { useEffect, useState } from 'react';
import { configApi } from '../api/client';
import type { ConfigItem } from '../types';

export default function Settings() {
  const [configs, setConfigs] = useState<ConfigItem[]>([]);
  const [editing, setEditing] = useState<Record<string, string>>({});

  const load = () => {
    configApi.list().then((res) => {
      setConfigs(res.data);
      const map: Record<string, string> = {};
      res.data.forEach((c: ConfigItem) => { map[c.key] = c.value; });
      setEditing(map);
    });
  };

  useEffect(() => {
    load();
  }, []);

  const save = async (key: string) => {
    await configApi.update(key, editing[key]);
    load();
  };

  return (
    <div>
      <h2 style={{ marginBottom: 20 }}>Settings</h2>
      <div style={{ background: '#fff', padding: 20, borderRadius: 8 }}>
        {configs.map((c) => (
          <div key={c.key} style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 12 }}>
            <div style={{ width: 220, fontWeight: 600 }}>{c.key}</div>
            <input
              value={editing[c.key] ?? c.value}
              onChange={(e) => setEditing({ ...editing, [c.key]: e.target.value })}
              style={{ flex: 1, padding: '8px 12px', border: '1px solid #ccc', borderRadius: 4 }}
            />
            <button onClick={() => save(c.key)} style={btnPrimary}>Save</button>
          </div>
        ))}
      </div>
    </div>
  );
}

const btnPrimary: React.CSSProperties = { padding: '8px 16px', background: '#1a1a2e', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer' };
