import { Input } from '@grafana/ui'
import { useDebounce } from 'hooks/useDebounce'
import React, { useEffect, useState } from 'react'

interface Props {
  jsonPath: string
  onChange: (v: string) => void
}

/**
 * JsonPathQueryField is an editor for JSON Path.
 */
export const JsonPathField: React.FC<Props> = ({ jsonPath, onChange }) => {
  const [value, setValue] = useState(jsonPath)
  const debouncedValue = useDebounce(value, 500, jsonPath)

  useEffect(() => {
    if (debouncedValue === jsonPath) {
      return
    }

    onChange(debouncedValue || '')
  }, [debouncedValue, jsonPath, onChange])

  return (
    <Input
      value={value}
      onChange={e => setValue(e.currentTarget.value)}
      onBlur={() => onChange(value || '')}
      placeholder='$.items[*].name'
      aria-label='JsonPath editor'
    />
  )
}
