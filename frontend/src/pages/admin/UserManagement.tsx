import React, { useState } from 'react';
import { List } from '@refinedev/antd';
import { Table, Tag, Space, Button, Switch, Modal, message, Tooltip, Avatar } from 'antd';
import {
  UserOutlined,
  CrownOutlined,
  LockOutlined,
  KeyOutlined,
  SafetyCertificateOutlined,
  TeamOutlined,
  CheckCircleOutlined,
  StopOutlined,
  PlusOutlined,
} from '@ant-design/icons';

export interface UserRow {
  id: string;
  username: string;
  full_name: string;
  email: string;
  role: string;
  is_superuser: boolean;
  bitmask_permissions: number;
  is_active: boolean;
  allowed_companies: string[];
  created_at: string;
}

const SAMPLE_USERS: UserRow[] = [
  {
    id: 'usr-01',
    username: 'admin',
    full_name: 'Tuhin (System Administrator & CFO)',
    email: 'admin@enterprise-erp.local',
    role: 'System Administrator (CFO)',
    is_superuser: true,
    bitmask_permissions: 15, // 1 | 2 | 4 | 8
    is_active: true,
    allowed_companies: ['Enterprise ERP Global Ltd', 'Subsidiary Tech SG'],
    created_at: '2026-05-10',
  },
  {
    id: 'usr-02',
    username: 'sales_officer',
    full_name: 'Rahim (Senior Sales Officer)',
    email: 'sales@enterprise-erp.local',
    role: 'Operational Sales Officer',
    is_superuser: false,
    bitmask_permissions: 3, // 1 (Read) | 2 (Write)
    is_active: true,
    allowed_companies: ['Enterprise ERP Global Ltd'],
    created_at: '2026-08-15',
  },
  {
    id: 'usr-03',
    username: 'auditor_visitor',
    full_name: 'External Financial Auditor',
    email: 'visitor@enterprise-erp.local',
    role: 'Financial Auditor',
    is_superuser: false,
    bitmask_permissions: 1, // 1 (Read-Only)
    is_active: true,
    allowed_companies: ['Enterprise ERP Global Ltd'],
    created_at: '2026-10-01',
  },
];

export const UserManagement: React.FC = () => {
  const [users, setUsers] = useState<UserRow[]>(SAMPLE_USERS);

  const handleToggleStatus = (user: UserRow) => {
    if (user.is_superuser) {
      message.error('Tryton Security Invariant: Superuser Administrator cannot be deactivated!');
      return;
    }

    setUsers((prev) =>
      prev.map((u) => (u.id === user.id ? { ...u, is_active: !u.is_active } : u))
    );

    message.success(
      `User ${user.username} is now ${!user.is_active ? 'Active' : 'Deactivated'}.`
    );
  };

  const handleResetPassword = (user: UserRow) => {
    Modal.confirm({
      title: `Reset Password for ${user.username}?`,
      content: `A secure Argon2id/bcrypt reset token will be dispatched to ${user.email}.`,
      okText: 'Send Reset Link',
      onOk: () => {
        message.success(`Password reset link dispatched to ${user.email}`);
      },
    });
  };

  const columns = [
    {
      title: 'User Profile',
      key: 'profile',
      render: (_: any, record: UserRow) => (
        <Space>
          <Avatar
            style={{
              backgroundColor: record.is_superuser ? '#2563eb' : '#334155',
            }}
            icon={record.is_superuser ? <CrownOutlined /> : <UserOutlined />}
          />
          <div>
            <div style={{ fontWeight: 600, color: '#f8fafc' }}>{record.full_name}</div>
            <div style={{ fontSize: '12px', color: '#64748b', fontFamily: 'monospace' }}>
              @{record.username} • {record.email}
            </div>
          </div>
        </Space>
      ),
    },
    {
      title: 'Assigned Role',
      dataIndex: 'role',
      key: 'role',
      render: (role: string, record: UserRow) => (
        <Space direction="vertical" size={2}>
          <Tag color={record.is_superuser ? 'purple' : 'blue'}>
            {record.is_superuser ? '👑 ' + role : role}
          </Tag>
          {record.is_superuser && (
            <span style={{ fontSize: '10px', color: '#a855f7' }}>Tryton Administrator</span>
          )}
        </Space>
      ),
    },
    {
      title: 'Effective Bitmask Permissions',
      dataIndex: 'bitmask_permissions',
      key: 'bitmask_permissions',
      render: (mask: number, record: UserRow) => (
        <Space size={4}>
          <Tag color={record.is_superuser || (mask & 1) ? 'green' : 'default'}>Read (1)</Tag>
          <Tag color={record.is_superuser || (mask & 2) ? 'blue' : 'default'}>Write (2)</Tag>
          <Tag color={record.is_superuser || (mask & 4) ? 'orange' : 'default'}>Delete (4)</Tag>
          <Tag color={record.is_superuser || (mask & 8) ? 'red' : 'default'}>Post (8)</Tag>
        </Space>
      ),
    },
    {
      title: 'Multi-Company Access',
      dataIndex: 'allowed_companies',
      key: 'allowed_companies',
      render: (companies: string[]) => (
        <Space wrap size={[0, 4]}>
          {companies.map((c) => (
            <Tag key={c} color="cyan">
              {c}
            </Tag>
          ))}
        </Space>
      ),
    },
    {
      title: 'Status',
      key: 'status',
      render: (_: any, record: UserRow) => (
        <Space>
          <Switch
            checked={record.is_active}
            onChange={() => handleToggleStatus(record)}
            checkedChildren={<CheckCircleOutlined />}
            unCheckedChildren={<StopOutlined />}
          />
          <span style={{ fontSize: '12px', color: record.is_active ? '#10b981' : '#ef4444' }}>
            {record.is_active ? 'Active' : 'Locked'}
          </span>
        </Space>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: UserRow) => (
        <Space size="small">
          <Tooltip title="Reset Password Token">
            <Button
              size="small"
              icon={<KeyOutlined />}
              onClick={() => handleResetPassword(record)}
            />
          </Tooltip>
          <Button size="small" icon={<SafetyCertificateOutlined />}>
            Roles
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <List
      title="Core Identity: User & Users Management (Tryton res.user Pattern)"
      headerButtons={({ defaultButtons }) => (
        <Space>
          <Tag color="purple">GAAP/IFRS Audit Trail Active</Tag>
          <Button type="primary" icon={<PlusOutlined />} style={{ backgroundColor: '#2563eb' }}>
            New User
          </Button>
        </Space>
      )}
    >
      <Table
        dataSource={users}
        columns={columns}
        rowKey="id"
        pagination={{ pageSize: 10 }}
        size="middle"
      />
    </List>
  );
};
