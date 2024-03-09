import { css } from '@emotion/css'
import logo from 'images/rimdian.png'
import CSS from 'utils/css'

export interface LayoutLandingProps {
  withLogo?: boolean
  children: JSX.Element
}

const landingCss = {
  // parent
  self: css({
    display: 'flex',
    height: '100vh',

    '& h1': css(
      {
        textAlign: 'center',
        padding: 0
      },
      CSS.margin_b_xl
    )
  }),

  left: css({
    display: 'flex',
    width: '50%',
    justifyContent: 'center',
    alignItems: 'center'
  }),

  right: css({
    display: 'flex',
    width: '50%',
    background:
      'url(https://images.pexels.com/photos/7925841/pexels-photo-7925841.jpeg?auto=compress&cs=tinysrgb&dpr=3&h=750&w=1260) no-repeat center center ',
    backgroundSize: 'cover'
  }),

  content: css({
    width: '350px',
    marginTop: '-150px'
  }),

  logo: css({
    textAlign: 'center',
    marginBottom: '50px',

    '& img': css({
      height: '100px'
    })
  })
}

const LayoutLanding = (props: LayoutLandingProps) => {
  return (
    <div className={landingCss.self}>
      <div className={landingCss.left}>
        <div className={landingCss.content}>
          {props.withLogo && (
            <div className={landingCss.logo}>
              <img src={logo} alt="" />
            </div>
          )}

          {props.children}
        </div>
      </div>
      <div className={landingCss.right}></div>
    </div>
  )
}

export default LayoutLanding
