export type FileItem = {
    filename: string
    dirId: number
    createdAt: string
    updatedAt: string
    id: number
}

export type DirItem = {
    nodeId: number
    path: string
    createdAt: string
    updatedAt: string
    id: number
    files: FileItem[]
}

export type NodeItem = {
    enName: string
    cnName: string
    datasource: string
    remark: string
    createdAt: string
    updatedAt: string
    id: number
    dirs: DirItem[]
}