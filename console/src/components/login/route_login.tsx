import LayoutLanding from 'components/app/layout_landing'
import LoginForm from './form_login'

const RouteLogin = () => {
  return (
    <LayoutLanding withLogo={true}>
      <LoginForm />
    </LayoutLanding>
  )
}

export default RouteLogin
