import { useState } from 'react'
import { Button, Drawer } from 'antd'
import { AppManifest, App } from 'interfaces'
import BlockAboutApp from './block_about'
import { css } from '@emotion/css'
import CSS, { backgroundColorBase } from 'utils/css'

type AboutAppButtonProps = {
  manifest: AppManifest
  installedApp?: App
}

const AboutAppButton = (props: AboutAppButtonProps) => {
  const [drawerVisible, setDrawerVisible] = useState(false)

  const closeDrawer = () => {
    setDrawerVisible(false)
  }

  // console.log('initialValues', initialValues);

  return (
    <>
      <Button type="primary" ghost size="small" onClick={() => setDrawerVisible(true)}>
        About
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
              About {props.manifest.name}
            </>
          }
          width={960}
          open={true}
          onClose={closeDrawer}
          headerStyle={{ backgroundColor: backgroundColorBase }}
          bodyStyle={{ backgroundColor: backgroundColorBase }}
        >
          <BlockAboutApp manifest={props.manifest} installedApp={props.installedApp} />
        </Drawer>
      )}
    </>
  )
}

export default AboutAppButton
