import { css } from '@emotion/css'
import { DataSourceSettings } from '@grafana/data'
import { Button, Icon, InlineField, Input, SecretInput } from '@grafana/ui'
import { uniqueId } from 'lodash'
import React, { PureComponent } from 'react'

export interface CustomQueryParam {
  id: string
  name: string
  value: string
  configured: boolean
}

export type CustomQueryParams = CustomQueryParam[]

export interface Props {
  dataSourceConfig: DataSourceSettings<any, any>
  onChange: (config: DataSourceSettings) => void
}

export interface State {
  queryParams: CustomQueryParams
}

interface CustomQueryParamRowProps {
  queryParam: CustomQueryParam
  onReset: (id: string) => void
  onRemove: (id: string) => void
  onChange: (value: CustomQueryParam) => void
  onBlur: () => void
}

const queryParamRowLayout = css`
  display: flex;
  align-items: center;
  margin-bottom: 4px;
  > * {
    margin-left: 4px;
    margin-bottom: 0;
    height: 100%;
    &:first-child,
    &:last-child {
      margin-left: 0;
    }
  }
`

const CustomQueryParamRow: React.FC<CustomQueryParamRowProps> = ({
  queryParam,
  onBlur,
  onChange,
  onRemove,
  onReset,
}) => {
  return (
    <div className={queryParamRowLayout}>
      <InlineField label='Key' labelWidth={8} grow>
        <Input
          value={queryParam.name || ''}
          placeholder='apiKey'
          onChange={e => onChange({ ...queryParam, name: e.currentTarget.value })}
          onBlur={onBlur}
        />
      </InlineField>
      <InlineField label='Value' labelWidth={5} grow>
        <SecretInput
          isConfigured={queryParam.configured}
          value={queryParam.value}
          width={queryParam.configured ? 11 : 12}
          placeholder='apiKeyValuexrCzlMqi6vJwlM5ijRgFL'
          onReset={() => onReset(queryParam.id)}
          onChange={e => onChange({ ...queryParam, value: e.currentTarget.value })}
          onBlur={onBlur}
        />
      </InlineField>
      <Button
        type='button'
        aria-label='Remove queryParam'
        variant='secondary'
        size='xs'
        onClick={() => onRemove(queryParam.id)}
      >
        <Icon name='trash-alt' />
      </Button>
    </div>
  )
}

CustomQueryParamRow.displayName = 'CustomQueryParamRow'

export class CustomQueryParamsSettings extends PureComponent<Props, State> {
  state: State = {
    queryParams: [],
  }

  constructor(props: Props) {
    super(props)
    const { jsonData, secureJsonData, secureJsonFields } = this.props.dataSourceConfig
    this.state = {
      queryParams: Object.keys(jsonData)
        .sort()
        .filter(key => key.startsWith('queryParamName'))
        .map((key, index) => {
          return {
            id: uniqueId(),
            name: jsonData[key],
            value: secureJsonData !== undefined ? secureJsonData[key] : '',
            configured: (secureJsonFields && secureJsonFields[`queryParamValue${index + 1}`]) || false,
          }
        }),
    }
  }

  updateSettings = () => {
    const { queryParams } = this.state

    // we remove every queryParamName* field
    const newJsonData = Object.fromEntries(
      Object.entries(this.props.dataSourceConfig.jsonData).filter(([key]) => !key.startsWith('queryParamName')),
    )

    // we remove every queryParamValue* field
    const newSecureJsonData = Object.fromEntries(
      Object.entries(this.props.dataSourceConfig.secureJsonData || {}).filter(
        ([key]) => !key.startsWith('queryParamValue'),
      ),
    )

    // then we add the current queryParam-fields
    for (const [index, queryParam] of queryParams.entries()) {
      newJsonData[`queryParamName${index + 1}`] = queryParam.name
      if (!queryParam.configured) {
        newSecureJsonData[`queryParamValue${index + 1}`] = queryParam.value
      }
    }

    this.props.onChange({
      ...this.props.dataSourceConfig,
      jsonData: newJsonData,
      secureJsonData: newSecureJsonData,
    })
  }

  onQueryParamAdd = () => {
    this.setState(prevState => {
      return {
        queryParams: [...prevState.queryParams, { id: uniqueId(), name: '', value: '', configured: false }],
      }
    })
  }

  onQueryParamChange = (queryParamIndex: number, value: CustomQueryParam) => {
    this.setState(({ queryParams }) => {
      return {
        queryParams: queryParams.map((item, index) => {
          if (queryParamIndex !== index) {
            return item
          }
          return { ...value }
        }),
      }
    })
  }

  onQueryParamReset = (queryParamId: string) => {
    this.setState(({ queryParams }) => {
      return {
        queryParams: queryParams.map(h => {
          if (h.id !== queryParamId) {
            return h
          }
          return {
            ...h,
            value: '',
            configured: false,
          }
        }),
      }
    })
  }

  onQueryParamRemove = (queryParamId: string) => {
    this.setState(
      ({ queryParams }) => ({
        queryParams: queryParams.filter(h => h.id !== queryParamId),
      }),
      this.updateSettings,
    )
  }

  render() {
    const { queryParams } = this.state
    return (
      <div className={'gf-form-group'}>
        <div className='gf-form'>
          <h6>Query Parameters</h6>
        </div>
        <div>
          {queryParams.map((queryParam, i) => (
            <CustomQueryParamRow
              key={queryParam.id}
              queryParam={queryParam}
              onChange={h => {
                this.onQueryParamChange(i, h)
              }}
              onBlur={this.updateSettings}
              onRemove={this.onQueryParamRemove}
              onReset={this.onQueryParamReset}
            />
          ))}
        </div>
        <div className='gf-form'>
          <Button
            variant='secondary'
            icon='plus'
            type='button'
            onClick={() => {
              this.onQueryParamAdd()
            }}
          >
            Add Parameter
          </Button>
        </div>
      </div>
    )
  }
}

export default CustomQueryParamsSettings
