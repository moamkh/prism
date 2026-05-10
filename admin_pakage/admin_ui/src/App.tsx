import { Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Providers from './pages/Providers';
import ModelsPage from './pages/Models';
import Tokens from './pages/Tokens';
import UsageLogs from './pages/UsageLogs';
import Logs from './pages/Logs';
import Playground from './pages/Playground';
import Settings from './pages/Settings';

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<Dashboard />} />
        <Route path="providers" element={<Providers />} />
        <Route path="models" element={<ModelsPage />} />
        <Route path="tokens" element={<Tokens />} />
        <Route path="usage" element={<UsageLogs />} />
        <Route path="logs" element={<Logs />} />
        <Route path="playground" element={<Playground />} />
        <Route path="settings" element={<Settings />} />
      </Route>
    </Routes>
  );
}

export default App;
