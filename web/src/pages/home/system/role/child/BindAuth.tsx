import authApi from '@/apis/home/system/auth'
import {
    ApiAuthSelectItem,
    ApiAuthTreeItem
} from '@/apis/home/system/auth/types'
import roleApi from '@/apis/home/system/role'
import { RoleListItem } from '@/apis/home/system/role/types'
import { StatusMap } from '@/apis/types'
import {
    Avatar,
    Badge,
    Button,
    Descriptions,
    Drawer,
    DrawerProps,
    Space,
    Tooltip,
    Tree,
    Typography,
    message
} from 'antd'
import { DescriptionsItemType } from 'antd/es/descriptions'
import dayjs from 'dayjs'
import React, { useEffect, useState } from 'react'
import { UserAvatar } from '../../user/child/UserAvatar'

export interface BindAuthProps extends DrawerProps {
    roleId?: number
}

const { Title } = Typography

const { roleDetail, roleRelateApi } = roleApi
const { authApiTree } = authApi

export const BindAuth: React.FC<BindAuthProps> = (props) => {
    const { roleId, onClose, open } = props

    const [loading, setLoading] = useState(false)
    const [submitLoading, setSubmitLoading] = useState<boolean>(false)
    const [roleApis, setRoleApis] = useState<ApiAuthSelectItem[]>([])
    const [authTree, setAuthTree] = useState<ApiAuthTreeItem[]>([])
    const [roleInfo, setRoleInfo] = useState<RoleListItem>()
    const [defaultCheckedKeys, setDefaultCheckedKeys] = useState<React.Key[]>(
        []
    )
    const [defaultExpandedKeys, setDefaultExpandedKeys] = useState<React.Key[]>(
        []
    )
    const [checkedKeys, setCheckedKeys] = useState<number[]>([])

    const fetchRoleDetail = async () => {
        if (!roleId) return
        setLoading(true)
        const {
            detail: {
                apis,
                name,
                remark,
                id,
                status,
                createdAt,
                updatedAt,
                deletedAt,
                users
            }
        } = await roleDetail({ id: roleId })
        setRoleInfo({
            name,
            remark,
            id,
            status,
            createdAt,
            updatedAt,
            deletedAt,
            users
        })
        if (!apis) return
        const selectKeys = apis.map((api) => {
            return api.value
        })
        setDefaultCheckedKeys(selectKeys)
        setCheckedKeys(selectKeys)
        setDefaultExpandedKeys(selectKeys)
        setRoleApis(apis)
        setLoading(false)
    }

    const fetchAuthTree = async () => {
        setLoading(true)
        const { tree } = await authApiTree()
        if (!tree) return
        setAuthTree(tree)
        setLoading(false)
    }

    const handleAuthConfig = (e: any) => {
        if (!roleId) return
        setSubmitLoading(true)
        roleRelateApi({ id: roleId, apiIds: checkedKeys })
            .then(() => {
                onClose?.(e)
                message.success('分配权限成功')
            })
            .finally(() => {
                setSubmitLoading(false)
            })
    }

    const buildTreeData = () => {
        const treeData = authTree.map((item) => {
            const { domain, domainName, module } = item
            // const checked = roleApis.some((api) => api.value === domain)
            return {
                ...item,
                title: domainName,
                key: domain,
                // checked,
                children: module.map((child) => {
                    const { module, apis, name } = child
                    return {
                        ...child,
                        title: name,
                        key: module,
                        children: apis.map((api) => {
                            const checked = roleApis.some(
                                (item) => api.value === item.value
                            )

                            return {
                                ...api,
                                checked,
                                title: api.label,
                                key: api.value
                            }
                        })
                    }
                })
            }
        })
        return treeData
    }

    const buildRoleDescItems = (): DescriptionsItemType[] => {
        if (!roleInfo) return []
        const { color } = StatusMap[roleInfo?.status]
        return [
            {
                label: '角色名称',
                children: <Badge color={color} text={roleInfo?.name} />,
                span: 2
            },
            {
                label: '角色描述',
                children: roleInfo?.remark,
                span: 2
            },
            {
                label: '相关人员',
                children: (
                    <Avatar.Group shape="square" maxCount={10}>
                        {roleInfo?.users?.map((user, index) => {
                            const userItem = {
                                ...user,
                                id: user.value,
                                nickname: user.nickname,
                                username: user.nickname,
                                email: '',
                                phone: '',
                                status: user.status,
                                remark: '',
                                createdAt: 0,
                                updatedAt: 0,
                                deletedAt: 0,
                                gender: 0
                            }
                            return (
                                <Tooltip title={user.label} key={index}>
                                    <UserAvatar {...userItem} />
                                </Tooltip>
                            )
                        })}
                    </Avatar.Group>
                ),
                span: 2
            },
            {
                label: '创建时间',
                children: dayjs(+(roleInfo?.createdAt || 0) * 1000).format(
                    'YYYY-MM-DD HH:mm:ss'
                )
            },

            {
                label: '更新时间',
                children: dayjs(+(roleInfo?.updatedAt || 0) * 1000).format(
                    'YYYY-MM-DD HH:mm:ss'
                )
            }
        ]
    }

    const handleClose = (e: any) => {
        onClose?.(e)
    }

    const handleOnCheck = (checkedKeys: React.Key[] | any) => {
        if (!Array.isArray(checkedKeys)) return
        setCheckedKeys(checkedKeys.filter((key) => typeof key !== 'string'))
    }

    useEffect(() => {
        if (!open) return
        fetchRoleDetail()
        fetchAuthTree()
    }, [open, roleId])

    return (
        <Drawer
            {...props}
            extra={
                <Space size={8}>
                    <Button type="default" onClick={handleClose}>
                        取消
                    </Button>
                    <Button
                        type="primary"
                        loading={submitLoading}
                        onClick={handleAuthConfig}
                    >
                        确定
                    </Button>
                </Space>
            }
        >
            <Descriptions
                title={
                    <Title level={5}>
                        <span>角色信息</span>
                    </Title>
                }
                column={2}
                items={buildRoleDescItems()}
            />
            <Title level={5}>
                <span>权限选择</span>
            </Title>
            {!loading && (
                <Tree
                    checkable
                    defaultCheckedKeys={defaultCheckedKeys}
                    defaultExpandedKeys={defaultExpandedKeys}
                    checkedKeys={checkedKeys}
                    autoExpandParent={true}
                    treeData={buildTreeData()}
                    onCheck={handleOnCheck}
                />
            )}
        </Drawer>
    )
}
