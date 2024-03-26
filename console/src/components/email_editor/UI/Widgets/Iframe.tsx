import { Component } from 'react'
import _ from 'lodash'
import { css } from '@emotion/css'

const cssIframe = css({
  position: 'relative',
  '&-normal': {
    height: '200px',
    width: '100%'
  },
  '&-actions': {
    position: 'absolute',
    top: '12px',
    right: '12px',
    zIndex: '1'
  }
})

class IframeSandbox extends Component<any, any> {
  ref: string

  constructor(props: any) {
    super(props)

    this.ref = _.uniqueId()
    this._resize = this._resize.bind(this)
  }

  _resize() {
    const el = document.querySelector(this.props.sizeSelector)
    const parentHeight = el ? parseInt(window.getComputedStyle(el).height) : 0

    const container: any = this.refs[this.ref + '-container']
    const thisRef: any = this.refs[this.ref]

    if (container) {
      container.style.height = parentHeight - 30 + 'px'
    }

    if (thisRef) {
      thisRef.style.height = parentHeight - 30 + 'px'
    }
  }

  componentDidMount() {
    // wait a bit to be sure parent element are well inserted in the dom
    // and height can be computed correctly
    window.setTimeout(() => {
      this._resize()
    }, 100)
  }

  shouldComponentUpdate(nextProps: any, nextState: any) {
    if (nextProps.id !== this.props.id) return true
    if (nextProps.ref !== this.ref) return true
    if (nextProps.className !== this.props.className) return true
    if (nextProps.content !== this.props.content) return true
    if (nextProps.sizeSelector !== this.props.sizeSelector) return true
    return false
  }

  render() {
    // console.log('render', this.props.className);
    return (
      <div className={cssIframe} ref={this.ref + '-container'}>
        <iframe
          style={this.props.style}
          title={'iframe-' + this.props.id}
          srcDoc={this.props.content}
          frameBorder="0"
          id={this.props.id}
          ref={this.ref}
          className={this.props.className}
        ></iframe>
      </div>
    )
  }
}

export default IframeSandbox
