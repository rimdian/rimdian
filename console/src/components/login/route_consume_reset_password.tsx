import { useState } from 'react'
import { Form, Input, Button, Alert, message } from 'antd'
import Axios from 'axios'
import { HandleAxiosError } from 'utils/errors'
import Messages from 'utils/formMessages'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { PasswordInput } from 'components/common/input_password_strength'
import { useAccount } from './context_account'
import LayoutLanding from 'components/app/layout_landing'
import CSS from 'utils/css'

const RouteConsumeResetPassword = () => {
  const accountCtx = useAccount()
  const [isLoading, setIsLoading] = useState(false)
  const [searchParams] = useSearchParams()
  const [form] = Form.useForm()
  const navigate = useNavigate()

  const token: string = searchParams.get('token') || ''

  const onSubmit = (values: any) => {
    setIsLoading(true)
    Axios.post(window.Config.API_ENDPOINT + '/account.consumeResetPassword', {
      token: token,
      new_password: values.password
    })
      .then((res) => {
        message.success('Your password has been changed.')
        form.resetFields()
        setIsLoading(false)
        accountCtx.login(res.data)
        navigate('/')
      })
      .catch((e) => {
        setIsLoading(false)
        HandleAxiosError(e)
      })
  }

  return (
    <LayoutLanding withLogo={true}>
      <>
        <h1>Set your new password</h1>

        {!token && (
          <Alert
            className={CSS.margin_b_l}
            type="error"
            message="A token is required to reset your password."
            showIcon
          />
        )}

        {token && (
          <>
            <Form form={form} layout="vertical" onFinish={onSubmit} requiredMark={false}>
              <Form.Item
                name="password"
                label="New password"
                rules={[
                  { required: true, type: 'string', min: 8, message: Messages.NewPasswordInvalid }
                ]}
                hasFeedback
              >
                <PasswordInput size="large" />
              </Form.Item>

              <Form.Item
                name="confirm"
                label="Confirm new password"
                dependencies={['password']}
                hasFeedback
                rules={[
                  { required: true, message: Messages.ConfirmPasswordRequired },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('password') === value) {
                        return Promise.resolve()
                      }
                      return Promise.reject(new Error(Messages.PasswordsDontMatch))
                    }
                  })
                ]}
              >
                <Input.Password size="large" />
              </Form.Item>

              <Form.Item>
                <Button loading={isLoading} type="primary" size="large" block htmlType="submit">
                  Confirm
                </Button>
              </Form.Item>
            </Form>
          </>
        )}
      </>
    </LayoutLanding>
  )
}

export default RouteConsumeResetPassword
