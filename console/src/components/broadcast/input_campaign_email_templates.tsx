import { faEdit, faTrashAlt } from '@fortawesome/free-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { Alert, Button, InputNumber, Modal, Select, Space, Table, Tag, Tooltip } from 'antd'
import { useQuery } from '@tanstack/react-query'
import ButtonUpsertEmailTemplate from 'components/assets/message_template/button_upsert_email'
import { MessageTemplate } from 'components/assets/message_template/interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { BroadcastCampaignMessageTemplate } from 'interfaces'
import { useMemo, useState } from 'react'
import CSS from 'utils/css'

type InputCampaignEmailTemplatesProps = {
  value?: BroadcastCampaignMessageTemplate[]
  onChange?: (value: BroadcastCampaignMessageTemplate[]) => void
  disabled?: boolean
  name?: string
  utmSource?: string
  utmMedium?: string
  utmCampaign?: string
}

const InputCampaignEmailTemplates = (props: InputCampaignEmailTemplatesProps) => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [modalVisible, setModalVisible] = useState(false)

  const {
    isLoading,
    data: messageTemplates,
    refetch,
    isFetching
  } = useQuery<MessageTemplate[]>(
    ['templates_email_campaign', workspaceCtx.workspace.id],
    (): Promise<MessageTemplate[]> => {
      return new Promise((resolve, reject) => {
        workspaceCtx
          .apiGET(
            '/messageTemplate.list?workspace_id=' +
              workspaceCtx.workspace.id +
              '&channel=email&category=campaign'
          )
          .then((data: any) => {
            // console.log(data)
            resolve(data as MessageTemplate[])
          })
          .catch((e) => {
            reject(e)
          })
      })
    }
  )

  const updatePercentage = (templateID: string, percentage: number | null) => {
    let diff = 0
    let isIncrement = false
    let templateIndex = 0
    const newTemplates = (props.value || []).map((x, i) => {
      if (x.id === templateID) {
        templateIndex = i
        if (percentage !== null && percentage > x.percentage) {
          isIncrement = true
          diff = (percentage || 0) - x.percentage
        } else {
          diff = x.percentage - (percentage || 0)
        }
        return { ...x, percentage: percentage || 0 }
      }
      return x
    })

    // if has a previous or next template, adjust the percentage
    const hasPrevious = templateIndex > 0
    const hasNext = templateIndex < newTemplates.length - 1

    if (hasPrevious) {
      if (isIncrement) {
        const newValue = newTemplates[templateIndex - 1].percentage - diff
        if (newValue >= 0 && newValue <= 100) {
          newTemplates[templateIndex - 1].percentage = newValue
        }
      } else {
        const newValue = newTemplates[templateIndex - 1].percentage + diff
        if (newValue >= 0 && newValue <= 100) {
          newTemplates[templateIndex - 1].percentage = newValue
        }
      }
    } else if (hasNext) {
      if (isIncrement) {
        const newValue = newTemplates[templateIndex + 1].percentage - diff
        if (newValue >= 0 && newValue <= 100) {
          newTemplates[templateIndex + 1].percentage = newValue
        }
      } else {
        const newValue = newTemplates[templateIndex + 1].percentage + diff
        if (newValue >= 0 && newValue <= 100) {
          newTemplates[templateIndex + 1].percentage = newValue
        }
      }
    }

    if (props.onChange) {
      props.onChange(newTemplates)
    }
  }

  const availableTemplates = useMemo(() => {
    if (!messageTemplates) {
      return []
    }
    return messageTemplates.filter((x) => !props.value?.find((y) => y.id === x.id))
  }, [messageTemplates, props.value])

  const equalSplit = (
    templates: BroadcastCampaignMessageTemplate[]
  ): BroadcastCampaignMessageTemplate[] => {
    if (templates.length === 0) return templates

    const arr = [...templates]

    if (arr.length === 1) {
      arr[0].percentage = 100
      return arr
    }

    let eachPercentage = 100
    let total = 0

    eachPercentage = Math.floor(100 / arr.length) // split percentages equally
    arr.forEach((x: BroadcastCampaignMessageTemplate, i: number) => {
      x.percentage = eachPercentage
      total += eachPercentage

      // last item has the rest
      if (i + 1 === arr.length) {
        const rest = total < 100 ? 100 - total : 0
        x.percentage += rest
      }
    })

    return arr
  }

  const addTemplate = (templateID: string) => {
    if (props.onChange) {
      let arr = props.value ? [...props.value] : []
      arr.push({
        id: templateID,
        percentage: 0
      })
      arr = equalSplit(arr)
      props.onChange(arr)
    }
  }

  const removeTemplate = (templateID: string) => {
    if (props.onChange) {
      let arr = (props.value || []).filter((x) => x.id !== templateID)
      arr = equalSplit(arr)
      props.onChange(arr)
    }
  }

  return (
    <>
      {props.value && props.value.length > 0 && (
        <Table
          size="middle"
          dataSource={props.value}
          className={CSS.margin_b_m}
          rowKey="id"
          // showHeader={false}
          pagination={false}
          loading={isLoading || isFetching}
          columns={[
            {
              title: 'A/B',
              width: 150,
              render: (record: BroadcastCampaignMessageTemplate) => {
                return (
                  <InputNumber
                    size="small"
                    value={record.percentage}
                    min={0}
                    max={100}
                    step={5}
                    onChange={updatePercentage.bind(null, record.id)}
                    formatter={(value) => `${value}%`}
                    parser={(value) => value?.replace('%', '') as unknown as number}
                  />
                )
              }
            },
            {
              title: 'Name',
              render: (record: BroadcastCampaignMessageTemplate) => {
                const template = messageTemplates?.find((x: MessageTemplate) => x.id === record.id)
                if (!template) {
                  return <div>{record.id}</div>
                }
                return <div>{template.name}</div>
              }
            },
            {
              title: '',
              className: CSS.text_right,
              render: (_value, record) => {
                return (
                  <Space>
                    <Tooltip title="Remove from campaign">
                      <Button
                        type="text"
                        size="small"
                        onClick={removeTemplate.bind(null, record.id)}
                      >
                        <FontAwesomeIcon icon={faTrashAlt} />
                      </Button>
                    </Tooltip>

                    <ButtonUpsertEmailTemplate
                      onSuccess={(newID: string) => {
                        refetch()
                      }}
                      btnProps={{ type: 'text', size: 'small' }}
                      template={messageTemplates?.find((x) => x.id === record.id)}
                    >
                      <Tooltip title="Edit template">
                        <FontAwesomeIcon icon={faEdit} />
                      </Tooltip>
                    </ButtonUpsertEmailTemplate>
                  </Space>
                )
              }
            }
          ]}
        />
      )}

      <Button type="primary" ghost block onClick={() => setModalVisible(true)}>
        Add a template
      </Button>

      {modalVisible && (
        <Modal
          title="Select a template"
          open={true}
          onCancel={() => setModalVisible(false)}
          footer={null}
          width={400}
        >
          <div className={CSS.margin_v_l}>
            {availableTemplates.length > 0 && (
              <>
                <Alert
                  className={CSS.margin_b_l}
                  message={
                    <>
                      Only templates from the <Tag color="purple">Campaign</Tag> category are
                      available.
                    </>
                  }
                  type="info"
                  showIcon
                />

                <Select
                  showSearch
                  style={{ width: '100%' }}
                  options={availableTemplates.map((x) => ({ label: x.name, value: x.id }))}
                  onChange={(value) => {
                    addTemplate(value)
                    setModalVisible(false)
                  }}
                />
                <div className={CSS.margin_v_m + ' ' + CSS.text_center}>or</div>
              </>
            )}
            <ButtonUpsertEmailTemplate
              onSuccess={(newID: string) => {
                refetch().then(() => {
                  addTemplate(newID)
                  setModalVisible(false)
                })
              }}
              btnProps={{ type: 'primary', ghost: true, block: true }}
              category="campaign"
              name={
                props.name
                  ? props.name + ' v' + (props.value ? props.value.length + 1 : 1)
                  : undefined
              }
              utmSource={props.utmSource}
              utmMedium={props.utmMedium}
              utmCampaign={props.utmCampaign}
            >
              Create a new template
            </ButtonUpsertEmailTemplate>
          </div>
        </Modal>
      )}
    </>
  )
}

export default InputCampaignEmailTemplates
