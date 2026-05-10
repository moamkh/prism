import { useToast } from '../hooks/useToast';

const typeStyles: Record<string, React.CSSProperties> = {
  success: { background: '#d4edda', color: '#155724', borderLeft: '4px solid #28a745' },
  error: { background: '#f8d7da', color: '#721c24', borderLeft: '4px solid #dc3545' },
  info: { background: '#d1ecf1', color: '#0c5460', borderLeft: '4px solid #17a2b8' },
};

export default function ToastContainer() {
  const { toasts, removeToast } = useToast();

  if (toasts.length === 0) return null;

  return (
    <div
      style={{
        position: 'fixed',
        bottom: 20,
        right: 20,
        zIndex: 9999,
        display: 'flex',
        flexDirection: 'column',
        gap: 10,
        minWidth: 280,
        maxWidth: 400,
      }}
    >
      {toasts.map((toast) => (
        <div
          key={toast.id}
          style={{
            padding: '12px 16px',
            borderRadius: 6,
            boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            gap: 12,
            animation: 'toastSlideIn 0.3s ease',
            ...typeStyles[toast.type],
          }}
        >
          <span style={{ fontSize: 14, fontWeight: 500 }}>{toast.message}</span>
          <button
            onClick={() => removeToast(toast.id)}
            style={{
              background: 'none',
              border: 'none',
              fontSize: 18,
              cursor: 'pointer',
              color: 'inherit',
              opacity: 0.6,
              lineHeight: 1,
            }}
          >
            ×
          </button>
        </div>
      ))}
      <style>{`
        @keyframes toastSlideIn {
          from { transform: translateX(100%); opacity: 0; }
          to { transform: translateX(0); opacity: 1; }
        }
      `}</style>
    </div>
  );
}
