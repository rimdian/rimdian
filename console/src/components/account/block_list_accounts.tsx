import { Button, Table, Tag } from 'antd'
import { useQuery } from '@tanstack/react-query'
import { Account, Organization, OrganizationInvitation } from 'interfaces'
import { useAccount } from 'components/login/context_account'
import AccountsInviteButton from './button_invite_account'
import CreateServiceAccountButton from './button_create_service_account'
import TransferOwnershipButton from './button_transfer_ownership'
import DeactivateAccountButton from './button_deactivate_account'
import CancelInvitationButton from './button_cancel_invitation'
import ResendInvitationButton from './button_resend_invitation'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faUser, faKey, faHourglassHalf, faShieldAlt } from '@fortawesome/free-solid-svg-icons'
import dayjs from 'dayjs'
import CSS from 'utils/css'
import Block from 'components/common/block'

type ListAccountsProps = {
  organization: Organization
}

const ListAccounts = (props: ListAccountsProps) => {
  const accountCtx = useAccount()

  // accounts
  const {
    isLoading: isLoadingAccounts,
    data: dataAccounts,
    refetch: refetchAccounts,
    isFetching: isFetchingAccounts
  } = useQuery<Account[]>(
    ['accounts', props.organization.id],
    (): Promise<Account[]> => {
      return new Promise((resolve, reject) => {
        accountCtx
          .apiGET('/organizationAccount.list?organization_id=' + props.organization.id)
          .then((data: any) => {
            resolve(data.accounts as Account[])
          })
          .catch((e) => {
            reject(e)
          })
      })
    },
    { enabled: !!props.organization.id }
  )

  // invitations
  const {
    isLoading: isLoadingInvitations,
    data: dataInvitations,
    refetch: refetchInvitations,
    isFetching: isFetchingInvitations
  } = useQuery<OrganizationInvitation[]>(
    ['invitations', props.organization.id],
    (): Promise<OrganizationInvitation[]> => {
      return new Promise((resolve, reject) => {
        accountCtx
          .apiGET('/organizationInvitation.list?organization_id=' + props.organization.id)
          .then((data: any) => {
            resolve(data.invitations as OrganizationInvitation[])
          })
          .catch((e) => {
            reject(e)
          })
      })
    },
    { enabled: !!props.organization.id }
  )

  const data: (Account | OrganizationInvitation)[] = []

  if (dataAccounts) {
    dataAccounts.forEach((x) => data.push(x))
  }
  if (dataInvitations) {
    dataInvitations.forEach((x) => data.push(x))
  }

  const isOrganizationOwner = props.organization.im_owner

  return (
    <Block
      classNames={[CSS.margin_b_l]}
      title="Accounts & API keys"
      extra={
        <>
          <CreateServiceAccountButton
            isOrganizationOwner={isOrganizationOwner}
            apiPOST={accountCtx.apiPOST}
            organizationId={props.organization.id}
            onComplete={() => {
              refetchAccounts()
              refetchInvitations()
            }}
          />
          <AccountsInviteButton
            isOrganizationOwner={isOrganizationOwner}
            apiPOST={accountCtx.apiPOST}
            organizationId={props.organization.id}
            onComplete={() => {
              refetchAccounts()
              refetchInvitations()
            }}
          />
        </>
      }
    >
      <Table
        showHeader={false}
        pagination={false}
        loading={
          isLoadingAccounts || isLoadingInvitations || isFetchingInvitations || isFetchingAccounts
        }
        dataSource={data}
        size="middle"
        rowKey={(r: any) => {
          if (r.id) return r.id
          return r.email
        }}
        locale={{
          emptyText: () => {
            // remove the default icon with "No Data"
            return <>&nbsp;</>
          }
        }}
        columns={[
          {
            key: 'roles',
            width: 220,
            render: (row: Account & OrganizationInvitation) => {
              // is account
              if (row.id) {
                const dpoTag =
                  props.organization.dpo_id === row.id ? (
                    <span>
                      <Tag color="green">
                        <FontAwesomeIcon icon={faShieldAlt} />
                        &nbsp; DPO
                      </Tag>
                    </span>
                  ) : (
                    ''
                  )

                let color = 'blue'
                let tag = (
                  <span>
                    <FontAwesomeIcon icon={faUser} />
                    &nbsp; Account{row.deactivated_at && ' - deactivated'}
                  </span>
                )
                if (row.is_owner) {
                  color = 'purple'
                  tag = (
                    <span>
                      <FontAwesomeIcon icon={faUser} />
                      &nbsp; Owner{row.deactivated_at && ' - deactivated'}
                    </span>
                  )
                }
                if (row.is_service_account) {
                  color = 'green'
                  tag = (
                    <span>
                      <FontAwesomeIcon icon={faKey} />
                      &nbsp; Service Account{row.deactivated_at && ' - deactivated'}
                    </span>
                  )
                }
                if (row.deactivated_at) {
                  color = 'red'
                }
                return (
                  <span>
                    <Tag color={color} className={CSS.margin_l_s}>
                      {tag}
                    </Tag>{' '}
                    {dpoTag}
                  </span>
                )
              }

              const invitationExpired = dayjs(row.expires_at).isBefore(dayjs())

              return (
                <span>
                  {invitationExpired && (
                    <Tag color="red" className={CSS.margin_l_s}>
                      <FontAwesomeIcon icon={faHourglassHalf} />
                      &nbsp; Invitation expired
                    </Tag>
                  )}

                  {!invitationExpired && (
                    <Tag color="cyan" className={CSS.margin_l_s}>
                      <FontAwesomeIcon icon={faHourglassHalf} />
                      &nbsp; Invitation sent
                    </Tag>
                  )}
                </span>
              )
            }
          },
          {
            key: 'name',
            width: 200,
            render: (row: Account & OrganizationInvitation) => row.full_name || ''
          },
          {
            key: 'email',
            width: 300,
            render: (row: Account & OrganizationInvitation) => row.email
          },
          // {
          //   key: 'workspaces_scopes',
          //   width: 300,
          //   render: (row: Account & OrganizationInvitation) => {
          //     if (row.id) {
          //       return row.workspaces_scopes.map((scope: WorkspaceScope) => {
          //         let workspace = scope.workspace_id

          //         if (scope.workspace_id === '*') {
          //           workspace = 'All workspaces'
          //         }

          //         return (
          //           <>
          //             <Tag color="magenta">{scope.role}</Tag> of &nbsp;
          //             <Tag color="cyan">{workspace}</Tag>
          //           </>
          //         )
          //       })
          //     }
          //     return ''
          //   }
          // },
          {
            key: 'actions',
            render: (row: Account & OrganizationInvitation) => (
              <div className={CSS.text_right}>
                <Button.Group>
                  {row.id && (
                    <>
                      {isOrganizationOwner &&
                        row.is_owner === false &&
                        !row.deactivated_at &&
                        !row.is_service_account && (
                          <TransferOwnershipButton
                            organizationId={props.organization.id || ''}
                            toAccountId={row.id}
                            apiPOST={accountCtx.apiPOST}
                            onComplete={() => {
                              // remove is_owner
                              const updatedAccount = { ...accountCtx.account?.account } as Account
                              updatedAccount.is_owner = false
                              accountCtx.updateAccountProfile(updatedAccount)

                              refetchAccounts()
                              refetchInvitations()
                            }}
                          />
                        )}

                      {isOrganizationOwner &&
                        row.id !== accountCtx.account?.account.id &&
                        !row.deactivated_at && (
                          <DeactivateAccountButton
                            organizationId={props.organization.id || ''}
                            deactivateAccountId={row.id}
                            apiPOST={accountCtx.apiPOST}
                            onComplete={() => {
                              refetchAccounts()
                              refetchInvitations()
                            }}
                          />
                        )}
                    </>
                  )}

                  {!row.id && isOrganizationOwner && (
                    <>
                      {dayjs(row.expires_at).isBefore(dayjs()) ? (
                        <ResendInvitationButton
                          organizationId={props.organization.id || ''}
                          email={row.email}
                          apiPOST={accountCtx.apiPOST}
                          onComplete={() => {
                            refetchInvitations()
                          }}
                        />
                      ) : (
                        <CancelInvitationButton
                          organizationId={props.organization.id || ''}
                          email={row.email}
                          apiPOST={accountCtx.apiPOST}
                          onComplete={() => {
                            refetchAccounts()
                            refetchInvitations()
                          }}
                        />
                      )}
                    </>
                  )}
                </Button.Group>
              </div>
            )
          }
        ]}
      />
    </Block>
  )
}

export default ListAccounts
