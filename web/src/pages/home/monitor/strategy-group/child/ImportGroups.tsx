import {
    Button,
    Form,
    Modal,
    ModalProps,
    Space,
    Steps,
    UploadProps,
    message
} from 'antd'
import React, { useContext, useEffect, useState } from 'react'
import {
    ImportModalStepsItemsParamsType,
    importModalStepsItems
} from '../options'
import {
    ImportGroupItemType,
    ImportGroupRequest
} from '@/apis/home/monitor/strategy-group/types'
import { UploadChangeParam, UploadFile } from 'antd/es/upload'
import yaml from 'js-yaml'
import { GlobalContext } from '@/context'
import strategyGroupApi from '@/apis/home/monitor/strategy-group'

export interface ImportGroupsProps extends ModalProps {}

export const ImportGroups: React.FC<ImportGroupsProps> = (props) => {
    const { sysTheme } = useContext(GlobalContext)
    const [form] = Form.useForm<ImportGroupRequest>()
    const [current, setCurrent] = useState(0)
    const [importData, setImportData] = useState<ImportGroupRequest>()
    const [uploadFileInfo, setUploadFileInfo] =
        useState<UploadChangeParam<UploadFile<any>>>()
    const [fileList, setFileList] = useState<UploadFile<any>[]>([])

    const uploadProps: UploadProps = {
        name: 'file',
        action: 'https://run.mocky.io/v3/435e224c-44fb-4773-9faf-380c5e6a2188',
        headers: {
            authorization: 'authorization-text'
        },
        // 文件类型限制
        accept: '.yaml',
        // 上传文件大小限制
        maxCount: 1,
        fileList: fileList,
        // 文件上传
        onChange(info) {
            setFileList(info.fileList)
            switch (info.file.status) {
                case 'uploading':
                    break
                case 'removed':
                    message.info(`${info.file.name} 文件已删除`)
                    break
                case 'done':
                    message.success(`${info.file.name} 文件上传成功`)
                    setUploadFileInfo(info)
                    setFileList(info.fileList)
                    break
                case 'error':
                    message.error(`${info.file.name} 文件上传失败`)
                    break
            }
        }
    }

    const params: ImportModalStepsItemsParamsType = {
        form,
        uploadProps,
        sysTheme,
        importData
    }

    const [maxCurrent] = useState(importModalStepsItems(params).length - 1)

    const handleNext = async () => {
        switch (current) {
            case 0:
                const defaultValues = await form.validateFields()
                if (!defaultValues) return
                setImportData({ ...defaultValues })
                break
            case 1:
                // 读取文件内容信息
                const file = uploadFileInfo?.file
                if (!file) {
                    message.error('请上传文件')
                    return
                }
                let reader = new FileReader()
                reader.onload = (e) => {
                    const doc = e?.target?.result
                    if (!doc) return
                    const yamlContent = yaml.load(`${doc}`) as {
                        groups: ImportGroupItemType[]
                    }
                    setImportData({ ...importData, groups: yamlContent.groups })
                }
                if (!file.originFileObj) return
                reader.readAsText(file.originFileObj)
                break
            case 2:
                break
            case 3:
                break
        }
        if (current < maxCurrent) {
            setCurrent(current + 1)
        }
    }

    const handleBack = () => {
        if (current > 0) {
            setCurrent(current - 1)
        }
    }

    const initFormData = () => {
        form?.resetFields()
        setCurrent(0)
        setImportData(undefined)
        setFileList([])
        setUploadFileInfo(undefined)
    }

    const handleDone = (e: any) => {
        if (!importData) return
        strategyGroupApi
            .batchImport({ ...importData, defaultAlarmNotifyIds: [1] })
            .then(() => {
                message.success('导入成功')
                // 清空所有表单
                initFormData()
                props.onOk?.(e)
            })
            .catch((err) => {
                message.error(`${err}`)
            })
    }

    useEffect(() => {
        console.log('importData', importData)
    }, [importData])

    return (
        <Modal
            {...props}
            footer={
                <>
                    <Space
                        size={8}
                        style={{
                            textAlign: 'center',
                            width: '100%',
                            justifyContent: 'center'
                        }}
                    >
                        <Steps.Step
                            title={
                                <Button type="primary" disabled={current === 0}>
                                    上一步
                                </Button>
                            }
                            onClick={handleBack}
                        />
                        {current === maxCurrent ? (
                            <Steps.Step
                                title={<Button type="primary">完成</Button>}
                                onClick={handleDone}
                            />
                        ) : (
                            <Steps.Step
                                title={<Button type="primary">下一步</Button>}
                                onClick={handleNext}
                            />
                        )}
                    </Space>
                </>
            }
        >
            <Steps
                style={{ marginTop: 20 }}
                current={current}
                items={importModalStepsItems(params)}
            />
            <div style={{ marginTop: 20 }}>
                {importModalStepsItems(params)[current].content}
            </div>
        </Modal>
    )
}
