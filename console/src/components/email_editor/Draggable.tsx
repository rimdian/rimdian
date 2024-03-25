import React, { Component } from 'react'
import PropTypes from 'prop-types'
import { constants } from './smooth-dnd'
const { wrapperClass } = constants

export interface DraggableProps {
  render?: () => React.ReactElement
  className?: string
  style?: object
  children?: React.ReactNode
}

class Draggable extends Component<DraggableProps> {
  public static propsTypes = {
    render: PropTypes.func,
    className: PropTypes.string,
    style: PropTypes.object
  }

  render() {
    if (this.props.render) {
      return React.cloneElement(this.props.render(), { className: wrapperClass })
    }

    const clsName = `${this.props.className ? this.props.className + ' ' : ''}`
    return (
      <div {...this.props} className={`${clsName}${wrapperClass}`}>
        {this.props.children}
      </div>
    )
  }
}

export default Draggable
