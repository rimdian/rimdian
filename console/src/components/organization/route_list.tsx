import { Link, Navigate } from 'react-router-dom'
import { useOrganizationsCtx } from './context_organizations'
import { Button, Form, Input, Modal, Select, Table } from 'antd'
import { Organization } from 'interfaces'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faArrowRight } from '@fortawesome/free-solid-svg-icons'
import { useState } from 'react'
import Messages from 'utils/formMessages'
import { Currencies, Currency } from 'utils/currencies'
import LayoutLanding from 'components/app/layout_landing'
import CSS from 'utils/css'

const RouteOrganizations = () => {
  const orgsCtx = useOrganizationsCtx()

  if (orgsCtx.organizations.length === 1 && !window.Config.MANAGED_RMD) {
    return <Navigate to={'/orgs/' + orgsCtx.organizations[0].id} />
  }

  const onCreateOrganization = (data: any) => {
    orgsCtx.apiPOST('/organization.create', data).then(() => {
      orgsCtx.refreshOrganizations()
    })
  }

  return (
    <LayoutLanding>
      <>
        <Table
          showHeader={false}
          pagination={false}
          dataSource={orgsCtx.organizations}
          rowKey="id"
          size="middle"
          className={CSS.margin_b_l}
          columns={[
            {
              key: 'name',
              render: (org: Organization) => (
                <div>
                  <div>
                    <b>{org.name}</b>
                    <br />
                    <span className={CSS.font_size_xs}>{org.id}</span>
                  </div>
                </div>
              )
            },
            {
              key: 'actions',
              render: (org: Organization) => {
                return (
                  <div className={CSS.text_right}>
                    <Link to={'/orgs/' + org.id}>
                      <Button type="primary">
                        View &nbsp;
                        <FontAwesomeIcon icon={faArrowRight} />
                      </Button>
                    </Link>
                  </div>
                )
              }
            }
          ]}
        />
        <AddOrganizationButton onComplete={onCreateOrganization} />
      </>
    </LayoutLanding>
  )
}

const AddOrganizationButton = ({ onComplete }: any) => {
  const [form] = Form.useForm()
  const [modalVisible, setModalVisible] = useState(false)

  const onClicked = () => {
    setModalVisible(true)
  }

  return (
    <>
      <Button type="primary" block onClick={onClicked}>
        Create organization
      </Button>
      <Modal
        open={modalVisible}
        title="Add an organization"
        okText="Confirm"
        cancelText="Cancel"
        onCancel={() => {
          setModalVisible(false)
        }}
        onOk={() => {
          form
            .validateFields()
            .then((values: any) => {
              form.resetFields()
              setModalVisible(false)
              onComplete(values)
            })
            .catch(console.error)
        }}
      >
        <Form form={form} labelCol={{ span: 10 }} wrapperCol={{ span: 14 }} layout="horizontal">
          <Form.Item
            name="id"
            label="ID"
            rules={[
              {
                required: true,
                type: 'string',
                pattern: /^([a-z0-9])+$/,
                message: Messages.InvalidOrganizationIDFormat
              }
            ]}
          >
            <Input placeholder="a-z0-9" />
          </Form.Item>

          <Form.Item
            name="name"
            label="Name"
            rules={[
              {
                required: false,
                type: 'string',
                message: Messages.RequiredField
              }
            ]}
          >
            <Input />
          </Form.Item>

          <Form.Item
            name="currency"
            label="Main currency"
            rules={[{ required: true, message: Messages.RequiredField }]}
          >
            <Select
              showSearch
              placeholder="Select a currency"
              optionFilterProp="children"
              filterOption={(input: any, option: any) =>
                option.value.toLowerCase().includes(input.toLowerCase())
              }
              options={Currencies.map((c: Currency) => {
                return { value: c.code, label: c.code + ' - ' + c.currency }
              })}
            />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}

export default RouteOrganizations
