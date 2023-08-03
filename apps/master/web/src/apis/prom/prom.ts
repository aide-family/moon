export type Map<T = string> = { [key: string]: T }

export type RuleItem = {
    group_id: number
    alert: string
    expr: string
    for: string
    labels: Map
    annotations: Map
    created_at: string
    updated_at: string
    id: number
}

export type GroupItem = {
    name: string
    createdAt: string
    updatedAt: string
    id: number
    file_id: number
    rules: RuleItem[]
}

export type FileItem = {
    filename: string
    dirId: number
    created_at: string
    updated_at: string
    id: number
    groups: GroupItem[]
}

export type DirItem = {
    nodeId: number
    path: string
    created_at: string
    updated_at: string
    id: number
    files: FileItem[]
}

export type NodeItem = {
    en_name: string
    cn_name: string
    datasource: string
    remark: string
    created_at: string
    updated_at: string
    id: number
    dirs: DirItem[]
}