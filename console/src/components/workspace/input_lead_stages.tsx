import { useState, useEffect } from 'react'
import {
  Button,
  Input,
  Radio,
  Select,
  Table,
  Tag,
  Divider,
  Form,
  Modal,
  message,
  Popconfirm
} from 'antd'
import { kebabCase } from 'lodash'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPenToSquare, faTrashCan } from '@fortawesome/free-regular-svg-icons'
import { faCaretUp, faCaretDown } from '@fortawesome/free-solid-svg-icons'
import Messages from 'utils/formMessages'
import { LeadStage } from 'interfaces'
import CSS from 'utils/css'

export const LeadsStageInput = ({ value, onChange }: any) => {
  const [stages, setStages] = useState(value)

  useEffect(() => {
    if (JSON.stringify(value) !== JSON.stringify(stages)) {
      setStages(value)
    }
  }, [stages, value])

  const moveStage = (fromIndex: number, toIndex: number) => {
    const updatedStages = [...stages]

    if (toIndex >= updatedStages.length) {
      var k = toIndex - updatedStages.length + 1
      while (k--) {
        updatedStages.push(undefined)
      }
    }
    updatedStages.splice(toIndex, 0, updatedStages.splice(fromIndex, 1)[0])

    setStages(updatedStages)
    onChange(updatedStages)
  }

  const addStage = (stage: LeadStage) => {
    if (stages.find((x: LeadStage) => x.id === stage.id)) {
      message.error('This stage ID already exists')
      return
    }

    const updatedStages = [...stages, stage]

    setStages(updatedStages)
    onChange(updatedStages)
  }

  const updateStage = (stage: LeadStage) => {
    const stageIndex = stages.findIndex((x: LeadStage) => x.id === stage.id)
    const updatedStages = [...stages]
    updatedStages[stageIndex] = stage

    setStages(updatedStages)
    onChange(updatedStages)
  }

  // const generateStageId = (stages: any, id: string, inc: number): string => {
  //   if (stages.find((s: any) => s.id === id)) {
  //     return generateStageId(stages, id + inc, inc + 1)
  //   }
  //   return id
  // }

  const deleteUnsavedStage = (stage: LeadStage) => {
    let updatedStages = stages.filter((s: LeadStage) => stage.id !== s.id)

    setStages(updatedStages)
    onChange(updatedStages)
  }

  const activeStages = stages.filter((s: LeadStage) => !s.deleted_at)
  const deletedStages = stages.filter((x: LeadStage) => x.deleted_at)

  // console.log('deletedStages', deletedStages)
  return (
    <>
      <Table
        size="middle"
        bordered={false}
        pagination={false}
        rowKey="id"
        // showHeader={false}
        className={CSS.margin_b_m}
        columns={[
          {
            title: 'ID',
            key: 'id',
            render: (x) => x.id
          },
          {
            title: 'Label',
            key: 'label',
            render: (x) => <Tag color={x.color !== 'grey' ? x.color : undefined}>{x.label}</Tag>
          },
          {
            title: 'Status',
            key: 'status',
            render: (x) => x.status
          },
          {
            title: '',
            key: 'remove',
            render: (x, _record, i) => {
              return (
                <div className={CSS.text_right}>
                  <Button.Group className={CSS.margin_r_m}>
                    {i !== 0 && (
                      <Button type="dashed" size="small" onClick={moveStage.bind(null, i, i - 1)}>
                        <FontAwesomeIcon icon={faCaretUp} />
                      </Button>
                    )}
                    {i !== activeStages.length - 1 && (
                      <Button type="dashed" size="small" onClick={moveStage.bind(null, i, i + 1)}>
                        <FontAwesomeIcon icon={faCaretDown} />
                      </Button>
                    )}
                  </Button.Group>

                  <Button.Group>
                    <AddOrUpdateStageModal onComplete={updateStage} stage={x}>
                      <Button type="dashed" size="small">
                        <FontAwesomeIcon icon={faPenToSquare} />
                      </Button>
                    </AddOrUpdateStageModal>

                    {!x.createdAt && (
                      <Popconfirm
                        onConfirm={deleteUnsavedStage.bind(null, x)}
                        title="Do you want to remove this stage?"
                      >
                        <Button type="dashed" size="small">
                          <FontAwesomeIcon icon={faTrashCan} />
                        </Button>
                      </Popconfirm>
                    )}

                    {x.createdAt && (
                      <DeleteSavedStage stage={x} stages={stages} onComplete={updateStage}>
                        <Button type="dashed" size="small">
                          <FontAwesomeIcon icon={faTrashCan} />
                        </Button>
                      </DeleteSavedStage>
                    )}
                  </Button.Group>
                </div>
              )
            }
          }
        ]}
        dataSource={activeStages}
      />

      <AddOrUpdateStageModal onComplete={addStage}>
        <Button type="primary" size="small" block ghost>
          Add a stage
        </Button>
      </AddOrUpdateStageModal>

      {deletedStages.length > 0 && (
        <>
          <Divider orientation="left" plain>
            Deleted stages
          </Divider>

          <Table
            size="middle"
            bordered={false}
            pagination={false}
            rowKey="id"
            className={CSS.margin_b_m}
            columns={[
              {
                title: 'ID',
                key: 'id',
                render: (x) => x.id
              },
              {
                title: 'Label',
                key: 'label',
                render: (x) => <Tag color={x.color !== 'grey' ? x.color : undefined}>{x.label}</Tag>
              },
              {
                title: 'Migrated to stage',
                key: 'status',
                render: (x) => x.migrate_to_id
              }
            ]}
            dataSource={deletedStages}
          />
        </>
      )}
    </>
  )
}

const AddOrUpdateStageModal = ({ stage, onComplete, children }: any) => {
  const [form] = Form.useForm()
  const [modalVisible, setModalVisible] = useState(false)

  const onClicked = () => {
    if (stage) {
      form.setFieldsValue({
        label: stage.label,
        id: stage.id,
        status: stage.status,
        color: stage.color
      })
    }
    setModalVisible(true)
  }

  return (<>
    <span onClick={onClicked}>{children}</span>
    {modalVisible && (
      <Modal
        open={modalVisible}
        title={stage ? 'Update stage' : 'Add a stage'}
        okText="Confirm"
        cancelText="Cancel"
        onCancel={() => {
          setModalVisible(false)
        }}
        width={700}
        onOk={() => {
          form
            .validateFields()
            .then((values: any) => {
              // console.log('onComplete', values);
              form.resetFields()
              setModalVisible(false)
              if (stage) {
                stage.label = values.label
                stage.color = values.color
                onComplete(stage)
              } else {
                onComplete(values)
              }
            })
            .catch(console.error)
        }}
      >
        <Form
          form={form}
          name="form_add_stage"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
          layout="horizontal"
        >
          <Form.Item
            name="label"
            label="Label"
            rules={[{ required: true, message: Messages.RequiredField }]}
          >
            <Input
              placeholder="i.e: New, Lost, Won, Signed..."
              onChange={(e: any) => {
                if (!stage) form.setFieldsValue({ id: kebabCase(e.target.value) })
              }}
            />
          </Form.Item>

          <Form.Item
            name="id"
            label="Stage ID"
            rules={[
              {
                required: true,
                type: 'string',
                pattern: /^[a-z0-9]+(-[a-z0-9]+)*$/,
                message: Messages.InvalidIdFormat
              }
            ]}
          >
            <Input placeholder="i.e: web" disabled={stage && stage.createdAt} />
          </Form.Item>

          <Form.Item
            name="status"
            label="Consider this stage as"
            rules={[{ required: true, message: Messages.RequiredField }]}
          >
            <Radio.Group style={{ width: '100%' }} disabled={stage && stage.createdAt}>
              <Radio.Button style={{ width: '33.33%', textAlign: 'center' }} value="open">
                Open
              </Radio.Button>
              <Radio.Button style={{ width: '33.33%', textAlign: 'center' }} value="converted">
                Converted
              </Radio.Button>
              <Radio.Button style={{ width: '33.33%', textAlign: 'center' }} value="lost">
                Lost
              </Radio.Button>
            </Radio.Group>
          </Form.Item>

          <Form.Item
            name="color"
            label="Color"
            rules={[{ required: true, message: Messages.RequiredField }]}
          >
            <Select>
              <Select.Option value="magenta">
                <Tag color="magenta">magenta</Tag>
              </Select.Option>
              <Select.Option value="red">
                <Tag color="red">red</Tag>
              </Select.Option>
              <Select.Option value="volcano">
                <Tag color="volcano">volcano</Tag>
              </Select.Option>
              <Select.Option value="orange">
                <Tag color="orange">orange</Tag>
              </Select.Option>
              <Select.Option value="gold">
                <Tag color="gold">gold</Tag>
              </Select.Option>
              <Select.Option value="lime">
                <Tag color="lime">lime</Tag>
              </Select.Option>
              <Select.Option value="green">
                <Tag color="green">green</Tag>
              </Select.Option>
              <Select.Option value="cyan">
                <Tag color="cyan">cyan</Tag>
              </Select.Option>
              <Select.Option value="blue">
                <Tag color="blue">blue</Tag>
              </Select.Option>
              <Select.Option value="geekblue">
                <Tag color="geekblue">geekblue</Tag>
              </Select.Option>
              <Select.Option value="purple">
                <Tag color="purple">purple</Tag>
              </Select.Option>
              <Select.Option value="grey">
                <Tag>grey</Tag>
              </Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    )}
  </>);
}

const DeleteSavedStage = ({ stage, stages, onComplete, children }: any) => {
  const [form] = Form.useForm()
  const [modalVisible, setModalVisible] = useState(false)

  const onClicked = () => {
    if (stage) {
      form.setFieldsValue({
        label: stage.label,
        id: stage.id,
        status: stage.status,
        color: stage.color,
        migrate_to_id: stage.migrate_to_id
      })
    }
    setModalVisible(true)
  }

  return (<>
    <span onClick={onClicked}>{children}</span>
    {modalVisible && (
      <Modal
        open={modalVisible}
        title="Delete stage"
        okText="Confirm"
        cancelText="Cancel"
        onCancel={() => {
          setModalVisible(false)
        }}
        width={700}
        onOk={() => {
          form
            .validateFields()
            .then((values: any) => {
              // console.log('onComplete', values);
              form.resetFields()
              setModalVisible(false)
              stage.migrate_to_id = values.migrate_to_id

              if (!stage.deleted_at) {
                stage.deleted_at = new Date().toISOString()
              }
              onComplete(stage)
            })
            .catch((info) => {
              console.log('Validate Failed:', info)
            })
        }}
      >
        <Form
          form={form}
          name="form_remove_stage"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
          layout="horizontal"
        >
          <Form.Item
            name="label"
            label="Label"
            rules={[{ required: true, message: Messages.RequiredField }]}
          >
            <Input
              placeholder="i.e: New, Lost, Won, Signed..."
              onChange={(e: any) => {
                form.setFieldsValue({ id: kebabCase(e.target.value) })
              }}
            />
          </Form.Item>

          <Form.Item
            name="id"
            label="Stage ID"
            rules={[
              {
                required: true,
                type: 'string',
                pattern: /^[a-z0-9]+(-[a-z0-9]+)*$/,
                message: Messages.InvalidIdFormat
              }
            ]}
          >
            <Input placeholder="i.e: web" disabled />
          </Form.Item>

          <Form.Item
            name="migrate_to_id"
            label="Migrate leads to stage"
            rules={[{ required: true, message: Messages.RequiredField }]}
          >
            <Select>
              {stages
                .filter((s: any) => !s.deleted_at && s.id !== stage.id)
                .map((s: any) => (
                  <Select.Option value={s.id} key={s.id}>
                    <Tag color={s.color !== 'grey' ? s.color : undefined}>{s.label}</Tag>
                  </Select.Option>
                ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="status"
            label="Consider this stage as"
            rules={[{ required: true, message: Messages.RequiredField }]}
          >
            <Radio.Group style={{ width: '100%' }} disabled={stage && stage.createdAt}>
              <Radio.Button style={{ width: '33.33%', textAlign: 'center' }} value="open">
                Open
              </Radio.Button>
              <Radio.Button style={{ width: '33.33%', textAlign: 'center' }} value="converted">
                Converted
              </Radio.Button>
              <Radio.Button style={{ width: '33.33%', textAlign: 'center' }} value="lost">
                Lost
              </Radio.Button>
            </Radio.Group>
          </Form.Item>

          <Form.Item
            name="color"
            label="Color"
            rules={[{ required: true, message: Messages.RequiredField }]}
          >
            <Select>
              <Select.Option value="magenta">
                <Tag color="magenta">magenta</Tag>
              </Select.Option>
              <Select.Option value="red">
                <Tag color="red">red</Tag>
              </Select.Option>
              <Select.Option value="volcano">
                <Tag color="volcano">volcano</Tag>
              </Select.Option>
              <Select.Option value="orange">
                <Tag color="orange">orange</Tag>
              </Select.Option>
              <Select.Option value="gold">
                <Tag color="gold">gold</Tag>
              </Select.Option>
              <Select.Option value="lime">
                <Tag color="lime">lime</Tag>
              </Select.Option>
              <Select.Option value="green">
                <Tag color="green">green</Tag>
              </Select.Option>
              <Select.Option value="cyan">
                <Tag color="cyan">cyan</Tag>
              </Select.Option>
              <Select.Option value="blue">
                <Tag color="blue">blue</Tag>
              </Select.Option>
              <Select.Option value="geekblue">
                <Tag color="geekblue">geekblue</Tag>
              </Select.Option>
              <Select.Option value="purple">
                <Tag color="purple">purple</Tag>
              </Select.Option>
              <Select.Option value="grey">
                <Tag>grey</Tag>
              </Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    )}
  </>);
}
