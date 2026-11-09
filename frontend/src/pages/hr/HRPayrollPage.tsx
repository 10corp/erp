import React, { useState, useEffect } from 'react';
import {
  Card,
  Tabs,
  Table,
  Tag,
  Button,
  Space,
  Typography,
  Row,
  Col,
  Statistic,
  Modal,
  Form,
  Input,
  Select,
  InputNumber,
  Badge,
  message,
  Popconfirm,
  Avatar,
  Tooltip,
} from 'antd';
import {
  DollarOutlined,
  TeamOutlined,
  CalendarOutlined,
  PlusOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  SendOutlined,
  AuditOutlined,
  UserOutlined,
  BankOutlined,
  ClockCircleOutlined,
  FileDoneOutlined,
  PaperClipOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';

const { Title, Text } = Typography;

export interface SalarySlipRow {
  id: string;
  slip_number: string;
  employee_id: string;
  employee_code: string;
  employee_name: string;
  designation: string;
  department: string;
  posting_date: string;
  month: number;
  year: number;
  base_salary: string;
  house_rent: string;
  medical: string;
  conveyance: string;
  gross_pay: string;
  provident_fund: string;
  tax_deduction: string;
  total_deduction: string;
  net_pay: string;
  docstatus: number; // 0=Draft, 1=Submitted/Posted, 2=Cancelled
  journal_entry_id?: string;
  voucher_no?: string;
  remarks: string;
}

export interface EmployeeRow {
  id: string;
  party_name: string;
  employee_code: string;
  designation: string;
  department: string;
  company_name: string;
  start_date: string;
  is_active: boolean;
  base_salary: string;
}

export interface LeaveRow {
  id: string;
  employee_name: string;
  leave_type: string;
  from_date: string;
  to_date: string;
  total_days: number;
  reason: string;
  docstatus: number; // 0=Draft, 1=Approved
  approved_by?: string;
}

export interface AttendanceRow {
  id: string;
  employee_id: string;
  employee_name: string;
  employee_code: string;
  department: string;
  attendance_date: string;
  check_in: string;
  check_out?: string;
  working_hours: string;
  overtime_hours: string;
  status: string;
  source: string;
}

export interface ClaimRow {
  id: string;
  claim_number: string;
  employee_id: string;
  employee_name: string;
  claim_date: string;
  category: string;
  amount: string;
  currency: string;
  description: string;
  receipt_ref?: string;
  docstatus: number; // 0=Draft, 1=Approved/GL Posted, 2=Cancelled
  voucher_no?: string;
}

export const HRPayrollPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<string>('salary-slips');
  const [slips, setSlips] = useState<SalarySlipRow[]>([]);
  const [employees, setEmployees] = useState<EmployeeRow[]>([]);
  const [leaves, setLeaves] = useState<LeaveRow[]>([]);
  const [attendance, setAttendance] = useState<AttendanceRow[]>([]);
  const [claims, setClaims] = useState<ClaimRow[]>([]);
  const [loading, setLoading] = useState<boolean>(false);

  // Modals
  const [slipModalOpen, setSlipModalOpen] = useState<boolean>(false);
  const [slipForm] = Form.useForm();

  const [leaveModalOpen, setLeaveModalOpen] = useState<boolean>(false);
  const [leaveForm] = Form.useForm();

  const [claimModalOpen, setClaimModalOpen] = useState<boolean>(false);
  const [claimForm] = Form.useForm();

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [slipRes, empRes, leaveRes, attRes, claimRes] = await Promise.all([
        fetch('/api/v1/hr/salary-slips'),
        fetch('/api/v1/hr/employees'),
        fetch('/api/v1/hr/leaves'),
        fetch('/api/v1/hr/attendance'),
        fetch('/api/v1/hr/claims'),
      ]);

      if (slipRes.ok) setSlips(await slipRes.json());
      if (empRes.ok) setEmployees(await empRes.json());
      if (leaveRes.ok) setLeaves(await leaveRes.json());
      if (attRes.ok) setAttendance(await attRes.json());
      if (claimRes.ok) setClaims(await claimRes.json());
    } catch (e) {
      console.error('Failed to load HR data', e);
    } finally {
      setLoading(false);
    }
  };

  // 1-Click GL Posting: Submit Slip (docstatus 0 -> 1)
  const handleSubmitSlip = async (id: string) => {
    try {
      const res = await fetch(`/api/v1/hr/salary-slips/${id}/submit`, { method: 'POST' });
      if (res.ok) {
        const updated = await res.json();
        message.success(`Posted ${updated.slip_number} to General Ledger (${updated.voucher_no})!`);
        setSlips((prev) => prev.map((s) => (s.id === id ? updated : s)));
      } else {
        const err = await res.json();
        message.error(err.error || 'Failed to submit salary slip');
      }
    } catch (e) {
      message.error('Network error during posting');
    }
  };

  // Reversal: Cancel Slip (docstatus 1 -> 2)
  const handleCancelSlip = async (id: string) => {
    try {
      const res = await fetch(`/api/v1/hr/salary-slips/${id}/cancel`, { method: 'POST' });
      if (res.ok) {
        const updated = await res.json();
        message.warning(`Cancelled ${updated.slip_number} & posted contra reversal.`);
        setSlips((prev) => prev.map((s) => (s.id === id ? updated : s)));
      }
    } catch (e) {
      message.error('Network error during cancellation');
    }
  };

  // Quick Web Punch Attendance
  const handleQuickPunch = async (employeeId: string) => {
    try {
      const res = await fetch('/api/v1/hr/attendance/check-in', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          company_id: '00000000-0000-0000-0000-000000000001',
          employee_id: employeeId,
          source: 'Web Punch',
        }),
      });

      if (res.ok) {
        const newRecord = await res.json();
        message.success(`Recorded check-in for ${newRecord.employee_name} at ${newRecord.check_in}!`);
        setAttendance([newRecord, ...attendance]);
      }
    } catch (e) {
      message.error('Failed to record check-in');
    }
  };

  // Approve Claim & 1-Click Auto-GL Posting
  const handleApproveClaim = async (id: string) => {
    try {
      const res = await fetch(`/api/v1/hr/claims/${id}/approve`, { method: 'POST' });
      if (res.ok) {
        const updated = await res.json();
        message.success(`Approved ${updated.claim_number} & auto-posted ${updated.voucher_no} to General Ledger!`);
        setClaims((prev) => prev.map((c) => (c.id === id ? updated : c)));
      }
    } catch (e) {
      message.error('Failed to approve claim');
    }
  };

  // Cancel Claim & Contra Reversal
  const handleCancelClaim = async (id: string) => {
    try {
      const res = await fetch(`/api/v1/hr/claims/${id}/cancel`, { method: 'POST' });
      if (res.ok) {
        const updated = await res.json();
        message.warning(`Cancelled ${updated.claim_number} & posted contra reversal.`);
        setClaims((prev) => prev.map((c) => (c.id === id ? updated : c)));
      }
    } catch (e) {
      message.error('Failed to cancel claim');
    }
  };

  // Approve Leave
  const handleApproveLeave = async (id: string) => {
    try {
      const res = await fetch(`/api/v1/hr/leaves/${id}/approve`, { method: 'POST' });
      if (res.ok) {
        const updated = await res.json();
        message.success(`Approved leave application for ${updated.employee_name}!`);
        setLeaves((prev) => prev.map((l) => (l.id === id ? updated : l)));
      }
    } catch (e) {
      message.error('Failed to approve leave');
    }
  };

  const handleCreateSalarySlip = async (values: any) => {
    try {
      const res = await fetch('/api/v1/hr/salary-slips', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          company_id: '00000000-0000-0000-0000-000000000001',
          employee_id: values.employee_id,
          posting_date: values.posting_date,
          month: values.month,
          year: values.year,
          base_salary: String(values.base_salary || 0),
          house_rent: String(values.house_rent || 0),
          medical: String(values.medical || 0),
          conveyance: String(values.conveyance || 0),
          provident_fund: String(values.provident_fund || 0),
          tax_deduction: String(values.tax_deduction || 0),
          remarks: values.remarks || 'Standard Monthly Payroll',
        }),
      });

      if (res.ok) {
        const created = await res.json();
        message.success(`Generated Draft Salary Slip ${created.slip_number}!`);
        setSlips([created, ...slips]);
        setSlipModalOpen(false);
        slipForm.resetFields();
      }
    } catch (e) {
      message.error('Failed to create salary slip');
    }
  };

  const handleCreateLeave = async (values: any) => {
    try {
      const res = await fetch('/api/v1/hr/leaves', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          company_id: '00000000-0000-0000-0000-000000000001',
          employee_id: values.employee_id,
          leave_type: values.leave_type,
          from_date: values.from_date,
          to_date: values.to_date,
          total_days: Number(values.total_days),
          reason: values.reason,
        }),
      });

      if (res.ok) {
        const created = await res.json();
        message.success('Submitted leave application for approval!');
        setLeaves([created, ...leaves]);
        setLeaveModalOpen(false);
        leaveForm.resetFields();
      }
    } catch (e) {
      message.error('Failed to submit leave request');
    }
  };

  const handleCreateClaim = async (values: any) => {
    try {
      const res = await fetch('/api/v1/hr/claims', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          company_id: '00000000-0000-0000-0000-000000000001',
          employee_id: values.employee_id,
          category: values.category,
          amount: String(values.amount),
          description: values.description,
          receipt_ref: values.receipt_ref || 'RECEIPT-ATTACHMENT.pdf',
        }),
      });

      if (res.ok) {
        const created = await res.json();
        message.success(`Submitted reimbursement claim ${created.claim_number}!`);
        setClaims([created, ...claims]);
        setClaimModalOpen(false);
        claimForm.resetFields();
      }
    } catch (e) {
      message.error('Failed to submit claim');
    }
  };

  // Totals
  const totalGross = slips.reduce((sum, s) => sum + Number(s.gross_pay || 0), 0);
  const totalDeductions = slips.reduce((sum, s) => sum + Number(s.total_deduction || 0), 0);
  const totalNet = slips.reduce((sum, s) => sum + Number(s.net_pay || 0), 0);
  const totalClaims = claims.reduce((sum, c) => sum + Number(c.amount || 0), 0);

  return (
    <div style={{ padding: '4px' }}>
      {/* Module Header Banner */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '20px',
        }}
      >
        <div>
          <Title level={3} style={{ margin: 0, color: '#f8fafc' }}>
            👥 Horilla-Enhanced Human Resources & Payroll
          </Title>
          <Text type="secondary" style={{ fontSize: '13px' }}>
            Horilla HRMS Workflows | Tryton Double-Entry docstatus (0=Draft $\rightarrow$ 1=GL Posted $\rightarrow$ 2=Cancelled)
          </Text>
        </div>
        <Space>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setSlipModalOpen(true)}
            style={{ backgroundColor: '#2563eb' }}
          >
            Generate Salary Slip
          </Button>
          <Button
            icon={<PaperClipOutlined />}
            onClick={() => setClaimModalOpen(true)}
            style={{ backgroundColor: '#1e293b', borderColor: '#334155', color: '#10b981' }}
          >
            Submit Expense Claim
          </Button>
          <Button
            icon={<CalendarOutlined />}
            onClick={() => setLeaveModalOpen(true)}
            style={{ backgroundColor: '#1e293b', borderColor: '#334155', color: '#38bdf8' }}
          >
            Apply for Leave
          </Button>
        </Space>
      </div>

      {/* KPI Cards */}
      <Row gutter={16} style={{ marginBottom: '20px' }}>
        <Col span={6}>
          <Card size="small" style={{ backgroundColor: '#0f172a', borderColor: '#1e293b' }}>
            <Statistic
              title={<span style={{ color: '#94a3b8' }}>Total Payroll Disbursed</span>}
              value={totalGross}
              precision={2}
              prefix="৳"
              valueStyle={{ color: '#38bdf8', fontFamily: 'monospace', fontWeight: 600 }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" style={{ backgroundColor: '#0f172a', borderColor: '#1e293b' }}>
            <Statistic
              title={<span style={{ color: '#94a3b8' }}>Statutory Deductions (Tax & PF)</span>}
              value={totalDeductions}
              precision={2}
              prefix="৳"
              valueStyle={{ color: '#f59e0b', fontFamily: 'monospace', fontWeight: 600 }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" style={{ backgroundColor: '#0f172a', borderColor: '#1e293b' }}>
            <Statistic
              title={<span style={{ color: '#94a3b8' }}>Net Transferred to Bank</span>}
              value={totalNet}
              precision={2}
              prefix="৳"
              valueStyle={{ color: '#10b981', fontFamily: 'monospace', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small" style={{ backgroundColor: '#0f172a', borderColor: '#1e293b' }}>
            <Statistic
              title={<span style={{ color: '#94a3b8' }}>Expense Claims Total</span>}
              value={totalClaims}
              precision={2}
              prefix="৳"
              valueStyle={{ color: '#ec4899', fontFamily: 'monospace', fontWeight: 600 }}
            />
          </Card>
        </Col>
      </Row>

      {/* Main Tabs */}
      <Card style={{ backgroundColor: '#0f172a', borderColor: '#1e293b' }}>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            {
              key: 'salary-slips',
              label: (
                <span>
                  <DollarOutlined style={{ marginRight: '6px' }} />
                  Salary Slips & GL Postings
                </span>
              ),
              children: (
                <Table
                  dataSource={slips}
                  loading={loading}
                  rowKey="id"
                  size="middle"
                  columns={[
                    {
                      title: 'Slip # / Period',
                      key: 'slip',
                      render: (_: any, r: SalarySlipRow) => (
                        <div>
                          <span style={{ fontWeight: 600, color: '#f8fafc', fontFamily: 'monospace' }}>
                            {r.slip_number}
                          </span>
                          <div style={{ fontSize: '11px', color: '#94a3b8' }}>
                            {r.year}-{String(r.month).padStart(2, '0')} ({r.posting_date})
                          </div>
                        </div>
                      ),
                    },
                    {
                      title: 'Employee & Designation',
                      key: 'emp',
                      render: (_: any, r: SalarySlipRow) => (
                        <Space>
                          <Avatar style={{ backgroundColor: '#2563eb' }} icon={<UserOutlined />} />
                          <div>
                            <div style={{ color: '#f8fafc', fontWeight: 500 }}>{r.employee_name}</div>
                            <div style={{ fontSize: '12px', color: '#38bdf8' }}>
                              {r.designation} • {r.department}
                            </div>
                          </div>
                        </Space>
                      ),
                    },
                    {
                      title: 'Gross Salary (৳)',
                      dataIndex: 'gross_pay',
                      align: 'right',
                      render: (val: string) => (
                        <span style={{ fontFamily: 'monospace', color: '#38bdf8' }}>
                          ৳ {Number(val).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                        </span>
                      ),
                    },
                    {
                      title: 'Deductions (৳)',
                      dataIndex: 'total_deduction',
                      align: 'right',
                      render: (val: string) => (
                        <span style={{ fontFamily: 'monospace', color: '#f59e0b' }}>
                          ৳ {Number(val).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                        </span>
                      ),
                    },
                    {
                      title: 'Net Salary (৳)',
                      dataIndex: 'net_pay',
                      align: 'right',
                      render: (val: string) => (
                        <span style={{ fontFamily: 'monospace', fontWeight: 700, color: '#10b981' }}>
                          ৳ {Number(val).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                        </span>
                      ),
                    },
                    {
                      title: 'DocStatus',
                      dataIndex: 'docstatus',
                      render: (st: number, r: SalarySlipRow) => (
                        <Space direction="vertical" size={2}>
                          {st === 0 && <Tag color="warning">0: Draft (Unposted)</Tag>}
                          {st === 1 && <Tag color="success">1: Submitted (GL Posted)</Tag>}
                          {st === 2 && <Tag color="error">2: Cancelled (Reversed)</Tag>}
                          {r.voucher_no && (
                            <Tag color="cyan" style={{ fontSize: '10px', fontFamily: 'monospace' }}>
                              {r.voucher_no}
                            </Tag>
                          )}
                        </Space>
                      ),
                    },
                    {
                      title: 'Workflow Actions',
                      key: 'actions',
                      align: 'right',
                      render: (_: any, r: SalarySlipRow) => (
                        <Space>
                          {r.docstatus === 0 && (
                            <Button
                              type="primary"
                              size="small"
                              icon={<SendOutlined />}
                              onClick={() => handleSubmitSlip(r.id)}
                              style={{ backgroundColor: '#10b981', borderColor: '#059669' }}
                            >
                              1-Click Post GL
                            </Button>
                          )}
                          {r.docstatus === 1 && (
                            <Popconfirm
                              title="Cancel & Post Contra Reversal"
                              description="Are you sure you want to cancel this payroll entry? A contra reversal voucher will be posted to the general ledger."
                              onConfirm={() => handleCancelSlip(r.id)}
                              okText="Yes, Reverse"
                              cancelText="No"
                            >
                              <Button size="small" danger icon={<CloseCircleOutlined />}>
                                Cancel & Reverse
                              </Button>
                            </Popconfirm>
                          )}
                        </Space>
                      ),
                    },
                  ]}
                />
              ),
            },
            {
              key: 'attendance',
              label: (
                <span>
                  <ClockCircleOutlined style={{ marginRight: '6px', color: '#38bdf8' }} />
                  Daily Attendance & Biometrics
                </span>
              ),
              children: (
                <div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                    <Text style={{ color: '#94a3b8' }}>
                      Horilla Biometric Sync (ZKTeco), Mobile GPS Geofence & Web Check-In with Grace Time Enforcement (09:15 AM)
                    </Text>
                    <Space>
                      <Select
                        placeholder="Quick Check-In as..."
                        style={{ width: 220 }}
                        onChange={(empId) => handleQuickPunch(empId)}
                        options={employees.map((e) => ({
                          label: `${e.party_name} (${e.employee_code})`,
                          value: e.id,
                        }))}
                      />
                    </Space>
                  </div>

                  <Table
                    dataSource={attendance}
                    loading={loading}
                    rowKey="id"
                    size="middle"
                    columns={[
                      {
                        title: 'Date',
                        dataIndex: 'attendance_date',
                        key: 'date',
                        render: (d: string) => <span style={{ fontFamily: 'monospace', color: '#94a3b8' }}>{d}</span>,
                      },
                      {
                        title: 'Employee',
                        key: 'emp',
                        render: (_: any, r: AttendanceRow) => (
                          <div>
                            <div style={{ color: '#f8fafc', fontWeight: 600 }}>{r.employee_name}</div>
                            <div style={{ fontSize: '11px', color: '#38bdf8' }}>{r.employee_code} • {r.department}</div>
                          </div>
                        ),
                      },
                      {
                        title: 'Check-In',
                        dataIndex: 'check_in',
                        render: (t: string) => (
                          <Tag color="blue" style={{ fontFamily: 'monospace' }}>
                            {t}
                          </Tag>
                        ),
                      },
                      {
                        title: 'Check-Out',
                        dataIndex: 'check_out',
                        render: (t?: string) =>
                          t ? (
                            <Tag color="cyan" style={{ fontFamily: 'monospace' }}>
                              {t}
                            </Tag>
                          ) : (
                            <Tag color="default">In-Progress</Tag>
                          ),
                      },
                      {
                        title: 'Hours Worked',
                        dataIndex: 'working_hours',
                        align: 'right',
                        render: (h: string) => <span style={{ fontFamily: 'monospace', color: '#cbd5e1' }}>{h} hrs</span>,
                      },
                      {
                        title: 'Overtime',
                        dataIndex: 'overtime_hours',
                        align: 'right',
                        render: (ot: string) =>
                          Number(ot) > 0 ? (
                            <Tag color="purple">+{ot} hrs OT</Tag>
                          ) : (
                            <span style={{ color: '#64748b' }}>-</span>
                          ),
                      },
                      {
                        title: 'Status',
                        dataIndex: 'status',
                        render: (st: string) => (
                          <Tag color={st.includes('Late') ? 'warning' : 'success'}>
                            {st}
                          </Tag>
                        ),
                      },
                      {
                        title: 'Source Device',
                        dataIndex: 'source',
                        render: (src: string) => (
                          <span style={{ fontSize: '12px', color: '#94a3b8' }}>
                            <ThunderboltOutlined style={{ marginRight: '4px' }} />
                            {src}
                          </span>
                        ),
                      },
                    ]}
                  />
                </div>
              ),
            },
            {
              key: 'claims',
              label: (
                <span>
                  <FileDoneOutlined style={{ marginRight: '6px', color: '#10b981' }} />
                  Expense Reimbursements
                </span>
              ),
              children: (
                <div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                    <Text style={{ color: '#94a3b8' }}>
                      Employee out-of-pocket claims (Travel, Medical, Client Entertainment) with 1-Click GL Auto-Posting
                    </Text>
                    <Button
                      type="primary"
                      icon={<PlusOutlined />}
                      onClick={() => setClaimModalOpen(true)}
                      style={{ backgroundColor: '#10b981', borderColor: '#059669' }}
                    >
                      New Expense Claim
                    </Button>
                  </div>

                  <Table
                    dataSource={claims}
                    loading={loading}
                    rowKey="id"
                    size="middle"
                    columns={[
                      {
                        title: 'Claim # / Date',
                        key: 'claim',
                        render: (_: any, r: ClaimRow) => (
                          <div>
                            <span style={{ fontFamily: 'monospace', fontWeight: 600, color: '#f8fafc' }}>
                              {r.claim_number}
                            </span>
                            <div style={{ fontSize: '11px', color: '#94a3b8' }}>{r.claim_date}</div>
                          </div>
                        ),
                      },
                      {
                        title: 'Employee',
                        dataIndex: 'employee_name',
                        render: (name: string) => <span style={{ color: '#f8fafc', fontWeight: 500 }}>{name}</span>,
                      },
                      {
                        title: 'Category & Description',
                        key: 'cat',
                        render: (_: any, r: ClaimRow) => (
                          <div>
                            <Tag color="cyan">{r.category}</Tag>
                            <div style={{ fontSize: '12px', color: '#cbd5e1', marginTop: '2px' }}>{r.description}</div>
                          </div>
                        ),
                      },
                      {
                        title: 'Amount (৳)',
                        dataIndex: 'amount',
                        align: 'right',
                        render: (val: string) => (
                          <span style={{ fontFamily: 'monospace', fontWeight: 700, color: '#10b981' }}>
                            ৳ {Number(val).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                          </span>
                        ),
                      },
                      {
                        title: 'Receipt',
                        dataIndex: 'receipt_ref',
                        render: (ref?: string) =>
                          ref ? (
                            <Tag color="geekblue" icon={<PaperClipOutlined />}>
                              {ref}
                            </Tag>
                          ) : (
                            <span style={{ color: '#64748b' }}>No file</span>
                          ),
                      },
                      {
                        title: 'Status',
                        dataIndex: 'docstatus',
                        render: (st: number, r: ClaimRow) => (
                          <Space direction="vertical" size={2}>
                            {st === 0 && <Tag color="warning">0: Draft / Pending Review</Tag>}
                            {st === 1 && <Tag color="success">1: Approved (GL Posted)</Tag>}
                            {st === 2 && <Tag color="error">2: Cancelled</Tag>}
                            {r.voucher_no && (
                              <Tag color="cyan" style={{ fontSize: '10px', fontFamily: 'monospace' }}>
                                {r.voucher_no}
                              </Tag>
                            )}
                          </Space>
                        ),
                      },
                      {
                        title: 'Actions',
                        key: 'act',
                        align: 'right',
                        render: (_: any, r: ClaimRow) => (
                          <Space>
                            {r.docstatus === 0 && (
                              <Button
                                type="primary"
                                size="small"
                                icon={<CheckCircleOutlined />}
                                onClick={() => handleApproveClaim(r.id)}
                                style={{ backgroundColor: '#10b981', borderColor: '#059669' }}
                              >
                                Approve & Post GL
                              </Button>
                            )}
                            {r.docstatus === 1 && (
                              <Button
                                danger
                                size="small"
                                icon={<CloseCircleOutlined />}
                                onClick={() => handleCancelClaim(r.id)}
                              >
                                Reverse
                              </Button>
                            )}
                          </Space>
                        ),
                      },
                    ]}
                  />
                </div>
              ),
            },
            {
              key: 'employees',
              label: (
                <span>
                  <TeamOutlined style={{ marginRight: '6px' }} />
                  Employee Directory
                </span>
              ),
              children: (
                <Table
                  dataSource={employees}
                  loading={loading}
                  rowKey="id"
                  size="middle"
                  columns={[
                    {
                      title: 'Code & Name',
                      key: 'name',
                      render: (_: any, r: EmployeeRow) => (
                        <Space>
                          <Avatar style={{ backgroundColor: '#10b981' }} icon={<UserOutlined />} />
                          <div>
                            <div style={{ color: '#f8fafc', fontWeight: 600 }}>{r.party_name}</div>
                            <div style={{ fontSize: '12px', color: '#38bdf8', fontFamily: 'monospace' }}>
                              {r.employee_code}
                            </div>
                          </div>
                        </Space>
                      ),
                    },
                    {
                      title: 'Designation & Department',
                      key: 'dept',
                      render: (_: any, r: EmployeeRow) => (
                        <div>
                          <div style={{ color: '#cbd5e1' }}>{r.designation}</div>
                          <div style={{ fontSize: '12px', color: '#94a3b8' }}>{r.department}</div>
                        </div>
                      ),
                    },
                    {
                      title: 'Assigned Branch',
                      dataIndex: 'company_name',
                      render: (b: string) => (
                        <span style={{ fontSize: '12px', color: '#94a3b8' }}>
                          <BankOutlined style={{ marginRight: '4px' }} />
                          {b}
                        </span>
                      ),
                    },
                    {
                      title: 'Base Gross (৳)',
                      dataIndex: 'base_salary',
                      align: 'right',
                      render: (val: string) => (
                        <span style={{ fontFamily: 'monospace', color: '#38bdf8' }}>
                          ৳ {Number(val).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                        </span>
                      ),
                    },
                    {
                      title: 'Start Date',
                      dataIndex: 'start_date',
                      render: (d: string) => <span style={{ color: '#94a3b8', fontSize: '12px' }}>{d}</span>,
                    },
                    {
                      title: 'Status',
                      dataIndex: 'is_active',
                      render: (act: boolean) => (
                        <Tag color={act ? 'success' : 'default'}>{act ? 'Active' : 'Terminated'}</Tag>
                      ),
                    },
                  ]}
                />
              ),
            },
            {
              key: 'leaves',
              label: (
                <span>
                  <CalendarOutlined style={{ marginRight: '6px' }} />
                  Leave Applications
                </span>
              ),
              children: (
                <Table
                  dataSource={leaves}
                  loading={loading}
                  rowKey="id"
                  size="middle"
                  columns={[
                    {
                      title: 'Employee',
                      dataIndex: 'employee_name',
                      key: 'emp',
                      render: (n: string) => <span style={{ fontWeight: 600, color: '#f8fafc' }}>{n}</span>,
                    },
                    {
                      title: 'Leave Type',
                      dataIndex: 'leave_type',
                      render: (t: string) => <Tag color="blue">{t}</Tag>,
                    },
                    {
                      title: 'Duration',
                      key: 'dates',
                      render: (_: any, r: LeaveRow) => (
                        <div>
                          <span style={{ fontSize: '12px', color: '#cbd5e1' }}>
                            {r.from_date} to {r.to_date}
                          </span>
                          <Tag color="cyan" style={{ marginLeft: '8px' }}>
                            {r.total_days} Days
                          </Tag>
                        </div>
                      ),
                    },
                    {
                      title: 'Reason Narration',
                      dataIndex: 'reason',
                      render: (re: string) => <span style={{ color: '#94a3b8', fontSize: '12px' }}>{re}</span>,
                    },
                    {
                      title: 'Status',
                      dataIndex: 'docstatus',
                      render: (st: number, r: LeaveRow) => (
                        <div>
                          {st === 0 && <Tag color="warning">Pending Approval</Tag>}
                          {st === 1 && (
                            <Tag color="success" icon={<CheckCircleOutlined />}>
                              Approved ({r.approved_by || 'HR'})
                            </Tag>
                          )}
                        </div>
                      ),
                    },
                    {
                      title: 'Action',
                      key: 'act',
                      align: 'right',
                      render: (_: any, r: LeaveRow) =>
                        r.docstatus === 0 ? (
                          <Button
                            type="primary"
                            size="small"
                            onClick={() => handleApproveLeave(r.id)}
                            style={{ backgroundColor: '#10b981', borderColor: '#059669' }}
                          >
                            Approve
                          </Button>
                        ) : (
                          <Tag color="default">Completed</Tag>
                        ),
                    },
                  ]}
                />
              ),
            },
            {
              key: 'audit-ledger',
              label: (
                <span>
                  <AuditOutlined style={{ marginRight: '6px' }} />
                  Payroll & Claims GL Rules
                </span>
              ),
              children: (
                <div style={{ padding: '16px' }}>
                  <Title level={5} style={{ color: '#f8fafc' }}>
                    GAAP/IFRS Payroll & Expense Claims Auto-Posting Invariant Rules
                  </Title>
                  <Text style={{ color: '#94a3b8' }}>
                    Unlike Horilla HRMS where payroll and claims are isolated silos, our ERP directly links HR operations to immutable double-entry ledgers:
                  </Text>

                  <div style={{ marginTop: '16px', background: '#1e293b', padding: '16px', borderRadius: '8px', border: '1px solid #334155' }}>
                    <div style={{ color: '#38bdf8', fontWeight: 600, marginBottom: '8px' }}>
                      1. Payroll Run: Dr 5020 (Salary Expense) = Cr 2010 (Net Salary Payable) + Cr 2020 (Tax/PF Withheld)
                    </div>
                    <div style={{ color: '#10b981', fontWeight: 600, marginBottom: '8px' }}>
                      2. Approved Claim: Dr 5020 (Travel/Office Expense) = Cr 2010 (Reimbursement Payable to Employee)
                    </div>
                    <div style={{ color: '#f59e0b', fontWeight: 600 }}>
                      3. Cancellation: Generates explicit reversal contra vouchers with 0.000000 variance invariant.
                    </div>
                  </div>
                </div>
              ),
            },
          ]}
        />
      </Card>

      {/* Generate Salary Slip Modal */}
      <Modal
        title="⚡ Generate Monthly Salary Slip"
        open={slipModalOpen}
        onCancel={() => setSlipModalOpen(false)}
        footer={null}
        styles={{
          header: { backgroundColor: '#1e293b' },
          content: { backgroundColor: '#0f172a' },
        }}
      >
        <Form form={slipForm} layout="vertical" onFinish={handleCreateSalarySlip} initialValues={{ year: 2026, month: 11, posting_date: '2026-11-30' }}>
          <Form.Item name="employee_id" label={<span style={{ color: '#94a3b8' }}>Employee Profile</span>} rules={[{ required: true }]}>
            <Select
              placeholder="Select employee"
              options={employees.map((e) => ({
                label: `${e.party_name} (${e.employee_code}) - Base ৳${e.base_salary}`,
                value: e.id,
              }))}
            />
          </Form.Item>

          <Row gutter={12}>
            <Col span={8}>
              <Form.Item name="year" label={<span style={{ color: '#94a3b8' }}>Fiscal Year</span>} rules={[{ required: true }]}>
                <InputNumber style={{ width: '100%', backgroundColor: '#1e293b', color: '#f8fafc' }} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="month" label={<span style={{ color: '#94a3b8' }}>Month (1-12)</span>} rules={[{ required: true }]}>
                <InputNumber min={1} max={12} style={{ width: '100%', backgroundColor: '#1e293b', color: '#f8fafc' }} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="posting_date" label={<span style={{ color: '#94a3b8' }}>Posting Date</span>} rules={[{ required: true }]}>
                <Input type="date" style={{ backgroundColor: '#1e293b', color: '#f8fafc' }} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={12}>
            <Col span={12}>
              <Form.Item name="base_salary" label={<span style={{ color: '#94a3b8' }}>Base Salary (৳)</span>} rules={[{ required: true }]}>
                <InputNumber style={{ width: '100%', backgroundColor: '#1e293b', color: '#f8fafc' }} placeholder="60000" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="house_rent" label={<span style={{ color: '#94a3b8' }}>House Rent (৳)</span>}>
                <InputNumber style={{ width: '100%', backgroundColor: '#1e293b', color: '#f8fafc' }} placeholder="30000" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="medical" label={<span style={{ color: '#94a3b8' }}>Medical Allowance (৳)</span>}>
                <InputNumber style={{ width: '100%', backgroundColor: '#1e293b', color: '#f8fafc' }} placeholder="5000" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="conveyance" label={<span style={{ color: '#94a3b8' }}>Conveyance (৳)</span>}>
                <InputNumber style={{ width: '100%', backgroundColor: '#1e293b', color: '#f8fafc' }} placeholder="5000" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="provident_fund" label={<span style={{ color: '#94a3b8' }}>Provident Fund Deduction (৳)</span>}>
                <InputNumber style={{ width: '100%', backgroundColor: '#1e293b', color: '#f8fafc' }} placeholder="6000" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="tax_deduction" label={<span style={{ color: '#94a3b8' }}>Tax Deduction (৳)</span>}>
                <InputNumber style={{ width: '100%', backgroundColor: '#1e293b', color: '#f8fafc' }} placeholder="4000" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="remarks" label={<span style={{ color: '#94a3b8' }}>Audit Remarks</span>}>
            <Input.TextArea rows={2} style={{ backgroundColor: '#1e293b', color: '#f8fafc' }} />
          </Form.Item>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
            <Button onClick={() => setSlipModalOpen(false)}>Cancel</Button>
            <Button type="primary" htmlType="submit" style={{ backgroundColor: '#2563eb' }}>
              Create Draft Slip (docstatus 0)
            </Button>
          </div>
        </Form>
      </Modal>

      {/* Apply for Leave Modal */}
      <Modal
        title="📅 Apply for Leave"
        open={leaveModalOpen}
        onCancel={() => setLeaveModalOpen(false)}
        footer={null}
        styles={{
          header: { backgroundColor: '#1e293b' },
          content: { backgroundColor: '#0f172a' },
        }}
      >
        <Form form={leaveForm} layout="vertical" onFinish={handleCreateLeave}>
          <Form.Item name="employee_id" label={<span style={{ color: '#94a3b8' }}>Employee</span>} rules={[{ required: true }]}>
            <Select
              placeholder="Select employee"
              options={employees.map((e) => ({
                label: `${e.party_name} (${e.employee_code})`,
                value: e.id,
              }))}
            />
          </Form.Item>

          <Form.Item name="leave_type" label={<span style={{ color: '#94a3b8' }}>Leave Type</span>} rules={[{ required: true }]}>
            <Select
              options={[
                { label: 'Annual Earned Leave', value: 'Annual Earned Leave' },
                { label: 'Casual Leave', value: 'Casual Leave' },
                { label: 'Sick / Medical Leave', value: 'Sick / Medical Leave' },
                { label: 'Maternity / Paternity Leave', value: 'Maternity / Paternity Leave' },
              ]}
            />
          </Form.Item>

          <Row gutter={12}>
            <Col span={8}>
              <Form.Item name="from_date" label={<span style={{ color: '#94a3b8' }}>From Date</span>} rules={[{ required: true }]}>
                <Input type="date" style={{ backgroundColor: '#1e293b', color: '#f8fafc' }} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="to_date" label={<span style={{ color: '#94a3b8' }}>To Date</span>} rules={[{ required: true }]}>
                <Input type="date" style={{ backgroundColor: '#1e293b', color: '#f8fafc' }} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="total_days" label={<span style={{ color: '#94a3b8' }}>Total Days</span>} rules={[{ required: true }]}>
                <InputNumber min={1} style={{ width: '100%', backgroundColor: '#1e293b', color: '#f8fafc' }} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="reason" label={<span style={{ color: '#94a3b8' }}>Reason</span>} rules={[{ required: true }]}>
            <Input.TextArea rows={2} style={{ backgroundColor: '#1e293b', color: '#f8fafc' }} placeholder="Specify reason for leave..." />
          </Form.Item>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
            <Button onClick={() => setLeaveModalOpen(false)}>Cancel</Button>
            <Button type="primary" htmlType="submit" style={{ backgroundColor: '#10b981' }}>
              Submit Application
            </Button>
          </div>
        </Form>
      </Modal>

      {/* New Expense Claim Modal */}
      <Modal
        title="🧾 Submit Expense Reimbursement Claim"
        open={claimModalOpen}
        onCancel={() => setClaimModalOpen(false)}
        footer={null}
        styles={{
          header: { backgroundColor: '#1e293b' },
          content: { backgroundColor: '#0f172a' },
        }}
      >
        <Form form={claimForm} layout="vertical" onFinish={handleCreateClaim}>
          <Form.Item name="employee_id" label={<span style={{ color: '#94a3b8' }}>Claimant Employee</span>} rules={[{ required: true }]}>
            <Select
              placeholder="Select employee"
              options={employees.map((e) => ({
                label: `${e.party_name} (${e.employee_code})`,
                value: e.id,
              }))}
            />
          </Form.Item>

          <Row gutter={12}>
            <Col span={12}>
              <Form.Item name="category" label={<span style={{ color: '#94a3b8' }}>Expense Category</span>} rules={[{ required: true }]}>
                <Select
                  options={[
                    { label: 'Travel & Transportation', value: 'Travel' },
                    { label: 'Client Entertainment / Dinner', value: 'Client Entertainment' },
                    { label: 'Office Supplies & Peripherals', value: 'Office Supplies' },
                    { label: 'Medical & Healthcare Claim', value: 'Medical' },
                  ]}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="amount" label={<span style={{ color: '#94a3b8' }}>Claim Amount (৳)</span>} rules={[{ required: true }]}>
                <InputNumber style={{ width: '100%', backgroundColor: '#1e293b', color: '#f8fafc' }} placeholder="5000" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="description" label={<span style={{ color: '#94a3b8' }}>Business Purpose Narration</span>} rules={[{ required: true }]}>
            <Input.TextArea rows={2} style={{ backgroundColor: '#1e293b', color: '#f8fafc' }} placeholder="Specify business purpose..." />
          </Form.Item>

          <Form.Item name="receipt_ref" label={<span style={{ color: '#94a3b8' }}>Receipt File Reference</span>} initialValue="RECEIPT-ATTACHMENT.pdf">
            <Input prefix={<PaperClipOutlined />} style={{ backgroundColor: '#1e293b', color: '#f8fafc' }} />
          </Form.Item>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
            <Button onClick={() => setClaimModalOpen(false)}>Cancel</Button>
            <Button type="primary" htmlType="submit" style={{ backgroundColor: '#10b981' }}>
              Submit Claim (docstatus 0)
            </Button>
          </div>
        </Form>
      </Modal>
    </div>
  );
};
