import React, { useState } from 'react';
import { List } from '@refinedev/antd';
import { Table, Tag, Space, Button, Radio, Typography, Avatar, Tooltip } from 'antd';
import {
  UserOutlined,
  ShopOutlined,
  BankOutlined,
  MailOutlined,
  PhoneOutlined,
  EnvironmentOutlined,
  SafetyCertificateOutlined,
  PlusOutlined,
  CheckCircleOutlined,
  EditOutlined,
} from '@ant-design/icons';
import { PartyDrawer } from './PartyDrawer';

const { Text } = Typography;

export interface PartyRow {
  id: string;
  code: string;
  name: string;
  is_customer: boolean;
  is_supplier: boolean;
  is_company: boolean;
  identifiers: { type: string; code: string; is_primary: boolean }[];
  address: string;
  contact: string;
  is_active: boolean;
}

const SAMPLE_PARTIES: PartyRow[] = [
  {
    id: '00000000-0000-0000-0000-000000000101',
    code: 'PRTY-CUST-001',
    name: 'Acme Holdings Global Ltd',
    is_customer: true,
    is_supplier: false,
    is_company: false,
    identifiers: [{ type: 'BIN', code: '002918274-0101', is_primary: true }],
    address: 'Gulshan Avenue, Road 11, Dhaka, Bangladesh',
    contact: 'billing@acmeholdings.com • +880 1711 000000',
    is_active: true,
  },
  {
    id: '00000000-0000-0000-0000-000000000102',
    code: 'PRTY-SUPP-001',
    name: 'Silicon Valley Chipsets Inc',
    is_customer: false,
    is_supplier: true,
    is_company: false,
    identifiers: [{ type: 'VAT', code: 'US-948271049', is_primary: true }],
    address: '440 Innovation Way, Suite 800, San Jose, USA',
    contact: 'orders@svchipsets.com',
    is_active: true,
  },
  {
    id: '00000000-0000-0000-0000-000000000103',
    code: 'PRTY-CORP-001',
    name: 'Enterprise ERP Global Ltd (HQ)',
    is_customer: false,
    is_supplier: false,
    is_company: true,
    identifiers: [{ type: 'TIN', code: '774829104918', is_primary: true }],
    address: 'Tejgaon I/A, Dhaka, Bangladesh',
    contact: 'contact@enterprise-erp.local',
    is_active: true,
  },
  {
    id: '00000000-0000-0000-0000-000000000104',
    code: 'PRTY-SUPP-002',
    name: 'Shenzhen Assemblies Ltd',
    is_customer: false,
    is_supplier: true,
    is_company: false,
    identifiers: [{ type: 'VAT', code: 'CN-849201948', is_primary: true }],
    address: 'Nanshan High-Tech Park, Shenzhen, China',
    contact: 'sales@sz-assemblies.cn',
    is_active: true,
  },
];

export const PartyList: React.FC = () => {
  const [parties, setParties] = useState<PartyRow[]>(SAMPLE_PARTIES);
  const [filterRole, setFilterRole] = useState<string>('all');
  const [drawerOpen, setDrawerOpen] = useState<boolean>(false);
  const [selectedParty, setSelectedParty] = useState<PartyRow | null>(null);

  const filteredData = parties.filter((p) => {
    if (filterRole === 'customers') return p.is_customer;
    if (filterRole === 'suppliers') return p.is_supplier;
    if (filterRole === 'companies') return p.is_company;
    return true;
  });

  const handlePartySaved = (saved: PartyRow) => {
    setParties((prev) => {
      const existsIndex = prev.findIndex((p) => p.id === saved.id || p.code === saved.code);
      if (existsIndex >= 0) {
        const copy = [...prev];
        copy[existsIndex] = saved;
        return copy;
      }
      return [saved, ...prev];
    });
  };

  const columns = [
    {
      title: 'Party Code & Name',
      key: 'name',
      render: (_: any, record: PartyRow) => (
        <Space
          style={{ cursor: 'pointer' }}
          onClick={() => {
            setSelectedParty(record);
            setDrawerOpen(true);
          }}
        >
          <Avatar
            style={{
              backgroundColor: record.is_company ? '#2563eb' : record.is_customer ? '#10b981' : '#f59e0b',
            }}
            icon={record.is_company ? <BankOutlined /> : record.is_customer ? <UserOutlined /> : <ShopOutlined />}
          />
          <div>
            <div style={{ fontWeight: 600, color: '#f8fafc' }}>{record.name}</div>
            <div style={{ fontSize: '12px', color: '#38bdf8', fontFamily: 'monospace' }}>
              {record.code}
            </div>
          </div>
        </Space>
      ),
    },
    {
      title: 'Classification',
      key: 'roles',
      render: (_: any, record: PartyRow) => (
        <Space size={4}>
          {record.is_customer && <Tag color="green">Customer</Tag>}
          {record.is_supplier && <Tag color="orange">Supplier</Tag>}
          {record.is_company && <Tag color="purple">Company Entity</Tag>}
        </Space>
      ),
    },
    {
      title: 'Legal Tax Identifiers',
      dataIndex: 'identifiers',
      key: 'identifiers',
      render: (ids: PartyRow['identifiers']) => (
        <Space wrap size={[0, 4]}>
          {ids.map((id, idx) => (
            <Tag key={idx} color="cyan" icon={<SafetyCertificateOutlined />}>
              {id.type}: {id.code}
            </Tag>
          ))}
        </Space>
      ),
    },
    {
      title: 'Primary Address',
      dataIndex: 'address',
      key: 'address',
      render: (addr: string) => (
        <span style={{ fontSize: '12px', color: '#94a3b8' }}>
          <EnvironmentOutlined style={{ marginRight: '4px' }} />
          {addr}
        </span>
      ),
    },
    {
      title: 'Contact Channels',
      dataIndex: 'contact',
      key: 'contact',
      render: (c: string) => (
        <span style={{ fontSize: '12px', color: '#cbd5e1' }}>{c}</span>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (active: boolean) => (
        <Tag color={active ? 'success' : 'default'} icon={<CheckCircleOutlined />}>
          {active ? 'Active' : 'Inactive'}
        </Tag>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      align: 'right' as const,
      render: (_: any, record: PartyRow) => (
        <Button
          size="small"
          icon={<EditOutlined />}
          onClick={() => {
            setSelectedParty(record);
            setDrawerOpen(true);
          }}
          style={{ backgroundColor: '#1e293b', borderColor: '#334155', color: '#38bdf8' }}
        >
          View / Edit
        </Button>
      ),
    },
  ];

  return (
    <List
      title="Universal Parties (Customers, Suppliers & Legal Entities - Tryton party.party)"
      headerButtons={() => (
        <Space>
          <Radio.Group
            value={filterRole}
            onChange={(e) => setFilterRole(e.target.value)}
            optionType="button"
            buttonStyle="solid"
          >
            <Radio.Button value="all">All Parties</Radio.Button>
            <Radio.Button value="customers">Customers</Radio.Button>
            <Radio.Button value="suppliers">Suppliers</Radio.Button>
            <Radio.Button value="companies">Company Entities</Radio.Button>
          </Radio.Group>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              setSelectedParty(null);
              setDrawerOpen(true);
            }}
            style={{ backgroundColor: '#2563eb' }}
          >
            New Party
          </Button>
        </Space>
      )}
    >
      <Table
        dataSource={filteredData}
        columns={columns}
        rowKey="id"
        pagination={{ pageSize: 10 }}
        size="middle"
      />

      <PartyDrawer
        isOpen={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        party={selectedParty}
        onSaved={handlePartySaved}
      />
    </List>
  );
};
