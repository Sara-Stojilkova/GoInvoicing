import { useState } from 'react'
import { useInvoices, useSummary, useCreateInvoice, useMarkPaid } from '../hooks/useInvoices'
import { SummaryCards } from '../components/SummaryCards'
import type { Invoice } from '../api/client'

const statusBadge: Record<Invoice['status'], string> = {
  paid:    'badge badge-green',
  overdue: 'badge badge-red',
  sent:    'badge badge-amber',
  draft:   'badge badge-gray',
}

const emptyForm = { customer_name: '', amount: '', currency: 'USD', due_date: '' }

export function InvoiceListPage() {
  const { data: invoices, isLoading, isError, error } = useInvoices()
  const { data: summary } = useSummary()
  const markPaid      = useMarkPaid()
  const createInvoice = useCreateInvoice()

  const [activeFilter, setActiveFilter] = useState<Invoice['status'] | null>(null)
  const [showForm, setShowForm]         = useState(false)
  const [form, setForm]                 = useState(emptyForm)

  const displayed = activeFilter
    ? invoices?.filter((inv) => inv.status === activeFilter)
    : invoices

  function handleCreate(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    createInvoice.mutate(
      {
        customer_name: form.customer_name,
        amount:        parseFloat(form.amount),
        currency:      form.currency,
        due_date:      new Date(form.due_date).toISOString(),
      },
      {
        onSuccess: () => {
          setShowForm(false)
          setForm(emptyForm)
        },
      },
    )
  }

  if (isLoading) return <p style={{ padding: '32px' }}>Loading invoices…</p>
  if (isError)   return <p style={{ padding: '32px' }}>Error: {(error as Error).message}</p>

  return (
    <div className="page">
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <h1 style={{ margin: '32px 0' }}>Invoices</h1>
        <button className="btn-pay" onClick={() => setShowForm((v) => !v)}>
          {showForm ? 'Cancel' : '+ New Invoice'}
        </button>
      </div>

      <SummaryCards
        summary={summary}
        activeFilter={activeFilter}
        onFilter={setActiveFilter}
      />

      {showForm && (
        <form className="create-form" onSubmit={handleCreate}>
          <div className="create-form__field">
            <label htmlFor="customer_name">Customer</label>
            <input
              id="customer_name"
              required
              value={form.customer_name}
              onChange={(e) => setForm({ ...form, customer_name: e.target.value })}
            />
          </div>
          <div className="create-form__field">
            <label htmlFor="amount">Amount</label>
            <input
              id="amount"
              type="number"
              min="0.01"
              step="0.01"
              required
              value={form.amount}
              onChange={(e) => setForm({ ...form, amount: e.target.value })}
            />
          </div>
          <div className="create-form__field">
            <label htmlFor="currency">Currency</label>
            <input
              id="currency"
              required
              value={form.currency}
              onChange={(e) => setForm({ ...form, currency: e.target.value })}
            />
          </div>
          <div className="create-form__field">
            <label htmlFor="due_date">Due Date</label>
            <input
              id="due_date"
              type="date"
              required
              value={form.due_date}
              onChange={(e) => setForm({ ...form, due_date: e.target.value })}
            />
          </div>
          <button className="btn-pay" type="submit" disabled={createInvoice.isPending}>
            {createInvoice.isPending ? 'Saving…' : 'Create'}
          </button>
        </form>
      )}

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
          {displayed?.map((inv) => (
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
          {displayed?.length === 0 && (
            <tr>
              <td colSpan={5} style={{ textAlign: 'center', color: 'var(--text)' }}>
                No invoices{activeFilter ? ` with status "${activeFilter}"` : ''}.
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  )
}
