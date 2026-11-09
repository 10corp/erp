import { AuthProvider } from '@refinedev/core';

// Bitmask Permissions matching Go backend internal/security/rbac.go
export const PERM_READ = 1 << 0;   // 1
export const PERM_WRITE = 1 << 1;  // 2
export const PERM_DELETE = 1 << 2; // 4
export const PERM_POST = 1 << 3;   // 8

export interface ERPUser {
  id: string;
  name: string;
  email: string;
  role: 'admin' | 'sales_rep' | 'visitor';
  role_label: string;
  permissions: number;
  avatar?: string;
  description: string;
}

export const PRESET_USERS: Record<string, ERPUser> = {
  admin: {
    id: 'usr-admin-01',
    name: 'Tuhin (System Administrator & CFO)',
    email: 'admin@enterprise-erp.local',
    role: 'admin',
    role_label: 'Admin & CFO',
    permissions: PERM_READ | PERM_WRITE | PERM_DELETE | PERM_POST, // 15
    avatar: 'https://api.dicebear.com/7.x/bottts/svg?seed=AdminCFO',
    description: 'Full bitmask 15: Post to General Ledger, Contra Reversals & CEO KPIs',
  },
  sales_rep: {
    id: 'usr-sales-02',
    name: 'Rahim (Operational Sales Officer)',
    email: 'sales@enterprise-erp.local',
    role: 'sales_rep',
    role_label: 'Sales Officer (User)',
    permissions: PERM_READ | PERM_WRITE, // 3
    avatar: 'https://api.dicebear.com/7.x/bottts/svg?seed=SalesExecutive',
    description: 'Bitmask 3: Create Sales Invoices, Review Drafts, Warehouses',
  },
  visitor: {
    id: 'usr-visitor-03',
    name: 'Visitor / External Auditor',
    email: 'visitor@enterprise-erp.local',
    role: 'visitor',
    role_label: 'Visitor (Auditor)',
    permissions: PERM_READ, // 1
    avatar: 'https://api.dicebear.com/7.x/bottts/svg?seed=AuditorVisitor',
    description: 'Bitmask 1: Read-only inspection of 60 FPS GL audit log without posting rights',
  },
};

export const erpAuthProvider: AuthProvider = {
  login: async ({ email, role = 'admin' }) => {
    let selectedUser = PRESET_USERS[role];

    if (!selectedUser && email) {
      if (email.includes('sales')) {
        selectedUser = PRESET_USERS.sales_rep;
      } else if (email.includes('visitor')) {
        selectedUser = PRESET_USERS.visitor;
      } else {
        selectedUser = PRESET_USERS.admin;
      }
    }

    if (!selectedUser) {
      selectedUser = PRESET_USERS.admin;
    }

    localStorage.setItem('erp_auth_user', JSON.stringify(selectedUser));
    return {
      success: true,
      redirectTo: '/admin/dashboard',
    };
  },

  logout: async () => {
    localStorage.removeItem('erp_auth_user');
    return {
      success: true,
      redirectTo: '/login',
    };
  },

  check: async () => {
    const user = localStorage.getItem('erp_auth_user');
    if (user) {
      return { authenticated: true };
    }
    return {
      authenticated: false,
      redirectTo: '/login',
    };
  },

  onError: async (error) => {
    return { error };
  },

  getIdentity: async (): Promise<ERPUser> => {
    const userStr = localStorage.getItem('erp_auth_user');
    if (userStr) {
      try {
        return JSON.parse(userStr);
      } catch (e) {}
    }
    return PRESET_USERS.admin;
  },

  getPermissions: async (): Promise<number> => {
    const userStr = localStorage.getItem('erp_auth_user');
    if (userStr) {
      try {
        const user: ERPUser = JSON.parse(userStr);
        return user.permissions;
      } catch (e) {}
    }
    return PRESET_USERS.admin.permissions;
  },
};
