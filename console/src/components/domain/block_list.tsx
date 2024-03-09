import { Button, Table, Tag } from 'antd'
import { Domain, DomainHost } from 'interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPenToSquare } from '@fortawesome/free-regular-svg-icons'
import DeleteDomainButton from './button_delete'
import UpsertDomainButton from './button_upsert'
import { faPlus } from '@fortawesome/free-solid-svg-icons'
import CSS from 'utils/css'
import { blockCss } from 'components/common/block'

const BlockDomains = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()

  return (
    <>
      <Table
        pagination={false}
        className={blockCss.self}
        dataSource={workspaceCtx.workspace.domains.filter((x: Domain) => !x.deleted_at)}
        rowKey="id"
        columns={[
          {
            title: 'ID',
            key: 'id',
            render: (dom) => dom.id
          },
          {
            title: 'Name',
            key: 'id',
            render: (dom) => dom.name
          },
          {
            title: 'Type',
            key: 'type',
            render: (dom) => dom.type
          },
          {
            title: 'Details',
            key: 'params',
            render: (dom: Domain) => (
              <div>
                {dom.type === 'web' && (
                  <>
                    <div className={CSS.margin_b_m}>
                      <b>Hosts:</b>{' '}
                      {dom.hosts &&
                        dom.hosts.map((x: DomainHost) => (
                          <div className={CSS.padding_b_s} key={x.host}>
                            - <Tag>{x.host}</Tag>{' '}
                            {x.path_prefix && x.path_prefix !== '' && (
                              <small>
                                rewrite URLs prefix with <Tag>{x.path_prefix}</Tag>
                              </small>
                            )}
                          </div>
                        ))}
                    </div>
                    <div className={CSS.margin_b_m}>
                      <b>Keep URL params:</b>{' '}
                      {dom.params_whitelist &&
                        dom.params_whitelist.map((p: any) => (
                          <div className={CSS.padding_b_s} key={p}>
                            - <Tag key={p}>{p}</Tag>
                          </div>
                        ))}
                    </div>
                  </>
                )}
              </div>
            )
          },
          {
            title: (
              <div className={CSS.text_right}>
                <UpsertDomainButton
                  organizationId={workspaceCtx.organization.id}
                  workspaceId={workspaceCtx.workspace.id}
                  btnSize="small"
                  btnType="primary"
                  btnContent={
                    <>
                      <FontAwesomeIcon icon={faPlus} />
                      &nbsp; New domain
                    </>
                  }
                  apiPOST={workspaceCtx.apiPOST}
                  refreshWorkspace={workspaceCtx.refreshWorkspace}
                />
              </div>
            ),
            key: 'actions',
            className: 'actions',
            render: (row: Domain) => (
              <div className={CSS.text_right}>
                <Button.Group>
                  <DeleteDomainButton
                    domainId={row.id}
                    workspaceId={workspaceCtx.workspace.id}
                    domains={workspaceCtx.workspace.domains}
                    apiPOST={workspaceCtx.apiPOST}
                    onComplete={() => {
                      workspaceCtx.refreshWorkspace()
                    }}
                    btnSize="small"
                    btnType="text"
                  />

                  <UpsertDomainButton
                    organizationId={workspaceCtx.organization.id}
                    workspaceId={workspaceCtx.workspace.id}
                    domain={row}
                    btnSize="small"
                    btnType="text"
                    btnContent={<FontAwesomeIcon icon={faPenToSquare} />}
                    apiPOST={workspaceCtx.apiPOST}
                    refreshWorkspace={workspaceCtx.refreshWorkspace}
                  />
                </Button.Group>
              </div>
            )
          }
        ]}
      />
    </>
  )
}

export default BlockDomains
