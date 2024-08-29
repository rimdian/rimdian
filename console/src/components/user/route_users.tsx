import { Row, Col, Popconfirm, Badge, Space, Tooltip, Tag, Button } from 'antd'
import { SubscriptionList } from 'interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import { useSearchParams } from 'react-router-dom'
import { useAccount } from 'components/login/context_account'
import { forEach } from 'lodash'
import { Segment } from 'components/segment/interfaces'
import BlockUsers from './table_users'
import BlockSidebarUsers from './block_sidebar'
import { useMemo } from 'react'
import { faTrashAlt } from '@fortawesome/free-regular-svg-icons'
import ButtonImportSubscriptionListUsers from 'components/subscription_list/button_import_users'
import ButtonUpsertSegment from 'components/segment/button_upsert'
import CSS from 'utils/css'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import SegmentMetrics from './segment_metrics'

const RouteUsers = () => {
  const accountCtx = useAccount()
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [searchParams] = useSearchParams()

  const segments: Segment[] = useMemo(() => {
    const list = [workspaceCtx.segmentsMap.anonymous, workspaceCtx.segmentsMap.authenticated]

    forEach(workspaceCtx.segmentsMap, (segment: Segment) => {
      if (segment.id !== '_all' && segment.id !== 'anonymous' && segment.id !== 'authenticated') {
        list.push(segment)
      }
    })

    return list
  }, [workspaceCtx.segmentsMap])

  const currentSegment = useMemo(() => {
    if (!searchParams.get('segment_id') && !searchParams.get('list_id')) {
      return workspaceCtx.segmentsMap['authenticated']
    }

    if (searchParams.get('segment_id'))
      return workspaceCtx.segmentsMap[searchParams.get('segment_id') as string]
    else return undefined
  }, [workspaceCtx.segmentsMap, searchParams])

  const currentList = useMemo(() => {
    if (!searchParams.get('list_id')) {
      return undefined
    }

    return workspaceCtx.subscriptionLists.find(
      (list: SubscriptionList) => list.id === searchParams.get('list_id')
    )
  }, [searchParams, workspaceCtx.subscriptionLists])

  return (
    <Layout currentOrganization={workspaceCtx.organization} currentWorkspaceCtx={workspaceCtx}>
      <Row gutter={[16, 16]}>
        <Col span={5}></Col>
        <Col span={19} className={CSS.margin_v_l}>
          {currentSegment && currentSegment.id !== '_all' && (
            <>
              <Space style={{ marginTop: 2, lineHeight: '24px', height: '24px' }}>
                <Tag color={currentSegment.color}>{currentSegment.name}</Tag>

                {currentSegment.status === 'building' && (
                  <Badge status="processing" text="Building" />
                )}
                {currentSegment.status === 'active' && <Badge status="success" text="Active" />}
                {currentSegment.status === 'deleted' && <Badge status="error" text="Deleted" />}
              </Space>

              <div className={CSS.pull_right}>
                {currentSegment &&
                  currentSegment.id !== 'anonymous' &&
                  currentSegment.id !== 'authenticated' && (
                    <Space>
                      <Popconfirm
                        title="Do you really want to delete this segment?"
                        okText="Delete"
                        okButtonProps={{ danger: true }}
                        cancelText="No"
                        placement="left"
                      >
                        <Button type="text" size="small">
                          <FontAwesomeIcon icon={faTrashAlt} />
                        </Button>
                      </Popconfirm>
                      <ButtonUpsertSegment segment={currentSegment} />
                    </Space>
                  )}
              </div>
            </>
          )}

          {currentList && (
            <>
              <Space size="large" style={{ marginTop: 2, lineHeight: '24px', height: '24px' }}>
                <Tag color={currentList.color}>{currentList.name}</Tag>

                <Tooltip title="Active users">
                  <Badge status="success" text={<>{currentList.active_users}</>} />
                </Tooltip>
                <Tooltip title="Paused users">
                  <Badge status="warning" text={<>{currentList.paused_users}</>} />
                </Tooltip>
                <Tooltip title="Unsubscribed users">
                  <Badge status="default" text={<>{currentList.unsubscribed_users}</>} />
                </Tooltip>
              </Space>

              <div className={CSS.pull_right}>
                {currentList && (
                  <Space>
                    <ButtonImportSubscriptionListUsers
                      btnProps={{
                        type: 'primary',
                        ghost: true
                      }}
                      segments={segments}
                      subscriptionList={currentList}
                    />
                    {/* <Popconfirm
                      title="Do you really want to delete this subscription list?"
                      okText="Delete"
                      okButtonProps={{ danger: true }}
                      cancelText="No"
                    >
                      <Button type="text" size="small">
                        <FontAwesomeIcon icon={faTrashAlt} />
                      </Button>
                    </Popconfirm> */}
                  </Space>
                )}
              </div>
            </>
          )}
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col span={5}>
          <BlockSidebarUsers
            segments={segments}
            currentSegment={currentSegment}
            currentList={currentList}
          />
        </Col>

        <Col span={19}>
          {currentSegment && (
            <SegmentMetrics
              timezone={accountCtx.account?.account.timezone || 'UTC'}
              currentSegment={currentSegment}
            />
          )}
          <BlockUsers
            timezone={accountCtx.account?.account.timezone || 'UTC'}
            segments={segments}
            currentSegment={currentSegment}
            currentList={currentList}
            limit={5}
          />
        </Col>
      </Row>
    </Layout>
  )
}

export default RouteUsers
