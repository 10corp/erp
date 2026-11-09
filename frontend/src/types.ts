export interface GLEntry {
  id: string;
  company_id: string;
  posting_date: string;
  voucher_type: string;
  voucher_id: string;
  voucher_no: string;
  account_id: string;
  account_code: string;
  party_type?: string;
  party_id?: string;
  currency: string;
  debit: string;
  credit: string;
  exchange_rate: string;
  debit_base: string;
  credit_base: string;
  remarks?: string;
  is_cancelled: boolean;
  created_at: string;
}

export interface Account {
  id: string;
  company_id: string;
  code: string;
  name: string;
  root_type: 'Asset' | 'Liability' | 'Equity' | 'Income' | 'Expense';
  currency: string;
  is_group: boolean;
  is_active: boolean;
}

export interface CEOKPIs {
  total_revenue: string;
  total_expense: string;
  net_profit: string;
  total_assets: string;
  total_liabilities: string;
  calculated_at: string;
}
