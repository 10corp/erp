import React, { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Outlet, Navigate } from 'react-router-dom';
import { Refine, Authenticated, useGetIdentity, useLogout, useInvalidate } from '@refinedev/core';
import { ThemedLayoutV2, useNotificationProvider } from '@refinedev/antd';
import routerProvider, { CatchAllNavigate } from '@refinedev/react-router-v6';
import { ConfigProvider, theme as antdTheme, Button, Space, Tag, notification, Dropdown, MenuProps } from 'antd';
import {
  DashboardOutlined,
  AuditOutlined,
  FileTextOutlined,
  InboxOutlined,
  DollarOutlined,
  ThunderboltOutlined,
  SearchOutlined,
  RocketOutlined,
  SyncOutlined,
  CheckCircleOutlined,
  UserOutlined,
  LogoutOutlined,
  SwapOutlined,
  CrownOutlined,
  EyeOutlined,
  ShoppingOutlined,
  FileProtectOutlined,
  TeamOutlined,
  ContactsOutlined,
  ApartmentOutlined,
} from '@ant-design/icons';

import '@refinedev/antd/dist/reset.css';

import { erpDataProvider } from './providers/dataProvider';
import { erpAuthProvider, ERPUser } from './providers/authProvider';
import { LoginPage } from './pages/auth/LoginPage';
import { AdminDashboard } from './pages/admin/Dashboard';
import { GeneralLedgerPage } from './pages/admin/GeneralLedger';
import { TrialBalancePage } from './pages/admin/TrialBalance';
import { UserManagement } from './pages/admin/UserManagement';
import { InvoiceList } from './pages/invoices/InvoiceList';
import { StockBalanceList } from './pages/inventory/StockBalanceList';
import { PaymentList } from './pages/payments/PaymentList';
import { PartyList } from './pages/parties/PartyList';
import { PurchaseOrderList } from './pages/procurement/PurchaseOrderList';
import { PurchaseInvoiceList } from './pages/procurement/PurchaseInvoiceList';
import { HRPayrollPage } from './pages/hr/HRPayrollPage';
import { SlideDrawer } from './components/SlideDrawer';
import { CommandPalette } from './components/CommandPalette';

// Custom Enterprise Header with Active Persona, Switch Role & SSE Heartbeat
const CustomHeader: React.FC<{
  onOpenPalette: () => void;
  onOpenDrawer: () => void;
  sseConnected: boolean;
}> = ({ onOpenPalette, onOpenDrawer, sseConnected }) => {
  const { data: user } = useGetIdentity<ERPUser>();
  const { mutate: logout } = useLogout();
  const invalidate = useInvalidate();
  const [activeBranch, setActiveBranch] = useState('Enterprise ERP Global Ltd (HQ)');

  const branchMenuItems: MenuProps['items'] = [
    {
      key: 'hq',
      label: 'Enterprise ERP Global Ltd (HQ Holding)',
      icon: <ApartmentOutlined style={{ color: '#38bdf8' }} />,
      onClick: () => {
        setActiveBranch('Enterprise ERP Global Ltd (HQ)');
        invalidate({ invalidates: ['all'] });
        notification.success({ message: 'Active Branch Switched', description: 'Switched session to Global HQ (Cache Invalidated)' });
      },
    },
    {
      key: 'dhk',
      label: 'Dhaka Principal Branch',
      icon: <ApartmentOutlined style={{ color: '#10b981' }} />,
      onClick: () => {
        setActiveBranch('Dhaka Principal Branch');
        invalidate({ invalidates: ['all'] });
        notification.success({ message: 'Active Branch Switched', description: 'Switched session to Dhaka Branch (Cache Invalidated)' });
      },
    },
    {
      key: 'ctg',
      label: 'Chittagong Port Distribution Branch',
      icon: <ApartmentOutlined style={{ color: '#f59e0b' }} />,
      onClick: () => {
        setActiveBranch('Chittagong Port Distribution Branch');
        invalidate({ invalidates: ['all'] });
        notification.success({ message: 'Active Branch Switched', description: 'Switched session to Chittagong Port Branch (Cache Invalidated)' });
      },
    },
  ];

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'persona-info',
      disabled: true,
      label: (
        <div style={{ padding: '4px 0' }}>
          <div style={{ fontWeight: 600, color: '#f8fafc' }}>{user?.name}</div>
          <div style={{ fontSize: '11px', color: '#94a3b8' }}>{user?.description}</div>
        </div>
      ),
    },
    { type: 'divider' },
    {
      key: 'switch-persona',
      icon: <SwapOutlined />,
      label: 'Switch Persona / Role',
      onClick: () => logout(),
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      danger: true,
      label: 'Sign Out to Login Panel',
      onClick: () => logout(),
    },
  ];

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '0 24px',
        height: '56px',
        backgroundColor: '#0f172a',
        borderBottom: '1px solid #1e293b',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
        <div
          onClick={onOpenPalette}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            backgroundColor: '#1e293b',
            border: '1px solid #334155',
            borderRadius: '6px',
            padding: '6px 14px',
            color: '#94a3b8',
            fontSize: '12px',
            cursor: 'pointer',
            transition: 'border-color 0.2s',
          }}
        >
          <SearchOutlined style={{ color: '#38bdf8' }} />
          <span>Ask Claude 3.5 Sonnet Action Brain...</span>
          <Tag color="#0f172a" style={{ border: '1px solid #475569', color: '#94a3b8', margin: 0, fontSize: '11px' }}>
            ⌘K
          </Tag>
        </div>
      </div>

      <Space size="middle">
        <Dropdown menu={{ items: branchMenuItems }} placement="bottomLeft" arrow>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              backgroundColor: '#1e293b',
              border: '1px solid #334155',
              borderRadius: '6px',
              padding: '5px 10px',
              color: '#38bdf8',
              fontSize: '12px',
              cursor: 'pointer',
            }}
          >
            <ApartmentOutlined />
            <span style={{ fontWeight: 600 }}>{activeBranch}</span>
            <SwapOutlined style={{ fontSize: '10px', color: '#94a3b8' }} />
          </div>
        </Dropdown>

        <Button
          type="primary"
          icon={<ThunderboltOutlined />}
          onClick={onOpenDrawer}
          style={{ backgroundColor: '#2563eb', borderColor: '#2563eb' }}
        >
          Gemini 1.5 IDP Scan
        </Button>

        <Tag
          color={sseConnected ? 'success' : 'default'}
          icon={sseConnected ? <CheckCircleOutlined /> : <SyncOutlined spin />}
        >
          {sseConnected ? 'SSE Live (0-Lag)' : 'SSE Standby'}
        </Tag>

        <Dropdown menu={{ items: userMenuItems }} placement="bottomRight" arrow>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              color: '#f8fafc',
              fontSize: '13px',
              cursor: 'pointer',
              padding: '4px 8px',
              borderRadius: '6px',
              background: '#1e293b',
              border: '1px solid #334155',
            }}
          >
            {user?.role === 'admin' ? (
              <CrownOutlined style={{ color: '#38bdf8' }} />
            ) : user?.role === 'visitor' ? (
              <EyeOutlined style={{ color: '#fbbf24' }} />
            ) : (
              <UserOutlined style={{ color: '#10b981' }} />
            )}
            <span style={{ fontWeight: 600 }}>{user?.role_label || user?.name || 'User'}</span>
            <Tag color={user?.role === 'admin' ? 'purple' : user?.role === 'visitor' ? 'gold' : 'green'}>
              RBAC: {user?.permissions || 1}
            </Tag>
          </div>
        </Dropdown>
      </Space>
    </div>
  );
};

// Custom Brand Title
const CustomTitle: React.FC<{ collapsed: boolean }> = ({ collapsed }) => (
  <div
    style={{
      display: 'flex',
      alignItems: 'center',
      gap: '10px',
      padding: '16px',
      color: '#f8fafc',
      fontWeight: 700,
      fontSize: '16px',
      letterSpacing: '-0.3px',
    }}
  >
    <div
      style={{
        width: '28px',
        height: '28px',
        borderRadius: '6px',
        background: 'linear-gradient(135deg, #2563eb 0%, #38bdf8 100%)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <RocketOutlined style={{ color: '#fff', fontSize: '16px' }} />
    </div>
    {!collapsed && <span>ENTERPRISE ERP</span>}
  </div>
);

export default function App() {
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [paletteOpen, setPaletteOpen] = useState(false);
  const [sseConnected, setSseConnected] = useState(true);

  // Keyboard shortcut: Cmd+K / Ctrl+K opens Claude 3.5 Sonnet Action Brain
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setPaletteOpen((prev) => !prev);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  // Server-Sent Events (SSE) Listener
  useEffect(() => {
    let eventSource: EventSource | null = null;
    try {
      eventSource = new EventSource('/api/v1/events');
      eventSource.addEventListener('connected', () => setSseConnected(true));
      eventSource.addEventListener('heartbeat', () => setSseConnected(true));
      eventSource.onerror = () => setSseConnected(false);
    } catch (e) {
      setSseConnected(false);
    }
    return () => {
      if (eventSource) eventSource.close();
    };
  }, []);

  const handleAction = (action: string, payload?: any) => {
    if (action === 'toggle_palette') {
      setPaletteOpen(true);
    } else if (action === 'new_invoice' || action === 'new_journal' || action === 'new_payment') {
      setDrawerOpen(true);
    } else if (action === 'ai_action' && payload) {
      notification.info({
        message: 'Claude 3.5 Sonnet Action Brain',
        description: payload.ui_message || payload.description,
        placement: 'bottomRight',
      });
      if (payload.action === 'open_drawer') {
        setDrawerOpen(true);
      }
    }
  };

  const handlePostSuccess = (newVoucher?: any) => {
    notification.success({
      message: 'Transaction Committed',
      description: `Posted ${newVoucher?.voucher_no || 'Document'} to partitioned ledger.`,
      placement: 'bottomRight',
    });
    setDrawerOpen(false);
  };

  return (
    <ConfigProvider
      theme={{
        algorithm: antdTheme.darkAlgorithm,
        token: {
          colorPrimary: '#2563eb',
          colorBgBase: '#0b1329',
          colorBgContainer: '#0f172a',
          colorBorder: '#1e293b',
          colorText: '#f8fafc',
          borderRadius: 6,
          fontFamily: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
        },
      }}
    >
      <BrowserRouter>
        <Refine
          dataProvider={erpDataProvider}
          authProvider={erpAuthProvider}
          routerProvider={routerProvider}
          notificationProvider={useNotificationProvider}
          options={{
            syncWithLocation: true,
            warnWhenUnsavedChanges: true,
          }}
          resources={[
            {
              name: 'dashboard',
              list: '/admin/dashboard',
              meta: {
                label: 'CEO Analytics',
                icon: <DashboardOutlined />,
              },
            },
            {
              name: 'general-ledger',
              list: '/admin/general-ledger',
              meta: {
                label: 'General Ledger',
                icon: <AuditOutlined />,
              },
            },
            {
              name: 'trial-balance',
              list: '/admin/trial-balance',
              meta: {
                label: 'Trial Balance',
                icon: <AuditOutlined />,
              },
            },
            {
              name: 'sales-invoices',
              list: '/sales-invoices',
              meta: {
                label: 'Sales Invoices',
                icon: <FileTextOutlined />,
              },
            },
            {
              name: 'inventory',
              list: '/inventory/balances',
              meta: {
                label: 'Stock Balances',
                icon: <InboxOutlined />,
              },
            },
            {
              name: 'payments',
              list: '/payments',
              meta: {
                label: 'Payment Receipts',
                icon: <DollarOutlined />,
              },
            },
            {
              name: 'purchase-orders',
              list: '/procurement/orders',
              meta: {
                label: 'Purchase Orders',
                icon: <ShoppingOutlined />,
              },
            },
            {
              name: 'purchase-invoices',
              list: '/procurement/invoices',
              meta: {
                label: 'Purchase Invoices (AP)',
                icon: <FileProtectOutlined />,
              },
            },
            {
              name: 'users',
              list: '/admin/users',
              meta: {
                label: 'User Management',
                icon: <TeamOutlined />,
              },
            },
            {
              name: 'parties',
              list: '/parties',
              meta: {
                label: 'Parties (Entities)',
                icon: <ContactsOutlined />,
              },
            },
            {
              name: 'hr-payroll',
              list: '/hr/payroll',
              meta: {
                label: 'HR & Payroll',
                icon: <TeamOutlined />,
              },
            },
          ]}
        >
          <Routes>
            {/* Public Login Route */}
            <Route path="/login" element={<LoginPage />} />

            {/* Authenticated ERP Application Routes */}
            <Route
              element={
                <Authenticated key="authenticated-routes" fallback={<CatchAllNavigate to="/login" />}>
                  <ThemedLayoutV2
                    Header={() => (
                      <CustomHeader
                        onOpenPalette={() => setPaletteOpen(true)}
                        onOpenDrawer={() => setDrawerOpen(true)}
                        sseConnected={sseConnected}
                      />
                    )}
                    Title={CustomTitle}
                    initialSiderCollapsed={false}
                  >
                    <Outlet />
                  </ThemedLayoutV2>
                </Authenticated>
              }
            >
              <Route index element={<Navigate to="/admin/dashboard" replace />} />
              <Route path="/admin/dashboard" element={<AdminDashboard />} />
              <Route path="/admin/general-ledger" element={<GeneralLedgerPage />} />
              <Route path="/admin/trial-balance" element={<TrialBalancePage />} />
              <Route path="/admin/users" element={<UserManagement />} />
              <Route path="/parties" element={<PartyList />} />
              <Route
                path="/sales-invoices"
                element={<InvoiceList onOpenCreate={() => setDrawerOpen(true)} />}
              />
              <Route path="/inventory/balances" element={<StockBalanceList />} />
              <Route path="/payments" element={<PaymentList />} />
              <Route path="/procurement/orders" element={<PurchaseOrderList />} />
              <Route path="/procurement/invoices" element={<PurchaseInvoiceList />} />
              <Route path="/hr/payroll" element={<HRPayrollPage />} />
            </Route>

            {/* Fallback to login */}
            <Route path="*" element={<CatchAllNavigate to="/login" />} />
          </Routes>
        </Refine>

        {/* Global Slide Drawer for Gemini 1.5 Flash Multimodal IDP */}
        <SlideDrawer
          isOpen={drawerOpen}
          onClose={() => setDrawerOpen(false)}
          onSuccess={handlePostSuccess}
        />

        {/* Global Command+K Palette for Claude 3.5 Sonnet Action Brain */}
        <CommandPalette
          isOpen={paletteOpen}
          onClose={() => setPaletteOpen(false)}
          onSelectAction={handleAction}
        />
      </BrowserRouter>
    </ConfigProvider>
  );
}
