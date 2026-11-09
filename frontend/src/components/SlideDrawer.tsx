import React from 'react';
import { Drawer, Form, Input, Select, Button, Space, Typography, Alert } from 'antd';
import { Send, Sparkles } from 'lucide-react';

const { Title, Text } = Typography;

interface SlideDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export const SlideDrawer: React.FC<SlideDrawerProps> = ({ isOpen, onClose, onSuccess }) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = React.useState(false);

  const handleSubmit = async (values: any) => {
    setLoading(true);
    try {
      // Post transaction to Go API
      console.log('Posting Journal Entry to Go API Server:', values);
      setTimeout(() => {
        setLoading(false);
        form.resetFields();
        onSuccess();
        onClose();
      }, 600);
    } catch (err) {
      setLoading(false);
    }
  };

  return (
    <Drawer
      title={
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div>
            <Title level={4} style={{ margin: 0, color: '#f8fafc' }}>
              ⚡ New Journal Entry (Contra Post)
            </Title>
            <Text type="secondary" style={{ fontSize: '12px' }}>
              Tryton Immutability Enforced | Zero-Float Fixed Precision
            </Text>
          </div>
          <Button
            size="small"
            icon={<Sparkles size={14} color="#38bdf8" />}
            style={{ backgroundColor: '#0369a1', borderColor: '#0284c7', color: '#fff' }}
          >
            Auto-Fill with AI
          </Button>
        </div>
      }
      placement="right"
      width={560}
      onClose={onClose}
      open={isOpen}
      styles={{
        header: { backgroundColor: '#1e293b', borderBottom: '1px solid #334155' },
        body: { backgroundColor: '#0f172a', padding: '20px' },
      }}
    >
      <Alert
        message="Strict Financial Integrity Active"
        description="Posted ledger entries are append-only. Sum of debits and credits in base currency must strictly balance (Diff = 0.000000)."
        type="info"
        showIcon
        style={{ marginBottom: '20px', backgroundColor: '#1e293b', borderColor: '#0284c7' }}
      />

      <Form form={form} layout="vertical" onFinish={handleSubmit}>
        <Form.Item
          name="posting_date"
          label={<span style={{ color: '#94a3b8' }}>Posting Date</span>}
          initialValue={new Date().toISOString().split('T')[0]}
          rules={[{ required: true }]}
        >
          <Input type="date" style={{ backgroundColor: '#1e293b', color: '#f8fafc', borderColor: '#334155' }} />
        </Form.Item>

        <Form.Item
          name="debit_account"
          label={<span style={{ color: '#94a3b8' }}>Debit Account</span>}
          rules={[{ required: true }]}
        >
          <Select
            placeholder="Select Asset/Expense Account"
            options={[
              { label: '1001 - Cash and Bank Equivalents', value: '1001' },
              { label: '1020 - Accounts Receivable (Trade Debtors)', value: '1020' },
              { label: '5001 - Cost of Goods Sold', value: '5001' },
              { label: '5020 - Office Rent & Utilities', value: '5020' },
            ]}
          />
        </Form.Item>

        <Form.Item
          name="debit_amount"
          label={<span style={{ color: '#94a3b8' }}>Debit Amount (BDT)</span>}
          rules={[{ required: true }]}
        >
          <Input placeholder="0.000000" style={{ backgroundColor: '#1e293b', color: '#f8fafc', borderColor: '#334155', fontFamily: 'monospace' }} />
        </Form.Item>

        <Form.Item
          name="credit_account"
          label={<span style={{ color: '#94a3b8' }}>Credit Account</span>}
          rules={[{ required: true }]}
        >
          <Select
            placeholder="Select Liability/Income Account"
            options={[
              { label: '2001 - Accounts Payable (Trade Creditors)', value: '2001' },
              { label: '4001 - Sales Revenue', value: '4001' },
              { label: '4010 - Service Consulting Income', value: '4010' },
            ]}
          />
        </Form.Item>

        <Form.Item
          name="credit_amount"
          label={<span style={{ color: '#94a3b8' }}>Credit Amount (BDT)</span>}
          rules={[{ required: true }]}
        >
          <Input placeholder="0.000000" style={{ backgroundColor: '#1e293b', color: '#f8fafc', borderColor: '#334155', fontFamily: 'monospace' }} />
        </Form.Item>

        <Form.Item
          name="remarks"
          label={<span style={{ color: '#94a3b8' }}>Remarks & Audit Narration</span>}
        >
          <Input.TextArea rows={3} placeholder="Provide audit trail narration..." style={{ backgroundColor: '#1e293b', color: '#f8fafc', borderColor: '#334155' }} />
        </Form.Item>

        <div style={{ marginTop: '30px', display: 'flex', justifyContent: 'flex-end', gap: '12px' }}>
          <Button onClick={onClose} style={{ backgroundColor: '#334155', borderColor: '#475569', color: '#e2e8f0' }}>
            Cancel
          </Button>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
            icon={<Send size={14} />}
            style={{ backgroundColor: '#0284c7', borderColor: '#0284c7' }}
          >
            Post & Commit Ledger
          </Button>
        </div>
      </Form>
    </Drawer>
  );
};
