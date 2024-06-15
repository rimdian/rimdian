import { useState } from 'react'
import { Button, message, Drawer, Space } from 'antd'
import { App, AppManifest } from 'interfaces'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import { useNavigate } from 'react-router-dom'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faArrowUp, faPlus } from '@fortawesome/free-solid-svg-icons'
import BlockAboutApp from './block_about'
import CSS, { backgroundColorBase } from 'utils/css'
import { css } from '@emotion/css'
import { SizeType } from 'antd/lib/config-provider/SizeContext'

type InstallAppButtonProps = {
  manifest: AppManifest
  workspaceCtx: CurrentWorkspaceCtxValue
  action: 'Install' | 'Reinstall' | 'Upgrade'
  size?: SizeType
  ghost?: boolean
}

const InstallAppButton = (props: InstallAppButtonProps) => {
  const [drawerVisible, setDrawerVisible] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const navigate = useNavigate()

  const closeDrawer = () => {
    setDrawerVisible(false)
  }

  const onSubmit = () => {
    if (isLoading) return
    setIsLoading(true)

    // launch a task to upgrade the app
    if (props.action === 'Upgrade') {
      props.workspaceCtx
        .apiPOST('/task.run', {
          id: 'system_upgrade_app',
          workspace_id: props.workspaceCtx.workspace.id,
          app_id: props.manifest.id,
          new_manifest: props.manifest
        })
        .then(() => {
          setIsLoading(false)
          message.success('The upgrade task has been launched!')
        })
        .catch((_) => {
          setIsLoading(false)
        })
    }

    if (props.action === 'Install' || props.action === 'Reinstall') {
      props.workspaceCtx
        .apiPOST('/app.install', {
          workspace_id: props.workspaceCtx.workspace.id,
          manifest: props.manifest,
          reinstall: props.action === 'Reinstall' ? true : false
        })
        .then((app: App) => {
          // reload apps list
          props.workspaceCtx
            .refetchApps()
            .then(() => {
              props.workspaceCtx
                .refreshWorkspace()
                .then(() => {
                  setIsLoading(false)
                  message.success(app.name + ' has been installed.')
                  closeDrawer()
                  // redirect to app
                  navigate(
                    '/orgs/' +
                      props.workspaceCtx.organization.id +
                      '/workspaces/' +
                      props.workspaceCtx.workspace.id +
                      '/apps/' +
                      app.id
                  )
                })
                .catch(() => {})
            })
            .catch(() => {})
        })
        .catch((e) => {
          setIsLoading(false)
        })
    }
  }

  // console.log('initialValues', initialValues);

  let icon = faPlus
  if (props.action === 'Upgrade') {
    icon = faArrowUp
  }

  return (
    <>
      <Button
        type="primary"
        ghost={props.ghost || false}
        size={props.size || 'small'}
        loading={isLoading}
        onClick={() => setDrawerVisible(true)}
        icon={<FontAwesomeIcon icon={icon} className={CSS.padding_r_xs} />}
      >
        {props.action}
      </Button>
      {drawerVisible && (
        <Drawer
          title={
            <>
              <img
                src={props.manifest.icon_url}
                className={css(CSS.appIcon, CSS.margin_r_m)}
                style={{ height: 30 }}
                alt=""
              />
              {props.action} {props.manifest.name}
            </>
          }
          width={960}
          open={true}
          onClose={closeDrawer}
          extra={
            <Space>
              <Button key="a" ghost type="primary" loading={isLoading} onClick={closeDrawer}>
                Cancel
              </Button>
              <Button key="b" loading={isLoading} onClick={onSubmit} type="primary">
                {props.action}
              </Button>
            </Space>
          }
          headerStyle={{ backgroundColor: backgroundColorBase }}
          bodyStyle={{ backgroundColor: backgroundColorBase }}
        >
          <BlockAboutApp manifest={props.manifest} />
        </Drawer>
      )}
    </>
  )
}

export default InstallAppButton
