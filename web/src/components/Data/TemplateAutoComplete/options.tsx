export const autoCompleteOption: { value: string }[] = [
    {
        value: `{{ $value }}`
    },
    {
        value: `{{ $labels }}`
    },
    {
        value: `{{ $annotations }}`
    },
    {
        value: `{{ $annotations.summary }}`
    },
    {
        value: `{{ $annotations.description }}`
    },
    {
        value: `{{ $labels.sverity }}`
    },
    {
        value: `{{ $labels.instance }}`
    },
    {
        value: `{{ $labels.alertname }}`
    },
    {
        value: `{{ $labels.app }}`
    },
    {
        value: `{{ $labels.namespace }}`
    },
    {
        value: `{{ $labels.node }}`
    },
    {
        value: `{{ $group_name }}`
    }
]
