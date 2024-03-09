import { Table, Tag } from 'antd'
import { LeadStage } from 'interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
// import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
// import { faPlus } from '@fortawesome/free-solid-svg-icons'
// import { faPenToSquare } from '@fortawesome/free-regular-svg-icons'
import CSS from 'utils/css'

const BlockConversions = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()

  return (
    <div className={CSS.margin_t_m}>
      TODO
      {workspaceCtx.workspace.has_leads && (
        <>
          <Table
            size="small"
            bordered={false}
            pagination={false}
            rowKey="id"
            className={CSS.margin_t_l}
            columns={[
              {
                title: 'Stage ID',
                key: 'id',
                render: (x: LeadStage) => x.id
              },
              {
                title: 'Label',
                key: 'label',
                render: (x: LeadStage) => (
                  <Tag color={x.color !== 'grey' ? x.color : undefined}>{x.label}</Tag>
                )
              },
              {
                title: 'Consider leads as',
                key: 'status',
                render: (x: LeadStage) => x.status
              },
              {
                title: '',
                key: 'deleted',
                render: (x: LeadStage) => {
                  if (!x.deleted_at) return ''
                  return <span>Migrated to stage: {x.migrate_to_id}</span>
                }
              }
            ]}
            dataSource={workspaceCtx.workspace.lead_stages}
          />
        </>
      )}
    </div>
  )
}

export default BlockConversions
