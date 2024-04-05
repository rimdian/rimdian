import mjml2html from 'mjml-browser'
import { Alert, Button, Space, Tabs } from 'antd'
import Prism from 'prismjs'
import 'prismjs/themes/prism-okaidia.css' /* or your own custom theme */
import Nunjucks from 'nunjucks'
import { BlockInterface } from '../Block'
import Iframe from './Widgets/Iframe'
import 'prismjs/components/prism-xml-doc'
import { kebabCase } from 'lodash'
import json2mjml from 'utils/json_to_mjml'
import { useState } from 'react'
import CSS from 'utils/css'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faDesktop, faMobileAlt, faPen } from '@fortawesome/free-solid-svg-icons'
import { DesktopWidth, MobileWidth } from './Layout'

const objectAsKebab = (obj: any) => {
  const newObj: any = {}
  // console.log('obj', obj)
  Object.keys(obj).forEach((key: string) => {
    newObj[kebabCase(key)] = obj[key]
  })
  return newObj
}

const trackURL = (url: string, urlParams: any) => {
  // parse href and append utm params
  const newURL = new URL(url)
  if (!newURL.searchParams.has('utm_source') && urlParams.utm_source) {
    newURL.searchParams.append('utm_source', urlParams.utm_source)
  }
  if (!newURL.searchParams.has('utm_medium') && urlParams.utm_medium) {
    newURL.searchParams.append('utm_medium', urlParams.utm_medium)
  }
  if (!newURL.searchParams.has('utm_campaign') && urlParams.utm_campaign) {
    newURL.searchParams.append('utm_campaign', urlParams.utm_campaign)
  }
  if (!newURL.searchParams.has('utm_content') && urlParams.utm_content) {
    newURL.searchParams.append('utm_content', urlParams.utm_content)
  }
  if (!newURL.searchParams.has('utm_id') && urlParams.utm_id) {
    newURL.searchParams.append('utm_id', urlParams.utm_id)
  }
  return newURL.toString()
}

const treeToMjmlJSON = (
  rootStyles: any,
  block: BlockInterface,
  templateData: string,
  urlParams: any,
  parent?: BlockInterface
) => {
  let children: any[] = []

  if (block.children && block.children.length) {
    children = block.children.map((child) => {
      return treeToMjmlJSON(rootStyles, child, templateData, urlParams, block)
    })
  }

  // console.log('block ', block)

  // let attrs: any

  switch (block.kind) {
    case 'root':
      const attrs = objectAsKebab(block.data.styles.body)

      delete attrs['margin']
      // console.log('body', attrs)

      return {
        tagName: 'mjml',
        attributes: {},
        children: [
          {
            tagName: 'mj-body',
            attributes: attrs,
            children: children
          }
        ]
      }

    case 'columns168':
    case 'columns204':
    case 'columns420':
    case 'columns816':
    case 'columns888':
    case 'columns1212':
    case 'columns6666':
    case 'oneColumn':
      const sectionAttrs: any = {
        'text-align': block.data.styles.textAlign
      }

      if (block.data.backgroundType === 'image') {
        sectionAttrs['background-url'] = block.data.styles.backgroundImage
        if (block.data.styles.backgroundSize) {
          sectionAttrs['background-size'] = block.data.styles.backgroundSize
        }
        if (block.data.styles.backgroundRepeat) {
          sectionAttrs['background-repeat'] = block.data.styles.backgroundRepeat
        }
      }

      if (block.data.backgroundType === 'color') {
        if (block.data.styles.backgroundColor)
          sectionAttrs['background-color'] = block.data.styles.backgroundColor
      }

      if (block.data.borderControl === 'all') {
        if (
          block.data.styles.borderStyle !== 'none' &&
          block.data.styles.borderWidth &&
          block.data.styles.borderColor
        ) {
          sectionAttrs['border'] =
            block.data.styles.borderWidth +
            ' ' +
            block.data.styles.borderStyle +
            ' ' +
            block.data.styles.borderColor
        }
      }

      if (block.data.styles.borderRadius && block.data.styles.borderRadius !== '0px') {
        sectionAttrs['border-radius'] = block.data.styles.borderRadius
      }

      if (block.data.borderControl === 'separate') {
        if (
          block.data.styles.borderTopStyle !== 'none' &&
          block.data.styles.borderTopWidth &&
          block.data.styles.borderTopColor
        ) {
          sectionAttrs['border-top'] =
            block.data.styles.borderTopWidth +
            ' ' +
            block.data.styles.borderTopStyle +
            ' ' +
            block.data.styles.borderTopColor
        }

        if (
          block.data.styles.borderRightStyle !== 'none' &&
          block.data.styles.borderRightWidth &&
          block.data.styles.borderRightColor
        ) {
          sectionAttrs['border-right'] =
            block.data.styles.borderRightWidth +
            ' ' +
            block.data.styles.borderRightStyle +
            ' ' +
            block.data.styles.borderRightColor
        }

        if (
          block.data.styles.borderBottomStyle !== 'none' &&
          block.data.styles.borderBottomWidth &&
          block.data.styles.borderBottomColor
        ) {
          sectionAttrs['border-bottom'] =
            block.data.styles.borderBottomWidth +
            ' ' +
            block.data.styles.borderBottomStyle +
            ' ' +
            block.data.styles.borderBottomColor
        }

        if (
          block.data.styles.borderLeftStyle !== 'none' &&
          block.data.styles.borderLeftWidth &&
          block.data.styles.borderLeftColor
        ) {
          sectionAttrs['border-left'] =
            block.data.styles.borderLeftWidth +
            ' ' +
            block.data.styles.borderLeftStyle +
            ' ' +
            block.data.styles.borderLeftColor
        }
      }

      if (block.data.paddingControl === 'all') {
        if (block.data.styles.padding && block.data.styles.padding !== '0px') {
          sectionAttrs['padding'] = block.data.styles.padding
        }
      }

      if (block.data.paddingControl === 'separate') {
        if (block.data.styles.paddingTop && block.data.styles.paddingTop !== '0px') {
          sectionAttrs['padding-top'] = block.data.styles.paddingTop
        }
        if (block.data.styles.paddingRight && block.data.styles.paddingRight !== '0px') {
          sectionAttrs['padding-right'] = block.data.styles.paddingRight
        }
        if (block.data.styles.paddingBottom && block.data.styles.paddingBottom !== '0px') {
          sectionAttrs['padding-bottom'] = block.data.styles.paddingBottom
        }
        if (block.data.styles.paddingLeft && block.data.styles.paddingLeft !== '0px') {
          sectionAttrs['padding-left'] = block.data.styles.paddingLeft
        }
      }

      // console.log('section', objectAsKebab(sectionAttrs))

      // wrap with mj-group if columnsOnMobile == true
      if (block.data.columnsOnMobile === true) {
        children = [
          {
            tagName: 'mj-group',
            attributes: {},
            children: children
          }
        ]
      }

      const sectionBlock = {
        tagName: 'mj-section',
        attributes: objectAsKebab(sectionAttrs),
        children: children
      }

      return sectionBlock

    case 'column':
      const columnAttrs: any = {
        'vertical-align': block.data.styles.verticalAlign
      }

      if (parent) {
        switch (parent.kind) {
          case 'columns168':
            if (parent.children[0].id === block.id) columnAttrs['width'] = '66.33%'
            else columnAttrs['width'] = '33.33%'
            break
          case 'columns204':
            if (parent.children[0].id === block.id) columnAttrs['width'] = '83.33%'
            else columnAttrs['width'] = '16.66%'
            break
          case 'columns420':
            if (parent.children[0].id === block.id) columnAttrs['width'] = '16.33%'
            else columnAttrs['width'] = '83.33%'
            break
          case 'columns816':
            if (parent.children[0].id === block.id) columnAttrs['width'] = '33.33%'
            else columnAttrs['width'] = '66.33%'
            break
          // default columns are equally divided
          case 'default':
        }
      }

      if (block.data.styles.backgroundColor) {
        if (block.data.styles.backgroundColor)
          columnAttrs['background-color'] = block.data.styles.backgroundColor
      }

      if (block.data.borderControl === 'all') {
        if (
          block.data.styles.borderStyle !== 'none' &&
          block.data.styles.borderWidth &&
          block.data.styles.borderColor
        ) {
          columnAttrs['border'] =
            block.data.styles.borderWidth +
            ' ' +
            block.data.styles.borderStyle +
            ' ' +
            block.data.styles.borderColor
        }
      }

      if (block.data.styles.borderRadius && block.data.styles.borderRadius !== '0px') {
        columnAttrs['border-radius'] = block.data.styles.borderRadius
      }

      if (block.data.borderControl === 'separate') {
        if (
          block.data.styles.borderTopStyle !== 'none' &&
          block.data.styles.borderTopWidth &&
          block.data.styles.borderTopColor
        ) {
          columnAttrs['border-top'] =
            block.data.styles.borderTopWidth +
            ' ' +
            block.data.styles.borderTopStyle +
            ' ' +
            block.data.styles.borderTopColor
        }

        if (
          block.data.styles.borderRightStyle !== 'none' &&
          block.data.styles.borderRightWidth &&
          block.data.styles.borderRightColor
        ) {
          columnAttrs['border-right'] =
            block.data.styles.borderRightWidth +
            ' ' +
            block.data.styles.borderRightStyle +
            ' ' +
            block.data.styles.borderRightColor
        }

        if (
          block.data.styles.borderBottomStyle !== 'none' &&
          block.data.styles.borderBottomWidth &&
          block.data.styles.borderBottomColor
        ) {
          columnAttrs['border-bottom'] =
            block.data.styles.borderBottomWidth +
            ' ' +
            block.data.styles.borderBottomStyle +
            ' ' +
            block.data.styles.borderBottomColor
        }

        if (
          block.data.styles.borderLeftStyle !== 'none' &&
          block.data.styles.borderLeftWidth &&
          block.data.styles.borderLeftColor
        ) {
          columnAttrs['border-left'] =
            block.data.styles.borderLeftWidth +
            ' ' +
            block.data.styles.borderLeftStyle +
            ' ' +
            block.data.styles.borderLeftColor
        }
      }

      if (block.data.paddingControl === 'all') {
        if (block.data.styles.padding && block.data.styles.padding !== '0px') {
          columnAttrs['padding'] = block.data.styles.padding
        }
      }

      if (block.data.paddingControl === 'separate') {
        if (block.data.styles.paddingTop && block.data.styles.paddingTop !== '0px') {
          columnAttrs['padding-top'] = block.data.styles.paddingTop
        }
        if (block.data.styles.paddingRight && block.data.styles.paddingRight !== '0px') {
          columnAttrs['padding-right'] = block.data.styles.paddingRight
        }
        if (block.data.styles.paddingBottom && block.data.styles.paddingBottom !== '0px') {
          columnAttrs['padding-bottom'] = block.data.styles.paddingBottom
        }
        if (block.data.styles.paddingLeft && block.data.styles.paddingLeft !== '0px') {
          columnAttrs['padding-left'] = block.data.styles.paddingLeft
        }
      }

      // console.log('column', objectAsKebab(columnAttrs))

      const columnBlock = {
        tagName: 'mj-column',
        attributes: objectAsKebab(columnAttrs),
        children: children
      }

      return columnBlock

    case 'text':
    case 'heading':
      const textAttrs: any = {
        align: block.data.align,
        padding: 0 // dont use default mjml 10px 25px
      }

      if (block.data.backgroundColor) {
        if (block.data.backgroundColor)
          textAttrs['container-background-color'] = block.data.backgroundColor
      }

      let content = ''

      block.data.editorData.forEach((line: any) => {
        // console.log('line', line)
        let lineContent = ''

        line.children.forEach((part: any) => {
          // console.log('part', part)

          // needs span
          if (
            part.bold === true ||
            part.italic === true ||
            part.underlined === true ||
            part.fontSize ||
            part.fontColor ||
            part.fontFamily ||
            part.hyperlink
          ) {
            const spanStyles = []
            if (part.bold) spanStyles.push('font-weight: bold')
            if (part.italic) spanStyles.push('font-style: italic')
            if (part.underlined) spanStyles.push('text-decoration: underline')
            if (part.fontSize) spanStyles.push('font-size: ' + part.fontSize)
            if (part.fontColor) spanStyles.push('color: ' + part.fontColor)
            if (part.fontFamily) spanStyles.push('font-family: ' + part.fontFamily)
            if (part.hyperlink) {
              const hyperlinkStyles = [
                'color: ' + block.data.hyperlinkStyles.color,
                'font-family: ' + block.data.hyperlinkStyles.fontFamily,
                'font-size: ' + block.data.hyperlinkStyles.fontSize,
                'font-style: ' + block.data.hyperlinkStyles.fontStyle,
                'font-weight: ' + block.data.hyperlinkStyles.fontWeight
              ]

              const finalURL = part.hyperlink.disable_tracking
                ? part.hyperlink.url
                : trackURL(part.hyperlink.url, urlParams)
              lineContent +=
                '<a style="' +
                hyperlinkStyles.join('; ') +
                '" href="' +
                finalURL +
                '">' +
                part.text +
                '</a>'
              return
            }

            lineContent +=
              '<span' +
              (spanStyles.length > 0 ? ' style="' + spanStyles.join('; ') + '"' : '') +
              '>' +
              part.text +
              '</span>'
          } else {
            lineContent += part.text
          }
        })

        // wrap
        const lineStyles = []
        if (rootStyles[line.type] && rootStyles[line.type].color)
          lineStyles.push('color: ' + rootStyles[line.type].color)
        if (rootStyles[line.type] && rootStyles[line.type].fontFamily)
          lineStyles.push('font-family: ' + rootStyles[line.type].fontFamily)
        if (rootStyles[line.type] && rootStyles[line.type].fontSize)
          lineStyles.push('font-size: ' + rootStyles[line.type].fontSize)
        if (rootStyles[line.type] && rootStyles[line.type].fontStyle)
          lineStyles.push('font-style: ' + rootStyles[line.type].fontStyle)
        if (rootStyles[line.type] && rootStyles[line.type].fontWeight)
          lineStyles.push('font-weight: ' + rootStyles[line.type].fontWeight)

        // padding
        if (rootStyles[line.type].paddingControl === 'all') {
          lineStyles.push('padding: ' + rootStyles[line.type].padding)
        } else {
          lineStyles.push('padding-top: ' + rootStyles[line.type].paddingTop)
          lineStyles.push('padding-right: ' + rootStyles[line.type].paddingRight)
          lineStyles.push('padding-bottom: ' + rootStyles[line.type].paddingBottom)
          lineStyles.push('padding-left: ' + rootStyles[line.type].paddingLeft)
        }

        // no margin there, reset default css
        lineStyles.push('margin: 0px')

        switch (line.type) {
          case 'h1':
            lineContent =
              '<h1' +
              (lineStyles.length > 0 ? ' style="' + lineStyles.join('; ') + '"' : '') +
              '>' +
              lineContent +
              '</h1>'
            break
          case 'h2':
            lineContent =
              '<h2' +
              (lineStyles.length > 0 ? ' style="' + lineStyles.join('; ') + '"' : '') +
              '>' +
              lineContent +
              '</h2>'
            break
          case 'h3':
            lineContent =
              '<h3' +
              (lineStyles.length > 0 ? ' style="' + lineStyles.join('; ') + '"' : '') +
              '>' +
              lineContent +
              '</h3>'
            break
          case 'paragraph':
            lineContent =
              '<p' +
              (lineStyles.length > 0 ? ' style="' + lineStyles.join('; ') + '"' : '') +
              '>' +
              lineContent +
              '</p>'
            break
          default:
            console.error('line type is not implemented', line)
        }
        // console.log('lineContent', lineContent)
        content += lineContent
      })
      // console.log('text', objectAsKebab(textAttrs))

      // exec nunjucks if has tags
      if (
        templateData &&
        templateData !== '' &&
        (content.includes('{{') || content.includes('{%'))
      ) {
        // console.log('got markup', content)
        // console.log('data', templateData)

        const jsonData = JSON.parse(templateData)

        try {
          const stringResult = Nunjucks.renderString(content, jsonData || {})
          // console.log('stringResult', stringResult)
          content = stringResult
        } catch (e) {
          // ignore error and templating
          console.error(e)
        }
      }

      const textBlock = {
        tagName: 'mj-text',
        attributes: objectAsKebab(textAttrs),
        content: content
      }

      return textBlock

    case 'image':
      // console.log('image', block)
      const imageAttrs: any = {
        align: block.data.wrapper.align,
        src: block.data.image.src,
        alt: block.data.image.alt,
        height: block.data.image.height,
        'fluid-on-mobile': block.data.image.fullWidthOnMobile,
        padding: 0
      }

      if (block.data.image.width !== '100%') {
        imageAttrs['width'] = block.data.image.width
      }

      if (block.data.image.borderRadius) {
        imageAttrs['border-radius'] = block.data.image.borderRadius
      }

      if (block.data.image.href) {
        imageAttrs['href'] = block.data.image.href

        if (!block.data.image.disable_tracking) {
          imageAttrs['href'] = trackURL(imageAttrs['href'], urlParams)
        }
      }

      if (block.data.wrapper.paddingControl === 'all') {
        imageAttrs['padding'] = block.data.wrapper.padding
      } else {
        if (block.data.wrapper.paddingTop) imageAttrs['padding-top'] = block.data.wrapper.paddingTop
        if (block.data.wrapper.paddingRight)
          imageAttrs['padding-right'] = block.data.wrapper.paddingRight
        if (block.data.wrapper.paddingBottom)
          imageAttrs['padding-bottom'] = block.data.wrapper.paddingBottom
        if (block.data.wrapper.paddingLeft)
          imageAttrs['padding-left'] = block.data.wrapper.paddingLeft
      }

      if (block.data.wrapper.borderControl === 'all' && block.data.wrapper.borderStyle !== 'none') {
        imageAttrs['border'] =
          block.data.wrapper.borderWidth +
          ' ' +
          block.data.wrapper.borderStyle +
          ' ' +
          block.data.wrapper.borderColor
      } else {
        if (
          block.data.wrapper.borderTopStyle !== 'none' &&
          block.data.wrapper.borderTopWidth &&
          block.data.wrapper.borderTopColor
        ) {
          imageAttrs['border-top'] =
            block.data.wrapper.borderTopWidth +
            ' ' +
            block.data.wrapper.borderTopStyle +
            ' ' +
            block.data.wrapper.borderTopColor
        }
        if (
          block.data.wrapper.borderRightStyle !== 'none' &&
          block.data.wrapper.borderRightWidth &&
          block.data.wrapper.borderRightColor
        ) {
          imageAttrs['border-right'] =
            block.data.wrapper.borderRightWidth +
            ' ' +
            block.data.wrapper.borderRightStyle +
            ' ' +
            block.data.wrapper.borderRightColor
        }
        if (
          block.data.wrapper.borderBottomStyle !== 'none' &&
          block.data.wrapper.borderBottomWidth &&
          block.data.wrapper.borderBottomColor
        ) {
          imageAttrs['border-bottom'] =
            block.data.wrapper.borderBottomWidth +
            ' ' +
            block.data.wrapper.borderBottomStyle +
            ' ' +
            block.data.wrapper.borderBottomColor
        }
        if (
          block.data.wrapper.borderLeftStyle !== 'none' &&
          block.data.wrapper.borderLeftWidth &&
          block.data.wrapper.borderLeftColor
        ) {
          imageAttrs['border-left'] =
            block.data.wrapper.borderLeftWidth +
            ' ' +
            block.data.wrapper.borderLeftStyle +
            ' ' +
            block.data.wrapper.borderLeftColor
        }
      }

      const imageBlock = {
        tagName: 'mj-image',
        attributes: objectAsKebab(imageAttrs)
      }

      return imageBlock

    case 'button':
      // console.log('button', block)

      const buttonAttrs: any = {
        align: block.data.wrapper.align,
        href: block.data.button.href,
        'background-color': block.data.button.backgroundColor,
        'font-family': block.data.button.fontFamily,
        'font-size': block.data.button.fontSize,
        'font-weight': block.data.button.fontWeight,
        'font-style': block.data.button.fontStyle,
        color: block.data.button.color,
        padding: 0,
        'inner-padding':
          block.data.button.innerVerticalPadding + ' ' + block.data.button.innerHorizontalPadding
      }

      if (!block.data.button.disable_tracking) {
        buttonAttrs['href'] = trackURL(buttonAttrs['href'], urlParams)
      }

      if (block.data.button.width !== 'auto') {
        buttonAttrs['width'] = block.data.button.width
      }

      if (block.data.button.textTransform !== 'none') {
        buttonAttrs['text-transform'] = block.data.button.textTransform
      }

      if (block.data.button.borderRadius) {
        buttonAttrs['border-radius'] = block.data.button.borderRadius
      }

      if (block.data.wrapper.paddingControl === 'all') {
        buttonAttrs['padding'] = block.data.wrapper.padding
      } else {
        if (block.data.wrapper.paddingTop)
          buttonAttrs['padding-top'] = block.data.wrapper.paddingTop
        if (block.data.wrapper.paddingRight)
          buttonAttrs['padding-right'] = block.data.wrapper.paddingRight
        if (block.data.wrapper.paddingBottom)
          buttonAttrs['padding-bottom'] = block.data.wrapper.paddingBottom
        if (block.data.wrapper.paddingLeft)
          buttonAttrs['padding-left'] = block.data.wrapper.paddingLeft
      }

      if (block.data.button.borderControl === 'all' && block.data.button.borderStyle !== 'none') {
        buttonAttrs['border'] =
          block.data.button.borderWidth +
          ' ' +
          block.data.button.borderStyle +
          ' ' +
          block.data.button.borderColor
      } else {
        if (
          block.data.button.borderTopStyle !== 'none' &&
          block.data.button.borderTopWidth &&
          block.data.button.borderTopColor
        ) {
          buttonAttrs['border-top'] =
            block.data.button.borderTopWidth +
            ' ' +
            block.data.button.borderTopStyle +
            ' ' +
            block.data.button.borderTopColor
        }
        if (
          block.data.button.borderRightStyle !== 'none' &&
          block.data.button.borderRightWidth &&
          block.data.button.borderRightColor
        ) {
          buttonAttrs['border-right'] =
            block.data.button.borderRightWidth +
            ' ' +
            block.data.button.borderRightStyle +
            ' ' +
            block.data.button.borderRightColor
        }
        if (
          block.data.button.borderBottomStyle !== 'none' &&
          block.data.button.borderBottomWidth &&
          block.data.button.borderBottomColor
        ) {
          buttonAttrs['border-bottom'] =
            block.data.button.borderBottomWidth +
            ' ' +
            block.data.button.borderBottomStyle +
            ' ' +
            block.data.button.borderBottomColor
        }
        if (
          block.data.button.borderLeftStyle !== 'none' &&
          block.data.button.borderLeftWidth &&
          block.data.button.borderLeftColor
        ) {
          buttonAttrs['border-left'] =
            block.data.button.borderLeftWidth +
            ' ' +
            block.data.button.borderLeftStyle +
            ' ' +
            block.data.button.borderLeftColor
        }
      }

      const buttonBlock = {
        tagName: 'mj-button',
        attributes: objectAsKebab(buttonAttrs),
        content: block.data.button.text
      }

      return buttonBlock

    case 'divider':
      // console.log('divider', block)

      const dividerAttrs: any = {
        align: block.data.align,
        'border-color': block.data.borderColor,
        'border-style': block.data.borderStyle,
        'border-width': block.data.borderWidth,
        padding: 0
      }

      if (block.data.backgroundColor) {
        dividerAttrs['container-background-color'] = block.data.backgroundColor
      }

      if (block.data.width !== '100%') {
        dividerAttrs['width'] = block.data.width
      }

      if (block.data.paddingControl === 'all') {
        dividerAttrs['padding'] = block.data.padding
      } else {
        if (block.data.paddingTop) dividerAttrs['padding-top'] = block.data.paddingTop
        if (block.data.paddingRight) dividerAttrs['padding-right'] = block.data.paddingRight
        if (block.data.paddingBottom) dividerAttrs['padding-bottom'] = block.data.paddingBottom
        if (block.data.paddingLeft) dividerAttrs['padding-left'] = block.data.paddingLeft
      }

      const dividerBlock = {
        tagName: 'mj-divider',
        attributes: objectAsKebab(dividerAttrs)
      }

      return dividerBlock

    case 'openTracking':
      return {
        tagName: 'mj-image',
        attributes: {
          src: '{{ open_tracking_pixel_src }}',
          alt: '',
          height: '1px',
          width: '1px'
        }
      }
      break

    default:
      console.log('mjml not implemented', block)
      return {
        tagName: 'not-implemented',
        attributes: {},
        children: children
      }
  }
}

export const ExportHTML = (editorData: any, urlParams: any) => {
  return mjml2html(
    json2mjml(treeToMjmlJSON(editorData.data.styles, editorData, '', urlParams, undefined))
  )
}

interface PreviewProps {
  tree: any
  templateData: string
  isMobile: boolean
  deviceWidth: number
  urlParams: any
  toggleDevice: () => void
  closePreview: () => void
}

const Preview = (props: PreviewProps) => {
  const [tab, setTab] = useState('html')
  const jsonMjml = treeToMjmlJSON(
    props.tree.data.styles,
    props.tree,
    props.templateData,
    props.urlParams,
    undefined
  )
  // console.log('json mjml', jsonMjml)
  const mjml = json2mjml(jsonMjml)
  // console.log('mjml', mjml)
  const mjmlBody = Prism.highlight(mjml, Prism.languages.xml, 'xml')
  const html = mjml2html(mjml)

  const iframeProps = {
    content: html.html,
    style: {
      width: props.isMobile ? '400px' : '100%',
      margin: '0 auto 0 auto',
      display: 'block',
      transition: 'all 0.1s'
    },
    sizeSelector: '.ant-drawer-body',
    id: 'htmlCompiled'
  }

  // console.log('html', html.errors)
  return (
    <div className="rmdeditor-layout-middle preview">
      <div className="rmdeditor-topbar">
        <span className={CSS.pull_right}>
          <Space>
            <Button.Group>
              <Button
                size="small"
                type="text"
                disabled={props.deviceWidth === MobileWidth}
                onClick={() => props.toggleDevice()}
              >
                <FontAwesomeIcon icon={faMobileAlt} />
              </Button>
              <Button
                size="small"
                type="text"
                disabled={props.deviceWidth === DesktopWidth}
                onClick={() => props.toggleDevice()}
              >
                <FontAwesomeIcon icon={faDesktop} />
              </Button>
            </Button.Group>
            <Button type="primary" size="small" ghost onClick={() => props.closePreview()}>
              <FontAwesomeIcon icon={faPen} />
              &nbsp; Edit
            </Button>
          </Space>
        </span>
        <Tabs
          activeKey={tab}
          centered
          onChange={(k) => setTab(k)}
          style={{ position: 'absolute', top: '6px' }}
          items={[
            {
              key: 'html',
              label: 'HTML'
            },
            {
              key: 'mjml',
              label: 'MJML'
            }
          ]}
        />
      </div>
      {tab === 'html' && (
        <div id="iframe-container">
          <div className="rmdeditor-transparent">
            <Iframe {...iframeProps} />
          </div>
        </div>
      )}

      {tab === 'mjml' && (
        <div className="rmdeditor-code-bg">
          {html.errors &&
            html.errors.length > 0 &&
            html.errors.map((err: any, i: number) => (
              <Alert
                key={i}
                className="rmdeditor-margin-b-s"
                message={err.formattedMessage}
                type="error"
              />
            ))}
          <pre
            className="language-xml"
            style={{
              background: 'none',
              margin: '0',
              padding: '0',
              wordWrap: 'break-word',
              whiteSpace: 'pre-wrap',
              wordBreak: 'normal'
            }}
          >
            <code dangerouslySetInnerHTML={{ __html: mjmlBody }} />
          </pre>
        </div>
      )}
    </div>
  )
}

export default Preview
