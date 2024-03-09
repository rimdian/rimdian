import { Avatar, Image, Button, Drawer, Row, Col, Tooltip, Tabs, Tag, Spin } from 'antd'
import { User, UserAlias, Device, UserSegment, Account } from 'interfaces'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useAccount } from 'components/login/context_account'
import { faCheck, faMapMarkerAlt, faSyncAlt, faTimes } from '@fortawesome/free-solid-svg-icons'
import PartialCustomColumns from 'components/common/partial_custom_column'
import { faUser } from '@fortawesome/free-regular-svg-icons'
import { PartialDeviceTypeIcon } from 'components/common/partial_device_icon'
import { RenderKPI } from 'components/common/partial_kpi'
import dayjs from 'dayjs'
import CSS, { backgroundColorBase } from 'utils/css'
import { css } from '@emotion/css'
import Block from 'components/common/block'
import Attribute from 'components/common/partial_attribute'
import { BlockUserTimeline } from 'components/item_timeline/block_user_timeline'
// import FormatCurrency from 'utils/format_currency'
// import FormatDuration from 'utils/format_duration'

// show a Drawer if the showUser URL parameter is set
type DrawerShowUserProps = {
  userExternalId: string
  workspaceCtx: CurrentWorkspaceCtxValue
}

type UserShowResult = {
  user: User
  user_segments: UserSegment[]
  devices: Device[]
  aliases: UserAlias[]
}

const DrawerShowUser = (props: DrawerShowUserProps) => {
  const [searchParams, setSearchParams] = useSearchParams()
  const accountCtx = useAccount()

  // fetch user profile
  const { isLoading, data, refetch, isFetching } = useQuery<UserShowResult>(
    ['user', props.userExternalId],
    (): Promise<UserShowResult> => {
      return new Promise((resolve, reject) => {
        props.workspaceCtx
          .apiGET(
            '/user.show?workspace_id=' +
              props.workspaceCtx.workspace.id +
              '&external_id=' +
              props.userExternalId
          )
          .then((data: any) => {
            resolve(data as UserShowResult)
          })
          .catch((e) => {
            reject(e)
          })
      })
    }
  )

  const refresh = () => {
    refetch()
  }

  const user = data?.user as User
  const devices = data?.devices as Device[]
  const aliases = data?.aliases as UserAlias[]
  const account = accountCtx.account?.account as Account
  const currency = props.workspaceCtx.workspace.currency

  return (
    <Drawer
      title={false}
      placement="right"
      closeIcon={<FontAwesomeIcon icon={faTimes} />}
      open={true}
      onClose={() => {
        // remove the showUser URL parameter
        searchParams.delete('showUser')
        setSearchParams(searchParams)
      }}
      width="90%"
      headerStyle={{ border: 'none', background: 'none', paddingTop: 0, paddingBottom: 0 }}
      bodyStyle={{ backgroundColor: backgroundColorBase }}
    >
      <Row gutter={24}>
        <Col span={6}>
          {isLoading === true && (
            <div className={css([CSS.text_center, CSS.margin_a_l])}>
              <Spin size="large" />
            </div>
          )}

          {!isLoading && data && (
            <div className={CSS.padding_b_xl}>
              <div className={CSS.text_center}>
                <div
                  className={CSS.padding_b_m}
                  style={{ fontSize: '16px', fontWeight: 500, textTransform: 'capitalize' }}
                >
                  {(user.first_name || '') + ' ' + (user.last_name || '')}
                </div>
                <div className={CSS.padding_b_m}>
                  {user.photo_url && (
                    <Avatar
                      size={100}
                      src={<Image src={user.photo_url} style={{ width: 100 }} />}
                    />
                  )}
                  {!user.photo_url && (
                    <Avatar
                      size={100}
                      icon={<FontAwesomeIcon icon={faUser} />}
                      style={{ backgroundColor: 'rgba(78, 108, 255, 0.2)' }}
                    />
                  )}
                </div>
                {user.timezone !== 'UTC' && (
                  <div className={css([CSS.padding_b_xl, CSS.font_size_xs])}>
                    <FontAwesomeIcon icon={faMapMarkerAlt} />
                    &nbsp;
                    {dayjs().tz(user.timezone).format('LT') +
                      ' in ' +
                      (user.city
                        ? user.state
                          ? user.city + ', ' + user.state
                          : user.city
                        : user.timezone) +
                      (user.country ? ', ' + user.country : '')}
                  </div>
                )}
              </div>

              <Tabs
                defaultActiveKey="profile"
                size="small"
                className={CSS.margin_t_s}
                items={[
                  {
                    key: 'profile',
                    label: 'Profile',
                    children: (
                      <>
                        <Block title="User details" small classNames={[CSS.margin_t_m]}>
                          <>
                            <Attribute classNames={[CSS.padding_t_m]} label="External ID">
                              <Tooltip title={user.id}>{user.external_id}</Tooltip>
                            </Attribute>
                            {user.last_interaction_at && (
                              <Attribute label="Last interaction">
                                <Tooltip
                                  title={
                                    dayjs(user.last_interaction_at)
                                      .tz(account.timezone)
                                      .format('lll') +
                                    ' (' +
                                    account.timezone +
                                    ')'
                                  }
                                >
                                  {dayjs(user.last_interaction_at).fromNow()}
                                </Tooltip>
                              </Attribute>
                            )}
                            {user.signed_up_at && (
                              <Attribute label="Signed up">
                                <Tooltip
                                  title={
                                    dayjs(user.signed_up_at).tz(account.timezone).format('lll') +
                                    ' (' +
                                    account.timezone +
                                    ')'
                                  }
                                >
                                  {dayjs(user.signed_up_at).fromNow()}
                                </Tooltip>
                              </Attribute>
                            )}
                            {user.first_name && (
                              <Attribute label="First name">{user.first_name}</Attribute>
                            )}
                            {user.last_name && (
                              <Attribute label="Last name">{user.last_name}</Attribute>
                            )}
                            {user.gender && <Attribute label="Gender">{user.gender}</Attribute>}
                            {user.birthday && (
                              <Attribute label="Birthday">
                                {dayjs(user.birthday, 'YYYY-MM-DD').format('LL')}
                              </Attribute>
                            )}
                            {user.email && (
                              <Attribute label={<FontAwesomeIcon icon="envelope" />}>
                                <a href={'mailto:' + user.email}>{user.email}</a>
                              </Attribute>
                            )}
                            {user.telephone && (
                              <Attribute label={<FontAwesomeIcon icon="phone" />}>
                                <a href={'tel:' + user.telephone}>{user.telephone}</a>
                              </Attribute>
                            )}
                            <Attribute label="Consent all">
                              {user.consent_all ? (
                                <span className={CSS.text_green}>
                                  <FontAwesomeIcon icon={faCheck} />
                                </span>
                              ) : (
                                <span className={CSS.text_orange}>
                                  <FontAwesomeIcon icon={faTimes} />
                                </span>
                              )}
                            </Attribute>
                            {!user.consent_all && (
                              <>
                                <Attribute label="Consent personalization">
                                  {user.consent_personalization ? (
                                    <span className={CSS.text_green}>
                                      <FontAwesomeIcon icon={faCheck} />
                                    </span>
                                  ) : (
                                    <span className={CSS.text_orange}>
                                      <FontAwesomeIcon icon={faTimes} />
                                    </span>
                                  )}
                                </Attribute>
                                <Attribute label="Consent statistic">
                                  {user.consent_statistic ? (
                                    <span className={CSS.text_green}>
                                      <FontAwesomeIcon icon={faCheck} />
                                    </span>
                                  ) : (
                                    <span className={CSS.text_orange}>
                                      <FontAwesomeIcon icon={faTimes} />
                                    </span>
                                  )}
                                </Attribute>
                                <Attribute label="Consent marketing">
                                  {user.consent_marketing ? (
                                    <span className={CSS.text_green}>
                                      <FontAwesomeIcon icon={faCheck} />
                                    </span>
                                  ) : (
                                    <span className={CSS.text_orange}>
                                      <FontAwesomeIcon icon={faTimes} />
                                    </span>
                                  )}
                                </Attribute>
                              </>
                            )}

                            {user.timezone && (
                              <Attribute label="Time zone">{user.timezone}</Attribute>
                            )}
                            {user.language && (
                              <Attribute label="Language">{user.language}</Attribute>
                            )}
                            {user.country && <Attribute label="Country">{user.country}</Attribute>}
                            {user.address_line_1 && (
                              <Attribute label="Address">{user.address_line_1}</Attribute>
                            )}
                            {user.address_line_2 && (
                              <Attribute label="Address line 2">{user.address_line_2}</Attribute>
                            )}
                            {user.city && <Attribute label="City">{user.city}</Attribute>}
                            {user.postal_code && (
                              <Attribute label="Postal code">{user.postal_code}</Attribute>
                            )}
                            {user.state && <Attribute label="State">{user.state}</Attribute>}
                            {user.region && <Attribute label="Region">{user.region}</Attribute>}
                            {user.last_ip && <Attribute label="Last IP">{user.last_ip}</Attribute>}
                            {user.latitude && (
                              <Attribute label="Latitude">{user.latitude}</Attribute>
                            )}
                            {user.longitude && (
                              <Attribute label="Longitude">{user.longitude}</Attribute>
                            )}
                            {/*user.latitude &&
                      user.longitude &&  <div>
                         TODO map <GoogleMap
                          googleMapURL={
                            'https://maps.googleapis.com/maps/api/js?key=' + Config().gmap_key
                          }
                          loadingElement={
                            <div className="loadingElement" style={{ height: '150px' }} />
                          }
                          containerElement={
                            <div className="containerElement" style={{ height: '150px' }} />
                          }
                          mapElement={<div className="mapElement" style={{ height: '100%' }} />}
                          latitude={user.latitude}
                          longitude={user.longitude}
                        /> 
                      </div>*/}

                            {Object.keys(user)
                              .filter(
                                (key: string) => key.startsWith('app_') || key.startsWith('appx_')
                              )
                              .map((key: string) => {
                                return (
                                  <Attribute key={key} label={key}>
                                    {PartialCustomColumns(user[key], user.timezone)}
                                  </Attribute>
                                )
                              })}
                          </>
                        </Block>

                        {/* {user.cart && user.cart.items && user.cart.items.length > 0 && (
                    <div className="block margin-t-m">
                      <h2 className="mini">{t('shopping_cart', 'Shopping cart')}</h2>
                      {user.cart.publicURL && (
                        <Attribute label="XXX">
                          Public URL</span>
                          
                            <a href={user.cart.publicURL} target="_blank" rel="noopener noreferrer">
                              {user.cart.publicURL}
                            </a>
                          </span>
                        </div>
                      )}

                      {user.cart.items.map((item: any, i: number) => (
                        <div key={i} className="subsider-attribute">
                          
                            <Tooltip title={item.external_id}>
                              {item.quantity}x {item.name + ' ' + item.brand}
                            </Tooltip>
                          </span>
                          
                            {Formatters.currency(
                              this.props.projectLayout.project.currency,
                              item.price,
                              item.priceSource,
                              item.currency
                            )}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}

                  {user.wishList && user.wishList.items && user.wishList.items.length > 0 && (
                    <div className="block margin-t-m">
                      <h2 className="mini">{t('wish_list', 'Wish list')}</h2>

                      {user.wishList.items.map((item: any) => (
                        <div key={item.id} className="subsider-attribute">
                          
                            <Tooltip title={item.external_id}>
                              {item.name + ' ' + item.brand}
                            </Tooltip>
                          </span>
                          
                            {Formatters.currency(
                              this.props.projectLayout.project.currency,
                              item.price,
                              item.priceSource,
                              item.currency
                            )}
                          </span>
                        </div>
                      ))}
                    </div>
                  )} */}

                        {aliases.length > 0 && (
                          <Block title="Merged user IDs" classNames={[CSS.margin_t_m]}>
                            <Attribute label="External ID">
                              <>
                                {aliases.map((alias: UserAlias) => (
                                  <div key={alias.from_user_external_id}>
                                    {alias.from_user_external_id}
                                  </div>
                                ))}
                              </>
                            </Attribute>
                          </Block>
                        )}
                      </>
                    )
                  },
                  {
                    key: 'devices',
                    label: 'Devices',
                    children: (
                      <>
                        {devices.map((device: Device) => {
                          return (
                            <Block
                              title={
                                <>
                                  {PartialDeviceTypeIcon(device.device_type)}&nbsp;&nbsp;
                                  {`${device.browser} ${device.browser_version_major}`}
                                </>
                              }
                              key={device.id}
                              classNames={[CSS.margin_t_m]}
                            >
                              {device.os && <Attribute label="OS">{device.os}</Attribute>}
                              {device.resolution && (
                                <Attribute label="Resolution">{device.resolution}</Attribute>
                              )}
                              <Attribute label="Ad block">
                                {device.ad_blocker ? 'yes' : 'no'}
                              </Attribute>
                              <Attribute label="Mobile Webview">
                                {device.in_webview ? 'yes' : 'no'}
                              </Attribute>
                              {device.language && (
                                <Attribute label="Language">{device.language}</Attribute>
                              )}
                            </Block>
                          )
                        })}
                      </>
                    )
                  }
                ]}
              />
            </div>
          )}
        </Col>

        <Col span={18}>
          <div className={CSS.padding_b_m} style={{ minHeight: '36px' }}>
            <span className={CSS.pull_right}>
              <Tooltip title="Refresh" placement="left">
                <Button
                  size="small"
                  type="text"
                  loading={isFetching}
                  icon={<FontAwesomeIcon icon={faSyncAlt} spin={isLoading} />}
                  onClick={refresh}
                />
              </Tooltip>
            </span>
            {!isLoading &&
              data?.user_segments &&
              data?.user_segments
                .filter((us: UserSegment) => !us.exit_at)
                .map((us: UserSegment) => {
                  const segment = props.workspaceCtx.segmentsMap[us.segment_id]
                  if (!segment) {
                    return null
                  }
                  return (
                    <Tooltip
                      title={dayjs(us.enter_at).fromNow()}
                      placement="bottom"
                      key={us.segment_id}
                    >
                      <Tag color={props.workspaceCtx.segmentsMap[us.segment_id].color}>
                        {props.workspaceCtx.segmentsMap[us.segment_id].name}
                      </Tag>
                    </Tooltip>
                  )
                })}
          </div>
          {props.workspaceCtx.workspace.has_orders && (
            <>
              <Block grid classNames={[CSS.margin_b_m]}>
                <RenderKPI
                  valueIsLoading={isLoading}
                  valueType="number"
                  title="Orders"
                  value={user && user.orders_count}
                />
                <RenderKPI
                  valueIsLoading={isLoading}
                  title="Orders LTV"
                  valueType="currency"
                  currency={currency}
                  tooltip="Sum of all orders subtotal"
                  // value={user && user.orders_ltv}
                  value={user && user.orders_ltv}
                />
                <RenderKPI
                  valueIsLoading={isLoading}
                  title="Last order"
                  valueType="relativeDate"
                  value={user && user.last_order_at ? dayjs(user.last_order_at).unix() : 0}
                />
                <RenderKPI
                  valueIsLoading={isLoading}
                  title="Avg. cart"
                  valueType="currency"
                  currency={currency}
                  tooltip="Average order subtotal accross all orders"
                  value={user && user.orders_avg_cart}
                />
              </Block>

              <Block grid classNames={[CSS.margin_b_m]}>
                <RenderKPI
                  valueIsLoading={isLoading}
                  title="#1 order"
                  valueType="currency"
                  currency={currency}
                  tooltip="First order subtotal"
                  value={user && user.first_order_subtotal}
                />
                <RenderKPI
                  valueIsLoading={isLoading}
                  title="#1 order TTC"
                  valueType="duration"
                  tooltip="First order time to conversion"
                  value={user && user.first_order_ttc}
                />
                <RenderKPI
                  valueIsLoading={isLoading}
                  title="Avg. repeat order"
                  valueType="currency"
                  currency={currency}
                  tooltip="Average repeat order subtotal, excluding first order"
                  value={user && user.avg_repeat_cart}
                />
                <RenderKPI
                  valueIsLoading={isLoading}
                  title="Avg. repeat order TTC"
                  valueType="duration"
                  tooltip="Average time to conversion for repeat orders, excluding first order"
                  value={user && user.avg_repeat_order_ttc}
                />
              </Block>
            </>
          )}

          <Tabs
            defaultActiveKey="1"
            size="small"
            items={[
              {
                key: '1',
                label: 'Timeline',
                children: (
                  <div className={CSS.margin_t_m}>
                    {user && (
                      <BlockUserTimeline
                        timezone={user.timezone}
                        workspaceCtx={props.workspaceCtx}
                        user={user}
                        devices={devices}
                      />
                    )}
                    {!user && <Spin />}
                  </div>
                )
              },
              {
                key: '2',
                label: 'Orders',
                children: (
                  <>
                    {/* <Orders */}
                    Work in progress...
                  </>
                )
              }
            ]}
          />
        </Col>
      </Row>
    </Drawer>
  )
}

export default DrawerShowUser
