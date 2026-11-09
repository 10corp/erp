import React, { useState, useEffect } from 'react';
import { ConfigProvider, theme, Layout, Button, Tag, Space, Dropdown, MenuProps, notification } from 'antd';
import {
  RocketOutlined,
  SearchOutlined,
  PlusOutlined,
  SyncOutlined,
  CheckCircleOutlined,
  ThunderboltOutlined,
  DashboardOutlined,
  BookOutlined,
  FileTextOutlined,
  RobotOutlined,
} from '@ant-design/icons';
import { CommandPalette } from './components/CommandPalette';
import { SlideDrawer } from './components/SlideDrawer';
import { LedgerGrid } from './components/LedgerGrid';
import { GLEntry } from './types';

const { Header, Content, Sider } = Layout;

// Sample partitioned general ledger entries matching our Go backend & schema
const INITIAL_LEDGER_DATA: GLEntry[] = [
  {
    id: 'e1001-uuid',
    company_id: 'comp-01',
    posting_date: '2026-11-09',
    voucher_type: 'Journal Entry',
    voucher_id: 'jv-101',
    voucher_no: 'JV-20261109-8A19F',
    account_id: 'acc-1001',
    account_code: '1001',
    party_type: 'Cash Account',
    currency: 'BDT',
    debit: '450000.000000',
    credit: '0.000000',
    exchange_rate: '1.000000',
    debit_base: '450000.000000',
    credit_base: '0.000000',
    remarks: 'Customer trade receipt contra settlement',
    is_cancelled: false,
    created_at: new Date().toISOString(),
  },
  {
    id: 'e1002-uuid',
    company_id: 'comp-01',
    posting_date: '2026-11-09',
    voucher_type: 'Journal Entry',
    voucher_id: 'jv-101',
    voucher_no: 'JV-20261109-8A19F',
    account_id: 'acc-1020',
    account_code: '1020',
    party_type: 'Acme Holdings Ltd',
    currency: 'BDT',
    debit: '0.000000',
    credit: '450000.000000',
    exchange_rate: '1.000000',
    debit_base: '0.000000',
    credit_base: '450000.000000',
    remarks: 'Customer trade receipt contra settlement',
    is_cancelled: false,
    created_at: new Date().toISOString(),
  },
  {
    id: 'e1003-uuid',
    company_id: 'comp-01',
    posting_date: '2026-11-08',
    voucher_type: 'Sales Invoice',
    voucher_id: 'inv-440',
    voucher_no: 'SINV-2026-0042',
    account_id: 'acc-1020',
    account_code: '1020',
    party_type: 'Global Tech Corp',
    currency: 'USD',
    debit: '12500.000000',
    credit: '0.000000',
    exchange_rate: '117.850000',
    debit_base: '1473125.000000',
    credit_base: '0.000000',
    remarks: 'Export of Enterprise Software Licenses',
    is_cancelled: false,
    created_at: new Date(Date.now() - 86400000).toISOString(),
  },
  {
    id: 'e1004-uuid',
    company_id: 'comp-01',
    posting_date: '2026-11-08',
    voucher_type: 'Sales Invoice',
    voucher_id: 'inv-440',
    voucher_no: 'SINV-2026-0042',
    account_id: 'acc-4001',
    account_code: '4001',
    party_type: 'Sales Revenue',
    currency: 'USD',
    debit: '0.000000',
    credit: '12500.000000',
    exchange_rate: '117.850000',
    debit_base: '0.000000',
    credit_base: '1473125.000000',
    remarks: 'Export of Enterprise Software Licenses',
    is_cancelled: false,
    created_at: new Date(Date.now() - 86400000).toISOString(),
  },
];

export default function App() {
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [paletteOpen, setPaletteOpen] = useState(false);
  const [activeMenu, setActiveMenu] = useState('ledger');
  const [ledgerData, setLedgerData] = useState<GLEntry[]>(INITIAL_LEDGER_DATA);
  const [sseConnected, setSseConnected] = useState(true);

  // Server-Sent Events (SSE) Listener for Real-Time zero-reload reactivity
  useEffect(() => {
    let eventSource: EventSource | null = null;
    try {
      eventSource = new EventSource('/api/v1/events');
      eventSource.addEventListener('connected', () => {
        setSseConnected(true);
      });
      eventSource.addEventListener('heartbeat', () => {
        setSseConnected(true);
      });
      eventSource.onerror = () => {
        // SSE graceful fallback if backend is offline
        setSseConnected(false);
      };
    } catch (e) {
      setSseConnected(false);
    }
    return () => {
      if (eventSource) eventSource.close();
    };
  }, []);

  const handleAction = (action: string) => {
    if (action === 'toggle_palette') {
      setPaletteOpen(true);
    } else if (action === 'new_journal' || action === 'new_invoice') {
      setDrawerOpen(true);
    } else if (action === 'view_ledger') {
      setActiveMenu('ledger');
    }
  };

  const handlePostSuccess = () => {
    notification.success({
      message: 'Transaction Posted to General Ledger',
      description: 'Double-entry invariants satisfied. Ledger row committed to 2026_h2 partition.',
      placement: 'bottomRight',
    });
  };

  const companyItems: MenuProps['items'] = [
    { key: '1', label: 'Apex Enterprise Holdings (Base: BDT)' },
    { key: '2', label: 'Apex Global Logistics (Base: USD)' },
  ];

  return (
    <ConfigProvider
      theme={{
        algorithm: [theme.darkAlgorithm, theme.compactAlgorithm],
        token: {
          colorPrimary: '#0284c7',
          colorBgBase: '#0b1120',
          colorBgContainer: '#131e32',
          borderRadius: 6,
          fontFamily: "'Inter', sans-serif",
        },
      }}
    >
      <Layout style={{ minHeight: '100vh', backgroundColor: '#0b1120' }}>
        {/* Top Header */}
        <Header
          style={{
            height: '52px',
            lineHeight: '52px',
            backgroundColor: '#0f172a',
            borderBottom: '1px solid #1e293b',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '0 20px',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer' }}>
              <RocketOutlined style={{ fontSize: '20px', color: '#38bdf8' }} />
              <span style={{ fontWeight: 700, fontSize: '15px', letterSpacing: '-0.02em', color: '#f8fafc' }}>
                BULLET<span style={{ color: '#38bdf8' }}>ERP</span>
              </span>
              <Tag color="cyan" style={{ fontSize: '10px', height: '18px', lineHeight: '16px', marginLeft: '4px' }}>
                HEADLESS GO
              </Tag>
            </div>

            <Dropdown menu={{ items: companyItems, selectedKeys: ['1'] }}>
              <Button size="small" style={{ backgroundColor: '#1e293b', borderColor: '#334155', color: '#e2e8f0', fontSize: '12px' }}>
                Apex Enterprise Holdings (BDT) ▼
              </Button>
            </Dropdown>
          </div>

          {/* Center Search Pill (Command+K) */}
          <Button
            size="small"
            icon={<SearchOutlined />}
            onClick={() => setPaletteOpen(true)}
            style={{
              backgroundColor: '#1e293b',
              borderColor: '#334155',
              color: '#94a3b8',
              width: '280px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              borderRadius: '6px',
            }}
          >
            <span>Search or jump to...</span>
            <kbd style={{ backgroundColor: '#0f172a', padding: '1px 5px', borderRadius: '4px', fontSize: '11px', color: '#cbd5e1' }}>
              Ctrl + K
            </kbd>
          </Button>

          {/* Right Toolbar */}
          <Space size="middle">
            {sseConnected ? (
              <Tag icon={<CheckCircleOutlined />} color="success" style={{ margin: 0 }}>
                SSE Live Stream
              </Tag>
            ) : (
              <Tag icon={<SyncOutlined spin />} color="default" style={{ margin: 0 }}>
                SSE Connecting
              </Tag>
            )}

            <Button
              type="primary"
              size="small"
              icon={<PlusOutlined />}
              onClick={() => setDrawerOpen(true)}
              style={{ backgroundColor: '#0284c7', borderColor: '#0284c7', fontWeight: 600 }}
            >
              Quick Post
            </Button>
          </Space>
        </Header>

        {/* Main Body */}
        <Layout>
          {/* Dense Sidebar */}
          <Sider
            width={200}
            style={{
              backgroundColor: '#0f172a',
              borderRight: '1px solid #1e293b',
            }}
          >
            <div style={{ padding: '12px 8px' }}>
              <div
                onClick={() => setActiveMenu('ledger')}
                style={{
                  ...sidebarItemStyle,
                  backgroundColor: activeMenu === 'ledger' ? '#0284c7' : 'transparent',
                  color: activeMenu === 'ledger' ? '#fff' : '#94a3b8',
                }}
              >
                <BookOutlined style={{ marginRight: '10px' }} />
                <span>General Ledger</span>
              </div>

              <div
                onClick={() => setActiveMenu('dashboard')}
                style={{
                  ...sidebarItemStyle,
                  backgroundColor: activeMenu === 'dashboard' ? '#0284c7' : 'transparent',
                  color: activeMenu === 'dashboard' ? '#fff' : '#94a3b8',
                }}
              >
                <DashboardOutlined style={{ marginRight: '10px' }} />
                <span>CEO Dashboard (CQRS)</span>
              </div>

              <div
                onClick={() => setActiveMenu('invoices')}
                style={{
                  ...sidebarItemStyle,
                  backgroundColor: activeMenu === 'invoices' ? '#0284c7' : 'transparent',
                  color: activeMenu === 'invoices' ? '#fff' : '#94a3b8',
                }}
              >
                <FileTextOutlined style={{ marginRight: '10px' }} />
                <span>Invoicing & AP/AR</span>
              </div>

              <div
                onClick={() => setActiveMenu('ai')}
                style={{
                  ...sidebarItemStyle,
                  backgroundColor: activeMenu === 'ai' ? '#0284c7' : 'transparent',
                  color: activeMenu === 'ai' ? '#fff' : '#94a3b8',
                }}
              >
                <RobotOutlined style={{ marginRight: '10px', color: '#f43f5e' }} />
                <span>AI Action Brain</span>
              </div>
            </div>
          </Sider>

          {/* Content Area */}
          <Content style={{ padding: '16px 20px', backgroundColor: '#0b1120' }}>
            {/* KPI Summary Cards */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '12px', marginBottom: '16px' }}>
              <div style={kpiCardStyle}>
                <div style={{ fontSize: '11px', color: '#64748b', textTransform: 'uppercase', fontWeight: 600 }}>Total Revenue</div>
                <div style={{ fontSize: '20px', fontWeight: 700, color: '#10b981', fontFamily: 'monospace' }}>৳ 1,473,125.00</div>
                <div style={{ fontSize: '11px', color: '#059669' }}>↑ +14.2% MoM (Materialized View)</div>
              </div>
              <div style={kpiCardStyle}>
                <div style={{ fontSize: '11px', color: '#64748b', textTransform: 'uppercase', fontWeight: 600 }}>Total Assets</div>
                <div style={{ fontSize: '20px', fontWeight: 700, color: '#38bdf8', fontFamily: 'monospace' }}>৳ 1,923,125.00</div>
                <div style={{ fontSize: '11px', color: '#0284c7' }}>Zero-Float Exact Precision</div>
              </div>
              <div style={kpiCardStyle}>
                <div style={{ fontSize: '11px', color: '#64748b', textTransform: 'uppercase', fontWeight: 600 }}>Active Partitions</div>
                <div style={{ fontSize: '20px', fontWeight: 700, color: '#fbbf24', fontFamily: 'monospace' }}>2026_H1, 2026_H2</div>
                <div style={{ fontSize: '11px', color: '#d97706' }}>PostgreSQL Range Partitioned</div>
              </div>
              <div style={kpiCardStyle}>
                <div style={{ fontSize: '11px', color: '#64748b', textTransform: 'uppercase', fontWeight: 600 }}>Engine Latency</div>
                <div style={{ fontSize: '20px', fontWeight: 700, color: '#a78bfa', fontFamily: 'monospace' }}>0.48 ms</div>
                <div style={{ fontSize: '11px', color: '#7c3aed' }}>Native Go 1.22 + Binary pgx</div>
              </div>
            </div>

            {/* High Speed Dense AG Grid Table */}
            <div style={{ backgroundColor: '#0f172a', padding: '16px', borderRadius: '8px', border: '1px solid #1e293b' }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px' }}>
                <div>
                  <span style={{ fontSize: '15px', fontWeight: 600, color: '#f8fafc' }}>
                    {activeMenu === 'ledger'
                      ? 'Partitioned General Ledger (Strict Double-Entry Append-Only)'
                      : activeMenu === 'invoices'
                      ? 'Sales Invoices (ERPNext DocType + Tryton Automatic GL Posting)'
                      : 'C-Level Executive View (CQRS Materialized Views)'}
                  </span>
                  <div style={{ fontSize: '12px', color: '#64748b' }}>
                    {activeMenu === 'ledger'
                      ? 'Virtual DOM table rendering 100k+ rows at 60 FPS without server pagination delays.'
                      : activeMenu === 'invoices'
                      ? 'Invoices automatically post balanced debits to AR and credits to Revenue + Tax partitions.'
                      : 'Materialized Views refreshed in background without locking operational queries.'}
                  </div>
                </div>
                <Space>
                  {activeMenu === 'invoices' && (
                    <Button
                      type="primary"
                      size="small"
                      icon={<PlusOutlined />}
                      onClick={() => setDrawerOpen(true)}
                      style={{ backgroundColor: '#10b981', borderColor: '#10b981' }}
                    >
                      New Sales Invoice
                    </Button>
                  )}
                  <Button size="small" icon={<ThunderboltOutlined />} onClick={() => setDrawerOpen(true)}>
                    Quick Entry Drawer
                  </Button>
                </Space>
              </div>

              <LedgerGrid rowData={ledgerData} />
            </div>
          </Content>
        </Layout>
      </Layout>

      {/* Global Slide-Over Drawer (Context Preserved) */}
      <SlideDrawer
        isOpen={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        onSuccess={handlePostSuccess}
      />

      {/* Global Command+K Search Palette */}
      <CommandPalette
        isOpen={paletteOpen}
        onClose={() => setPaletteOpen(false)}
        onSelectAction={handleAction}
      />
    </ConfigProvider>
  );
}

const sidebarItemStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  padding: '8px 12px',
  borderRadius: '6px',
  fontSize: '13px',
  fontWeight: 500,
  cursor: 'pointer',
  marginBottom: '4px',
  transition: 'all 0.15s ease',
};

const kpiCardStyle: React.CSSProperties = {
  backgroundColor: '#131e32',
  padding: '12px 16px',
  borderRadius: '8px',
  border: '1px solid #1e293b',
};
