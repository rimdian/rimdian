import { faTrashAlt } from '@fortawesome/free-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { Button, Table, Tag } from 'antd'
import ButtonUpsertSubscriptionList from 'components/subscription_list/button_upsert'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { BroadcastCampaignSubscriptionList, SubscriptionList } from 'interfaces'
import { useMemo } from 'react'
import CSS from 'utils/css'

type SubscriptionListsProps = {
  channel: 'email'
  value?: BroadcastCampaignSubscriptionList[]
  onChange?: (value: BroadcastCampaignSubscriptionList[]) => void
  disabled?: boolean
}

const InputSubscriptionLists = (props: SubscriptionListsProps) => {
  const workspaceCtx = useCurrentWorkspaceCtx()

  const channelLists = useMemo(() => {
    return workspaceCtx.subscriptionLists.filter((x) => x.channel === props.channel)
  }, [workspaceCtx.subscriptionLists, props.channel])

  const addList = (list: SubscriptionList) => {
    if (props.onChange) {
      props.onChange([...(props.value || []), list])
    }
  }

  const removeList = (list: SubscriptionList) => {
    if (props.onChange) {
      props.onChange((props.value || []).filter((x) => x.id !== list.id))
    }
  }

  return (
    <>
      {channelLists && channelLists.length > 0 && (
        <Table
          size="middle"
          dataSource={channelLists}
          rowKey="id"
          showHeader={false}
          pagination={false}
          className={CSS.margin_b_m}
          columns={[
            {
              title: 'id',
              render: (_value, record) => {
                return (
                  <>
                    <Tag color={record.color}>{record.name}</Tag>
                  </>
                )
              }
            },

            {
              title: 'action',
              width: 70,
              className: CSS.text_right,
              render: (_value, record) => {
                const isSelected = props.value?.find(
                  (x: BroadcastCampaignSubscriptionList) => x.id === record.id
                )
                if (isSelected) {
                  return (
                    <Button type="text" size="small" onClick={removeList.bind(null, record)}>
                      <FontAwesomeIcon icon={faTrashAlt} />
                    </Button>
                  )
                }

                return (
                  <Button type="primary" ghost size="small" onClick={addList.bind(null, record)}>
                    Add
                  </Button>
                )
              }
            }
          ]}
        />
      )}
      {channelLists.length === 0 && <ButtonUpsertSubscriptionList channel={props.channel} />}
    </>
  )
}

export default InputSubscriptionLists
