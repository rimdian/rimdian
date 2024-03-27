import { useState } from 'react'
import {
  Dropdown,
  // Avatar,
  Select,
  // Badge,
  Drawer,
  Button,
  Input,
  Form,
  message,
  Spin,
  Tooltip,
  Space,
  Popover
} from 'antd'
import { AccountContextValue, useAccount } from 'components/login/context_account'
import { useNavigate, useMatch, useParams } from 'react-router-dom'
import { truncate } from 'lodash'
import Axios from 'axios'
import { HandleAxiosError } from 'utils/errors'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faPlus,
  faChevronDown,
  faCircleUser,
  faCog,
  faPowerOff,
  faWaveSquare,
  faUserGroup,
  faGear,
  faBorderAll,
  faDatabase,
  faAnglesDown,
  faFolderOpen
} from '@fortawesome/free-solid-svg-icons'
import { faSquareCheck } from '@fortawesome/free-regular-svg-icons'
import { Timezones } from 'utils/countries_timezones'
import Messages from 'utils/formMessages'
import { App, Organization, Workspace } from 'interfaces'
import { css } from '@emotion/css'
import CSS from 'utils/css'

const topbarCss = {
  // parent
  self: css({
    flex: '0 0 auto',
    height: '60px',
    overflow: 'hidden',
    backgroundImage: 'linear-gradient(90deg, #4138be, #5e79ff)'
  }),

  fixed: css({
    position: 'fixed',
    top: 0,
    left: 0,
    right: 0
  }),

  logo: css(
    {
      display: 'inline-block',
      verticalAlign: 'top',
      paddingTop: '12px',
      height: '60px',
      lineHeight: '60px'
    },
    CSS.margin_h_l
  ),

  block: css({
    display: 'inline-block',
    color: 'white',
    padding: '0 24px',
    transition: 'all 0.3s ease',
    cursor: 'pointer',
    borderLeft: '1px solid rgba(255, 255, 255, 0.3)',
    borderRight: '1px solid rgba(255, 255, 255, 0.3)',
    '&:hover': {
      backgroundColor: 'rgba(0, 0, 0, 0.1)'
    }
  }),

  itemName: css({
    display: 'table-cell',
    lineHeight: '60px'
  }),

  itemNameTop: css({
    display: 'inline-block',
    lineHeight: '16px',
    verticalAlign: 'middle'
  }),

  itemIcon: css({
    fontSize: '10px',
    display: 'table-cell',
    lineHeight: '60px'
  }),

  userWrapper: css({
    float: 'right',
    height: '60px'
  }),

  user: css(
    {
      display: 'table',
      color: 'white',
      transition: 'all 0.3s ease',
      cursor: 'pointer',
      '&:hover': {
        backgroundColor: 'rgba(0, 0, 0, 0.1)'
      }
    },
    CSS.padding_h_l
  ),

  userPicture: css({
    display: 'table-cell',
    lineHeight: '60px',
    paddingRight: CSS.M
  }),

  userName: css({
    display: 'table-cell',
    lineHeight: '60px'
  }),

  userNameTop: css({
    display: 'inline-block',
    lineHeight: '16px',
    verticalAlign: 'middle'
  }),

  userNameBottom: css({
    fontSize: '10px'
  }),

  userButton: css({
    fontSize: '10px',
    display: 'table-cell',
    lineHeight: '60px',
    paddingLeft: CSS.L,
    marginLeft: CSS.S
  })
}

type LayoutProps = {
  loadingText?: string
  currentOrganization?: Organization
  currentWorkspace?: Workspace
  children?: JSX.Element[] | JSX.Element
  beforeContent?: JSX.Element[] | JSX.Element
  hasIframe?: boolean // app iframe
}

const layoutContentCss = css(CSS.margin_t_s, CSS.margin_r_l, CSS.margin_l_l, CSS.padding_b_xl)

const loaderCss = css({
  textAlign: 'center',
  paddingTop: 200
})
const dotSelected = (
  <div
    style={{
      width: 4,
      height: 4,
      borderRadius: 4,
      background: '#FFF',
      position: 'absolute',
      marginLeft: '12px',
      marginTop: '4px',
      opacity: 0.5
    }}
  ></div>
)

const Layout: React.FC<LayoutProps> = (props) => {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <Topbar
        currentOrganization={props.currentOrganization}
        currentWorkspace={props.currentWorkspace}
        beforeContent={props.beforeContent}
      />
      {props.hasIframe && props.children}

      {!props.hasIframe && (
        <div className={props.hasIframe ? '' : layoutContentCss}>
          {props.loadingText && (
            <div className={loaderCss}>
              <Spin size="large" tip={props.loadingText} />
            </div>
          )}
          {!props.loadingText && props.children}
        </div>
      )}
    </div>
  )
}

type TopbarProps = {
  currentOrganization?: Organization
  currentWorkspace?: Workspace
  beforeContent?: JSX.Element[] | JSX.Element
}

type AppItemProps = {
  icon: JSX.Element
  title: string
  route: string
  inTooltip?: boolean
}

const appItemCss = {
  self: css({
    display: 'inline-block',
    verticalAlign: 'top'
    // marginLeft: CSS.L
  }),

  icon: css({ marginTop: '15px' }),

  selected: css({
    // borderBottom: '1px solid #5e79ff',
    boxShadow: 'none'
  })
}

const AppItem = (props: AppItemProps) => {
  const params = useParams()
  const navigate = useNavigate()

  const onClick = () => {
    navigate(
      props.route
        .replace(':organizationId', params.organizationId as string)
        .replace(':workspaceId', params.workspaceId as string)
    )
  }

  const matchRoute = useMatch({ path: props.route, end: true })

  return (
    <Tooltip title={props.title} placement="bottom">
      <div className={appItemCss.self} onClick={onClick}>
        <div
          className={css([
            CSS.appIcon,
            !props.inTooltip ? appItemCss.icon : null,
            matchRoute && appItemCss.selected
          ])}
        >
          {props.icon}
          {matchRoute && (
            <div
              style={{
                width: 4,
                height: 4,
                borderRadius: 4,
                background: props.inTooltip ? '#000' : '#FFF',
                position: 'absolute',
                marginLeft: '12px',
                marginTop: '4px',
                opacity: 0.5
              }}
            ></div>
          )}
        </div>
      </div>
    </Tooltip>
  )
}

const Topbar = (props: TopbarProps) => {
  const accountCtx = useAccount()
  const [accountSettingsVisible, setAccountSettingsVisible] = useState(false)
  const navigate = useNavigate()

  const toggleAccountSettings = () => {
    setAccountSettingsVisible(!accountSettingsVisible)
  }

  // const organizationMenuSelected = useMatch({ path: '/orgs/:organizationId', end: true })

  const apps: AppItemProps[] = []

  const matchSystemRoute = useMatch({
    path: '/orgs/:organizationId/workspaces/:workspaceId/system',
    end: false
  })

  if (props.currentWorkspace) {
    props.currentWorkspace.apps.forEach((app: App) => {
      apps.push({
        icon: <img src={app.manifest.icon_url} alt="" />,
        title: app.name,
        route: `/orgs/:organizationId/workspaces/:workspaceId/apps/${app.id}`
      })
    })
  }

  const spacer = <div style={{ display: 'inline-block', width: CSS.M }}></div>

  return (
    <div>
      <div className={topbarCss.self}>
        <div className={topbarCss.logo}>
          <svg width="30px" height="30px" viewBox="0 0 291 291" version="1.1">
            <g stroke="none" strokeWidth="1" fill="none" fillRule="evenodd">
              <g transform="translate(33.000000, 0.000000)" fill="#FFFFFF">
                <path d="M20.0893628,294 L4.20365898,291.335975 C13.8972054,263.997971 24.4157564,243.599795 35.759312,230.141446 C8.4108739,166.340529 -3.25066284,132.368731 0.774701765,128.226052 C4.80006637,124.083374 10.5017178,120.008023 17.879656,116 C35.7984323,170.220018 48.1718803,206.85737 55,225.912058 C40.7591683,250.60389 29.1222892,273.29987 20.0893628,294 Z"></path>
                <path d="M30,102.880345 L63.580817,209 L74,191.265718 C69.7875085,165.585733 65.9113307,147.257089 62.3714664,136.279785 C57.0616701,119.81383 53.567225,88.5169457 53.567225,87.1069275 C53.567225,86.1669154 45.7114833,91.4247212 30,102.880345 Z"></path>
                <path d="M85.7035061,171.138454 C75.234502,119.473808 70,88.3177125 70,77.6701664 L98.5618478,61.9324265 L95.462367,70.6601375 C155.286055,34.9795232 198.465266,11.426144 225,0 L204.506339,100.894098 C173.238495,117.485471 152.740595,130.472341 143.012641,139.85471 C170.189171,131.968097 188.017101,126.365194 196.496433,123.046001 L180.869752,176.393724 C127.969474,198.071583 95.3065982,210.607008 82.8811242,214 L77.2421885,214 L146.174641,94.9137251 L143.012641,94.0366108 L85.7035061,171.138454 Z"></path>
                <polygon points="111 221 168.324927 216.800661 174 197"></polygon>
                <path d="M53,264 C110.298712,259.995813 144.174551,256.790582 154.627518,254.384308 L160,234 C109.512576,237.705399 78.5611317,239.136176 67.1456658,238.29233 L53,264 Z"></path>
              </g>
            </g>
          </svg>
        </div>

        {props.currentOrganization && (
          <div
            className={topbarCss.block}
            onClick={() => navigate('/orgs/' + props.currentOrganization?.id)}
          >
            <div className={topbarCss.itemName}>
              <div className={topbarCss.itemNameTop}>
                {truncate(props.currentOrganization.name, { length: 10 })}
              </div>
            </div>
          </div>
        )}

        {props.currentWorkspace && (
          <>
            {spacer}
            <AppItem
              icon={
                <div
                  style={{
                    color: '#FFF',
                    background: 'linear-gradient(to bottom, #00c6ff, #0072ff)'
                  }}
                >
                  <FontAwesomeIcon icon={faBorderAll} />
                </div>
              }
              title="Dashboard"
              route="/orgs/:organizationId/workspaces/:workspaceId"
            />
            {spacer}
            <AppItem
              icon={
                <div
                  style={{
                    color: '#FFF',
                    background: 'linear-gradient(to bottom, #f953c6, #b91d73)'
                  }}
                >
                  <FontAwesomeIcon icon={faWaveSquare} />
                </div>
              }
              title="Attribution"
              route="/orgs/:organizationId/workspaces/:workspaceId/attribution"
            />
            {spacer}
            <AppItem
              icon={
                <div
                  style={{
                    color: '#FFF',
                    background: 'linear-gradient(to top, #004D40, #26A69A)'
                  }}
                >
                  <FontAwesomeIcon icon={faUserGroup} />
                </div>
              }
              title="Users"
              route="/orgs/:organizationId/workspaces/:workspaceId/users"
            />
            {spacer}
            <AppItem
              icon={
                <div
                  style={{
                    color: '#FFF',
                    background: 'linear-gradient(to top, #FF5F6D, #FFC371)'
                  }}
                >
                  <FontAwesomeIcon icon={faFolderOpen} />
                </div>
              }
              title="Files & templates"
              route="/orgs/:organizationId/workspaces/:workspaceId/assets"
            />
            {spacer}
            <div className={appItemCss.self}>
              <Popover
                content={
                  <div>
                    <AppItem
                      inTooltip={true}
                      icon={
                        <div
                          style={{
                            color: '#212121',
                            background: 'linear-gradient(to top, #FF8F00, #FFCA28)'
                          }}
                        >
                          <FontAwesomeIcon icon={faSquareCheck} />
                        </div>
                      }
                      title="Tasks"
                      route="/orgs/:organizationId/workspaces/:workspaceId/system/tasks"
                    />
                    {spacer}
                    <AppItem
                      inTooltip={true}
                      icon={
                        <div
                          style={{
                            color: '#FFF',
                            background: 'linear-gradient(to top, #4A148C, #AB47BC)'
                          }}
                        >
                          <FontAwesomeIcon icon={faAnglesDown} />
                        </div>
                      }
                      title="Data logs"
                      route="/orgs/:organizationId/workspaces/:workspaceId/system/data-logs"
                    />
                    {spacer}
                    <AppItem
                      inTooltip={true}
                      icon={
                        <div
                          style={{
                            color: '#FFF',
                            background: 'linear-gradient(to top, #283048, #859398)'
                          }}
                        >
                          <FontAwesomeIcon icon={faDatabase} />
                        </div>
                      }
                      title="Database"
                      route="/orgs/:organizationId/workspaces/:workspaceId/system/database"
                    />
                    {spacer}
                    <AppItem
                      inTooltip={true}
                      icon={
                        <div
                          style={{
                            color: '#FFF',
                            background: 'linear-gradient(to top, #1A237E, #5C6BC0)'
                          }}
                        >
                          <FontAwesomeIcon icon={faGear} />
                        </div>
                      }
                      title="Configuration"
                      route="/orgs/:organizationId/workspaces/:workspaceId/system/configuration"
                    />
                  </div>
                }
                placement="bottom"
              >
                <div
                  className={css([
                    CSS.appIcon,
                    appItemCss.icon,
                    matchSystemRoute && appItemCss.selected
                  ])}
                >
                  <div
                    style={{
                      color: '#FFF',
                      background: 'linear-gradient(to top, #1A237E, #5C6BC0)'
                    }}
                  >
                    <FontAwesomeIcon icon={faGear} />
                    {matchSystemRoute && dotSelected}
                  </div>
                </div>
              </Popover>
            </div>
            {spacer}
            {apps.map((app) => (
              <span key={app.route}>
                <AppItem icon={app.icon} title={app.title} route={app.route} />
                {spacer}
              </span>
            ))}
            <AppItem
              icon={
                <div
                  style={{
                    color: '#4E6CFF',
                    background: 'white'
                  }}
                >
                  <FontAwesomeIcon icon={faPlus} />
                </div>
              }
              title="Add an app"
              route="/orgs/:organizationId/workspaces/:workspaceId/apps"
            />
          </>
        )}

        <div className={topbarCss.userWrapper}>
          {accountSettingsVisible && (
            <UserSettings toggleSettings={toggleAccountSettings} accountCtx={accountCtx} />
          )}
          <Dropdown
            menu={{
              items: [
                {
                  key: 'user-settings',
                  label: (
                    <div className={CSS.margin_a_s} onClick={toggleAccountSettings}>
                      <FontAwesomeIcon icon={faCog} />
                      &nbsp; My settings
                    </div>
                  )
                },
                {
                  key: 'logout',
                  label: (
                    <div className={CSS.margin_a_s} onClick={() => navigate('/logout')}>
                      <FontAwesomeIcon icon={faPowerOff} />
                      &nbsp; Sign out
                    </div>
                  )
                }
              ]
            }}
          >
            <div className={topbarCss.user}>
              <div className={topbarCss.userPicture}>
                <FontAwesomeIcon
                  icon={faCircleUser}
                  style={{ fontSize: 20, lineHeight: 60, verticalAlign: 'middle' }}
                />
                {/* <Avatar style={{ backgroundColor: '#1565C0' }} icon={} /> */}
              </div>
              <div className={topbarCss.userName}>
                <div className={topbarCss.userNameTop}>
                  {accountCtx.account?.account.full_name ||
                    truncate(accountCtx.account?.account.email, { length: 15 })}
                  <br />
                  <span className={topbarCss.userNameBottom}>
                    {accountCtx.account?.account.timezone}
                  </span>
                </div>
              </div>
              <div className={topbarCss.userButton}>
                <FontAwesomeIcon icon={faChevronDown} />
              </div>
            </div>
          </Dropdown>
        </div>
      </div>
      {props.beforeContent && <div>{props.beforeContent}</div>}
    </div>
  )
}

// import { useAppContext, AppContextValue } from '../app'

type UserSettingsProps = {
  toggleSettings: () => void
  accountCtx: AccountContextValue
}

const UserSettings = (props: UserSettingsProps) => {
  //   const appCtx: AppContextValue = useAppContext()
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  const onFinish = (values: any) => {
    setLoading(true)

    Axios.post(window.Config.API_ENDPOINT + '/account.setProfile', values, {
      headers: { Authorization: 'Bearer ' + props.accountCtx.account?.access_token }
    })
      .then((_res) => {
        setLoading(false)
        message.success('Your profile has been updated!')
        props.accountCtx.updateAccountProfile(values)
        props.toggleSettings()
      })
      .catch((e) => {
        HandleAxiosError(e)
        setLoading(false)
      })
  }

  const initialValues = {
    full_name: props.accountCtx.account?.account.full_name,
    timezone: props.accountCtx.account?.account.timezone,
    locale: props.accountCtx.account?.account.locale
  }

  return (
    <Drawer
      title="My settings"
      open={true}
      onClose={props.toggleSettings}
      width={600}
      extra={
        <Space>
          <Button loading={loading} onClick={props.toggleSettings}>
            Cancel
          </Button>
          <Button
            loading={loading}
            onClick={() => {
              form
                .validateFields()
                .then(onFinish)
                .catch(() => {})
            }}
            type="primary"
          >
            Save
          </Button>
        </Space>
      }
    >
      <Form form={form} initialValues={initialValues} onFinish={onFinish} layout="vertical">
        <Form.Item
          name="full_name"
          label="Full name"
          rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
        >
          <Input />
        </Form.Item>

        <Form.Item
          name="timezone"
          label="Time zone"
          rules={[{ required: true, type: 'string', message: Messages.InvalidTimezone }]}
        >
          <Select
            placeholder="Select a time zone"
            allowClear={false}
            showSearch={true}
            filterOption={(searchText: any, option: any) => {
              return (
                searchText !== '' && option.name.toLowerCase().includes(searchText.toLowerCase())
              )
            }}
            options={Timezones}
            fieldNames={{
              label: 'name',
              value: 'name'
            }}
          />
        </Form.Item>

        <Form.Item
          name="locale"
          label="Numbers &amp; dates format"
          rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
        >
          <Select
            placeholder="Select a locale"
            allowClear={false}
            showSearch={true}
            filterOption={(searchText: any, option: any) => {
              return (
                searchText !== '' && option.key.toLowerCase().includes(searchText.toLowerCase())
              )
            }}
            options={[
              { value: 'en-US', label: 'en-US' },
              { value: 'fr-FR', label: 'fr-FR' }
            ]}
          />
        </Form.Item>
      </Form>
    </Drawer>
  )
}

export default Layout
