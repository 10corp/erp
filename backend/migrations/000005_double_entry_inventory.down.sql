-- 000005_double_entry_inventory.down.sql
DROP TABLE IF EXISTS delivery_note_items CASCADE;
DROP TABLE IF EXISTS delivery_notes CASCADE;
DROP TABLE IF EXISTS stock_ledger_entries CASCADE;
DROP TABLE IF EXISTS item_warehouse_balances CASCADE;
DROP TABLE IF EXISTS warehouses CASCADE;
DROP TYPE IF EXISTS warehouse_type_enum CASCADE;
