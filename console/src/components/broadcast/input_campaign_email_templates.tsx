import { faTrashAlt } from '@fortawesome/free-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { Button, InputNumber, Modal, Select, Table } from 'antd'
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
    ['templates', workspaceCtx.workspace.id],
    (): Promise<MessageTemplate[]> => {
      return new Promise((resolve, reject) => {
        workspaceCtx
          .apiGET(
            '/messageTemplate.list?workspace_id=' + workspaceCtx.workspace.id + '&channel=email'
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
    if (props.onChange) {
      let templateIndex = 0
      const newTemplates = (props.value || []).map((x, i) => {
        if (x.message_template_id === templateID) {
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
          newTemplates[templateIndex - 1].percentage -= diff
        } else {
          newTemplates[templateIndex - 1].percentage += diff
        }
      } else if (hasNext) {
        if (isIncrement) {
          newTemplates[templateIndex + 1].percentage -= diff
        } else {
          newTemplates[templateIndex + 1].percentage += diff
        }
      }

      props.onChange(newTemplates)
    }
  }

  const availableTemplates = useMemo(() => {
    if (!messageTemplates) {
      return []
    }
    return messageTemplates.filter((x) => !props.value?.find((y) => y.message_template_id === x.id))
  }, [messageTemplates, props.value])

  const addTemplate = (templateID: string) => {
    if (props.onChange) {
      let percentageAvailable = 100

      if (props.value && props.value.length > 0) {
        const totalPercentage = props.value.reduce((acc, x) => acc + x.percentage, 0)
        percentageAvailable = 100 - totalPercentage
      }

      props.onChange([
        ...(props.value || []),
        { message_template_id: templateID, percentage: percentageAvailable }
      ])
    }
  }

  const removeTemplate = (templateID: string) => {
    if (props.onChange) {
      props.onChange((props.value || []).filter((x) => x.message_template_id !== templateID))
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
                    onChange={updatePercentage.bind(null, record.message_template_id)}
                    formatter={(value) => `${value}%`}
                    parser={(value) => value?.replace('%', '') as unknown as number}
                  />
                )
              }
            },
            {
              title: 'Name',
              render: (record: BroadcastCampaignMessageTemplate) => {
                const template = messageTemplates?.find(
                  (x: MessageTemplate) => x.id === record.message_template_id
                )
                if (!template) {
                  return <div>{record.message_template_id}</div>
                }
                return <div>{template.name}</div>
              }
            },
            {
              title: (
                <div className={CSS.text_right}>
                  {/* <Button type="primary" ghost size="small">
                    Change A/B
                  </Button> */}
                </div>
              ),
              // width: 20,
              className: CSS.text_right,
              render: (_value, record) => {
                return (
                  <Button
                    type="text"
                    size="small"
                    onClick={removeTemplate.bind(null, record.message_template_id)}
                  >
                    <FontAwesomeIcon icon={faTrashAlt} />
                  </Button>
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
          <div className={CSS.margin_v_m}>
            {availableTemplates.length > 0 && (
              <>
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
              onSuccess={refetch}
              btnProps={{ type: 'primary', ghost: true, block: true }}
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
