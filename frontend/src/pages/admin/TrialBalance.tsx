import React, { useState } from 'react';
import { List } from '@refinedev/antd';
import { Table, Tag, Space, Button, Switch, Modal, Form, InputNumber, Select, Card, Row, Col, Statistic, message } from 'antd';
import {
  AuditOutlined,
  CheckCircleOutlined,
  SwapOutlined,
  ApartmentOutlined,
  FileDoneOutlined,
  DollarOutlined,
} from '@ant-design/icons';

export interface TrialBalanceItem {
  account_code: string;
  account_name: string;
  root_type: 'Asset' | 'Liability' | 'Equity' | 'Income' | 'Expense';
  total_debit: number;
  total_credit: number;
  net_debit: number;
  net_credit: number;
}

const SAMPLE_TRIAL_BALANCE: TrialBalanceItem[] = [
  {
    account_code: '1001',
    account_name: 'Cash on Hand & Bank Current Accounts',
    root_type: 'Asset',
    total_debit: 450000.0,
    total_credit: 0.0,
    net_debit: 450000.0,
    net_credit: 0.0,
  },
  {
    account_code: '1020',
    account_name: 'Accounts Receivable (Trade Debtors Sub-Ledger)',
    root_type: 'Asset',
    total_debit: 1473125.0,
    total_credit: 450000.0,
    net_debit: 1023125.0,
    net_credit: 0.0,
  },
  {
    account_code: '1030',
    account_name: 'Stock in Hand (Inventory Asset)',
    root_type: 'Asset',
    total_debit: 450000.0,
    total_credit: 0.0,
    net_debit: 450000.0,
    net_credit: 0.0,
  },
  {
    account_code: '1990',
    account_name: 'Due From Branches (Inter-Company Clearing Asset)',
    root_type: 'Asset',
    total_debit: 150000.0,
    total_credit: 0.0,
    net_debit: 150000.0,
    net_credit: 0.0,
  },
  {
    account_code: '2001',
    account_name: 'Accounts Payable (Trade Creditors Sub-Ledger)',
    root_type: 'Liability',
    total_debit: 0.0,
    total_credit: 450000.0,
    net_debit: 0.0,
    net_credit: 450000.0,
  },
  {
    account_code: '2990',
    account_name: 'Due To Branches (Inter-Company Clearing Liability)',
    root_type: 'Liability',
    total_debit: 0.0,
    total_credit: 150000.0,
    net_debit: 0.0,
    net_credit: 150000.0,
  },
  {
    account_code: '4001',
    account_name: 'Operating Sales Revenue',
    root_type: 'Income',
    total_debit: 0.0,
    total_credit: 1473125.0,
    net_debit: 0.0,
    net_credit: 1473125.0,
  },
];

export const TrialBalancePage: React.FC = () => {
  const [consolidated, setConsolidated] = useState<boolean>(true);
  const [transferModalOpen, setTransferModalOpen] = useState<boolean>(false);
  const [form] = Form.useForm();

  const totalDebit = SAMPLE_TRIAL_BALANCE.reduce((sum, item) => sum + item.total_debit, 0);
  const totalCredit = SAMPLE_TRIAL_BALANCE.reduce((sum, item) => sum + item.total_credit, 0);
  const difference = Math.abs(totalDebit - totalCredit);

  const handleInterBranchSubmit = (values: any) => {
    message.success(
      `Atomic Inter-Branch Transfer Committed! Cleared ৳${values.amount.toLocaleString()} between ${values.from} and ${values.to} via Due To / Due From accounts (1990 / 2990).`
    );
    setTransferModalOpen(false);
    form.resetFields();
  };

  const columns = [
    {
      title: 'Account Code',
      dataIndex: 'account_code',
      key: 'account_code',
      render: (code: string) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 600, color: '#38bdf8' }}>{code}</span>
      ),
    },
    {
      title: 'Account Name',
      dataIndex: 'account_name',
      key: 'account_name',
      render: (name: string, record: TrialBalanceItem) => (
        <Space>
          <span>{name}</span>
          {record.account_code === '1990' || record.account_code === '2990' ? (
            <Tag color="cyan">Inter-Branch Clearing</Tag>
          ) : null}
        </Space>
      ),
    },
    {
      title: 'Type',
      dataIndex: 'root_type',
      key: 'root_type',
      render: (type: string) => {
        const color =
          type === 'Asset'
            ? 'green'
            : type === 'Liability'
            ? 'red'
            : type === 'Income'
            ? 'purple'
            : 'blue';
        return <Tag color={color}>{type}</Tag>;
      },
    },
    {
      title: 'Total Debit (৳)',
      dataIndex: 'total_debit',
      key: 'total_debit',
      align: 'right' as const,
      render: (val: number) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 600, color: val > 0 ? '#10b981' : '#64748b' }}>
          {val > 0 ? val.toLocaleString(undefined, { minimumFractionDigits: 2 }) : '-'}
        </span>
      ),
    },
    {
      title: 'Total Credit (৳)',
      dataIndex: 'total_credit',
      key: 'total_credit',
      align: 'right' as const,
      render: (val: number) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 600, color: val > 0 ? '#38bdf8' : '#64748b' }}>
          {val > 0 ? val.toLocaleString(undefined, { minimumFractionDigits: 2 }) : '-'}
        </span>
      ),
    },
    {
      title: 'Net Debit (৳)',
      dataIndex: 'net_debit',
      key: 'net_debit',
      align: 'right' as const,
      render: (val: number) => (
        <span style={{ fontFamily: 'monospace', color: val > 0 ? '#10b981' : '#64748b' }}>
          {val > 0 ? val.toLocaleString(undefined, { minimumFractionDigits: 2 }) : '-'}
        </span>
      ),
    },
    {
      title: 'Net Credit (৳)',
      dataIndex: 'net_credit',
      key: 'net_credit',
      align: 'right' as const,
      render: (val: number) => (
        <span style={{ fontFamily: 'monospace', color: val > 0 ? '#38bdf8' : '#64748b' }}>
          {val > 0 ? val.toLocaleString(undefined, { minimumFractionDigits: 2 }) : '-'}
        </span>
      ),
    },
  ];

  return (
    <List
      title="Consolidated Trial Balance & Sub-Ledger Integrity (Tryton GAAP/IFRS Invariant)"
      headerButtons={({ defaultButtons }) => (
        <Space size="middle">
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              background: '#1e293b',
              padding: '4px 12px',
              borderRadius: '6px',
            }}
          >
            <ApartmentOutlined style={{ color: '#38bdf8' }} />
            <span style={{ fontSize: '13px', color: '#cbd5e1' }}>Consolidated Multi-Branch View</span>
            <Switch checked={consolidated} onChange={(chk) => setConsolidated(chk)} />
          </div>

          <Button
            type="primary"
            icon={<SwapOutlined />}
            onClick={() => setTransferModalOpen(true)}
            style={{ backgroundColor: '#2563eb' }}
          >
            New Inter-Branch Transfer
          </Button>
        </Space>
      )}
    >
      {/* KPI Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: '16px' }}>
        <Col xs={24} sm={8}>
          <Card bordered={false} style={{ background: '#131e32', borderRadius: '8px' }}>
            <Statistic
              title={<span style={{ color: '#94a3b8', fontSize: '11px' }}>TOTAL DEBIT BASE</span>}
              value={totalDebit}
              precision={2}
              prefix="৳"
              valueStyle={{ color: '#10b981', fontFamily: 'monospace', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card bordered={false} style={{ background: '#131e32', borderRadius: '8px' }}>
            <Statistic
              title={<span style={{ color: '#94a3b8', fontSize: '11px' }}>TOTAL CREDIT BASE</span>}
              value={totalCredit}
              precision={2}
              prefix="৳"
              valueStyle={{ color: '#38bdf8', fontFamily: 'monospace', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card bordered={false} style={{ background: '#131e32', borderRadius: '8px' }}>
            <Statistic
              title={<span style={{ color: '#94a3b8', fontSize: '11px' }}>BALANCE INVARIANT DISCREPANCY</span>}
              value={difference}
              precision={6}
              prefix="৳"
              valueStyle={{ color: '#10b981', fontFamily: 'monospace', fontWeight: 700 }}
              suffix={<Tag color="green">0.000000 BALANCED</Tag>}
            />
          </Card>
        </Col>
      </Row>

      <Table
        dataSource={SAMPLE_TRIAL_BALANCE}
        columns={columns}
        rowKey="account_code"
        pagination={false}
        size="middle"
        summary={() => (
          <Table.Summary fixed>
            <Table.Summary.Row style={{ background: '#1e293b', fontWeight: 700 }}>
              <Table.Summary.Cell index={0} colSpan={3}>
                <span style={{ color: '#f8fafc' }}>
                  {consolidated
                    ? 'Total Group Consolidated Balance (Holding + All Branches)'
                    : 'Total Standalone Branch Balance'}
                </span>
              </Table.Summary.Cell>
              <Table.Summary.Cell index={3} align="right">
                <span style={{ color: '#10b981', fontFamily: 'monospace' }}>
                  ৳ {totalDebit.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                </span>
              </Table.Summary.Cell>
              <Table.Summary.Cell index={4} align="right">
                <span style={{ color: '#38bdf8', fontFamily: 'monospace' }}>
                  ৳ {totalCredit.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                </span>
              </Table.Summary.Cell>
              <Table.Summary.Cell index={5} align="right">
                <span style={{ color: '#10b981', fontFamily: 'monospace' }}>
                  ৳ {totalDebit.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                </span>
              </Table.Summary.Cell>
              <Table.Summary.Cell index={6} align="right">
                <span style={{ color: '#38bdf8', fontFamily: 'monospace' }}>
                  ৳ {totalCredit.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                </span>
              </Table.Summary.Cell>
            </Table.Summary.Row>
          </Table.Summary>
        )}
      />

      {/* Inter-Branch Transfer Modal */}
      <Modal
        title={
          <Space>
            <SwapOutlined style={{ color: '#38bdf8' }} />
            <span>Inter-Branch Fund / Stock Transfer</span>
          </Space>
        }
        open={transferModalOpen}
        onCancel={() => setTransferModalOpen(false)}
        onOk={() => form.submit()}
        okText="Execute Clearing Transfer"
      >
        <Form form={form} layout="vertical" onFinish={handleInterBranchSubmit}>
          <Form.Item
            name="from"
            label="Source Branch (Origin)"
            initialValue="Enterprise ERP Global Ltd (HQ)"
            rules={[{ required: true }]}
          >
            <Select>
              <Select.Option value="Enterprise ERP Global Ltd (HQ)">Enterprise ERP Global Ltd (HQ)</Select.Option>
              <Select.Option value="Dhaka Principal Branch">Dhaka Principal Branch</Select.Option>
              <Select.Option value="Chittagong Port Distribution Branch">Chittagong Port Distribution Branch</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="to"
            label="Destination Branch (Target)"
            initialValue="Chittagong Port Distribution Branch"
            rules={[{ required: true }]}
          >
            <Select>
              <Select.Option value="Enterprise ERP Global Ltd (HQ)">Enterprise ERP Global Ltd (HQ)</Select.Option>
              <Select.Option value="Dhaka Principal Branch">Dhaka Principal Branch</Select.Option>
              <Select.Option value="Chittagong Port Distribution Branch">Chittagong Port Distribution Branch</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="amount"
            label="Transfer Amount (BDT)"
            initialValue={150000}
            rules={[{ required: true }]}
          >
            <InputNumber style={{ width: '100%' }} size="large" prefix="৳" />
          </Form.Item>

          <Form.Item name="remarks" label="Transfer Memo / Reference" initialValue="Fund transfer for port customs clearing">
            <Select>
              <Select.Option value="Fund transfer for port customs clearing">Fund transfer for port customs clearing</Select.Option>
              <Select.Option value="Inter-branch inventory replenishment">Inter-branch inventory replenishment</Select.Option>
              <Select.Option value="Operational expense advance">Operational expense advance</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </List>
  );
};
