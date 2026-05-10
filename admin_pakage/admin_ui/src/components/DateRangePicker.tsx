interface DateRangePickerProps {
  startDate: string;
  endDate: string;
  onStartChange: (value: string) => void;
  onEndChange: (value: string) => void;
}

export default function DateRangePicker({
  startDate,
  endDate,
  onStartChange,
  onEndChange,
}: DateRangePickerProps) {
  const presets = [
    { label: 'Today', days: 0 },
    { label: 'Last 7 days', days: 7 },
    { label: 'Last 30 days', days: 30 },
    { label: 'Last 90 days', days: 90 },
  ];

  const applyPreset = (days: number) => {
    const end = new Date();
    end.setHours(23, 59, 59, 999);
    const start = new Date();
    start.setDate(start.getDate() - days);
    start.setHours(0, 0, 0, 0);
    onEndChange(end.toISOString().slice(0, 16));
    onStartChange(start.toISOString().slice(0, 16));
  };

  const clearRange = () => {
    onStartChange('');
    onEndChange('');
  };

  return (
    <div>
      <label style={{ fontSize: 12, fontWeight: 600, color: '#555', marginBottom: 4, display: 'block' }}>
        Date Range
      </label>
      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center' }}>
        <input
          type="datetime-local"
          value={startDate}
          onChange={(e) => onStartChange(e.target.value)}
          style={{
            padding: '8px 10px',
            border: '1px solid #ccc',
            borderRadius: 4,
            fontSize: 13,
            fontFamily: 'inherit',
          }}
        />
        <span style={{ color: '#888', fontSize: 13 }}>to</span>
        <input
          type="datetime-local"
          value={endDate}
          onChange={(e) => onEndChange(e.target.value)}
          style={{
            padding: '8px 10px',
            border: '1px solid #ccc',
            borderRadius: 4,
            fontSize: 13,
            fontFamily: 'inherit',
          }}
        />
        {presets.map((p) => (
          <button
            key={p.label}
            onClick={() => applyPreset(p.days)}
            style={{
              padding: '6px 10px',
              background: '#f0f0f0',
              border: 'none',
              borderRadius: 4,
              cursor: 'pointer',
              fontSize: 12,
              color: '#444',
            }}
          >
            {p.label}
          </button>
        ))}
        {(startDate || endDate) && (
          <button
            onClick={clearRange}
            style={{
              padding: '6px 10px',
              background: 'transparent',
              border: '1px solid #ccc',
              borderRadius: 4,
              cursor: 'pointer',
              fontSize: 12,
              color: '#666',
            }}
          >
            Clear
          </button>
        )}
      </div>
    </div>
  );
}
