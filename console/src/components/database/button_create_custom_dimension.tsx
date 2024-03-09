export default {}
// import { useState } from 'react'
// import { Form, Input, Button, Select, message, Drawer, Switch, InputNumber } from 'antd'
// import { ButtonType } from 'antd/lib/button'
// import { SizeType } from 'antd/lib/config-provider/SizeContext'
// import { CustomColumns } from 'interfaces'
// import Messages from 'utils/formMessages'

// type CreateCustomColumnsButtonProps = {
//   workspaceId: string
//   customDimensions: CustomColumns[]
//   tableName: string
//   btnContent: JSX.Element
//   apiPOST: (endpoint: string, data: any) => Promise<any>
//   onComplete: () => void
//   btnType?: ButtonType
//   btnSize?: SizeType
//   btnBlock?: boolean
// }

// const CreateCustomColumnsButton = (props: CreateCustomColumnsButtonProps) => {
//   const [form] = Form.useForm()
//   const [drawerVisible, setDrawerVisible] = useState(false)
//   const [loading, setLoading] = useState(false)

//   const closeDrawer = () => {
//     if (drawerVisible) form.resetFields()
//     setDrawerVisible(false)
//   }

//   const onFinish = (values: any) => {
//     // console.log('values', values);

//     if (loading) return

//     setLoading(true)

//     values.workspace_id = props.workspaceId
//     values.table = props.tableName

//     props
//       .apiPOST('/customDimension.create', values)
//       .then(() => {
//         message.success('The custom dimension has successfully been created.')
//         form.resetFields()

//         setLoading(false)
//         setDrawerVisible(false)
//         props.onComplete()
//       })
//       .catch((_) => {
//         setLoading(false)
//       })
//   }

//   const initialValues = {}

//   // console.log('initialValues', initialValues);

//   return (
//     <>
//       <Button
//         type={props.btnType}
//         block={props.btnBlock}
//         size={props.btnSize}
//         onClick={() => setDrawerVisible(true)}
//       >
//         {props.btnContent}
//       </Button>

//       {drawerVisible && (
//         <Drawer
//           title="Add a new custom dimension"
//           width="90%"
//           visible={true}
//           onClose={closeDrawer}
//           footer={[
//             <Button key="a" loading={loading} onClick={closeDrawer} style={{ marginRight: 8 }}>
//               Cancel
//             </Button>,
//             <Button
//               key="b"
//               loading={loading}
//               onClick={() => {
//                 form.submit()
//               }}
//               type="primary"
//             >
//               Confirm
//             </Button>
//           ]}
//         >
//           <div className="block-cta padding-v-l margin-b-xl">
//             <p>In Rimdian, a Custom Dimension is a custom column persisted in a database table.</p>
//             <p>It is used to store extra data specific to your business case.</p>
//           </div>

//           <Form
//             form={form}
//             name="custom_column"
//             initialValues={initialValues}
//             labelCol={{ span: 6 }}
//             wrapperCol={{ span: 18 }}
//             onFinish={onFinish}
//           >
//             {/* <Form.Item name="table" label="Table" rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}>
//                     <Select options={[
//                         { value: 'users', label: 'Users' },
//                         { value: 'timeline_sessions', label: 'Sessions' },
//                         // TODO: add more tables
//                     ]} />
//                 </Form.Item> */}

//             <Form.Item
//               name="name"
//               label="Column name"
//               extra='Column names of custom dimensions will always be prefixed with a "_"'
//               rules={[
//                 {
//                   required: true,
//                   type: 'string',
//                   pattern: /^([a-z])+([a-z0-9_])+$/,
//                   message: Messages.InvalidTableName
//                 }
//               ]}
//             >
//               <Input addonBefore="_" />
//             </Form.Item>

//             <Form.Item
//               name="type"
//               label="Data type"
//               rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
//             >
//               <Select
//                 options={[
//                   { value: 'varchar', label: 'Text - less than 21,845 characters' },
//                   { value: 'longtext', label: 'Long text - more than 21,845 characters' },
//                   { value: 'boolean', label: 'True or False' },
//                   { value: 'number', label: 'Number' },
//                   { value: 'date', label: 'Date (YYYY-MM-DD)' },
//                   { value: 'datetime', label: 'Date & time (YYYY-MM-DD HH:mm:ss)' },
//                   { value: 'timestamp', label: 'Timestamp (secs)' },
//                   { value: 'json', label: 'JSON object' }
//                 ]}
//                 onChange={(value) => {
//                   if (value === 'varchar') {
//                     form.setFieldsValue({ size: 50 })
//                   } else {
//                     form.setFieldsValue({ size: undefined })
//                   }
//                 }}
//               />
//             </Form.Item>

//             <Form.Item noStyle shouldUpdate>
//               {(funcs) => {
//                 if (funcs.getFieldValue('type') === 'varchar') {
//                   return (
//                     <Form.Item
//                       name="size"
//                       label="Text max characters"
//                       rules={[{ required: true, type: 'integer', min: 1, max: 21845 }]}
//                     >
//                       <InputNumber />
//                     </Form.Item>
//                   )
//                 }
//               }}
//             </Form.Item>

//             <Form.Item
//               valuePropName="checked"
//               name="is_required"
//               label="Is required?"
//               rules={[{ required: false, type: 'boolean' }]}
//             >
//               <Switch />
//             </Form.Item>

//             <Form.Item noStyle shouldUpdate>
//               {(funcs) => {
//                 const type = funcs.getFieldValue('type')
//                 if (!type) return

//                 switch (type) {
//                   case 'varchar':
//                     return (
//                       <Form.Item
//                         name="default_text"
//                         label="Default text value"
//                         rules={[{ required: false, type: 'string' }]}
//                       >
//                         <Input
//                           onChange={(e) => {
//                             // remove emojis
//                             if (e.target.value) {
//                               funcs.setFieldsValue({
//                                 defaultText: e.target.value.replace(/\p{Emoji}/gu, '')
//                               })
//                             }
//                           }}
//                         />
//                       </Form.Item>
//                     )
//                   case 'boolean':
//                     return (
//                       <Form.Item
//                         name="default_boolean"
//                         label="Default value"
//                         rules={[{ required: false, type: 'boolean' }]}
//                       >
//                         <Select
//                           allowClear
//                           options={[
//                             { value: true, label: 'True' },
//                             { value: false, label: 'False' }
//                           ]}
//                         />
//                       </Form.Item>
//                     )
//                   case 'number':
//                     return (
//                       <Form.Item
//                         name="default_number"
//                         label="Default number value"
//                         rules={[{ required: false, type: 'number' }]}
//                       >
//                         <InputNumber style={{ width: '100%' }} />
//                       </Form.Item>
//                     )
//                   case 'timestamp':
//                     return (
//                       <>
//                         <Form.Item
//                           name="default_timestamp"
//                           label="Default value"
//                           rules={[{ required: false, type: 'string' }]}
//                         >
//                           <Select
//                             options={[{ value: 'CURRENT_TIMESTAMP', label: 'CURRENT_TIMESTAMP' }]}
//                           />
//                         </Form.Item>

//                         <Form.Item
//                           name="extra"
//                           label="Extra"
//                           rules={[{ required: false, type: 'string' }]}
//                         >
//                           <Select
//                             options={[
//                               {
//                                 value: 'ON UPDATE CURRENT_TIMESTAMP',
//                                 label: 'ON UPDATE CURRENT_TIMESTAMP'
//                               }
//                             ]}
//                           />
//                         </Form.Item>
//                       </>
//                     )
//                   default:
//                 }
//               }}
//             </Form.Item>

//             <Form.Item
//               name="description"
//               label="Description"
//               rules={[{ required: false, type: 'string' }]}
//             >
//               <Input />
//             </Form.Item>
//           </Form>
//         </Drawer>
//       )}
//     </>
//   )
// }

// export default CreateCustomColumnsButton
