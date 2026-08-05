import React from 'react';
import { Card, Tag, Empty, Table } from 'antd';
import { WarningOutlined, InfoCircleOutlined, AlertOutlined, CheckCircleOutlined } from '@ant-design/icons';
import { useScanStore } from '../stores/scanStore';
import type { Finding } from '../types';

const severityConfig: Record<string, { color: string; icon: React.ReactNode }> = {
  critical: { color: 'red', icon: <AlertOutlined /> },
  high: { color: 'orange', icon: <WarningOutlined /> },
  medium: { color: 'gold', icon: <InfoCircleOutlined /> },
  low: { color: 'blue', icon: <InfoCircleOutlined /> },
  info: { color: 'green', icon: <CheckCircleOutlined /> },
};

const FindingsPage: React.FC = () => {
  const { findings } = useScanStore();

  const columns = [
    {
      title: 'Severity', dataIndex: 'severity', width: 100,
      render: (s: string) => {
        const cfg = severityConfig[s] || severityConfig.info;
        return <Tag color={cfg.color} icon={cfg.icon}>{s.toUpperCase()}</Tag>;
      },
    },
    { title: 'Type', dataIndex: 'type', width: 140 },
    { title: 'Title', dataIndex: 'title', ellipsis: true },
    { title: 'Description', dataIndex: 'description', ellipsis: true },
    {
      title: 'Recommendation', dataIndex: 'recommendation', ellipsis: true,
      render: (r: string) => r ? <span className="text-green-400">{r}</span> : '-',
    },
  ];

  return (
    <div className="h-full p-2">
      <Card title="Findings" className="h-full" bodyStyle={{ height: 'calc(100% - 57px)' }}>
        {findings.length === 0 ? (
          <Empty description="No findings. Run a scan to discover issues." />
        ) : (
          <Table dataSource={findings} columns={columns} rowKey="id" size="small" scroll={{ y: 'calc(100vh - 280px)' }} />
        )}
      </Card>
    </div>
  );
};

export default FindingsPage;
