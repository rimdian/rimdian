import { useState, useEffect, useCallback } from 'react'
import Axios from 'axios'
import { DataLogBatch, Organization } from 'interfaces'
import { useAccount } from 'components/login/context_account'
import { Spin } from 'antd'
import { Outlet, useOutletContext, useParams } from 'react-router-dom'
import LoginScreen from 'components/login/route_login'

export interface OrganizationsCtxValue {
  organizations: Organization[]
  updateOrganization: (org: Organization) => void
  currentOrganizationId: string | undefined
  refreshOrganizations: () => Promise<void>
  apiGET: (endpoint: string) => Promise<any>
  apiPOST: (endpoint: string, data: any) => Promise<any>
  collectorPOST: (sync: boolean, batch: DataLogBatch) => Promise<any>
}

export const OrganizationsCtx = () => {
  const [organizations, setOrganizations] = useState<Organization[]>([])
  const [loadingOrganizations, setLoadingOrganizations] = useState(true)
  const accountCtx = useAccount()
  const params = useParams()
  const [currentOrganizationId, setCurrentOrganizationId] = useState<string | undefined>()

  const updateOrganization = (updatedOrg: Organization) => {
    const orgs = organizations.filter((x) => x.id !== updatedOrg.id)
    orgs.push(updatedOrg)
    setOrganizations(orgs)
  }

  const loadOrganizations = useCallback(
    (accessToken: string): Promise<void> => {
      return new Promise<void>((resolve, reject) => {
        Axios.get(window.Config.API_ENDPOINT + '/organization.list', {
          headers: { Authorization: 'Bearer ' + accessToken }
        })
          .then((res) => {
            const organizations = res.data.organizations as Organization[]

            setOrganizations(organizations)
            resolve()
          })
          .catch((e) => {
            reject(e)
          })
      })
    },
    [setOrganizations]
  )

  const refreshOrganizations = (): Promise<void> => {
    return loadOrganizations(accountCtx.account?.access_token as string)
  }

  // load organization when account is accountCtxenticated
  useEffect(() => {
    if (!accountCtx.account) {
      return
    }

    loadOrganizations(accountCtx.account.access_token)
      .then(() => setLoadingOrganizations(false))
      .catch((_e) => setLoadingOrganizations(false))
  }, [loadOrganizations, accountCtx])

  // load/clean workspaces when we enter/change/leave an organization
  useEffect(() => {
    if (!accountCtx.account) {
      return
    }

    // clean on org exit
    if (!params.organizationId || params.organizationId === '') {
      if (currentOrganizationId) {
        setCurrentOrganizationId(undefined)
      }
      return
    }

    // }, [accountCtx.account, currentOrganizationId, loadWorkspaces, params.organizationId])
  }, [accountCtx.account, currentOrganizationId, params.organizationId])

  const ctx: OrganizationsCtxValue = {
    organizations,
    updateOrganization,
    currentOrganizationId,
    refreshOrganizations,
    apiGET: accountCtx.apiGET,
    apiPOST: accountCtx.apiPOST,
    collectorPOST: accountCtx.collectorPOST
  }

  if (!accountCtx.account) {
    return <LoginScreen />
  }

  if (loadingOrganizations) {
    return (
      <div style={{ textAlign: 'center', paddingTop: 200 }}>
        <Spin size="large" tip="Fetching organization..." />
      </div>
    )
  }

  return <Outlet context={ctx} />
}

export function useOrganizationsCtx() {
  return useOutletContext<OrganizationsCtxValue>()
}
