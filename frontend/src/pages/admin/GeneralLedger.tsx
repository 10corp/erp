import React, { useState } from 'react';
import { List } from '@refinedev/antd';
import { Button, Space, Tag, Modal, message } from 'antd';
import { SyncOutlined, CheckCircleOutlined, RollbackOutlined, FileTextOutlined } from '@ant-design/icons';
import { LedgerGrid } from '../../components/LedgerGrid';
import { GLEntry } from '../../types';

const SAMPLE_GL_DATA: GLEntry[] = [
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

export const GeneralLedgerPage: React.FC = () => {
  const [data, setData] = useState<GLEntry[]>(SAMPLE_GL_DATA);

  const handleContraReversal = (voucherNo: string) => {
    Modal.confirm({
      title: 'Generate Tryton-Style Contra Reversal?',
      content: `In accordance with GAAP/IFRS, posted entries in voucher ${voucherNo} cannot be updated or deleted. A mirror-image contra reversal entry will be appended to the ledger.`,
      okText: 'Post Contra Reversal',
      okType: 'danger',
      onOk: () => {
        message.success(`Contra reversal generated successfully for ${voucherNo}`);
      },
    });
  };

  return (
    <List
      title="Partitioned General Ledger (Strict Double-Entry Append-Only)"
      headerButtons={({ defaultButtons }) => (
        <Space direction="horizontal">
          <Tag color="cyan">Range: gl_entries_2026_h1</Tag>
          <Tag color="green" icon={<CheckCircleOutlined />}>
            Invariant: 0.000000
          </Tag>
          <Button
            icon={<RollbackOutlined />}
            onClick={() => handleContraReversal('SINV-2026-0042')}
          >
            Post Contra Reversal
          </Button>
          <Button icon={<SyncOutlined />} onClick={() => message.info('Ledger synced')}>
            Sync Partitions
          </Button>
        </Space>
      )}
    >
      <div style={{ background: '#0f172a', padding: '12px', borderRadius: '8px' }}>
        <LedgerGrid rowData={data} />
      </div>
    </List>
  );
};
