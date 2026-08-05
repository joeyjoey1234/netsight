import React, { useState } from 'react';
import { Card, Form, Input, Button, Switch, Space, message, Typography } from 'antd';
import { useThemeStore } from '../stores/themeStore';
import { createProject } from '../hooks/api';

const { Text } = Typography;

const SettingsPage: React.FC = () => {
  const { isDark, toggle } = useThemeStore();
  const [projectName, setProjectName] = useState('');

  const handleSave = async () => {
    if (!projectName) return;
    try {
      const project = await createProject(projectName);
      message.success(`Project "${projectName}" created`);
    } catch (err: any) {
      message.error(err?.message || 'Failed to create project');
    }
  };

  return (
    <div className="h-full p-2 overflow-y-auto">
      <Card title="Project" size="small" className="mb-4">
        <Form layout="vertical">
          <Form.Item label="Project Name">
            <Input value={projectName} onChange={e => setProjectName(e.target.value)} placeholder="My Network Survey" />
          </Form.Item>
          <Button type="primary" onClick={handleSave}>Save Project</Button>
        </Form>
      </Card>

      <Card title="Appearance" size="small" className="mb-4">
        <Space><span>Dark Mode:</span><Switch checked={isDark} onChange={toggle} /></Space>
      </Card>

      <Card title="About" size="small">
        <Text type="secondary">
          NetSight v0.1.0<br />
          Portable Network Analysis Tool for Windows<br />
          Built with Go + Wails v2 + React
        </Text>
      </Card>
    </div>
  );
};

export default SettingsPage;
