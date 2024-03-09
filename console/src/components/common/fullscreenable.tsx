// create a React component that wraps the given component
// and provides the fullscreen functionality with a button in the top right corner

import * as React from 'react'
import { Button, Tooltip } from 'antd'
import { FullscreenOutlined, FullscreenExitOutlined } from '@ant-design/icons'
import { backgroundColorBase } from 'utils/css'

interface FullscreenableProps {
  children: React.ReactNode
}

interface FullscreenableState {
  fullscreen: boolean
}

export class Fullscreenable extends React.Component<FullscreenableProps, FullscreenableState> {
  constructor(props: FullscreenableProps) {
    super(props)
    this.state = {
      fullscreen: false
    }
  }

  private toggleFullscreen = () => {
    this.setState({
      fullscreen: !this.state.fullscreen
    })
  }

  render() {
    const { children } = this.props
    const { fullscreen } = this.state

    return (
      <div
        style={{
          backgroundColor: fullscreen ? backgroundColorBase : 'inherit',
          padding: fullscreen ? 20 : 0,
          position: fullscreen ? 'fixed' : 'relative',
          top: 0,
          left: 0,
          width: fullscreen ? '100%' : 'auto',
          height: fullscreen ? '100%' : 'auto',
          zIndex: 1000
        }}
      >
        <div style={{ position: 'absolute', top: 10, right: 10, zIndex: 100 }}>
          <Tooltip title={fullscreen ? 'Exit Fullscreen' : 'Fullscreen'} placement="left">
            <Button
              // type="primary"
              shape="circle"
              icon={fullscreen ? <FullscreenExitOutlined /> : <FullscreenOutlined />}
              onClick={this.toggleFullscreen}
            />
          </Tooltip>
        </div>
        <div className="fullscreenable__content">{children}</div>
      </div>
    )
  }
}
