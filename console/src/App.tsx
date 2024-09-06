// reset CSS
import 'antd/dist/reset.css'

import { useEffect, useState } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom'
import { AuthProvider, useAccount } from 'components/login/context_account'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import RouteLogin from 'components/login/route_login'
import RouteResetPassword from 'components/login/route_reset_password'
import RouteConsumeResetPassword from 'components/login/route_consume_reset_password'
import RouteOrganizations from 'components/organization/route_list'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { ConfigProvider, Spin } from 'antd'
import RouteOrganizationDashboard from 'components/organization/route_dashboard'
import { OrganizationsCtx } from 'components/organization/context_organizations'
import { CurrentOrganizationCtx } from 'components/organization/context_current_organization'
import RouteWorkspaceDashboard from 'components/workspace/route_dashboard'
import { CurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import RouteAcceptInvitation from 'components/account/route_accept_invitation'
import RouteTasks from 'components/task_exec/route_list'
import RouteDataLogs from 'components/data_log/route_list'
import RouteWorkspaceConfiguration from 'components/workspace/route_configuration'
import RouteUsers from 'components/user/route_users'
import RouteAttribution from 'components/attribution/route_attribution'
import RouteListApps from 'components/app/route_list'
import RouteShowApp from 'components/app/route_show'

import { css, CSSInterpolation, injectGlobal } from '@emotion/css'
import CSS from 'utils/css'
import RouteDatabase from 'components/database/route_database'
import RouteAssets from 'components/assets/route_assets'
import RouteBroadcasts from 'components/broadcast/route_broadcast'
// init dayjs
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import timezone from 'dayjs/plugin/timezone'
import relativeTime from 'dayjs/plugin/relativeTime'
import duration from 'dayjs/plugin/duration'
import localizedFormat from 'dayjs/plugin/localizedFormat'
import advancedFormat from 'dayjs/plugin/advancedFormat'
import customParseFormat from 'dayjs/plugin/customParseFormat'
import localeData from 'dayjs/plugin/localeData'
import weekday from 'dayjs/plugin/weekday'
import weekOfYear from 'dayjs/plugin/weekOfYear'
import weekYear from 'dayjs/plugin/weekYear'
dayjs.extend(duration)
dayjs.extend(relativeTime)
dayjs.extend(utc)
dayjs.extend(timezone)
dayjs.extend(localizedFormat)
dayjs.extend(customParseFormat)
dayjs.extend(advancedFormat)
dayjs.extend(weekday)
dayjs.extend(localeData)
dayjs.extend(weekOfYear)
dayjs.extend(weekYear)

// the config is declared in public/config.js in dev
// and is replaced with production variables
// by the API server while serving the Console
export interface Config {
  ENV: 'development' | 'test' | 'production'
  API_ENDPOINT: string
  COLLECTOR_ENDPOINT: string
  CUBEJS_ENDPOINT: string
  MANAGED_RMD: boolean
}

declare global {
  interface Window {
    // cmAgent: any;
    // FB: any;
    MutexCubeJS: {}
    Config: Config
    CM: any
    google: any
  }
  interface Navigator {
    userLanguage: any
  }
}

const queryClient = new QueryClient()

injectGlobal(CSS.GLOBAL as CSSInterpolation)

const appCss = css({ height: '100%' })

const App = () => {
  return (
    <ConfigProvider theme={CSS.AntD}>
      <div className={appCss}>
        <QueryClientProvider client={queryClient}>
          <ReactQueryDevtools initialIsOpen={false} />
          <AuthProvider>
            <BrowserRouter>
              <Routes>
                <Route path="/" element={<RouteHomepage />} />
                <Route path="/reset-password" element={<RouteResetPassword />} />
                <Route path="/consume-reset-password" element={<RouteConsumeResetPassword />} />
                <Route path="/accept-invitation" element={<RouteAcceptInvitation />} />
                <Route path="/logout" element={<RouteLogout />} />

                {/* requires authenticated account */}
                <Route element={<OrganizationsCtx />}>
                  <Route path="/orgs" element={<RouteOrganizations />} />

                  {/* inside an organization */}
                  <Route element={<CurrentOrganizationCtx />}>
                    <Route path="/orgs/:organizationId" element={<RouteOrganizationDashboard />} />

                    {/* inside a workspace */}
                    <Route element={<CurrentWorkspaceCtx />}>
                      <Route
                        path="/orgs/:organizationId/workspaces/:workspaceId"
                        element={<RouteWorkspaceDashboard />}
                      />
                      <Route
                        path="/orgs/:organizationId/workspaces/:workspaceId/attribution"
                        element={<RouteAttribution />}
                      />
                      <Route
                        path="/orgs/:organizationId/workspaces/:workspaceId/users"
                        element={<RouteUsers />}
                      />
                      <Route
                        path="/orgs/:organizationId/workspaces/:workspaceId/broadcasts"
                        element={<RouteBroadcasts />}
                      />
                      <Route
                        path="/orgs/:organizationId/workspaces/:workspaceId/assets"
                        element={<RouteAssets />}
                      />
                      <Route
                        path="/orgs/:organizationId/workspaces/:workspaceId/system/configuration"
                        element={<RouteWorkspaceConfiguration />}
                      />
                      <Route
                        path="/orgs/:organizationId/workspaces/:workspaceId/system/tasks"
                        element={<RouteTasks />}
                      />
                      <Route
                        path="/orgs/:organizationId/workspaces/:workspaceId/system/data-logs"
                        element={<RouteDataLogs />}
                      />
                      <Route
                        path="/orgs/:organizationId/workspaces/:workspaceId/system/database"
                        element={<RouteDatabase />}
                      />
                      <Route
                        path="/orgs/:organizationId/workspaces/:workspaceId/apps/:appId"
                        element={<RouteShowApp />}
                      />
                      <Route
                        path="/orgs/:organizationId/workspaces/:workspaceId/apps"
                        element={<RouteListApps />}
                      />
                    </Route>
                  </Route>
                </Route>
              </Routes>
            </BrowserRouter>
          </AuthProvider>
        </QueryClientProvider>
      </div>
    </ConfigProvider>
  )
}

const RouteLogout = () => {
  const [loading, setLoading] = useState<boolean>(true)
  const accountCtx = useAccount()
  const navigate = useNavigate()

  useEffect(() => {
    if (accountCtx.account) {
      setLoading(true)
      accountCtx
        .logout(accountCtx.account)
        .then(() => {
          setLoading(false)
        })
        .catch(() => {})
    } else {
      navigate('/')
    }
  }, [accountCtx, navigate])

  return (
    <Spin spinning={loading} tip="Logout...">
      <RouteLogin />
    </Spin>
  )
}

const RouteHomepage = () => {
  const accountCtx = useAccount()

  if (accountCtx.account) {
    return <Navigate to="/orgs" replace={true} />
  }

  return <RouteLogin />
}

export default App
