import React, { useState, useEffect } from 'react';
import { Card, Form, Input, Button, Switch, Space, Select, Modal, message, Typography } from 'antd';
import { useThemeStore } from '../stores/themeStore';
import { createProject, loadProject, listProjects } from '../hooks/api';

const { Text } = Typography;

const SettingsPage: React.FC = () => {
  const { isDark, toggle } = useThemeStore();
  const [projectName, setProjectName] = useState('');
  const [projects, setProjects] = useState<any[]>([]);
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(null);

  useEffect(() => {
    listProjects().then(setProjects).catch(() => setProjects([]));
  }, []);

  const handleSave = async () => {
    if (!projectName) return;
    const existing = projects.find((p: any) => p.name === projectName);
    if (existing) {
      Modal.confirm({
        title: 'Overwrite Project?',
        content: `A project named "${projectName}" already exists. Do you want to overwrite it?`,
        okText: 'Overwrite',
        okType: 'danger',
        cancelText: 'Cancel',
        onOk: async () => {
          try {
            await createProject(projectName);
            message.success(`Project "${projectName}" overwritten`);
            listProjects().then(setProjects).catch(() => {});
          } catch (err: any) {
            message.error(err?.message || 'Failed to overwrite project');
          }
        },
      });
      return;
    }
    try {
      await createProject(projectName);
      message.success(`Project "${projectName}" created`);
      listProjects().then(setProjects).catch(() => {});
    } catch (err: any) {
      message.error(err?.message || 'Failed to create project');
    }
  };

  const handleLoad = async () => {
    if (!selectedProjectId) return;
    try {
      const project = await loadProject(selectedProjectId);
      message.success(`Project loaded: ${project?.name || selectedProjectId}`);
    } catch (err: any) {
      message.error(err?.message || 'Failed to load project');
    }
  };

  return (
    <div className="h-full p-2 overflow-y-auto">
      <Card title="Project" size="small" className="mb-4">
        <Form layout="vertical">
          <Form.Item label="Existing Projects">
            <Space>
              <Select
                placeholder="Select a project"
                value={selectedProjectId}
                onChange={setSelectedProjectId}
                style={{ width: 280 }}
                options={projects.map((p: any) => ({ value: p.id, label: p.name }))}
                notFoundContent={<Text type="secondary">No projects found</Text>}
                allowClear
              />
              <Button type="primary" onClick={handleLoad} disabled={!selectedProjectId}>Load</Button>
            </Space>
          </Form.Item>
          <Form.Item label="New Project Name">
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
