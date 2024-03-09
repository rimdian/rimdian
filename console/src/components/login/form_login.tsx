import { useState } from 'react'
import { useAccount } from './context_account'
import { Link } from 'react-router-dom'
import { Form, Input, Button, Spin } from 'antd'
import Axios from 'axios'
import { HandleAxiosError } from 'utils/errors'
import { AccountLoginResult } from 'interfaces'
import Messages from 'utils/formMessages'

type LoginFormProps = {
  email?: string
  onComplete?: () => void
}

const LoginForm = (props: LoginFormProps) => {
  const accountCtx = useAccount()
  const [isLoading, setIsLoading] = useState(false)

  const onSubmit = (values: any) => {
    setIsLoading(true)
    Axios.post(window.Config.API_ENDPOINT + '/account.login', {
      email: values.email,
      password: values.password
    })
      .then((res) => {
        setIsLoading(false)
        // console.log('res', res)
        accountCtx.login(res.data as AccountLoginResult)

        if (props.onComplete) {
          props.onComplete()
        }
      })
      .catch((e) => {
        setIsLoading(false)
        HandleAxiosError(e)
      })
  }

  return (
    <Spin size="large" tip="Initializing..." spinning={accountCtx.initializing}>
      <Form
        layout="vertical"
        initialValues={{ email: props.email }}
        onFinish={onSubmit}
        requiredMark={false}
      >
        <Form.Item
          label="Email"
          name="email"
          hasFeedback={true}
          rules={[{ required: true, type: 'email', message: Messages.EmailRequired }]}
        >
          <Input size="large" disabled={props.email ? true : false} />
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
          <Button loading={isLoading} type="primary" size="large" block htmlType="submit">
            Sign in
          </Button>
        </Form.Item>
      </Form>

      <Link to="/reset-password">Forgot your password?</Link>
    </Spin>
  )
}

export default LoginForm
