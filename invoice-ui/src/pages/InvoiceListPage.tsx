import { useInvoices, useMarkPaid } from '../hooks/useInvoices'
import type { Invoice } from '../api/client'

const statusBadge: Record<Invoice['status'], string> = {
  paid: 'badge badge-green',
  overdue: 'badge badge-red',
  sent: 'badge badge-amber',
  draft: 'badge badge-gray',
}

export function InvoiceListPage() {
  const { data: invoices, isLoading, isError, error } = useInvoices()
  const markPaid = useMarkPaid()

  if (isLoading) return <p>Loading invoices…</p>
  if (isError) return <p>Error: {(error as Error).message}</p>

  return (
    <div className="page">
      <h1>Invoices</h1>
      <table className="invoice-table">
        <thead>
          <tr>
            <th>Customer</th>
            <th>Amount</th>
            <th>Due Date</th>
            <th>Status</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {invoices?.map((inv) => (
            <tr key={inv.id}>
              <td>{inv.customer_name}</td>
              <td className="col-amount">{inv.amount.toFixed(2)} {inv.currency}</td>
              <td>{new Date(inv.due_date).toLocaleDateString()}</td>
              <td>
                <span className={statusBadge[inv.status] ?? 'badge badge-gray'}>
                  {inv.status}
                </span>
              </td>
              <td className="col-action">
                {inv.status !== 'paid' && (
                  <button
                    className="btn-pay"
                    onClick={() => markPaid.mutate(inv.id)}
                    disabled={markPaid.isPending}
                  >
                    Mark as Paid
                  </button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
