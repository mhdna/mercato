// Reasons the append-only stock ledger records a movement under. Used to
// populate the Stock Movements filter; the ledger itself is read through
// ServerSideTable against GET /stock_movements.
export const STOCK_MOVEMENT_REASONS = [
  'purchase',
  'sale',
  'return',
  'transfer_out',
  'transfer_in',
  'adjustment',
  'count',
]
