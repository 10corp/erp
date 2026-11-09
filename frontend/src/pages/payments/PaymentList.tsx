import React from 'react';
import { List } from '@refinedev/antd';
import { Table, Tag, Space, Button } from 'antd';
import { PlusOutlined, DollarOutlined, ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons';

export interface PaymentEntryRow {
  id: string;
  payment_no: string;
  party_name: string;
  posting_date: string;
  currency: string;
  paid_amount: string;
  exchange_rate: string;
  realized_forex_gain_loss: string;
  docstatus: 0 | 1 | 2;
}

const SAMPLE_PAYMENTS: PaymentEntryRow[] = [
  {
    id: 'pay-01',
    payment_no: 'PAY-20261109-001',
    party_name: 'Global Tech Corp',
    posting_date: '2026-11-09',
    currency: 'USD',
    paid_amount: '12500.000000',
    exchange_rate: '119.500000', // Invoice rate was 117.85 -> Forex Gain!
    realized_forex_gain_loss: '+20625.000000',
    docstatus: 1,
  },
  {
    id: 'pay-02',
    payment_no: 'PAY-20261109-002',
    party_name: 'Acme Holdings Ltd',
    posting_date: '2026-11-09',
    currency: 'BDT',
    paid_amount: '450000.000000',
    exchange_rate: '1.000000',
    realized_forex_gain_loss: '0.000000',
    docstatus: 1,
  },
];

export const PaymentList: React.FC = () => {
  const columns = [
    {
      title: 'Payment No',
      dataIndex: 'payment_no',
      key: 'payment_no',
      render: (text: string) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 600, color: '#38bdf8' }}>{text}</span>
      ),
    },
    {
      title: 'Customer / Party',
      dataIndex: 'party_name',
      key: 'party_name',
    },
    {
      title: 'Posting Date',
      dataIndex: 'posting_date',
      key: 'posting_date',
    },
    {
      title: 'Paid Amount',
      dataIndex: 'paid_amount',
      key: 'paid_amount',
      render: (amt: string, record: PaymentEntryRow) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 600 }}>
          {record.currency === 'USD' ? '$' : '৳'} {parseFloat(amt).toLocaleString()}
        </span>
      ),
    },
    {
      title: 'Exchange Rate',
      dataIndex: 'exchange_rate',
      key: 'exchange_rate',
      render: (rate: string) => <Tag color="blue">{parseFloat(rate).toFixed(4)}</Tag>,
    },
    {
      title: 'Realized Forex Gain/Loss',
      dataIndex: 'realized_forex_gain_loss',
      key: 'realized_forex_gain_loss',
      render: (val: string) => {
        const num = parseFloat(val);
        if (num > 0) {
          return (
            <Tag color="green" icon={<ArrowUpOutlined />}>
              +৳{num.toLocaleString()} (Account 4200)
            </Tag>
          );
        }
        if (num < 0) {
          return (
            <Tag color="red" icon={<ArrowDownOutlined />}>
              -৳{Math.abs(num).toLocaleString()} (Account 5200)
            </Tag>
          );
        }
        return <Tag>৳0.00</Tag>;
      },
    },
    {
      title: 'Status',
      dataIndex: 'docstatus',
      key: 'docstatus',
      render: () => <Tag color="success">Submitted</Tag>,
    },
  ];

  return (
    <List
      title="Payment Receipts & Forex Reconciliation (Multi-Currency)"
      headerButtons={({ defaultButtons }) => (
        <Button type="primary" icon={<PlusOutlined />} style={{ backgroundColor: '#2563eb' }}>
          New Payment Entry
        </Button>
      )}
    >
      <Table
        dataSource={SAMPLE_PAYMENTS}
        columns={columns}
        rowKey="id"
        pagination={false}
        size="middle"
      />
    </List>
  );
};
