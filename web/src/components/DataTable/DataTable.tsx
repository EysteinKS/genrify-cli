// Generic sortable data table component

import { useState } from 'react'
import styles from './DataTable.module.css'

export interface Column<T> {
  key: string
  header: string
  render: (item: T) => React.ReactNode
  sortable?: boolean
}

interface DataTableProps<T> {
  columns: Column<T>[]
  data: T[]
  keyExtractor: (item: T) => string
  onRowClick?: (item: T) => void
  emptyMessage?: string
}

export function DataTable<T>({
  columns,
  data,
  keyExtractor,
  onRowClick,
  emptyMessage = 'No data',
}: DataTableProps<T>) {
  const [sortKey, setSortKey] = useState<string | null>(null)
  const [sortAsc, setSortAsc] = useState(true)

  const handleSort = (key: string) => {
    if (sortKey === key) {
      setSortAsc(!sortAsc)
    } else {
      setSortKey(key)
      setSortAsc(true)
    }
  }

  const sortedData = [...data]
  if (sortKey) {
    const col = columns.find((c) => c.key === sortKey)
    if (col) {
      sortedData.sort((a, b) => {
        const aVal = String(col.render(a))
        const bVal = String(col.render(b))
        const cmp = aVal.localeCompare(bVal)
        return sortAsc ? cmp : -cmp
      })
    }
  }

  if (data.length === 0) {
    return <div className={styles.empty}>{emptyMessage}</div>
  }

  return (
    <div className={styles.tableContainer}>
      <table className={styles.table}>
        <thead>
          <tr>
            {columns.map((col) => (
              <th
                key={col.key}
                className={col.sortable ? styles.sortable : ''}
                onClick={() => col.sortable && handleSort(col.key)}
              >
                {col.header}
                {col.sortable && sortKey === col.key && (
                  <span className={styles.sortIcon}>{sortAsc ? ' ▲' : ' ▼'}</span>
                )}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {sortedData.map((item) => (
            <tr
              key={keyExtractor(item)}
              className={onRowClick ? styles.clickable : ''}
              onClick={() => onRowClick?.(item)}
            >
              {columns.map((col) => (
                <td key={col.key}>{col.render(item)}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
