import React from 'react';
import { List } from '@refinedev/antd';
import { Table, Tag, Space, Button, Card, Row, Col, Statistic } from 'antd';
import {
  InboxOutlined,
  ThunderboltOutlined,
  SwapOutlined,
  HomeOutlined,
} from '@ant-design/icons';

export interface StockBalanceRow {
  id: string;
  item_code: string;
  item_name: string;
  warehouse_name: string;
  warehouse_type: 'Internal' | 'Supplier' | 'Customer' | 'Transit';
  current_qty: string;
  moving_avg_rate: string;
  total_valuation: string;
}

const SAMPLE_STOCK: StockBalanceRow[] = [
  {
    id: 'stk-01',
    item_code: 'ITEM-SRV-01',
    item_name: 'High Performance NVMe Server 1U',
    warehouse_name: 'Main Central Warehouse',
    warehouse_type: 'Internal',
    current_qty: '45.000000',
    moving_avg_rate: '1100.000000',
    total_valuation: '49500.000000',
  },
  {
    id: 'stk-02',
    item_code: 'ITEM-SW-02',
    item_name: 'Managed 10GbE Switch 24-Port',
    warehouse_name: 'Main Central Warehouse',
    warehouse_type: 'Internal',
    current_qty: '80.000000',
    moving_avg_rate: '450.000000',
    total_valuation: '36000.000000',
  },
  {
    id: 'stk-03',
    item_code: 'ITEM-SRV-01',
    item_name: 'High Performance NVMe Server 1U',
    warehouse_name: 'Supplier Virtual Location',
    warehouse_type: 'Supplier',
    current_qty: '120.000000',
    moving_avg_rate: '1000.000000',
    total_valuation: '120000.000000',
  },
  {
    id: 'stk-04',
    item_code: 'ITEM-SW-02',
    item_name: 'Managed 10GbE Switch 24-Port',
    warehouse_name: 'Transit Distribution Dock',
    warehouse_type: 'Transit',
    current_qty: '15.000000',
    moving_avg_rate: '450.000000',
    total_valuation: '6750.000000',
  },
];

export const StockBalanceList: React.FC = () => {
  const columns = [
    {
      title: 'Item Code',
      dataIndex: 'item_code',
      key: 'item_code',
      render: (text: string) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 600, color: '#38bdf8' }}>{text}</span>
      ),
    },
    {
      title: 'Item Name',
      dataIndex: 'item_name',
      key: 'item_name',
    },
    {
      title: 'Warehouse',
      dataIndex: 'warehouse_name',
      key: 'warehouse_name',
      render: (text: string, record: StockBalanceRow) => (
        <Space>
          <HomeOutlined style={{ color: '#64748b' }} />
          <span>{text}</span>
          <Tag color={record.warehouse_type === 'Internal' ? 'green' : 'default'}>
            {record.warehouse_type}
          </Tag>
        </Space>
      ),
    },
    {
      title: 'Quantity On Hand',
      dataIndex: 'current_qty',
      key: 'current_qty',
      render: (qty: string) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 600, color: '#10b981' }}>
          {parseFloat(qty).toFixed(2)}
        </span>
      ),
    },
    {
      title: 'Moving Avg Cost ($)',
      dataIndex: 'moving_avg_rate',
      key: 'moving_avg_rate',
      render: (rate: string) => (
        <span style={{ fontFamily: 'monospace' }}>${parseFloat(rate).toFixed(2)}</span>
      ),
    },
    {
      title: 'Valuation ($)',
      dataIndex: 'total_valuation',
      key: 'total_valuation',
      render: (val: string) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 700, color: '#fbbf24' }}>
          ${parseFloat(val).toLocaleString(undefined, { minimumFractionDigits: 2 })}
        </span>
      ),
    },
  ];

  return (
    <List
      title="Double-Entry Inventory & Stock Valuation (O(1) Moving Average)"
      headerButtons={({ defaultButtons }) => (
        <Space>
          <Button icon={<SwapOutlined />} style={{ backgroundColor: '#10b981', color: '#fff', borderColor: '#10b981' }}>
            New Stock Movement
          </Button>
          <Button icon={<ThunderboltOutlined />}>Recalculate Moving Average</Button>
        </Space>
      )}
    >
      <Row gutter={[16, 16]} style={{ marginBottom: '16px' }}>
        <Col xs={24} sm={8}>
          <Card bordered={false} style={{ background: '#131e32', borderRadius: '8px' }}>
            <Statistic
              title={<span style={{ color: '#94a3b8', fontSize: '11px' }}>TOTAL INVENTORY VALUATION</span>}
              value={92250.00}
              precision={2}
              prefix="$"
              valueStyle={{ color: '#10b981', fontFamily: 'monospace', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card bordered={false} style={{ background: '#131e32', borderRadius: '8px' }}>
            <Statistic
              title={<span style={{ color: '#94a3b8', fontSize: '11px' }}>COSTING COMPLEXITY</span>}
              value="O(1)"
              valueStyle={{ color: '#38bdf8', fontFamily: 'monospace', fontWeight: 700 }}
              suffix={<Tag color="blue">Constant Time</Tag>}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card bordered={false} style={{ background: '#131e32', borderRadius: '8px' }}>
            <Statistic
              title={<span style={{ color: '#94a3b8', fontSize: '11px' }}>AUTO COGS POSTED</span>}
              value="Perpetual (gl_entries)"
              valueStyle={{ color: '#fbbf24', fontSize: '18px', fontWeight: 600 }}
            />
          </Card>
        </Col>
      </Row>

      <Table
        dataSource={SAMPLE_STOCK}
        columns={columns}
        rowKey="id"
        pagination={false}
        size="middle"
      />
    </List>
  );
};
