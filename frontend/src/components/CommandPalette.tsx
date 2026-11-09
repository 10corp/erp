import React, { useEffect, useState } from 'react';
import { Command } from 'cmdk';
import { Search, PlusCircle, BookOpen, FileText, Cpu, TrendingUp, Sparkles, CornerDownLeft } from 'lucide-react';

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
  onSelectAction: (action: string, payload?: any) => void;
}

export const CommandPalette: React.FC<CommandPaletteProps> = ({ isOpen, onClose, onSelectAction }) => {
  const [search, setSearch] = useState('');
  const [isProcessingAI, setIsProcessingAI] = useState(false);

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

  const handleExecuteAI = async () => {
    if (!search.trim()) return;
    setIsProcessingAI(true);
    try {
      const resp = await fetch('/api/v1/ai/action', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query: search }),
      });
      if (resp.ok) {
        const actionData = await resp.json();
        onSelectAction('ai_action', actionData);
        onClose();
      }
    } catch (err) {
      console.error('Claude Action Brain call failed:', err);
    } finally {
      setIsProcessingAI(false);
    }
  };

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
          width: '640px',
          maxWidth: '92vw',
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
              value={search}
              onValueChange={setSearch}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && search.length > 5) {
                  e.preventDefault();
                  handleExecuteAI();
                }
              }}
              placeholder="Ask Claude AI or jump to... (e.g. 'Show me overdue invoices for Acme and draft receipt')"
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
            {search.length > 3 && (
              <button
                onClick={handleExecuteAI}
                disabled={isProcessingAI}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '4px',
                  backgroundColor: '#0284c7',
                  border: 'none',
                  color: '#fff',
                  padding: '4px 8px',
                  borderRadius: '6px',
                  fontSize: '12px',
                  fontWeight: 600,
                  cursor: 'pointer',
                  marginLeft: '8px',
                }}
              >
                <Sparkles size={13} />
                <span>{isProcessingAI ? 'Thinking...' : 'Claude Action'}</span>
              </button>
            )}
          </div>

          <Command.List style={{ maxHeight: '340px', overflowY: 'auto', padding: '8px' }}>
            <Command.Empty style={{ padding: '16px', textAlign: 'center', color: '#94a3b8', fontSize: '13px' }}>
              <div>Press <b>Enter</b> to let Claude 3.5 Sonnet parse intent and execute action.</div>
            </Command.Empty>

            {search.length > 3 && (
              <Command.Group heading="🤖 Natural Language Intent Parser" style={{ padding: '6px 8px', fontSize: '11px', color: '#38bdf8', fontWeight: 600 }}>
                <Command.Item
                  onSelect={handleExecuteAI}
                  style={{ ...itemStyle, backgroundColor: '#0f172a', border: '1px solid #0284c7' }}
                >
                  <Sparkles size={16} color="#38bdf8" style={{ marginRight: '10px' }} />
                  <span>Execute natural language command: <i>"{search}"</i></span>
                  <CornerDownLeft size={14} color="#64748b" style={{ marginLeft: 'auto' }} />
                </Command.Item>
              </Command.Group>
            )}

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

              <Command.Item
                onSelect={() => { onSelectAction('new_delivery'); onClose(); }}
                style={itemStyle}
              >
                <PlusCircle size={16} color="#06b6d4" style={{ marginRight: '10px' }} />
                <span>Dispatch Delivery Note (Perpetual COGS Auto-Post)</span>
                <span style={badgeStyle}>DN</span>
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
