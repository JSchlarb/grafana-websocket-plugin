import { DataSourceJsonData, FieldType } from '@grafana/data'
import { DataQuery } from '@grafana/schema'

export type QueryLanguage = 'jsonpath'

export interface QueryField {
  name?: string
  jsonPath: string
  type?: FieldType
  language?: QueryLanguage
}

export type Pair<T, K> = [T, K]

export interface Query extends DataQuery {
  path: string
  fields: QueryField[]
  queryParams?: Record<string, string> | Array<Pair<string, string>>
}

export const defaultQuery: Partial<Query> = {
  path: '',
  fields: [{ jsonPath: '', language: 'jsonpath', name: '' }],
  queryParams: [],
}

export interface DataSourceOptions extends DataSourceJsonData {
  url?: string
}

export interface SecureJsonData {
  apiKey?: string
}
