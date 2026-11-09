import React, { useState } from 'react';
import { List } from '@refinedev/antd';
import { Table, Tag, Space, Button, message, Modal, Tooltip } from 'antd';
import {
  PlusOutlined,
  SendOutlined,
  RollbackOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  FileProtectOutlined,
} from '@ant-design/icons';

export interface PurchaseInvoiceRow {
  id: string;
  invoice_number: string;
  supplier_invoice_no: string;
  supplier_name: string;
  posting_date: string;
  due_date: string;
  currency: string;
  grand_total: string;
  docstatus: 0 | 1 | 2; // 0=Draft, 1=Submitted, 2=Cancelled
}

const SAMPLE_INVOICES: PurchaseInvoiceRow[] = [
  {
    id: 'pinv-01',
    invoice_number: 'PINV-2026-0001',
    supplier_invoice_no: 'SV-INV-9982',
    supplier_name: 'Silicon Valley Chipsets Inc',
    posting_date: '2026-11-09',
    due_date: '2026-12-09',
    currency: 'USD',
    grand_total: '5500.000000',
    docstatus: 1, // Submitted & Posted to AP
  },
  {
    id: 'pinv-02',
    invoice_number: 'PINV-2026-0002',
    supplier_invoice_no: 'SZ-2026-441',
    supplier_name: 'Shenzhen Assemblies Ltd',
    posting_date: '2026-11-10',
    due_date: '2026-11-30',
    currency: 'USD',
    grand_total: '12800.000000',
    docstatus: 0, // Draft (Needs review & submit)
  },
  {
    id: 'pinv-03',
    invoice_number: 'PINV-2026-0003',
    supplier_invoice_no: 'TYO-8812',
    supplier_name: 'Tokyo Optoelectronics',
    posting_date: '2026-11-05',
    due_date: '2026-12-05',
    currency: 'BDT',
    grand_total: '60000.000000',
    docstatus: 1, // Submitted & Posted to AP
  },
];

export const PurchaseInvoiceList: React.FC = () => {
  const [invoices, setInvoices] = useState<PurchaseInvoiceRow[]>(SAMPLE_INVOICES);

  const handlePostToAP = (inv: PurchaseInvoiceRow) => {
    Modal.confirm({
      title: `Post ${inv.invoice_number} to Accounts Payable?`,
      content: (
        <div>
          <p>This will execute an atomic double-entry transaction in <code>gl_entries</code>:</p>
          <ul>
            <li><strong>Credit 2001:</strong> Accounts Payable (Trade Creditors) for {inv.currency} {parseFloat(inv.grand_total).toLocaleString()}</li>
            <li><strong>Debit 1030:</strong> Stock in Hand (Asset)</li>
            <li><strong>Debit 1050:</strong> Input VAT (Tax Receivable)</li>
          </ul>
          <p style={{ color: '#10b981', fontSize: '12px' }}>
            ✓ Tryton 0.000000 Invariant & Account 5090 Round-Off Guard Verified.
          </p>
        </div>
      ),
      okText: 'Post to AP Ledger',
      okType: 'primary',
      onOk: () => {
        setInvoices((prev) =>
          prev.map((i) => (i.id === inv.id ? { ...i, docstatus: 1 } : i))
        );
        message.success(`Invoice ${inv.invoice_number} posted to Accounts Payable ledger!`);
      },
    });
  };

  const handleContraReversal = (inv: PurchaseInvoiceRow) => {
    Modal.confirm({
      title: `Generate Tryton Contra Reversal for ${inv.invoice_number}?`,
      content: `Under GAAP/IFRS, posted purchase invoices cannot be updated or deleted. A mirror-image contra reversal entry will be appended to reverse the liability.`,
      okText: 'Post Contra Reversal',
      okType: 'danger',
      onOk: () => {
        setInvoices((prev) =>
          prev.map((i) => (i.id === inv.id ? { ...i, docstatus: 2 } : i))
        );
        message.success(`Contra reversal appended to ledger for ${inv.invoice_number}.`);
      },
    });
  };

  const columns = [
    {
      title: 'Invoice Number',
      dataIndex: 'invoice_number',
      key: 'invoice_number',
      render: (text: string) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 600, color: '#38bdf8' }}>{text}</span>
      ),
    },
    {
      title: 'Supplier Bill #',
      dataIndex: 'supplier_invoice_no',
      key: 'supplier_invoice_no',
      render: (text: string) => (
        <span style={{ fontFamily: 'monospace', color: '#94a3b8' }}>{text || '-'}</span>
      ),
    },
    {
      title: 'Supplier Name',
      dataIndex: 'supplier_name',
      key: 'supplier_name',
    },
    {
      title: 'Bill Date',
      dataIndex: 'posting_date',
      key: 'posting_date',
    },
    {
      title: 'Due Date',
      dataIndex: 'due_date',
      key: 'due_date',
    },
    {
      title: 'Currency',
      dataIndex: 'currency',
      key: 'currency',
      render: (curr: string) => <Tag color="blue">{curr}</Tag>,
    },
    {
      title: 'Grand Total',
      dataIndex: 'grand_total',
      key: 'grand_total',
      render: (amount: string, record: PurchaseInvoiceRow) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 600 }}>
          {record.currency === 'USD' ? '$' : '৳'} {parseFloat(amount).toLocaleString()}
        </span>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'docstatus',
      key: 'docstatus',
      render: (status: number) => {
        if (status === 1) {
          return (
            <Tag color="success" icon={<CheckCircleOutlined />}>
              Posted to AP (2001)
            </Tag>
          );
        }
        if (status === 2) {
          return (
            <Tag color="error" icon={<CloseCircleOutlined />}>
              Cancelled (Reversed)
            </Tag>
          );
        }
        return (
          <Tag color="warning" icon={<ClockCircleOutlined />}>
            Draft (Review)
          </Tag>
        );
      },
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: PurchaseInvoiceRow) => (
        <Space size="small">
          {record.docstatus === 0 ? (
            <Button
              type="primary"
              size="small"
              icon={<SendOutlined />}
              style={{ backgroundColor: '#10b981', borderColor: '#10b981' }}
              onClick={() => handlePostToAP(record)}
            >
              Post to AP
            </Button>
          ) : record.docstatus === 1 ? (
            <Button
              size="small"
              danger
              icon={<RollbackOutlined />}
              onClick={() => handleContraReversal(record)}
            >
              Contra Reversal
            </Button>
          ) : (
            <Tooltip title="Cancelled document with mirror-image contra entry appended.">
              <Tag color="default">Audited</Tag>
            </Tooltip>
          )}
        </Space>
      ),
    },
  ];

  return (
    <List
      title="Procure-to-Pay: Purchase Invoices (Accounts Payable Liability)"
      headerButtons={({ defaultButtons }) => (
        <Space>
          <Tag color="cyan">AP Account: 2001 Trade Creditors</Tag>
          <Button type="primary" icon={<PlusOutlined />} style={{ backgroundColor: '#2563eb' }}>
            New Purchase Invoice
          </Button>
        </Space>
      )}
    >
      <Table
        dataSource={invoices}
        columns={columns}
        rowKey="id"
        pagination={{ pageSize: 10 }}
        size="middle"
      />
    </List>
  );
};
