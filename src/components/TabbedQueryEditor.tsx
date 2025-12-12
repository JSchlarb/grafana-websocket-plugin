import { TimeRange } from '@grafana/data'
import { Field, InlineField, InlineFieldRow, RadioButtonGroup } from '@grafana/ui'
import { DataSource } from 'datasource'
import defaults from 'lodash/defaults'
import React, { useState } from 'react'
import { defaultQuery, Query } from '../types'
import { PathField } from './fields/PathField'
import { QueryParamsEditor } from './fields/QueryParamsEditor'

interface Props {
  onChange: (query: Query) => void
  onRunQuery: () => void
  editorContext: string
  query: Query
  limitFields?: number
  datasource: DataSource
  range?: TimeRange
  fieldsTab: React.ReactNode
}

export const TabbedQueryEditor = ({ query, onChange, onRunQuery, fieldsTab }: Props) => {
  const [tabIndex, setTabIndex] = useState(0)

  const q = defaults(query, defaultQuery)

  const onChangePath = (value: string) => {
    onChange({ ...q, path: value })
    onRunQuery()
  }

  const tabs = [
    {
      title: 'Fields',
      content: fieldsTab,
    },
    {
      title: 'Path',
      content: (
        <>
          <Field label='Path' description='Websocket URL path to connect'>
            <PathField path={q.path} onChange={onChangePath} />
          </Field>
          <Field label='Query parameters' description='Optional query string parameters for the websocket connection'>
            <QueryParamsEditor
              values={Array.isArray(q.queryParams) ? q.queryParams : Object.entries(q.queryParams || {})}
              onChange={qp => {
                onChange({ ...q, queryParams: qp })
                onRunQuery()
              }}
            />
          </Field>
        </>
      ),
    },
  ]

  return (
    <>
      <InlineFieldRow>
        <InlineField>
          <RadioButtonGroup
            onChange={e => setTabIndex(e ?? 0)}
            value={tabIndex}
            options={tabs.map((tab, idx) => ({ label: tab.title, value: idx }))}
          />
        </InlineField>
      </InlineFieldRow>
      {tabs[tabIndex].content}
    </>
  )
}
