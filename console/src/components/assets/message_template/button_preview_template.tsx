import { useQuery } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { MessageTemplate } from './interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { Descriptions, Drawer, Space, Spin, Tabs } from 'antd'
import IframeSandbox from 'components/email_editor/UI/Widgets/Iframe'
import Nunjucks from 'nunjucks'
import CSS from 'utils/css'

type PreviewMessageTemplate = {
  id: string
  percentage?: number
  version?: number
}

type ButtonPreviewMessageTemplateProps = {
  templates: PreviewMessageTemplate[]
  selectedID?: string
  children: JSX.Element
}

const ButtonPreviewMessageTemplate = (props: ButtonPreviewMessageTemplateProps) => {
  const [visible, setVisible] = useState(false)

  return (
    <>
      <span onClick={() => setVisible(true)}>{props.children}</span>
      {visible && (
        <DrawerPreviewMessageTemplate
          templates={props.templates}
          selectedID={props.selectedID}
          setVisible={setVisible}
        />
      )}
    </>
  )
}

type DrawerPreviewMessageTemplateProps = {
  templates: PreviewMessageTemplate[]
  selectedID?: string
  setVisible: (value: boolean) => void
}
const DrawerPreviewMessageTemplate = (props: DrawerPreviewMessageTemplateProps) => {
  const hasMany = props.templates.length > 1
  return (
    <Drawer
      title={!hasMany ? 'Preview template' : 'Preview templates'}
      placement="right"
      closable={true}
      onClose={() => props.setVisible(false)}
      open={true}
      width={800}
    >
      {hasMany && (
        <Tabs
          defaultActiveKey={props.selectedID}
          items={props.templates.map((template) => {
            return {
              key: template.id,
              label: (
                <Space>
                  {template.percentage !== undefined && template.percentage !== 100 && (
                    <span className={CSS.font_size_xs + ' ' + CSS.font_weight_bold}>
                      {template.percentage}%
                    </span>
                  )}
                  {template.id}
                </Space>
              ),
              children: (
                <MessageTemplatePreview
                  key={template.id}
                  templateID={template.id}
                  version={template.version}
                />
              )
            }
          })}
        />
      )}
      {!hasMany && (
        <MessageTemplatePreview
          key={props.templates[0].id}
          templateID={props.templates[0].id}
          version={props.templates[0].version}
        />
      )}
    </Drawer>
  )
}

type MessageTemplatePreviewProps = {
  templateID: string
  version?: number
}

const MessageTemplatePreview = (props: MessageTemplatePreviewProps) => {
  const workspaceCtx = useCurrentWorkspaceCtx()

  const { isLoading, data, isFetching } = useQuery<MessageTemplate>(
    ['message_template', workspaceCtx.workspace.id, props.templateID],
    (): Promise<MessageTemplate> => {
      return new Promise((resolve, reject) => {
        workspaceCtx
          .apiGET(
            '/messageTemplate.get?workspace_id=' +
              workspaceCtx.workspace.id +
              '&id=' +
              props.templateID +
              '&version=' +
              (props.version || '')
          )
          .then((data: any) => {
            // console.log(data)
            resolve(data as MessageTemplate)
          })
          .catch((e) => {
            reject(e)
          })
      })
    }
  )

  return (
    <div>
      {(isLoading || isFetching) && <Spin />}
      {data?.channel === 'email' && <EmailMessageTemplatePreview data={data} />}
    </div>
  )
}

type EmailMessageTemplatePreviewProps = {
  data: MessageTemplate
}

const EmailMessageTemplatePreview = (props: EmailMessageTemplatePreviewProps) => {
  // const [isMobile, setIsMobile] = useState(false)

  const html = useMemo(() => {
    let content = props.data.email.content

    // exec nunjucks if has tags
    if (
      props.data.test_data &&
      props.data.test_data !== '' &&
      (content.includes('{{') || content.includes('{%'))
    ) {
      // console.log('got markup', content)
      // console.log('data', props.data.test_data)

      const jsonData = JSON.parse(props.data.test_data)

      try {
        const stringResult = Nunjucks.renderString(content, jsonData || {})
        // console.log('stringResult', stringResult)
        content = stringResult
      } catch (e) {
        // ignore error and templating
        console.error(e)
      }
    }

    return content
  }, [props.data.test_data, props.data.email.content])

  const subject = useMemo(() => {
    let content = props.data.email.subject

    // exec nunjucks if has tags
    if (
      props.data.test_data &&
      props.data.test_data !== '' &&
      (content.includes('{{') || content.includes('{%'))
    ) {
      // console.log('got markup', content)
      // console.log('data', props.data.test_data)

      const jsonData = JSON.parse(props.data.test_data)

      try {
        const stringResult = Nunjucks.renderString(content, jsonData || {})
        // console.log('stringResult', stringResult)
        content = stringResult
      } catch (e) {
        // ignore error and templating
        console.error(e)
      }
    }

    return content
  }, [props.data.test_data, props.data.email.subject])

  return (
    <div>
      <Descriptions size="small" layout="vertical" className={CSS.margin_b_m}>
        <Descriptions.Item label="Template name">{props.data.name}</Descriptions.Item>
        <Descriptions.Item label="utm_source / medium / campaign / content">
          {props.data.utm_source} / {props.data.utm_medium} / {props.data.utm_campaign} /{' '}
          {props.data.id}
        </Descriptions.Item>
      </Descriptions>
      <Descriptions size="small" layout="vertical">
        <Descriptions.Item label="Subject">{subject}</Descriptions.Item>
        <Descriptions.Item label="From">
          {props.data.email.from_name + ' <' + props.data.email.from_address + '>'}{' '}
        </Descriptions.Item>
        {props.data.email.reply_to && (
          <Descriptions.Item label="Reply to">{props.data.email.reply_to}</Descriptions.Item>
        )}
      </Descriptions>

      <IframeSandbox
        {...{
          content: html,
          style: {
            // width: isMobile ? '400px' : '100%',
            width: '100%',
            margin: '0 auto 0 auto',
            display: 'block',
            transition: 'all 0.1s'
          },
          sizeSelector: '.ant-drawer-body',
          id: 'htmlCompiled'
        }}
      />
    </div>
  )
}
export default ButtonPreviewMessageTemplate
