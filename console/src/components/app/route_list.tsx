import { Button } from 'antd'
import { AppManifest } from 'interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faArrowRight } from '@fortawesome/free-solid-svg-icons'
import { useNavigate } from 'react-router-dom'
import InstallAppButton from './button_install'
import CSS from 'utils/css'
import { css } from '@emotion/css'
import Block from 'components/common/block'
import customApp from 'images/custom_app.png'
import PrivateAppButton from './button_private_app'
import manifests from './manifests'

const appCss = {
  name: css({
    fontWeight: 600,
    fontSize: '16px'
  }),

  description: css({
    overflow: 'auto'
  })
}

const appIcon = css({
  width: '70px',
  height: '70px'
})
const RouteApps = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const navigate = useNavigate()

  // new apps are native apps rewritten outside of the console
  // in dev we rewrite the endpoints to localhost
  if (window.Config.ENV === 'development') {
    manifests.forEach((app) => {
      app.ui_endpoint = 'https://localhost:3000'
      app.webhook_endpoint = 'https://localhost:3000/api/webhooks'
    })
  }

  return (
    <Layout
      currentOrganization={workspaceCtx.organization}
      currentWorkspace={workspaceCtx.workspace}
    >
      <div className={CSS.container} style={{ width: 600 }}>
        <div className={CSS.top}>
          <h1>Apps</h1>
        </div>

        <>
          <div>
            <Block classNames={[CSS.padding_a_l, CSS.margin_r_m]}>
              <table>
                <tbody>
                  <tr>
                    <td style={{ width: 90 }}>
                      <img src={customApp} alt="Private app" className={appIcon} />
                    </td>
                    <td>
                      <p className={appCss.name}>Private app</p>
                      <p className={appCss.description}>Create a private app from a config file.</p>
                    </td>
                    <td className={CSS.text_right + ' ' + CSS.padding_a_m}>
                      <PrivateAppButton workspaceCtx={workspaceCtx} />
                    </td>
                  </tr>
                </tbody>
              </table>
            </Block>
          </div>
          {manifests.map((app: AppManifest) => {
            const existingApp = workspaceCtx.workspace.apps.find((x) => x.id === app.id)

            return (
              <div key={app.id}>
                <Block classNames={[CSS.padding_a_l, CSS.margin_r_m]}>
                  <table>
                    <tbody>
                      <tr>
                        <td style={{ width: 90 }}>
                          <img src={app.icon_url} alt={app.name} className={appIcon} />
                        </td>
                        <td>
                          <p className={appCss.name}>{app.name}</p>
                          <p className={appCss.description}>{app.short_description}</p>
                        </td>
                        <td className={CSS.text_right + ' ' + CSS.padding_a_m}>
                          {!existingApp && (
                            <InstallAppButton
                              size="middle"
                              ghost={false}
                              workspaceCtx={workspaceCtx}
                              manifest={app}
                              action="Install"
                            />
                          )}
                          {existingApp && existingApp.status === 'stopped' && (
                            <InstallAppButton
                              workspaceCtx={workspaceCtx}
                              manifest={app}
                              action="Reinstall"
                              size="middle"
                            />
                          )}
                          {existingApp && existingApp.status !== 'stopped' && (
                            <Button
                              type="primary"
                              ghost
                              icon={
                                <FontAwesomeIcon icon={faArrowRight} className={CSS.padding_r_xs} />
                              }
                              onClick={() =>
                                navigate(
                                  '/orgs/' +
                                    workspaceCtx.organization.id +
                                    '/workspaces/' +
                                    workspaceCtx.workspace.id +
                                    '/apps/' +
                                    existingApp.id
                                )
                              }
                            >
                              View
                            </Button>
                          )}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </Block>
              </div>
            )
          })}
        </>
      </div>
    </Layout>
  )
}

export default RouteApps
