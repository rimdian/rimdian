import { useQuery } from '@tanstack/react-query'
import { MessageTemplate } from './interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { Divider, Select, Spin } from 'antd'
import ButtonUpsertTemplate from './button_upsert_email'

// Form input to select an email template
interface EmailTemplateInputProps {
  value?: string
  onChange?: (value: string) => void
}

const EmailTemplateInput = (props: EmailTemplateInputProps) => {
  const { value, onChange } = props
  const workspaceCtx = useCurrentWorkspaceCtx()

  // email tempaltes
  const { data, isLoading, refetch, isFetching } = useQuery<MessageTemplate[]>(
    ['emailTemplates'],
    (): Promise<MessageTemplate[]> => {
      return new Promise((resolve, reject) => {
        workspaceCtx
          .apiGET(
            '/messageTemplate.list?workspace_id=' + workspaceCtx.workspace.id + '&channel=email'
          )
          .then((data: any) => {
            resolve(data as MessageTemplate[])
          })
          .catch((e) => {
            reject(e)
          })
      })
    }
  )

  if (isLoading) return <Spin size="small" />

  if (data && data.length === 0) {
    return (
      <ButtonUpsertTemplate
        btnProps={{ isLoading: isFetching, type: 'primary', block: true }}
        onSuccess={() => refetch()}
      >
        New template
      </ButtonUpsertTemplate>
    )
  }

  return (
    <Select
      value={value}
      onChange={(value) => onChange && onChange(value)}
      showSearch
      loading={isFetching}
      style={{ width: '100%' }}
      options={data?.map((emailTemplate: MessageTemplate) => ({
        label: emailTemplate.name,
        value: emailTemplate.id
      }))}
      dropdownRender={(menu) => (
        <>
          {menu}
          <Divider style={{ margin: '8px 0' }} />
          <ButtonUpsertTemplate
            btnProps={{ type: 'primary', ghost: true, block: true }}
            onSuccess={() => refetch()}
          >
            New template
          </ButtonUpsertTemplate>
        </>
      )}
    />
  )
}

export default EmailTemplateInput
