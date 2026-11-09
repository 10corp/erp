import React from 'react';
import { Card, Row, Col, Typography, Tag, Progress, Statistic, Button, Space } from 'antd';
import {
  ArrowUpOutlined,
  ThunderboltOutlined,
  AuditOutlined,
  DollarOutlined,
  DatabaseOutlined,
  SafetyCertificateOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';

const { Title, Text } = Typography;

export const AdminDashboard: React.FC = () => {
  return (
    <div style={{ padding: '4px' }}>
      {/* Top Banner */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '20px',
          background: 'linear-gradient(135deg, #1e293b 0%, #0f172a 100%)',
          padding: '16px 20px',
          borderRadius: '8px',
          border: '1px solid #334155',
        }}
      >
        <div>
          <Title level={4} style={{ margin: 0, color: '#f8fafc' }}>
            <SafetyCertificateOutlined style={{ color: '#10b981', marginRight: '8px' }} />
            CEO & Financial Controller Executive Dashboard (CQRS)
          </Title>
          <Text style={{ color: '#94a3b8', fontSize: '13px' }}>
            Non-locking real-time analytics powered by PostgreSQL 16 Materialized Views & Native Go 1.22
          </Text>
        </div>
        <Space>
          <Tag color="cyan" icon={<DatabaseOutlined />}>
            MV: mv_ceo_financial_kpis
          </Tag>
          <Tag color="green" icon={<CheckCircleOutlined />}>
            Audit Invariant Balanced: 0.000000
          </Tag>
        </Space>
      </div>

      {/* KPI Cards */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false} style={{ background: '#131e32', border: '1px solid #1e293b', borderRadius: '8px' }}>
            <Statistic
              title={<span style={{ color: '#94a3b8', fontSize: '12px', textTransform: 'uppercase' }}>Total Revenue</span>}
              value={1473125.00}
              precision={2}
              valueStyle={{ color: '#10b981', fontFamily: 'monospace', fontWeight: 700 }}
              prefix="৳"
              suffix={<ArrowUpOutlined style={{ fontSize: '14px', color: '#10b981' }} />}
            />
            <div style={{ marginTop: '8px', fontSize: '11px', color: '#059669' }}>
              +14.2% MoM (Automated GL Invoicing)
            </div>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false} style={{ background: '#131e32', border: '1px solid #1e293b', borderRadius: '8px' }}>
            <Statistic
              title={<span style={{ color: '#94a3b8', fontSize: '12px', textTransform: 'uppercase' }}>Gross Profit Margin</span>}
              value={78.4}
              precision={1}
              valueStyle={{ color: '#38bdf8', fontFamily: 'monospace', fontWeight: 700 }}
              suffix="%"
            />
            <div style={{ marginTop: '8px' }}>
              <Progress percent={78.4} size="small" status="active" strokeColor="#38bdf8" />
            </div>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false} style={{ background: '#131e32', border: '1px solid #1e293b', borderRadius: '8px' }}>
            <Statistic
              title={<span style={{ color: '#94a3b8', fontSize: '12px', textTransform: 'uppercase' }}>Total Current Assets</span>}
              value={1923125.00}
              precision={2}
              valueStyle={{ color: '#fbbf24', fontFamily: 'monospace', fontWeight: 700 }}
              prefix="৳"
            />
            <div style={{ marginTop: '8px', fontSize: '11px', color: '#d97706' }}>
              AR: ৳1.47M | Inventory: ৳450K
            </div>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card bordered={false} style={{ background: '#131e32', border: '1px solid #1e293b', borderRadius: '8px' }}>
            <Statistic
              title={<span style={{ color: '#94a3b8', fontSize: '12px', textTransform: 'uppercase' }}>Engine Latency</span>}
              value={0.48}
              precision={2}
              valueStyle={{ color: '#a78bfa', fontFamily: 'monospace', fontWeight: 700 }}
              suffix="ms"
            />
            <div style={{ marginTop: '8px', fontSize: '11px', color: '#7c3aed' }}>
              Native Go + Goroutines (0-lag)
            </div>
          </Card>
        </Col>
      </Row>

      {/* Financial Health Summary */}
      <Row gutter={[16, 16]} style={{ marginTop: '16px' }}>
        <Col xs={24} lg={16}>
          <Card
            title={<span style={{ color: '#f8fafc' }}>Accounting Compliance & Ledger Partitions</span>}
            bordered={false}
            style={{ background: '#0f172a', border: '1px solid #1e293b', borderRadius: '8px' }}
          >
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '16px' }}>
              <div style={{ padding: '12px', background: '#1e293b', borderRadius: '6px' }}>
                <Text style={{ color: '#94a3b8', fontSize: '12px' }}>Active Partition Range</Text>
                <div style={{ fontSize: '16px', fontWeight: 600, color: '#f8fafc', marginTop: '4px' }}>
                  gl_entries_2026_h1
                </div>
                <Tag color="blue" style={{ marginTop: '8px' }}>PostgreSQL Range Partition</Tag>
              </div>

              <div style={{ padding: '12px', background: '#1e293b', borderRadius: '6px' }}>
                <Text style={{ color: '#94a3b8', fontSize: '12px' }}>Forex Variance Guard</Text>
                <div style={{ fontSize: '16px', fontWeight: 600, color: '#10b981', marginTop: '4px' }}>
                  Account 4200 / 5200
                </div>
                <Tag color="green" style={{ marginTop: '8px' }}>Auto Contra Reversal</Tag>
              </div>

              <div style={{ padding: '12px', background: '#1e293b', borderRadius: '6px' }}>
                <Text style={{ color: '#94a3b8', fontSize: '12px' }}>Round-Off Tolerence</Text>
                <div style={{ fontSize: '16px', fontWeight: 600, color: '#fbbf24', marginTop: '4px' }}>
                  Account 5090 (≤0.05)
                </div>
                <Tag color="gold" style={{ marginTop: '8px' }}>Precision Protected</Tag>
              </div>
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card
            title={<span style={{ color: '#f8fafc' }}>Automated Fast Controls</span>}
            bordered={false}
            style={{ background: '#0f172a', border: '1px solid #1e293b', borderRadius: '8px' }}
          >
            <Space direction="vertical" style={{ width: '100%' }}>
              <Button type="primary" block icon={<ThunderboltOutlined />} style={{ backgroundColor: '#2563eb' }}>
                Refresh Materialized Views (Concurrent)
              </Button>
              <Button block icon={<AuditOutlined />} style={{ borderColor: '#334155', color: '#94a3b8' }}>
                Run Trial Balance Audit Check
              </Button>
              <Button block icon={<DollarOutlined />} style={{ borderColor: '#334155', color: '#94a3b8' }}>
                Export CEO Monthly Financial Pack
              </Button>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  );
};
