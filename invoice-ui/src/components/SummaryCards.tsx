import type { Summary, Invoice } from '../api/client'

interface Props {
  summary: Summary | undefined
  activeFilter: Invoice['status'] | null
  onFilter: (status: Invoice['status'] | null) => void
}

interface CardDef {
  label: string
  value: number | string | undefined
  status: Invoice['status'] | null
  colorClass: string
}

export function SummaryCards({ summary, activeFilter, onFilter }: Props) {
  const cards: CardDef[] = [
    { label: 'Paid',    value: summary?.paid_count,    status: 'paid',    colorClass: 'card-green' },
    { label: 'Unpaid',  value: summary?.unpaid_count,  status: 'sent',    colorClass: 'card-amber' },
    { label: 'Overdue', value: summary?.overdue_count, status: 'overdue', colorClass: 'card-red'   },
    { label: 'Total',   value: summary ? `$${summary.total_amount.toFixed(2)}` : undefined, status: null, colorClass: 'card-gray' },
  ]

  return (
    <div className="summary-cards">
      {cards.map((card) => (
        <button
          key={card.label}
          className={[
            'summary-card',
            card.colorClass,
            card.status && activeFilter === card.status ? 'summary-card--active' : '',
          ].join(' ')}
          onClick={() => onFilter(activeFilter === card.status ? null : card.status)}
          disabled={card.status === null}
        >
          <span className="summary-card__value">{card.value ?? '—'}</span>
          <span className="summary-card__label">{card.label}</span>
        </button>
      ))}
    </div>
  )
}
