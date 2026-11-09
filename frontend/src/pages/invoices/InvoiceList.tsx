import React, { useState } from 'react';
import { List } from '@refinedev/antd';
import { Table, Tag, Space, Button, message, Tooltip } from 'antd';
import {
  PlusOutlined,
  SendOutlined,
  FileSearchOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';

export interface SalesInvoiceRow {
  id: string;
  invoice_no: string;
  customer_name: string;
  posting_date: string;
  currency: string;
  grand_total: string;
  outstanding_amount: string;
  docstatus: 0 | 1 | 2; // 0=Draft, 1=Submitted, 2=Cancelled
}

const SAMPLE_INVOICES: SalesInvoiceRow[] = [
  {
    id: 'inv-440',
    invoice_no: 'SINV-2026-0042',
    customer_name: 'Global Tech Corp',
    posting_date: '2026-11-08',
    currency: 'USD',
    grand_total: '12500.000000',
    outstanding_amount: '0.000000',
    docstatus: 1, // Submitted/Posted to GL
  },
  {
    id: 'inv-441',
    invoice_no: 'SINV-2026-0043',
    customer_name: 'Acme Holdings Ltd',
    posting_date: '2026-11-09',
    currency: 'BDT',
    grand_total: '450000.000000',
    outstanding_amount: '450000.000000',
    docstatus: 0, // Draft (Ready for Human Review & Submit)
  },
  {
    id: 'inv-442',
    invoice_no: 'SINV-2026-0044',
    customer_name: 'Apex Innovations',
    posting_date: '2026-11-09',
    currency: 'USD',
    grand_total: '3450.000000',
    outstanding_amount: '3450.000000',
    docstatus: 0, // Draft scanned by Gemini IDP
  },
];

interface InvoiceListProps {
  onOpenCreate?: () => void;
}

export const InvoiceList: React.FC<InvoiceListProps> = ({ onOpenCreate }) => {
  const [invoices, setInvoices] = useState<SalesInvoiceRow[]>(SAMPLE_INVOICES);

  const handleSubmitToLedger = (id: string, invoiceNo: string) => {
    setInvoices((prev) =>
      prev.map((inv) =>
        inv.id === id ? { ...inv, docstatus: 1 } : inv
      )
    );
    message.success(`Invoice ${invoiceNo} posted to General Ledger! GL entries created with Round-off Guard.`);
  };

  const columns = [
    {
      title: 'Invoice No',
      dataIndex: 'invoice_no',
      key: 'invoice_no',
      render: (text: string) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 600, color: '#38bdf8' }}>{text}</span>
      ),
    },
    {
      title: 'Customer',
      dataIndex: 'customer_name',
      key: 'customer_name',
    },
    {
      title: 'Posting Date',
      dataIndex: 'posting_date',
      key: 'posting_date',
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
      render: (amount: string, record: SalesInvoiceRow) => (
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
              Submitted (Posted)
            </Tag>
          );
        }
        if (status === 2) {
          return (
            <Tag color="error" icon={<CloseCircleOutlined />}>
              Cancelled
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
      render: (_: any, record: SalesInvoiceRow) => (
        <Space size="small">
          {record.docstatus === 0 ? (
            <Button
              type="primary"
              size="small"
              icon={<SendOutlined />}
              style={{ backgroundColor: '#10b981', borderColor: '#10b981' }}
              onClick={() => handleSubmitToLedger(record.id, record.invoice_no)}
            >
              Post to GL
            </Button>
          ) : (
            <Tooltip title="Strict GAAP/IFRS: Posted entries are immutable.">
              <Button size="small" icon={<FileSearchOutlined />} disabled>
                Posted
              </Button>
            </Tooltip>
          )}
        </Space>
      ),
    },
  ];

  return (
    <List
      title="Sales Invoices (ERPNext DocType + Tryton Strict GL Posting)"
      headerButtons={({ defaultButtons }) => (
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={onOpenCreate}
          style={{ backgroundColor: '#2563eb' }}
        >
          New Sales Invoice
        </Button>
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
