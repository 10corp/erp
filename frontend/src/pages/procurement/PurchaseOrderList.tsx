import React, { useState } from 'react';
import { List } from '@refinedev/antd';
import { Table, Tag, Space, Button, message, Modal, InputNumber, Select } from 'antd';
import {
  PlusOutlined,
  InboxOutlined,
  FileDoneOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ShoppingOutlined,
} from '@ant-design/icons';

export interface PurchaseOrderRow {
  id: string;
  po_number: string;
  supplier_name: string;
  transaction_date: string;
  delivery_date: string;
  currency: string;
  grand_total: string;
  status: 'Draft' | 'To Receive and Bill' | 'Completed' | 'Cancelled';
  docstatus: 0 | 1 | 2;
  items_count: number;
}

const SAMPLE_ORDERS: PurchaseOrderRow[] = [
  {
    id: 'po-101',
    po_number: 'PO-2026-0045',
    supplier_name: 'Silicon Valley Chipsets Inc',
    transaction_date: '2026-11-09',
    delivery_date: '2026-11-15',
    currency: 'USD',
    grand_total: '5500.000000',
    status: 'To Receive and Bill',
    docstatus: 1,
    items_count: 10,
  },
  {
    id: 'po-102',
    po_number: 'PO-2026-0046',
    supplier_name: 'Shenzhen Assemblies Ltd',
    transaction_date: '2026-11-10',
    delivery_date: '2026-11-20',
    currency: 'USD',
    grand_total: '12800.000000',
    status: 'Draft',
    docstatus: 0,
    items_count: 50,
  },
  {
    id: 'po-103',
    po_number: 'PO-2026-0042',
    supplier_name: 'Tokyo Optoelectronics',
    transaction_date: '2026-11-05',
    delivery_date: '2026-11-09',
    currency: 'BDT',
    grand_total: '60000.000000',
    status: 'Completed',
    docstatus: 1,
    items_count: 50,
  },
];

export const PurchaseOrderList: React.FC = () => {
  const [orders, setOrders] = useState<PurchaseOrderRow[]>(SAMPLE_ORDERS);

  const handleReceiveGoods = (order: PurchaseOrderRow) => {
    Modal.confirm({
      title: `Receive Goods for ${order.po_number}?`,
      content: (
        <div>
          <p>
            This will record a <strong>Goods Receipt Note (GRN)</strong> moving goods from{' '}
            <code>Supplier Virtual Location</code> to <code>Main Central Warehouse</code>.
          </p>
          <p>
            Valuation will be updated instantly at <strong>$O(1)$ Moving Average Cost</strong>.
          </p>
        </div>
      ),
      okText: 'Receive Stock & Update Moving Avg',
      okType: 'primary',
      onOk: () => {
        setOrders((prev) =>
          prev.map((o) => (o.id === order.id ? { ...o, status: 'Completed' } : o))
        );
        message.success(
          `Stock received into Main Warehouse! Moving average cost updated for ${order.po_number}.`
        );
      },
    });
  };

  const columns = [
    {
      title: 'PO Number',
      dataIndex: 'po_number',
      key: 'po_number',
      render: (text: string) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 600, color: '#38bdf8' }}>{text}</span>
      ),
    },
    {
      title: 'Supplier',
      dataIndex: 'supplier_name',
      key: 'supplier_name',
    },
    {
      title: 'PO Date',
      dataIndex: 'transaction_date',
      key: 'transaction_date',
    },
    {
      title: 'Target Delivery',
      dataIndex: 'delivery_date',
      key: 'delivery_date',
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
      render: (amount: string, record: PurchaseOrderRow) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 600 }}>
          {record.currency === 'USD' ? '$' : '৳'} {parseFloat(amount).toLocaleString()}
        </span>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        if (status === 'Completed') {
          return (
            <Tag color="success" icon={<CheckCircleOutlined />}>
              Completed (Received)
            </Tag>
          );
        }
        if (status === 'To Receive and Bill') {
          return (
            <Tag color="processing" icon={<InboxOutlined />}>
              To Receive & Bill
            </Tag>
          );
        }
        if (status === 'Cancelled') {
          return <Tag color="error">Cancelled</Tag>;
        }
        return (
          <Tag color="warning" icon={<ClockCircleOutlined />}>
            Draft
          </Tag>
        );
      },
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: PurchaseOrderRow) => (
        <Space size="small">
          {record.status === 'To Receive and Bill' && (
            <Button
              type="primary"
              size="small"
              icon={<InboxOutlined />}
              style={{ backgroundColor: '#10b981', borderColor: '#10b981' }}
              onClick={() => handleReceiveGoods(record)}
            >
              Receive Stock
            </Button>
          )}
          {record.status === 'Draft' && (
            <Button
              type="primary"
              size="small"
              icon={<FileDoneOutlined />}
              style={{ backgroundColor: '#2563eb' }}
              onClick={() => {
                setOrders((prev) =>
                  prev.map((o) => (o.id === record.id ? { ...o, status: 'To Receive and Bill', docstatus: 1 } : o))
                );
                message.success(`Purchase Order ${record.po_number} submitted to Supplier.`);
              }}
            >
              Submit PO
            </Button>
          )}
          {record.status === 'Completed' && (
            <Tag color="green" icon={<CheckCircleOutlined />}>
              Stock Inwarded
            </Tag>
          )}
        </Space>
      ),
    },
  ];

  return (
    <List
      title="Procure-to-Pay: Purchase Orders (Procurement Lifecycle)"
      headerButtons={({ defaultButtons }) => (
        <Button type="primary" icon={<PlusOutlined />} style={{ backgroundColor: '#2563eb' }}>
          New Purchase Order
        </Button>
      )}
    >
      <Table
        dataSource={orders}
        columns={columns}
        rowKey="id"
        pagination={{ pageSize: 10 }}
        size="middle"
      />
    </List>
  );
};
