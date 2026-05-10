import { Link, Outlet, useLocation } from 'react-router-dom';
import ToastContainer from './ToastContainer';

const navItems = [
  { path: '/', label: 'Dashboard' },
  { path: '/providers', label: 'Providers' },
  { path: '/models', label: 'Models' },
  { path: '/tokens', label: 'Tokens' },
  { path: '/usage', label: 'Usage Logs' },
  { path: '/logs', label: 'Proxy Logs' },
  { path: '/settings', label: 'Settings' },
];

export default function Layout() {
  const location = useLocation();

  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      <aside style={{ width: 220, background: '#1a1a2e', color: '#fff', padding: 20 }}>
        <h1 style={{ fontSize: 18, marginBottom: 24 }}>Proxy Manager</h1>
        <nav style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {navItems.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              style={{
                padding: '10px 14px',
                borderRadius: 6,
                textDecoration: 'none',
                color: location.pathname === item.path ? '#fff' : '#aaa',
                background: location.pathname === item.path ? '#16213e' : 'transparent',
                fontWeight: location.pathname === item.path ? 600 : 400,
              }}
            >
              {item.label}
            </Link>
          ))}
        </nav>
      </aside>
      <main style={{ flex: 1, padding: 24 }}>
        <Outlet />
      </main>
      <ToastContainer />
    </div>
  );
}
