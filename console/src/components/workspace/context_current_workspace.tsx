import { useEffect, useMemo, useRef, useState } from 'react'
import { Organization, Workspace, App, CubeSchema, DataLogBatch } from 'interfaces'
import { Outlet, useOutletContext, useParams, useSearchParams } from 'react-router-dom'
import { useCurrentOrganizationCtx } from 'components/organization/context_current_organization'
import RouteWorkspaceSetup from './route_setup'
import { QueryObserverResult, useQuery } from '@tanstack/react-query'
import Layout from 'components/common/layout'
import { BadgeRunningTasks } from 'components/task_exec/badge_running_tasks'
import { useAccount } from 'components/login/context_account'
import DrawerShowUser from 'components/user/drawer_show'
import { forEach } from 'lodash'
import cubejs, { CubejsApi } from '@cubejs-client/core'
import { CubeProvider } from '@cubejs-client/react'
import { Segment } from 'components/segment/interfaces'
import LicenseWarning from './block_license_warning'

interface SegmentList {
  segments: Segment[]
}

interface CubeSchemaList {
  schemas: CubeSchema[]
}

export interface CurrentWorkspaceCtxValue {
  workspace: Workspace
  refreshWorkspace: () => Promise<QueryObserverResult<Workspace, unknown>>
  isRefreshingWorkspace: boolean
  // forward currentOrg context
  organizations: Organization[]
  // refreshOrganizations: () => Promise<void>
  organization: Organization
  updateOrganization: (org: Organization) => void
  workspaces: Workspace[]
  segmentsMap: { [key: string]: Segment }
  refetchSegments: () => Promise<QueryObserverResult<SegmentList, unknown>>
  cubeSchemasMap: { [key: string]: CubeSchema }
  refetchApps: () => Promise<QueryObserverResult<App[], unknown>>
  // refreshWorkspaces: Promise<QueryObserverResult<Workspace[], unknown>>
  apiGET: (endpoint: string) => Promise<any>
  apiPOST: (endpoint: string, data: any) => Promise<any>
  collectorPOST: (sync: boolean, batch: DataLogBatch) => Promise<any>
}

export const CurrentWorkspaceCtx = () => {
  const currentOrgCtx = useCurrentOrganizationCtx()
  const accountCtx = useAccount()
  const params = useParams()
  const [searchParams] = useSearchParams()
  const [segmentsMap, setSegmentsMap] = useState<{ [key: string]: Segment }>({})
  const [cubeSchemasMap, setCubeSchemasMap] = useState<{ [key: string]: CubeSchema }>({})
  const cubejsApiRef = useRef<CubejsApi | null>(null)

  const { isLoading, data, refetch, isFetching } = useQuery<Workspace>(
    ['workspace', params.workspaceId],
    (): Promise<Workspace> => {
      return new Promise((resolve, reject) => {
        currentOrgCtx
          .apiGET('/workspace.show?workspace_id=' + params.workspaceId)
          .then((data: any) => {
            resolve(data.workspace as Workspace)
          })
          .catch((e) => {
            reject(e)
          })
      })
    },
    { enabled: params.workspaceId && currentOrgCtx.workspaces.length > 0 ? true : false }
  )

  // segments
  const { isLoading: isLoadingSegments, refetch: refetchSegments } = useQuery<SegmentList>(
    ['segments', params.workspaceId],
    (): Promise<SegmentList> => {
      return new Promise((resolve, reject) => {
        currentOrgCtx
          .apiGET('/segment.list?with_users_count=true&workspace_id=' + params.workspaceId)
          .then((data: any) => {
            // convert data.segments to segmentsMap
            const segmentsMap: { [key: string]: Segment } = {}
            forEach(data.segments, (segment: Segment) => {
              segmentsMap[segment.id] = segment
            })
            setSegmentsMap(segmentsMap)
            resolve(data as SegmentList)
          })
          .catch((e) => {
            reject(e)
          })
      })
    }
  )

  // Cube schemas
  const { isLoading: isLoadingCubeSchemas } = useQuery<CubeSchemaList>(
    ['cube_schemas', params.workspaceId],
    (): Promise<CubeSchemaList> => {
      return new Promise((resolve, reject) => {
        currentOrgCtx
          .apiGET('/cubejs.schemas?workspace_id=' + params.workspaceId)
          .then((data: any) => {
            // convert data.segments to segmentsMap
            const schemasMap: { [key: string]: CubeSchema } = {}
            forEach(data, (entry: any) => {
              try {
                const schema = JSON.parse(entry.content)
                schemasMap[entry.fileName] = schema
              } catch (e) {
                console.error(e)
              }
            })
            setCubeSchemasMap(schemasMap)
            resolve(data as CubeSchemaList)
          })
          .catch((e) => {
            reject(e)
          })
      })
    }
  )

  // apps
  const {
    isLoading: isLoadingApps,
    data: apps,
    refetch: refetchApps
  } = useQuery<App[]>(['apps', params.workspaceId], (): Promise<App[]> => {
    return new Promise((resolve, reject) => {
      currentOrgCtx
        .apiGET('/app.list?workspace_id=' + params.workspaceId)
        .then((data: any) => {
          resolve(data.apps)
        })
        .catch((e) => {
          reject(e)
        })
    })
  })

  // merge apps into workspace data
  const workspace = useMemo(() => {
    if (!data) return null
    return { ...data, apps: apps || [] }
  }, [data, apps])

  useEffect(() => {
    if (!workspace) return
    if (cubejsApiRef.current) return
    cubejsApiRef.current = cubejs(workspace.cubejs_token || '', {
      apiUrl: window.Config.CUBEJS_ENDPOINT + '/cubejs-api/v1'
    })
  }, [workspace])

  if (isLoading || !workspace || isLoadingSegments || isLoadingCubeSchemas || isLoadingApps) {
    return (
      <Layout currentOrganization={currentOrgCtx.organization} loadingText="Loading workspace..." />
    )
  }

  // const cubejsApi = cubejs('Bearer ' + accountCtx.account?.access_token, {
  //   apiUrl: window.Config.API_ENDPOINT + '/cubejs',
  //   headers: {
  //     'X-Rmd-Workspace-Id': params.workspaceId
  //   } as any
  // })

  let isReady = true
  if (workspace.domains.length === 0) isReady = false
  if (workspace.has_orders === false && workspace.has_leads === false) isReady = false
  if (!cubejsApiRef.current) isReady = false

  if (!isReady) {
    return (
      <RouteWorkspaceSetup
        workspace={workspace as Workspace}
        organization={currentOrgCtx.organization}
        refreshWorkspace={refetch}
        apiGET={currentOrgCtx.apiGET}
        apiPOST={currentOrgCtx.apiPOST}
        collectorPOST={currentOrgCtx.collectorPOST}
      />
    )
  }

  const ctx: CurrentWorkspaceCtxValue = {
    // undefined currentWorkspace won't display props.children
    workspace: workspace as Workspace,
    isRefreshingWorkspace: isFetching,
    // forward currentOrg context
    organization: currentOrgCtx.organization,
    organizations: currentOrgCtx.organizations,
    updateOrganization: currentOrgCtx.updateOrganization,
    workspaces: currentOrgCtx.workspaces,
    segmentsMap: segmentsMap,
    refetchSegments: refetchSegments,
    cubeSchemasMap: cubeSchemasMap,
    refetchApps: refetchApps,
    refreshWorkspace: refetch,
    apiGET: currentOrgCtx.apiGET,
    apiPOST: currentOrgCtx.apiPOST,
    collectorPOST: currentOrgCtx.collectorPOST
  }

  const showUserId = searchParams.get('showUser')

  return (
    <>
      <CubeProvider cubejsApi={cubejsApiRef.current}>
        <Outlet context={ctx} />

        {showUserId && showUserId !== '' && (
          <DrawerShowUser workspaceCtx={ctx} userExternalId={showUserId} />
        )}
        <LicenseWarning workspaceCtx={ctx} />
        <BadgeRunningTasks
          organizationId={currentOrgCtx.organization.id}
          workspaceId={workspace.id}
          workspaces={currentOrgCtx.workspaces}
          refetchSegments={refetchSegments}
          accountTimezone={accountCtx.account?.account.timezone as string}
          apiGET={currentOrgCtx.apiGET}
        />
      </CubeProvider>
    </>
  )
}

export function useCurrentWorkspaceCtx() {
  return useOutletContext<CurrentWorkspaceCtxValue>()
}
