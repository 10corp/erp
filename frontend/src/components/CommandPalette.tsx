import React, { useEffect, useState } from 'react';
import { Command } from 'cmdk';
import { Search, PlusCircle, BookOpen, FileText, Cpu, TrendingUp } from 'lucide-react';

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
  onSelectAction: (action: string) => void;
}

export const CommandPalette: React.FC<CommandPaletteProps> = ({ isOpen, onClose, onSelectAction }) => {
  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if ((e.key === 'k' && (e.metaKey || e.ctrlKey))) {
        e.preventDefault();
        if (isOpen) {
          onClose();
        } else {
          onSelectAction('toggle_palette');
        }
      }
    };
    document.addEventListener('keydown', down);
    return () => document.removeEventListener('keydown', down);
  }, [isOpen, onClose, onSelectAction]);

  if (!isOpen) return null;

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        backgroundColor: 'rgba(0, 0, 0, 0.65)',
        backdropFilter: 'blur(4px)',
        zIndex: 1000,
        display: 'flex',
        alignItems: 'flex-start',
        justifyContent: 'center',
        paddingTop: '15vh',
      }}
      onClick={onClose}
    >
      <div
        style={{
          width: '600px',
          maxWidth: '90vw',
          backgroundColor: '#1e293b',
          borderRadius: '12px',
          boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)',
          border: '1px solid #334155',
          overflow: 'hidden',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <Command label="Global Command Palette">
          <div style={{ display: 'flex', alignItems: 'center', padding: '12px 16px', borderBottom: '1px solid #334155' }}>
            <Search size={18} color="#94a3b8" style={{ marginRight: '10px' }} />
            <Command.Input
              placeholder="Type a command, search accounts, or ask AI (e.g. 'new invoice', 'p&l')..."
              style={{
                width: '100%',
                background: 'transparent',
                border: 'none',
                outline: 'none',
                color: '#f8fafc',
                fontSize: '15px',
                fontFamily: 'Inter, sans-serif',
              }}
              autoFocus
            />
          </div>

          <Command.List style={{ maxHeight: '320px', overflowY: 'auto', padding: '8px' }}>
            <Command.Empty style={{ padding: '16px', textAlign: 'center', color: '#64748b', fontSize: '13px' }}>
              No matching commands or documents found.
            </Command.Empty>

            <Command.Group heading="⚡ Rapid Quick Actions" style={{ padding: '6px 8px', fontSize: '11px', color: '#64748b', fontWeight: 600 }}>
              <Command.Item
                onSelect={() => { onSelectAction('new_journal'); onClose(); }}
                style={itemStyle}
              >
                <PlusCircle size={16} color="#38bdf8" style={{ marginRight: '10px' }} />
                <span>New Journal Entry (Double-Entry Reversal)</span>
                <span style={badgeStyle}>JV</span>
              </Command.Item>

              <Command.Item
                onSelect={() => { onSelectAction('new_invoice'); onClose(); }}
                style={itemStyle}
              >
                <FileText size={16} color="#34d399" style={{ marginRight: '10px' }} />
                <span>Create Sales Invoice</span>
                <span style={badgeStyle}>INV</span>
              </Command.Item>

              <Command.Item
                onSelect={() => { onSelectAction('new_payment'); onClose(); }}
                style={itemStyle}
              >
                <TrendingUp size={16} color="#f59e0b" style={{ marginRight: '10px' }} />
                <span>Record Customer Payment (Forex Auto-Reconciled)</span>
                <span style={badgeStyle}>PAY</span>
              </Command.Item>
            </Command.Group>

            <Command.Group heading="📊 Reports & Ledgers" style={{ padding: '6px 8px', fontSize: '11px', color: '#64748b', fontWeight: 600 }}>
              <Command.Item
                onSelect={() => { onSelectAction('view_ledger'); onClose(); }}
                style={itemStyle}
              >
                <BookOpen size={16} color="#fbbf24" style={{ marginRight: '10px' }} />
                <span>Partitioned General Ledger (Real-Time)</span>
              </Command.Item>

              <Command.Item
                onSelect={() => { onSelectAction('view_pnl'); onClose(); }}
                style={itemStyle}
              >
                <TrendingUp size={16} color="#a78bfa" style={{ marginRight: '10px' }} />
                <span>Profit & Loss (CQRS Materialized View)</span>
              </Command.Item>
            </Command.Group>

            <Command.Group heading="🤖 AI Brain Commands" style={{ padding: '6px 8px', fontSize: '11px', color: '#64748b', fontWeight: 600 }}>
              <Command.Item
                onSelect={() => { onSelectAction('scan_receipt'); onClose(); }}
                style={itemStyle}
              >
                <Cpu size={16} color="#f43f5e" style={{ marginRight: '10px' }} />
                <span>Scan Receipt / Invoice PDF (Gemini 1.5 Flash IDP)</span>
                <span style={{ ...badgeStyle, backgroundColor: '#831843' }}>AI</span>
              </Command.Item>
            </Command.Group>
          </Command.List>
        </Command>
      </div>
    </div>
  );
};

const itemStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  padding: '10px 12px',
  borderRadius: '8px',
  color: '#e2e8f0',
  fontSize: '13px',
  cursor: 'pointer',
  transition: 'background-color 0.1s ease',
};

const badgeStyle: React.CSSProperties = {
  marginLeft: 'auto',
  fontSize: '11px',
  backgroundColor: '#334155',
  padding: '2px 6px',
  borderRadius: '4px',
  fontFamily: 'monospace',
};
