import { Button, Popconfirm, message, Space, Tooltip } from 'antd'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import { useParams } from 'react-router-dom'
import { useEffect, useMemo, useRef, useState } from 'react'
import DeleteAppButton from './button_delete'
import AboutAppButton from './button_about'
import TasksAppButton from './button_tasks'
import { css } from '@emotion/css'
import AppTest from './app_test'
import InstallAppButton from './button_install'
import { App } from 'interfaces'
import { QueryObserverResult } from '@tanstack/react-query'
import manifests from './manifests'
import BlockAppUpgrading from './block_upgrading'

// bottom right fixed
const appMenu = css({
  position: 'fixed',
  bottom: 12,
  right: 12,
  padding: '10px',
  background: 'rgba(255,255,255,0.8)',
  borderRadius: '5px',
  boxShadow: '0 0 10px rgba(0,0,0,0.1)',
  zIndex: 1000
})

const RouteApp = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const params = useParams()
  const [loading, setLoading] = useState(false)
  const [currentApp, setCurrentApp] = useState<App | undefined>(
    workspaceCtx.workspace.apps.find((x) => x.id === params.appId)
  )
  const updatedAtRef = useRef(
    workspaceCtx.workspace.apps.find((x) => x.id === params.appId)?.updated_at
  )
  const refetchAppsRef = useRef(workspaceCtx.refetchApps)

  // compare the app value with the current app to trigger a re-render
  useEffect(() => {
    const app = workspaceCtx.workspace.apps.find((x) => x.id === params.appId)
    if (app?.updated_at !== updatedAtRef.current) {
      updatedAtRef.current = app?.updated_at
      setCurrentApp(app)
    }
  }, [workspaceCtx.workspace.apps, params.appId])

  const stopApp = (app: App) => {
    if (loading) return
    setLoading(true)
    workspaceCtx
      .apiPOST('/app.stop', { id: app.id, workspace_id: workspaceCtx.workspace.id })
      .then(() => {
        workspaceCtx
          .refetchApps()
          .then(() => {
            setLoading(false)
            message.success(app.name + ' has been stopped.')
          })
          .catch(() => {})
      })
      .catch((e) => {
        setLoading(false)
        message.error(e.message)
      })
  }

  const recentManifest = useMemo(() => {
    return manifests.find((x) => x.id === params.appId)
  }, [params.appId])

  return (
    <Layout
      hasIframe={true}
      currentOrganization={workspaceCtx.organization}
      currentWorkspaceCtx={workspaceCtx}
    >
      <>
        {!currentApp && <>App not found</>}
        {currentApp && (
          <>
            <div className={appMenu}>
              <Space>
                {currentApp.status === 'active' && (
                  <Popconfirm
                    title="Stopping the app will disable its tasks, webhooks & processing. Are you sure?"
                    okText="Stop app"
                    placement="bottomRight"
                    onConfirm={stopApp.bind(null, currentApp)}
                  >
                    <Tooltip title="Stop app" placement="left">
                      <Button size="small" danger type="text" loading={loading}>
                        ■
                      </Button>
                    </Tooltip>
                  </Popconfirm>
                )}

                {['stopped', 'initializing'].includes(currentApp.status) && (
                  <DeleteAppButton workspaceCtx={workspaceCtx} manifest={currentApp.manifest} />
                )}
                {recentManifest &&
                  recentManifest.version !== currentApp.manifest.version &&
                  currentApp.status !== 'stopped' &&
                  currentApp.status !== 'upgrading' && (
                    <InstallAppButton
                      workspaceCtx={workspaceCtx}
                      manifest={recentManifest}
                      action={'Upgrade'}
                    />
                  )}
                {currentApp.status === 'stopped' && (
                  <InstallAppButton
                    workspaceCtx={workspaceCtx}
                    manifest={currentApp.manifest}
                    action={'Reinstall'}
                  />
                )}
                {/* {['stopped'].includes(currentApp.status) && (
                <ReactivateAppButton workspaceCtx={workspaceCtx} app={app} />
              )} */}
                <AboutAppButton manifest={currentApp.manifest} installedApp={currentApp} />
                {currentApp.manifest.tasks && (
                  <TasksAppButton app={currentApp} workspaceCtx={workspaceCtx} />
                )}
              </Space>
            </div>

            {currentApp.id === 'app_test' && <AppTest app={currentApp} />}
            {currentApp.status === 'upgrading' && <BlockAppUpgrading app={currentApp} />}
            {!currentApp.is_native && currentApp.status !== 'upgrading' && (
              <AppIframe app={currentApp} refetchApps={refetchAppsRef.current} />
            )}
          </>
        )}
      </>
    </Layout>
  )
}

interface AppIframeProps {
  app: App
  refetchApps: () => Promise<QueryObserverResult<App[], unknown>>
}

const AppIframe = ({ app, refetchApps }: AppIframeProps) => {
  // listen to iframe messages
  useEffect(() => {
    const handler = (ev: MessageEvent<{ type: string; data: any }>) => {
      if (typeof ev.data !== 'object') return
      if (!ev.data.type) return
      if (!ev.data.data) return

      // console.log(ev.data.type, ev.data.data)
      switch (ev.data.type) {
        case 'refreshApp':
          refetchApps()
          break
        default:
      }
    }

    window.addEventListener('message', handler)

    // Don't forget to remove addEventListener
    return () => window.removeEventListener('message', handler)
  }, [refetchApps])

  const url = new URL(app.manifest.ui_endpoint)
  url.searchParams.append('token', app.ui_token || '')

  return (
    <iframe
      title="ui"
      src={url.toString()}
      style={{
        flex: 1,
        width: '100%',
        border: 'none'
      }}
    ></iframe>
  )
}

export default RouteApp
