import React, { useState } from 'react';
import { useLogin } from '@refinedev/core';
import { Card, Form, Input, Button, Typography, Tag, Space, Divider, Alert } from 'antd';
import {
  RocketOutlined,
  CrownOutlined,
  UserOutlined,
  EyeOutlined,
  LockOutlined,
  MailOutlined,
  SafetyCertificateOutlined,
  ThunderboltOutlined,
  ArrowRightOutlined,
} from '@ant-design/icons';
import { PRESET_USERS } from '../../providers/authProvider';

const { Title, Text, Paragraph } = Typography;

export const LoginPage: React.FC = () => {
  const { mutate: login, isLoading } = useLogin();
  const [activeTab, setActiveTab] = useState<'quick' | 'form'>('quick');

  const handleQuickLogin = (roleKey: 'admin' | 'sales_rep' | 'visitor') => {
    login({ role: roleKey, email: PRESET_USERS[roleKey].email });
  };

  const onFinish = (values: any) => {
    login({ email: values.email, password: values.password });
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        background: 'radial-gradient(circle at 50% 20%, #1e293b 0%, #0b1329 100%)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '24px',
        fontFamily: "'Inter', sans-serif",
      }}
    >
      <div style={{ maxWidth: '480px', width: '100%' }}>
        {/* Brand Header */}
        <div style={{ textAlign: 'center', marginBottom: '24px' }}>
          <div
            style={{
              width: '56px',
              height: '56px',
              borderRadius: '12px',
              background: 'linear-gradient(135deg, #2563eb 0%, #38bdf8 100%)',
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              boxShadow: '0 8px 24px rgba(37, 99, 235, 0.4)',
              marginBottom: '12px',
            }}
          >
            <RocketOutlined style={{ fontSize: '28px', color: '#fff' }} />
          </div>
          <Title level={2} style={{ color: '#f8fafc', margin: 0, fontWeight: 800, letterSpacing: '-0.5px' }}>
            ENTERPRISE ERP
          </Title>
          <Text style={{ color: '#94a3b8', fontSize: '14px' }}>
            Ultra-Fast Headless ERP with Refine, Hexya (Go), Tryton & ERPNext
          </Text>
        </div>

        {/* Login Box */}
        <Card
          bordered={false}
          style={{
            background: 'rgba(15, 23, 42, 0.85)',
            backdropFilter: 'blur(16px)',
            border: '1px solid #1e293b',
            borderRadius: '12px',
            boxShadow: '0 20px 40px rgba(0, 0, 0, 0.5)',
          }}
        >
          <div style={{ display: 'flex', gap: '8px', marginBottom: '20px' }}>
            <Button
              type={activeTab === 'quick' ? 'primary' : 'default'}
              block
              onClick={() => setActiveTab('quick')}
              icon={<ThunderboltOutlined />}
              style={{
                backgroundColor: activeTab === 'quick' ? '#2563eb' : 'transparent',
                borderColor: '#334155',
                color: activeTab === 'quick' ? '#fff' : '#94a3b8',
              }}
            >
              1-Click Role Switch
            </Button>
            <Button
              type={activeTab === 'form' ? 'primary' : 'default'}
              block
              onClick={() => setActiveTab('form')}
              icon={<LockOutlined />}
              style={{
                backgroundColor: activeTab === 'form' ? '#2563eb' : 'transparent',
                borderColor: '#334155',
                color: activeTab === 'form' ? '#fff' : '#94a3b8',
              }}
            >
              Standard Credentials
            </Button>
          </div>

          {activeTab === 'quick' ? (
            <div>
              <Alert
                message="Choose how you wish to experience the ERP:"
                type="info"
                showIcon
                style={{
                  background: '#131e32',
                  border: '1px solid #1e293b',
                  color: '#94a3b8',
                  marginBottom: '16px',
                  borderRadius: '6px',
                }}
              />

              {/* Persona 1: Admin / CFO */}
              <div
                onClick={() => handleQuickLogin('admin')}
                style={{
                  padding: '14px',
                  borderRadius: '8px',
                  background: '#1e293b',
                  border: '1px solid #334155',
                  marginBottom: '12px',
                  cursor: 'pointer',
                  transition: 'all 0.2s ease',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                }}
                onMouseEnter={(e) => (e.currentTarget.style.borderColor = '#38bdf8')}
                onMouseLeave={(e) => (e.currentTarget.style.borderColor = '#334155')}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                  <div
                    style={{
                      width: '38px',
                      height: '38px',
                      borderRadius: '8px',
                      background: 'rgba(56, 189, 248, 0.1)',
                      color: '#38bdf8',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '18px',
                    }}
                  >
                    <CrownOutlined />
                  </div>
                  <div>
                    <div style={{ color: '#f8fafc', fontWeight: 600, fontSize: '14px' }}>
                      Admin / CFO Panel
                    </div>
                    <div style={{ color: '#94a3b8', fontSize: '11px' }}>
                      Bitmask 15: Full Post, Contra Reversal & CEO KPIs
                    </div>
                  </div>
                </div>
                <Tag color="purple">Sign In <ArrowRightOutlined /></Tag>
              </div>

              {/* Persona 2: Operational User (Sales Officer) */}
              <div
                onClick={() => handleQuickLogin('sales_rep')}
                style={{
                  padding: '14px',
                  borderRadius: '8px',
                  background: '#1e293b',
                  border: '1px solid #334155',
                  marginBottom: '12px',
                  cursor: 'pointer',
                  transition: 'all 0.2s ease',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                }}
                onMouseEnter={(e) => (e.currentTarget.style.borderColor = '#10b981')}
                onMouseLeave={(e) => (e.currentTarget.style.borderColor = '#334155')}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                  <div
                    style={{
                      width: '38px',
                      height: '38px',
                      borderRadius: '8px',
                      background: 'rgba(16, 185, 129, 0.1)',
                      color: '#10b981',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '18px',
                    }}
                  >
                    <UserOutlined />
                  </div>
                  <div>
                    <div style={{ color: '#f8fafc', fontWeight: 600, fontSize: '14px' }}>
                      User / Operational Panel
                    </div>
                    <div style={{ color: '#94a3b8', fontSize: '11px' }}>
                      Bitmask 3: Draft Sales Invoices, Stock & Deliveries
                    </div>
                  </div>
                </div>
                <Tag color="green">Sign In <ArrowRightOutlined /></Tag>
              </div>

              {/* Persona 3: Visitor / Auditor */}
              <div
                onClick={() => handleQuickLogin('visitor')}
                style={{
                  padding: '14px',
                  borderRadius: '8px',
                  background: '#1e293b',
                  border: '1px solid #334155',
                  marginBottom: '16px',
                  cursor: 'pointer',
                  transition: 'all 0.2s ease',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                }}
                onMouseEnter={(e) => (e.currentTarget.style.borderColor = '#fbbf24')}
                onMouseLeave={(e) => (e.currentTarget.style.borderColor = '#334155')}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                  <div
                    style={{
                      width: '38px',
                      height: '38px',
                      borderRadius: '8px',
                      background: 'rgba(251, 191, 36, 0.1)',
                      color: '#fbbf24',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '18px',
                    }}
                  >
                    <EyeOutlined />
                  </div>
                  <div>
                    <div style={{ color: '#f8fafc', fontWeight: 600, fontSize: '14px' }}>
                      Visitor / External Auditor
                    </div>
                    <div style={{ color: '#94a3b8', fontSize: '11px' }}>
                      Bitmask 1: Read-Only 60 FPS General Ledger Inspection
                    </div>
                  </div>
                </div>
                <Tag color="gold">Sign In <ArrowRightOutlined /></Tag>
              </div>
            </div>
          ) : (
            <Form
              layout="vertical"
              initialValues={{ email: 'admin@enterprise-erp.local', password: 'password123' }}
              onFinish={onFinish}
            >
              <Form.Item
                label={<span style={{ color: '#cbd5e1' }}>Email Address</span>}
                name="email"
                rules={[{ required: true, message: 'Please input your email!' }]}
              >
                <Input
                  prefix={<MailOutlined style={{ color: '#64748b' }} />}
                  placeholder="admin@enterprise-erp.local"
                  size="large"
                  style={{ background: '#1e293b', borderColor: '#334155', color: '#f8fafc' }}
                />
              </Form.Item>

              <Form.Item
                label={<span style={{ color: '#cbd5e1' }}>Password</span>}
                name="password"
                rules={[{ required: true, message: 'Please input your password!' }]}
              >
                <Input.Password
                  prefix={<LockOutlined style={{ color: '#64748b' }} />}
                  size="large"
                  style={{ background: '#1e293b', borderColor: '#334155', color: '#f8fafc' }}
                />
              </Form.Item>

              <Button
                type="primary"
                htmlType="submit"
                block
                size="large"
                loading={isLoading}
                style={{ backgroundColor: '#2563eb', borderColor: '#2563eb', marginTop: '8px' }}
              >
                Sign In to Workstation
              </Button>
            </Form>
          )}

          {/* Engine Highlights */}
          <Divider style={{ borderColor: '#1e293b', margin: '16px 0' }} />
          <div style={{ display: 'flex', justifyContent: 'space-between', textAlign: 'center' }}>
            <div>
              <div style={{ fontSize: '12px', fontWeight: 700, color: '#10b981' }}>0.48 ms</div>
              <div style={{ fontSize: '10px', color: '#64748b' }}>Native Go Engine</div>
            </div>
            <div>
              <div style={{ fontSize: '12px', fontWeight: 700, color: '#38bdf8' }}>0.000000</div>
              <div style={{ fontSize: '10px', color: '#64748b' }}>Tryton Balance Invariant</div>
            </div>
            <div>
              <div style={{ fontSize: '12px', fontWeight: 700, color: '#fbbf24' }}>Claude 3.5 & Gemini</div>
              <div style={{ fontSize: '10px', color: '#64748b' }}>Dual AI Core</div>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
};
