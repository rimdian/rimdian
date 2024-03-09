import { DataLogBatch, Organization, Workspace } from 'interfaces'
import { useOutletContext, useParams, Outlet } from 'react-router-dom'
import { useOrganizationsCtx } from './context_organizations'
import { QueryObserverResult, useQuery } from '@tanstack/react-query'
import Layout from 'components/common/layout'

// This context forwards organizationsContext values,
// provides the current organization
// and blocks until the current organization is ready

export interface CurrentOrganizationCtxValue {
  organization: Organization
  // forwards orgs ctx
  organizations: Organization[]
  updateOrganization: (org: Organization) => void
  workspaces: Workspace[]
  refreshOrganizations: () => Promise<void>
  refreshWorkspaces: () => Promise<QueryObserverResult<Workspace[], unknown>>
  apiGET: (endpoint: string) => Promise<any>
  apiPOST: (endpoint: string, data: any) => Promise<any>
  collectorPOST: (sync: boolean, batch: DataLogBatch) => Promise<any>
}

export const CurrentOrganizationCtx = () => {
  const orgCtx = useOrganizationsCtx()
  const params = useParams()

  const organization = orgCtx.organizations.find(
    (x: Organization) => x.id === params.organizationId
  )

  const { isLoading, data, refetch } = useQuery<Workspace[]>(
    ['workspaces', params.organizationId],
    (): Promise<Workspace[]> => {
      return new Promise((resolve, reject) => {
        orgCtx
          .apiGET('/workspace.list?organization_id=' + params.organizationId)
          .then((data: any) => {
            resolve(data.workspaces as Workspace[])
          })
          .catch((e) => {
            reject(e)
          })
      })
    },
    { enabled: organization ? true : false }
  )

  const ctx: CurrentOrganizationCtxValue = {
    // undefined currentOrganization will block childrens rendering
    organization: organization as Organization,
    // forwards orgs ctx
    organizations: orgCtx.organizations,
    updateOrganization: orgCtx.updateOrganization,
    workspaces: data as Workspace[],
    refreshWorkspaces: refetch,
    refreshOrganizations: orgCtx.refreshOrganizations,
    apiGET: orgCtx.apiGET,
    apiPOST: orgCtx.apiPOST,
    collectorPOST: orgCtx.collectorPOST
  }

  if (!organization) {
    return <Layout loadingText="Loading organization..." />
  }

  if (isLoading) {
    return <Layout loadingText="Loading workspaces..." />
  }

  return <Outlet context={ctx} />
}

export function useCurrentOrganizationCtx() {
  return useOutletContext<CurrentOrganizationCtxValue>()
}
