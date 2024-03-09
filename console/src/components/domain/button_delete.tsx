import { useState } from 'react'
import { Button, message, Modal, Form, Spin, Tooltip, Select } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faTrashCan } from '@fortawesome/free-regular-svg-icons'
import { SizeType } from 'antd/lib/config-provider/SizeContext'
import Messages from 'utils/formMessages'
import { Domain } from 'interfaces'
import { ButtonType } from 'antd/es/button/buttonHelpers'

type Props = {
  domainId: string
  workspaceId: string
  domains: Domain[]
  apiPOST: (endpoint: string, data: any) => Promise<any>
  onComplete: () => void
  btnSize?: SizeType
  btnType?: ButtonType
}

const DeleteDomainButton = (props: Props) => {
  const [visible, setVisible] = useState(false)
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  const toggleModal = () => {
    setVisible(!visible)
  }

  const onSubmit = () => {
    form
      .validateFields()
      .then((values: any) => {
        form.resetFields()
        setLoading(true)

        props
          .apiPOST('/domain.delete', {
            workspace_id: props.workspaceId,
            id: props.domainId,
            migrate_to_domain_id: values.migrate_to_domain_id
          })
          .then(() => {
            setLoading(false)
            message.success('This domain has been deleted!')
            props.onComplete()
            toggleModal()
          })
          .catch((_) => {
            setLoading(false)
          })
      })
      .catch((_) => {})
  }

  return (
    <>
      <Tooltip title="Delete domain" placement="bottom">
        <Button onClick={toggleModal} type={props.btnType} size={props.btnSize} loading={loading}>
          <FontAwesomeIcon icon={faTrashCan} />
        </Button>
      </Tooltip>
      {visible && (
        <Modal
          title="Delete domain"
          open={true}
          onCancel={toggleModal}
          footer={[
            <Button key="back" loading={loading} onClick={toggleModal}>
              Cancel
            </Button>,
            <Button key="submit" danger type="primary" loading={loading} onClick={onSubmit}>
              Delete domain
            </Button>
          ]}
        >
          <Spin tip="Loading..." spinning={loading}>
            <Form form={form} layout="vertical">
              <Form.Item
                name="migrate_to_domain_id"
                label="Migrate existing data to new domain"
                rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
              >
                <Select
                  options={props.domains.filter((d: Domain) => d.id !== props.domainId)}
                  fieldNames={{ label: 'name', value: 'id' }}
                />
              </Form.Item>

              {/* <Alert type="warning" message="Save this password in a secure place, it cannot be changed or recovered if you loose it!" showIcon /> */}
            </Form>
          </Spin>
        </Modal>
      )}
    </>
  )
}

export default DeleteDomainButton
