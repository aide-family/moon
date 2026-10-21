export interface AlertRuleGroupItem {
  uid: string
  name: string
  status: string
}

export interface AlertRuleLevelItem {
  uid: string
  condition: string
  sample: string
  threshold: string
  forSeconds: number
  levelUid: string
  levelName: string
}

export interface AlertRuleLevelInput {
  condition: string
  sample: string
  threshold: string
  forSeconds: number
  levelUid: string
}

export interface AlertRuleItem {
  uid: string
  groupUid: string
  name: string
  driver: string
  status: string
  expr: string
  labels: string
  datasourceUid: string
  datasourceLabels: string
  summary: string
  description: string
  levels: AlertRuleLevelItem[]
}

export interface ListAlertRuleGroupsParams {
  page?: number
  pageSize?: number
  name?: string
}

export interface ListAlertRuleGroupsResult {
  items: AlertRuleGroupItem[]
  total: number
  page: number
  pageSize: number
}

export interface ListAlertRulesParams {
  page?: number
  pageSize?: number
  groupUid?: string
  name?: string
}

export interface ListAlertRulesResult {
  items: AlertRuleItem[]
  total: number
  page: number
  pageSize: number
}

export interface SaveAlertRuleParams {
  uid?: string
  groupUid: string
  name: string
  driver: string
  expr?: string
  labels?: string
  datasourceUid?: string
  datasourceLabels?: string
  summary?: string
  description?: string
  levels: AlertRuleLevelInput[]
}
