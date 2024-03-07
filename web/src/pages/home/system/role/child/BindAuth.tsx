import authApi from '@/apis/home/system/auth'
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
    Spin,
    Tree,
    TreeDataNode,
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
    const [roleInfo, setRoleInfo] = useState<RoleListItem>()
    const [defaultCheckedKeys, setDefaultCheckedKeys] = useState<React.Key[]>(
        []
    )
    const [defaultExpandedKeys, setDefaultExpandedKeys] = useState<React.Key[]>(
        []
    )
    const [checkedKeys, setCheckedKeys] = useState<number[]>([])
    const [treeData, setTreeData] = useState<TreeDataNode[]>([])
    const [autoExpandParent, setAutoExpandParent] = useState<boolean>(true)

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
        setLoading(false)
    }

    const fetchAuthTree = async () => {
        setLoading(true)
        const { tree } = await authApiTree()
        if (!tree) return
        const treeData: TreeDataNode[] = tree.map((item) => {
            const { domain, domainName, module } = item
            return {
                title: domainName,
                key: `domain-${domain}`,
                children: module.map((child) => {
                    const { module, apis, name } = child
                    return {
                        title: name,
                        key: `module-${module}`,
                        children: apis.map((api) => {
                            return {
                                title: api.label,
                                key: api.value
                            }
                        })
                    }
                })
            }
        })
        setTreeData(treeData)
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
                                <UserAvatar {...userItem} key={index} toolTip />
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
        setAutoExpandParent(true)
        onClose?.(e)
    }

    const handleOnCheck = (checkedKeys: React.Key[] | any) => {
        if (!Array.isArray(checkedKeys)) return
        setCheckedKeys(checkedKeys.filter((key) => typeof key !== 'string'))
    }

    const handleExpand = (expandedKeysValue: React.Key[]) => {
        setDefaultExpandedKeys(expandedKeysValue)
        setAutoExpandParent(false)
    }

    useEffect(() => {
        if (!open) return
        setAutoExpandParent(true)
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
            <Spin spinning={loading}>
                <Tree
                    checkable
                    defaultCheckedKeys={defaultCheckedKeys}
                    defaultExpandedKeys={defaultExpandedKeys}
                    expandedKeys={defaultExpandedKeys}
                    checkedKeys={checkedKeys}
                    autoExpandParent={autoExpandParent}
                    treeData={treeData}
                    onCheck={handleOnCheck}
                    onExpand={handleExpand}
                />
            </Spin>
        </Drawer>
    )
}
