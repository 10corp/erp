import React, { useMemo } from 'react';
import { AgGridReact } from 'ag-grid-react';
import { ColDef, ModuleRegistry } from 'ag-grid-community';
import { ClientSideRowModelModule } from 'ag-grid-community';
import 'ag-grid-community/styles/ag-grid.css';
import 'ag-grid-community/styles/ag-theme-alpine.css';
import { GLEntry } from '../types';
import { Tag } from 'antd';

ModuleRegistry.registerModules([ClientSideRowModelModule]);

interface LedgerGridProps {
  rowData: GLEntry[];
  loading?: boolean;
}

export const LedgerGrid: React.FC<LedgerGridProps> = ({ rowData, loading }) => {
  const columnDefs: ColDef<GLEntry>[] = useMemo(
    () => [
      {
        field: 'posting_date',
        headerName: 'Posting Date',
        width: 130,
        sortable: true,
        filter: true,
        pinned: 'left',
      },
      {
        field: 'voucher_no',
        headerName: 'Voucher Ref',
        width: 180,
        filter: true,
        cellRenderer: (params: any) => (
          <span style={{ fontFamily: 'monospace', fontWeight: 600, color: '#38bdf8' }}>
            {params.value}
          </span>
        ),
      },
      {
        field: 'voucher_type',
        headerName: 'Type',
        width: 140,
        cellRenderer: (params: any) => {
          const color = params.value === 'Journal Entry' ? 'geekblue' : params.value === 'Sales Invoice' ? 'green' : 'orange';
          return <Tag color={color}>{params.value}</Tag>;
        },
      },
      {
        field: 'account_code',
        headerName: 'Account',
        width: 220,
        valueGetter: (params) => `${params.data?.account_code} - ${params.data?.party_type || 'Main Ledger'}`,
      },
      {
        field: 'debit_base',
        headerName: 'Debit (Base BDT)',
        width: 160,
        cellStyle: { textAlign: 'right', fontFamily: 'monospace', color: '#f43f5e' },
        valueFormatter: (params) => (params.value && params.value !== '0.000000' ? `৳ ${params.value}` : '-'),
      },
      {
        field: 'credit_base',
        headerName: 'Credit (Base BDT)',
        width: 160,
        cellStyle: { textAlign: 'right', fontFamily: 'monospace', color: '#10b981' },
        valueFormatter: (params) => (params.value && params.value !== '0.000000' ? `৳ ${params.value}` : '-'),
      },
      {
        field: 'currency',
        headerName: 'Curr',
        width: 80,
      },
      {
        field: 'remarks',
        headerName: 'Narration / Remarks',
        flex: 1,
        minWidth: 220,
      },
    ],
    []
  );

  const defaultColDef = useMemo<ColDef>(
    () => ({
      resizable: true,
      sortable: true,
    }),
    []
  );

  return (
    <div className="ag-theme-alpine-dark" style={{ height: '620px', width: '100%', borderRadius: '8px', overflow: 'hidden' }}>
      <AgGridReact<GLEntry>
        rowData={rowData}
        columnDefs={columnDefs}
        defaultColDef={defaultColDef}
        pagination={true}
        paginationPageSize={25}
        paginationPageSizeSelector={[25, 50, 100]}
        rowSelection="single"
        loading={loading}
      />
    </div>
  );
};
