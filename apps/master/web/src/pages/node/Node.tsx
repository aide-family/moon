import React, {useEffect, useState} from "react";
import {Button, Collapse, Modal} from '@arco-design/web-react';
import NodeList from "@/apis/prom/node/list.api";
import type {NodeItem} from "@/apis/prom/prom";
import NodeDetail from "@/apis/prom/node/detail.api";

const CollapseItem = Collapse.Item;

const Node = () => {
    const [nodes, setNodes] = useState<NodeItem[]>([]);

    const getNodes = async () => {
        const data = await NodeList()
        setNodes(data.list);
    }

    const getNode = async (id: string, keys: string[], event: any) => {
        const data = await NodeDetail(+id)
        let nodeTemp = nodes.map(node => {
            if (node.id === +id) {
                node.dirs = data.node.dirs
            }
            return node
        })

        setNodes(nodeTemp);
    }

    useEffect(() => {
        getNodes()
    }, []);

    const [visible, setVisible] = useState(false);


    return <div>
        <Collapse defaultActiveKey={['1']} bordered onChange={getNode} triggerRegion="icon">
            {
                nodes.map((node) => (
                        <CollapseItem
                            key={node.id}
                            header={<div>
                                <span>{`${node.cn_name}(${node.en_name})  ${node.datasource}`}</span>
                            </div>}
                            name={node.id + ""}
                        >
                            {node.remark}
                            <Collapse>
                                {
                                    node.dirs.map((dir) => (
                                            <CollapseItem key={dir.id} header={dir.path} name={dir.id + ""}>
                                                {
                                                    dir.files.map(file => (
                                                        <Button onClick={() => setVisible(true)} type="text"
                                                                key={file.id}>{file.filename}</Button>
                                                    ))
                                                }
                                            </CollapseItem>
                                        )
                                    )
                                }
                            </Collapse>
                        </CollapseItem>
                    )
                )
            }

        </Collapse>
        <Modal visible={visible} onCancel={() => setVisible(false)} style={{
            height: '100%',
            width: '100%',
        }}>
            <h1>
                123
            </h1>
        </Modal>
    </div>;
}

export default Node;