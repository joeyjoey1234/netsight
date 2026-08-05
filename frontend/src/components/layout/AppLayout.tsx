import React, { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Space, Badge } from 'antd';
import {
  AimOutlined,
  ThunderboltOutlined,
  ToolOutlined,
  CloudServerOutlined,
  WarningOutlined,
  SettingOutlined,
  DashboardOutlined,
} from '@ant-design/icons';
import { useScanStore } from '../../stores/scanStore';
import { useThemeStore } from '../../stores/themeStore';

const { Sider, Content } = Layout;

const AppLayout: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { findings } = useScanStore();
  const { isDark } = useThemeStore();

  const menuItems = [
    { key: '/', icon: <AimOutlined />, label: 'Topology' },
    { key: '/capture', icon: <DashboardOutlined />, label: 'Capture' },
    { key: '/tools', icon: <ToolOutlined />, label: 'Tools' },
    { key: '/servers', icon: <CloudServerOutlined />, label: 'Servers' },
    {
      key: '/findings',
      icon: <WarningOutlined />,
      label: (
        <Space>
          Findings
          {findings.length > 0 && (
            <Badge count={findings.length} size="small" overflowCount={99} />
          )}
        </Space>
      ),
    },
    { key: '/settings', icon: <SettingOutlined />, label: 'Settings' },
  ];

  return (
    <Layout className="h-screen">
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        theme={isDark ? 'dark' : 'light'}
        width={200}
        className="border-r border-gray-700"
      >
        <div className="h-12 flex items-center justify-center text-lg font-bold text-blue-400 select-none">
          {collapsed ? 'NS' : 'NetSight'}
        </div>
        <Menu
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          theme={isDark ? 'dark' : 'light'}
        />
      </Sider>
      <Layout>
        <Content className="overflow-hidden bg-transparent">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default AppLayout;
