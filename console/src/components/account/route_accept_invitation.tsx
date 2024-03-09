import { useState } from 'react'
import { useAccount } from 'components/login/context_account'
import { Form, Input, Button, Spin, Alert, Tabs, message } from 'antd'
import Axios from 'axios'
import { HandleAxiosError } from 'utils/errors'
import { AccountLoginResult } from 'interfaces'
import LoginForm from 'components/login/form_login'
import { useQuery } from '@tanstack/react-query'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useForm } from 'antd/lib/form/Form'
import Messages from 'utils/formMessages'
import LayoutLanding from 'components/app/layout_landing'
import CSS from 'utils/css'

interface InvitationRead {
  email: string
  organization_id: string
  organization_name: string
}

const RouteAcceptInvitation = () => {
  const accountCtx = useAccount()
  const [isConsuming, setIsConsuming] = useState(false)
  const [searchParams] = useSearchParams()
  const [form] = useForm()
  const navigate = useNavigate()

  // 'aW52aXRlZDNAY2FwdGFpbm1ldHJpY3MuY29tfmFjbWU=.dda6e515c1d533dc606537da5ab68137d67ef7020ddefb804747ce450c062c93'
  const token: string = searchParams.get('token') || ''

  const { isLoading, data, isFetching } = useQuery<InvitationRead>(
    ['organizationInvitation', token],
    (): Promise<InvitationRead> => {
      return new Promise((resolve, reject) => {
        Axios.post(window.Config.API_ENDPOINT + '/organizationInvitation.read', {
          token: token
        })
          .then((res) => {
            // console.log('res', res)
            form.setFieldsValue({ email: res.data.email })
            resolve(res.data as InvitationRead)
          })
          .catch((e) => {
            HandleAxiosError(e)
            reject(e)
          })
      })
    },
    { enabled: !!token }
  )

  const onSignup = (values: any) => {
    setIsConsuming(true)
    Axios.post(window.Config.API_ENDPOINT + '/organizationInvitation.consume', {
      token: token,
      name: values.name,
      password: values.password
    })
      .then((res) => {
        message.success('Your account has been created, welcome!')
        setIsConsuming(false)
        accountCtx.login(res.data as AccountLoginResult)
        navigate('/orgs/' + data?.organization_id)
      })
      .catch((e) => {
        setIsConsuming(false)
        HandleAxiosError(e)
      })
  }

  const onAccept = () => {
    setIsConsuming(true)
    Axios.post(
      window.Config.API_ENDPOINT + '/organizationInvitation.consume',
      { token: token },
      { headers: { Authorization: 'Bearer ' + accountCtx.account?.access_token } }
    )
      .then((res) => {
        message.success('Welcome to ' + data?.organization_name + '!')
        setIsConsuming(false)
        navigate('/orgs/' + data?.organization_id)
      })
      .catch((e) => {
        setIsConsuming(false)
        HandleAxiosError(e)
      })
  }

  return (
    <LayoutLanding withLogo={true}>
      <>
        {token === '' && (
          <Alert
            type="error"
            message="Oops!"
            description="An invitation token is required."
            showIcon
          />
        )}

        {token !== '' && (
          <Spin
            size="large"
            tip="Initializing..."
            spinning={accountCtx.initializing || isLoading || isFetching}
          >
            {data && (
              <div style={{ fontSize: 18, textAlign: 'center' }} className={CSS.margin_b_l}>
                Welcome, you have been invited to join <b>{data.organization_name}</b> on Rimdian!
              </div>
            )}

            {accountCtx.account && (
              <>
                <Button block type="primary" onClick={onAccept} loading={isConsuming} size="large">
                  Accept invitation
                </Button>
              </>
            )}

            {!accountCtx.account && (
              <Tabs
                defaultActiveKey={accountCtx.account ? 'login' : 'createAccount'}
                items={[
                  {
                    key: 'createAccount',
                    label: <span>Create Account</span>,
                    disabled: accountCtx.account ? true : false,
                    children: (
                      <>
                        <Form
                          form={form}
                          className={CSS.margin_t_l}
                          layout="vertical"
                          onFinish={onSignup}
                          requiredMark={false}
                        >
                          <Form.Item
                            label="Email"
                            name="email"
                            hasFeedback={true}
                            rules={[
                              { required: true, type: 'email', message: Messages.EmailRequired }
                            ]}
                          >
                            <Input size="large" value={data?.email} disabled />
                          </Form.Item>
                          <Form.Item
                            label="Your name"
                            name="name"
                            hasFeedback={true}
                            rules={[
                              {
                                required: true,
                                type: 'string',
                                message: Messages.YourNameIsRequired
                              }
                            ]}
                          >
                            <Input size="large" />
                          </Form.Item>

                          <Form.Item
                            label="Password"
                            name="password"
                            hasFeedback={true}
                            rules={[{ required: true, message: Messages.PasswordRequired }]}
                          >
                            <Input.Password size="large" />
                          </Form.Item>

                          {/* <Form.Item name="remember" valuePropName="checked" wrapperCol={{ offset: 8, span: 16 }}>
                                    <Checkbox>Remember me</Checkbox>
                                </Form.Item> */}

                          <Form.Item>
                            <Button
                              loading={isConsuming}
                              type="primary"
                              size="large"
                              block
                              htmlType="submit"
                            >
                              Create account
                            </Button>
                          </Form.Item>
                        </Form>
                      </>
                    )
                  },
                  {
                    key: 'login',
                    label: <span>Sign in</span>,
                    children: (
                      <div className={CSS.margin_t_l}>
                        <LoginForm email={data?.email} />
                      </div>
                    )
                  }
                ]}
              />
            )}
          </Spin>
        )}
      </>
    </LayoutLanding>
  )
}

export default RouteAcceptInvitation
