import { useState, useRef, useEffect } from 'react';

interface Option {
  value: string;
  label: string;
}

interface SearchableDropdownProps {
  label: string;
  options: Option[];
  value: string | null;
  onChange: (value: string | null) => void;
  placeholder?: string;
}

export default function SearchableDropdown({
  label,
  options,
  value,
  onChange,
  placeholder = 'Search...',
}: SearchableDropdownProps) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const containerRef = useRef<HTMLDivElement>(null);

  const selected = options.find((o) => o.value === value);
  const filtered = options.filter((o) =>
    o.label.toLowerCase().includes(search.toLowerCase())
  );

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  return (
    <div ref={containerRef} style={{ position: 'relative', minWidth: 200 }}>
      <label style={{ fontSize: 12, fontWeight: 600, color: '#555', marginBottom: 4, display: 'block' }}>
        {label}
      </label>
      <div
        onClick={() => setOpen(!open)}
        style={{
          padding: '8px 12px',
          border: '1px solid #ccc',
          borderRadius: 4,
          background: '#fff',
          cursor: 'pointer',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          fontSize: 14,
        }}
      >
        <span style={{ color: selected ? '#333' : '#888' }}>
          {selected ? selected.label : `All ${label}s`}
        </span>
        <span style={{ fontSize: 10, color: '#888' }}>{open ? '▲' : '▼'}</span>
      </div>
      {open && (
        <div
          style={{
            position: 'absolute',
            top: '100%',
            left: 0,
            right: 0,
            marginTop: 4,
            background: '#fff',
            border: '1px solid #ccc',
            borderRadius: 4,
            boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
            zIndex: 100,
            maxHeight: 280,
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <div style={{ padding: 8, borderBottom: '1px solid #eee' }}>
            <input
              autoFocus
              placeholder={placeholder}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              style={{
                width: '100%',
                padding: '6px 10px',
                border: '1px solid #ddd',
                borderRadius: 4,
                fontSize: 13,
                boxSizing: 'border-box',
              }}
            />
          </div>
          <div style={{ overflowY: 'auto', flex: 1 }}>
            <div
              onClick={() => { onChange(null); setOpen(false); setSearch(''); }}
              style={{
                padding: '8px 12px',
                cursor: 'pointer',
                background: value === null ? '#f0f4ff' : '#fff',
                fontWeight: value === null ? 600 : 400,
                fontSize: 13,
              }}
            >
              All {label}s
            </div>
            {filtered.map((opt) => (
              <div
                key={opt.value}
                onClick={() => { onChange(opt.value); setOpen(false); setSearch(''); }}
                style={{
                  padding: '8px 12px',
                  cursor: 'pointer',
                  background: value === opt.value ? '#f0f4ff' : '#fff',
                  fontWeight: value === opt.value ? 600 : 400,
                  fontSize: 13,
                  borderTop: '1px solid #f5f5f5',
                }}
              >
                {opt.label}
              </div>
            ))}
            {filtered.length === 0 && (
              <div style={{ padding: '12px', color: '#888', fontSize: 13, textAlign: 'center' }}>
                No matches
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
