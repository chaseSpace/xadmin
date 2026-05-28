import { Table } from 'antd'
import type { TableProps } from 'antd'

export type UiTableProps<T extends object> = TableProps<T>

export function UiTable<T extends object>(props: UiTableProps<T>) {
  return <Table<T> {...props} />
}
