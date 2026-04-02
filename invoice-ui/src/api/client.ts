export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })
  if (!res.ok) {
    const body = await res.text()
    let message = body || res.statusText
    try { message = JSON.parse(body).error } catch {}
    throw new ApiError(res.status, message)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export interface Invoice {
  id: string
  customer_name: string
  amount: number
  currency: string
  issued_at: string
  due_date: string
  paid_at: string | null
  status: 'draft' | 'sent' | 'paid' | 'overdue'
}

export interface CreateInvoiceInput {
  customer_name: string
  amount: number
  currency: string
  due_date: string
}

export interface Summary {
  paid_count: number
  unpaid_count: number
  overdue_count: number
  total_amount: number
}

export function fetchInvoices(): Promise<Invoice[]> {
  return request<Invoice[]>('/api/invoices/')
}

export function createInvoice(input: CreateInvoiceInput): Promise<Invoice> {
  return request<Invoice>('/api/invoices/', { method: 'POST', body: JSON.stringify(input) })
}

export function markInvoicePaid(id: string): Promise<void> {
  return request<void>(`/api/invoices/${id}/pay`, { method: 'POST' })
}

export function fetchSummary(): Promise<Summary> {
  return request<Summary>('/api/invoices/summary')
}
