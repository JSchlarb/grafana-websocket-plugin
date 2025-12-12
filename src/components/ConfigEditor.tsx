import { DataSourcePluginOptionsEditorProps } from '@grafana/data'
import { InlineField, Input } from '@grafana/ui'
import React, { ChangeEvent, FC } from 'react'
import { DataSourceOptions } from '../types'
import CustomHeadersSettings from './fields/CustomHeadersField'
import CustomQueryParamsSettings from './fields/CustomQueryParamsField'

type Props = DataSourcePluginOptionsEditorProps<DataSourceOptions>

export const ConfigEditor: FC<Props> = ({ options, onOptionsChange }) => {
  const { url, jsonData } = options

  const onHostChange = (event: ChangeEvent<HTMLInputElement>) => {
    const newOptions = {
      ...options,
      url: event.currentTarget.value,
    }
    onOptionsChange({ ...newOptions, jsonData })
  }

  return (
    <div className='gf-form-group'>
      <h3 className='page-heading'>WebSocket</h3>
      <div className='gf-form-group'>
        <InlineField label='Host' labelWidth={10} grow>
          <Input
            id='config-editor-host'
            onChange={onHostChange}
            value={url || ''}
            placeholder='wss://api.domain.io/v1/ws/'
            width={30}
          />
        </InlineField>
      </div>

      <CustomHeadersSettings dataSourceConfig={options} onChange={onOptionsChange} />
      <CustomQueryParamsSettings dataSourceConfig={options} onChange={onOptionsChange} />
    </div>
  )
}
