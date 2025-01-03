{{- $status := .status -}}
{{- $labels := .labels -}}
{{- $annotations := .annotations -}}

{
"msgtype": "markdown",
"markdown": {
"title": "平台状态通知",
"text": "### {{if eq $status `resolved`}}✅ 告警已恢复{{else}}🚨 紧急告警通知{{end}}\n\n  \n**时间**: `{{ .startsAt }}` 至 `{{ .endsAt }}`  \n\n<hr/>\n\n**摘要**:  \n`{{ $annotations.summary }}`  \n\n**描述**:  \n`{{ $annotations.description }}`  \n\n<hr/>\n\n**标签**:  \n- **数据源 ID**: {{ index $labels "__moon__datasource_id__" }}  \n- **数据源 URL**: [链接]({{ index $labels "__moon__datasource_url__" }})  \n- **级别 ID**: {{ index $labels "__moon__level_id__" }}  \n- **策略 ID**: {{ index $labels "__moon__strategy_id__" }}  \n- **团队 ID**: {{ index $labels "__moon__team_id__" }}  \n- **实例**: `{{ index $labels "instance" }}`  \n- **IP**: `{{ index $labels "ip" }}`  \n- **作业**: `{{ index $labels "job" }}`  \n\n<hr/>\n\n请根据以上信息进行后续处理！"
}
}