import { Button, Input } from '@grafana/ui'
import React from 'react'

interface Props {
  values: Array<[string, string]>
  onChange: (rows: Array<[string, string]>) => void
}

export const QueryParamsEditor: React.FC<Props> = ({ values, onChange }) => {
  const updateRow = (idx: number, col: 0 | 1, val: string) => {
    onChange(
      values.map((row, i) => {
        if (i !== idx) {
          return row
        }
        const next: [string, string] = [...row] as [string, string]
        next[col] = val
        return next
      }),
    )
  }

  const addRow = () => {
    onChange([...values, ['', '']])
  }

  const removeRow = (idx: number) => {
    const next = [...values]
    next.splice(idx, 1)
    onChange(next)
  }

  if (values.length === 0) {
    return (
      <Button variant='secondary' onClick={() => onChange([['', '']])} icon='plus'>
        Add query parameter
      </Button>
    )
  }

  return (
    <table style={{ width: '100%', marginTop: '8px' }}>
      <thead>
        <tr>
          <th style={{ textAlign: 'left' }}>Key</th>
          <th style={{ textAlign: 'left' }}>Value</th>
          <th />
        </tr>
      </thead>
      <tbody>
        {values.map(([k, v], idx) => (
          <tr key={idx}>
            <td>
              <Input
                value={k}
                onChange={e => updateRow(idx, 0, e.currentTarget.value)}
                placeholder='param'
                aria-label='Query parameter key'
                style={{ width: '100%' }}
              />
            </td>
            <td>
              <Input
                value={v}
                onChange={e => updateRow(idx, 1, e.currentTarget.value)}
                placeholder='value'
                aria-label='Query parameter value'
                style={{ width: '100%' }}
              />
            </td>
            <td>
              <div style={{ display: 'flex', gap: 4 }}>
                <Button
                  variant='destructive'
                  size='sm'
                  icon='trash-alt'
                  aria-label='Remove query parameter'
                  onClick={() => removeRow(idx)}
                />
              </div>
            </td>
          </tr>
        ))}
      </tbody>
      <tfoot>
        <tr>
          <td colSpan={3}>
            <Button
              variant='secondary'
              size='sm'
              icon='plus'
              aria-label='Add query parameter'
              onClick={addRow}
            >
              Add query parameter
            </Button>
          </td>
        </tr>
      </tfoot>
    </table>
  )
}
