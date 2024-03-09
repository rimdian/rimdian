import { useState } from 'react'
import { Form, Input, Button } from 'antd'
import Axios from 'axios'
import { HandleAxiosError } from 'utils/errors'
import Messages from 'utils/formMessages'
import LayoutLanding from 'components/app/layout_landing'

const RouteResetPassword = () => {
  const [isLoading, setIsLoading] = useState(false)
  const [isSent, setIsSent] = useState(false)

  const onSubmit = (values: any) => {
    setIsLoading(true)
    Axios.post(window.Config.API_ENDPOINT + '/account.resetPassword', {
      email: values.email
    })
      .then((_res) => {
        setIsLoading(false)
        setIsSent(true)
      })
      .catch((e) => {
        setIsLoading(false)
        HandleAxiosError(e)
      })
  }

  return (
    <LayoutLanding withLogo={true}>
      <>
        <h1>Reset your password</h1>

        {isSent && (
          <div>
            <p>You will soon receive an email with instructions on how to reset your password.</p>
            <p>
              If you don't receive this email it might be because there's no account associated with
              the provided email address, or because our email ended up in your spam folder.{' '}
            </p>
          </div>
        )}

        {!isSent && (
          <>
            <p>Provide your Rimdian email address and we'll send you a password reset link.</p>
            <Form layout="vertical" onFinish={onSubmit} requiredMark={false}>
              <Form.Item
                label=""
                name="email"
                hasFeedback={true}
                rules={[{ required: true, type: 'email', message: Messages.EmailRequired }]}
              >
                <Input size="large" placeholder="Email address" />
              </Form.Item>

              <Form.Item>
                <Button loading={isLoading} type="primary" size="large" block htmlType="submit">
                  Send instructions
                </Button>
              </Form.Item>
            </Form>
          </>
        )}
      </>
    </LayoutLanding>
  )
}

export default RouteResetPassword
