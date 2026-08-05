import React from 'react';
import { Card, Form, Input, Button, Select, Switch, Space } from 'antd';
import { useThemeStore } from '../stores/themeStore';

const SettingsPage: React.FC = () => {
  const { isDark, toggle } = useThemeStore();

  return (
    <div className="h-full p-2 overflow-y-auto">
      <Card title="Project Settings" size="small" className="mb-4">
        <Form layout="vertical">
          <Form.Item label="Default Subnet">
            <Input placeholder="192.168.1.0/24" />
          </Form.Item>
          <Form.Item label="Excluded IPs">
            <Select mode="tags" placeholder="Add IPs to exclude" />
          </Form.Item>
          <Button type="primary">Save</Button>
        </Form>
      </Card>

      <Card title="Appearance" size="small" className="mb-4">
        <Space>
          <span>Dark Mode:</span>
          <Switch checked={isDark} onChange={toggle} />
        </Space>
      </Card>

      <Card title="About" size="small">
        <p className="text-gray-400 text-sm">
          NetSight v0.1.0<br />
          Portable Network Analysis Tool for Windows<br />
          Built with Go + Wails v2 + React
        </p>
      </Card>
    </div>
  );
};

export default SettingsPage;
