import { useEffect, useState } from 'react';
import { usageApi, tokensApi, providersApi } from '../api/client';
import type { UsageLog, Token, Provider } from '../types';
import SearchableDropdown from '../components/SearchableDropdown';
import DateRangePicker from '../components/DateRangePicker';

interface Totals {
  count: number;
  total_input_tokens: number;
  total_output_tokens: number;
  total_tokens: number;
}

type SuccessFilter = 'all' | 'successful' | 'failed';

export default function UsageLogs() {
  const [logs, setLogs] = useState<UsageLog[]>([]);
  const [totals, setTotals] = useState<Totals | null>(null);
  const [page, setPage] = useState(0);
  const limit = 50;

  const [tokens, setTokens] = useState<Token[]>([]);
  const [providers, setProviders] = useState<Provider[]>([]);

  const [selectedToken, setSelectedToken] = useState<string | null>(null);
  const [selectedProvider, setSelectedProvider] = useState<string | null>(null);
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [successFilter, setSuccessFilter] = useState<SuccessFilter>('successful');

  const load = () => {
    const params: Record<string, unknown> = { skip: page * limit, limit };
    if (selectedToken) params.token_name = selectedToken;
    if (selectedProvider) params.provider_name = selectedProvider;
    if (startDate) params.start_date = new Date(startDate).toISOString();
    if (endDate) params.end_date = new Date(endDate).toISOString();
    if (successFilter !== 'all') {
      params.is_successful = successFilter === 'successful';
    }

    usageApi.filtered(params).then((res) => {
      setLogs(res.data.logs || []);
      setTotals(res.data.totals || null);
    });
  };

  useEffect(() => {
    tokensApi.list().then((res) => setTokens(res.data));
    providersApi.list().then((res) => setProviders(res.data));
  }, []);

  useEffect(() => {
    load();
  }, [page, selectedToken, selectedProvider, startDate, endDate, successFilter]);

  const tokenOptions = tokens.map((t) => ({ value: t.name, label: t.name }));
  const providerOptions = providers.map((p) => ({ value: p.name, label: p.name }));

  return (
    <div>
      <h2 style={{ marginBottom: 20 }}>Usage Logs</h2>

      {/* Filters */}
      <div style={{ background: '#fff', padding: 16, borderRadius: 8, marginBottom: 16, display: 'flex', flexWrap: 'wrap', gap: 16, alignItems: 'flex-start' }}>
        <SearchableDropdown
          label="Token"
          options={tokenOptions}
          value={selectedToken}
          onChange={(v) => { setSelectedToken(v); setPage(0); }}
          placeholder="Search tokens..."
        />
        <SearchableDropdown
          label="Provider"
          options={providerOptions}
          value={selectedProvider}
          onChange={(v) => { setSelectedProvider(v); setPage(0); }}
          placeholder="Search providers..."
        />
        <DateRangePicker
          startDate={startDate}
          endDate={endDate}
          onStartChange={(v) => { setStartDate(v); setPage(0); }}
          onEndChange={(v) => { setEndDate(v); setPage(0); }}
        />
        <div>
          <label style={{ display: 'block', fontSize: 12, fontWeight: 600, marginBottom: 4, color: '#555' }}>Status</label>
          <select
            value={successFilter}
            onChange={(e) => { setSuccessFilter(e.target.value as SuccessFilter); setPage(0); }}
            style={{ padding: '8px 12px', borderRadius: 4, border: '1px solid #ddd', fontSize: 14, minWidth: 140 }}
          >
            <option value="successful">Successful</option>
            <option value="failed">Failed</option>
            <option value="all">All</option>
          </select>
        </div>
      </div>

      {/* Totals */}
      {totals && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(160px, 1fr))', gap: 12, marginBottom: 16 }}>
          <TotalCard label="Requests" value={totals.count.toLocaleString()} />
          <TotalCard label="Input Tokens" value={totals.total_input_tokens.toLocaleString()} color="#3498db" />
          <TotalCard label="Output Tokens" value={totals.total_output_tokens.toLocaleString()} color="#2ecc71" />
          <TotalCard label="Total Tokens" value={totals.total_tokens.toLocaleString()} color="#9b59b6" />
        </div>
      )}

      {/* Table */}
      <table style={{ width: '100%', borderCollapse: 'collapse', background: '#fff', borderRadius: 8, overflow: 'hidden' }}>
        <thead>
          <tr style={{ background: '#eee' }}>
            <th style={thStyle}>Time</th>
            <th style={thStyle}>Token</th>
            <th style={thStyle}>Provider</th>
            <th style={thStyle}>Path</th>
            <th style={thStyle}>Input</th>
            <th style={thStyle}>Output</th>
            <th style={thStyle}>Total</th>
            <th style={thStyle}>Latency</th>
            <th style={thStyle}>Status</th>
          </tr>
        </thead>
        <tbody>
          {logs.length === 0 && (
            <tr>
              <td colSpan={9} style={{ padding: 40, textAlign: 'center', color: '#888' }}>
                No usage logs found for the selected filters.
              </td>
            </tr>
          )}
          {logs.map((log) => {
            const failed = !log.is_successful;
            const rowStyle: React.CSSProperties = failed
              ? { borderBottom: '1px solid #eee', background: '#fff5f5' }
              : { borderBottom: '1px solid #eee' };
            return (
              <tr
                key={log.id}
                style={rowStyle}
                title={failed && log.error_message ? log.error_message : undefined}
              >
                <td style={tdStyle}>{new Date(log.created_at).toLocaleString()}</td>
                <td style={tdStyle}>{log.token_id ? tokens.find((t) => t.id === log.token_id)?.name || log.token_id.slice(0, 8) : '-'}</td>
                <td style={tdStyle}>{log.provider_name || '-'}</td>
                <td style={tdStyle}>{log.request_path}</td>
                <td style={tdStyle}>{log.input_tokens.toLocaleString()}</td>
                <td style={tdStyle}>{log.output_tokens.toLocaleString()}</td>
                <td style={tdStyle}><strong>{log.total_tokens.toLocaleString()}</strong></td>
                <td style={tdStyle}>{log.latency_ms}ms</td>
                <td style={tdStyle}>
                  {failed ? (
                    <span style={{ color: '#e74c3c', fontWeight: 600 }}>{log.status_code}</span>
                  ) : (
                    log.status_code
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>

      {/* Pagination */}
      <div style={{ marginTop: 16, display: 'flex', alignItems: 'center', gap: 12 }}>
        <button onClick={() => setPage((p) => Math.max(0, p - 1))} disabled={page === 0} style={{ ...btnPrimary, opacity: page === 0 ? 0.5 : 1 }}>
          Previous
        </button>
        <span style={{ fontSize: 14, color: '#555' }}>Page {page + 1}</span>
        <button onClick={() => setPage((p) => p + 1)} disabled={logs.length < limit} style={{ ...btnPrimary, opacity: logs.length < limit ? 0.5 : 1 }}>
          Next
        </button>
        {logs.length > 0 && (
          <span style={{ fontSize: 13, color: '#888', marginLeft: 'auto' }}>
            Showing {logs.length} records
          </span>
        )}
      </div>
    </div>
  );
}

function TotalCard({ label, value, color = '#1a1a2e' }: { label: string; value: string; color?: string }) {
  return (
    <div style={{ background: '#fff', padding: 16, borderRadius: 8, textAlign: 'center', borderLeft: `4px solid ${color}` }}>
      <div style={{ fontSize: 12, color: '#888', textTransform: 'uppercase', letterSpacing: 1, marginBottom: 4 }}>{label}</div>
      <div style={{ fontSize: 22, fontWeight: 700, color }}>{value}</div>
    </div>
  );
}

const btnPrimary: React.CSSProperties = { padding: '8px 16px', background: '#1a1a2e', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer' };
const thStyle: React.CSSProperties = { padding: '12px 16px', textAlign: 'left', fontWeight: 600 };
const tdStyle: React.CSSProperties = { padding: '12px 16px' };
