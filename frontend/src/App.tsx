import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { ConfigProvider, theme as antTheme, App as AntApp } from 'antd';
import { useThemeStore } from './stores/themeStore';
import AppLayout from './components/layout/AppLayout';
import TopologyPage from './pages/TopologyPage';
import CapturePage from './pages/CapturePage';
import ToolsPage from './pages/ToolsPage';
import ServersPage from './pages/ServersPage';
import FindingsPage from './pages/FindingsPage';
import SettingsPage from './pages/SettingsPage';

function App() {
  const { isDark } = useThemeStore();

  return (
    <ConfigProvider
      theme={{
        algorithm: isDark ? antTheme.darkAlgorithm : antTheme.defaultAlgorithm,
        token: {
          colorPrimary: '#1677ff',
          borderRadius: 6,
        },
      }}
    >
      <AntApp>
        <BrowserRouter>
          <Routes>
            <Route element={<AppLayout />}>
              <Route path="/" element={<TopologyPage />} />
              <Route path="/capture" element={<CapturePage />} />
              <Route path="/tools" element={<ToolsPage />} />
              <Route path="/servers" element={<ServersPage />} />
              <Route path="/findings" element={<FindingsPage />} />
              <Route path="/settings" element={<SettingsPage />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </AntApp>
    </ConfigProvider>
  );
}

export default App;
