import { formatMoney } from '@/utils/money'

// Which admin-socket message types surface as branch activity (a toast or
// the app-bar ticker). Mirrors the message types documented on
// adminWSMessage in api/admin_ws.go.
export const BRANCH_ACTIVITY_TYPES = [
  'branch_invoice_created',
  'branch_expense_created',
  'branch_loan_created',
  'branch_shift_closed',
  'branch_settlement_changed',
  'branch_attendance_event',
  'branch_attendance_events',
  'branch_attendance_changed',
  'branch_visitor_event',
]

export function isBranchActivityMessage (message) {
  return BRANCH_ACTIVITY_TYPES.includes(message?.type)
}

function amountColor (amount) {
  if (amount > 0) {
    return 'success'
  }
  if (amount < 0) {
    return 'error'
  }
  return 'warning'
}

function invoiceKindWord (message) {
  if (message.type === 'branch_expense_created') {
    return 'expense'
  }
  if (message.type === 'branch_loan_created') {
    return 'loan'
  }
  if (message.kind === 'exchange') {
    return 'exchange'
  }
  if (message.kind === 'return') {
    return 'return'
  }
  if (message.kind) {
    return 'revenue'
  }
  return message.amount < 0 ? 'return' : 'revenue'
}

function invoiceIcon (message) {
  if (message.type === 'branch_expense_created') {
    return 'mdi-cash-minus'
  }
  if (message.type === 'branch_loan_created') {
    // Borrowed money coming in -- a hand passing a coin, not a refund.
    return 'mdi-hand-coin'
  }
  if (message.kind === 'exchange') {
    return 'mdi-swap-horizontal'
  }
  if (message.kind === 'return') {
    // Money refunded to the customer.
    return 'mdi-cash-refund'
  }
  return 'mdi-sale'
}

// The manager-approval workflow has three outcomes that read very
// differently -- a filed request, an approved edit, a rejected one. The
// backend sends `kind` as "complaint" | "approved" | "rejected" (see
// attendanceChangeKind in api/branch_attendance.go).
const ATTENDANCE_CHANGE_ICON = {
  complaint: 'mdi-clock-edit-outline',
  approved: 'mdi-clock-check-outline',
  rejected: 'mdi-clock-remove-outline',
}

function attendanceChangeIcon (kind) {
  return ATTENDANCE_CHANGE_ICON[kind] ?? 'mdi-clock-edit-outline'
}

// describeBranchActivity turns one admin-socket message into the shape both
// the toast (BranchActivityToast) and the app-bar ticker (SyncCard) render:
//
//   text      one-line summary, branch name included
//   color     'success' | 'error' | 'warning' | 'info'
//   trendIcon triangle up/down, for the money types
//   icon      leading mdi glyph
//   amount    signed cents when the message carries money, else undefined
//   money     true when the toast should use its amount+currency+kind layout
//
// Returns null for a type that isn't branch activity.
export function describeBranchActivity (message, branchName) {
  const branch = branchName(message.branch_id)

  switch (message.type) {
    case 'branch_invoice_created':
    case 'branch_expense_created':
    case 'branch_loan_created': {
      const kind = invoiceKindWord(message)
      return {
        text: `${branch}: ${formatMoney(message.amount)} ${message.currency_code ?? ''} ${kind}`.replace(/\s+/g, ' ').trim(),
        color: amountColor(message.amount),
        trendIcon: message.amount < 0 ? 'mdi-triangle-down' : 'mdi-triangle',
        icon: invoiceIcon(message),
        amount: message.amount,
        currencyCode: message.currency_code,
        kindWord: kind,
        money: true,
      }
    }

    case 'branch_shift_closed': {
      const balanced = !message.amount
      const detail = balanced
        ? 'balanced'
        : `${formatMoney(message.amount)} ${message.currency_code ?? ''}`.trim()
      const who = message.label ? ` by ${message.label}` : ''
      return {
        text: `${branch}: shift closed${who} (${detail})`,
        color: balanced ? 'success' : amountColor(message.amount),
        trendIcon: message.amount < 0 ? 'mdi-triangle-down' : 'mdi-triangle',
        icon: 'mdi-cash-register',
        amount: message.amount,
        money: false,
      }
    }

    case 'branch_settlement_changed': {
      return {
        text: `${branch}: settlement changed${message.label ? ` on ${message.label}` : ''}`,
        color: 'info',
        icon: 'mdi-cash-edit',
        money: false,
      }
    }

    case 'branch_attendance_event': {
      return {
        text: `${branch}: attendance${message.label ? ` — ${message.label}` : ''}`,
        color: 'info',
        icon: 'mdi-fingerprint',
        money: false,
      }
    }

    case 'branch_attendance_events': {
      const n = message.amount ?? 0
      return {
        text: `${branch}: ${n} attendance event${n === 1 ? '' : 's'}`,
        color: 'info',
        icon: 'mdi-fingerprint',
        money: false,
      }
    }

    case 'branch_attendance_changed': {
      const verb = {
        complaint: 'time-change requested',
        approved: 'time change approved',
        rejected: 'time change rejected',
      }[message.kind] ?? 'attendance change'
      return {
        text: `${branch}: ${verb}${message.label ? ` — ${message.label}` : ''}`,
        color: message.kind === 'rejected' ? 'warning' : 'info',
        icon: attendanceChangeIcon(message.kind),
        money: false,
      }
    }

    case 'branch_visitor_event': {
      // kind is the counter direction: "in" (▲) or "out" (▼).
      const isOut = message.kind === 'out'
      return {
        text: `${branch}: customer ${isOut ? 'out' : 'in'}`,
        color: 'info',
        icon: isOut ? 'mdi-account-arrow-left' : 'mdi-account-arrow-right',
        money: false,
      }
    }

    default: {
      return null
    }
  }
}
