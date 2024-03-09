export default {}
// import { FC, useState } from 'react'
// import {
//   Form,
//   Input,
//   Button,
//   message,
//   Drawer,
//   Table,
//   Modal,
//   Radio,
//   Select,
//   InputNumber,
//   Switch,
//   Row,
//   Col,
//   Tag
// } from 'antd'
// import { ButtonType } from 'antd/lib/button'
// import { SizeType } from 'antd/lib/config-provider/SizeContext'
// import { CustomTable, TableColumn } from 'interfaces'
// import Messages from 'utils/formMessages'
// import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
// import { faAngleDown, faAngleUp } from '@fortawesome/free-solid-svg-icons'
// import { faTrashCan } from '@fortawesome/free-regular-svg-icons'

// type CreateCustomTableButtonProps = {
//   workspaceId: string
//   customTables: CustomTable[]
//   btnContent: JSX.Element
//   apiPOST: (endpoint: string, data: any) => Promise<any>
//   onComplete: () => void
//   btnType?: ButtonType
//   btnSize?: SizeType
//   btnBlock?: boolean
// }

// const CreateCustomTableButton = (props: CreateCustomTableButtonProps) => {
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
//     values.name = '_' + values.name

//     props
//       .apiPOST('/customTable.create', values)
//       .then((res) => {
//         message.success('The custom table has successfully been created.')
//         form.resetFields()

//         setLoading(false)
//         setDrawerVisible(false)
//         props.onComplete()
//       })
//       .catch((_) => {
//         setLoading(false)
//       })
//   }

//   const initialValues = {
//     storage_type: 'columnstore',
//     columns: [
//       {
//         name: 'id',
//         label: 'ID',
//         type: 'varchar',
//         size: 60,
//         is_required: true,
//         // description?: string
//         created_at: new Date(),
//         updated_at: new Date()
//       },
//       {
//         name: 'db_created_at',
//         label: 'Created at in DB',
//         type: 'timestamp',
//         is_required: true,
//         default_timestamp: 'CURRENT_TIMESTAMP',
//         // description?: string
//         created_at: new Date(),
//         updated_at: new Date()
//       },
//       {
//         name: 'db_updated_at',
//         label: 'Updated at in DB',
//         type: 'timestamp',
//         is_required: true,
//         default_timestamp: 'CURRENT_TIMESTAMP',
//         extra: 'ON UPDATE CURRENT_TIMESTAMP',
//         // description?: string
//         created_at: new Date(),
//         updated_at: new Date()
//       }
//     ],
//     unique_key: ['id']
//   }

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
//           title="Create a custom table"
//           width="95%"
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
//             <p>In Rimdian, a Custom Table is a table persisted in the database.</p>
//             <p>It is used to store data that is specific to your business case.</p>
//             <p>
//               You can then read/insert/update/delete its data via API, and use it to produce
//               Analytics reports.
//             </p>
//           </div>

//           <Form
//             form={form}
//             name="custom_table"
//             initialValues={initialValues}
//             labelCol={{ span: 8 }}
//             wrapperCol={{ span: 16 }}
//             onFinish={onFinish}
//           >
//             <Row>
//               <Col span={12}>
//                 <Form.Item
//                   name="name"
//                   label="Table name"
//                   extra='To differenciate a custom table from a native table, its name will always be prefixed with an underscore "_"'
//                   rules={[
//                     {
//                       required: true,
//                       type: 'string',
//                       pattern: /^([a-z])+([a-z0-9_])+$/,
//                       message: Messages.InvalidTableName
//                     }
//                   ]}
//                 >
//                   <Input addonBefore="_" />
//                 </Form.Item>

//                 <Form.Item
//                   name="storage_type"
//                   label="Storage type"
//                   rules={[{ required: true, type: 'string' }]}
//                 >
//                   <Radio.Group style={{ width: '100%' }}>
//                     <Radio.Button style={{ textAlign: 'center', width: '50%' }} value="columnstore">
//                       Columnstore
//                     </Radio.Button>
//                     <Radio.Button
//                       style={{ textAlign: 'center', width: '50%' }}
//                       disabled
//                       value="rowstore"
//                     >
//                       In-memory Rowstore
//                     </Radio.Button>
//                   </Radio.Group>
//                 </Form.Item>

//                 <Form.Item
//                   name="description"
//                   label="Description"
//                   rules={[{ required: false, type: 'string' }]}
//                 >
//                   <Input />
//                 </Form.Item>

//                 <Form.Item noStyle shouldUpdate>
//                   {(funcs) => {
//                     const columns = funcs.getFieldValue('columns')
//                     if (!columns || columns.length === 0) return

//                     return (
//                       <>
//                         <Form.Item
//                           name="shard_key"
//                           label="Shard key"
//                           extra="Rows will be distributed to the DB nodes according to this key. Make sure this values are as random as possible."
//                           rules={[
//                             {
//                               required: true,
//                               type: 'array',
//                               min: 1,
//                               message: Messages.RequiredField
//                             }
//                           ]}
//                         >
//                           <Select
//                             options={columns}
//                             fieldNames={{
//                               value: 'name',
//                               label: 'name'
//                             }}
//                             mode="multiple"
//                           />
//                         </Form.Item>

//                         <Form.Item
//                           name="sort_key"
//                           label="Storage sort key"
//                           rules={[
//                             {
//                               required: true,
//                               type: 'array',
//                               min: 1,
//                               message: Messages.RequiredField
//                             }
//                           ]}
//                           extra="Rows will be sorted and stored by default by this key."
//                         >
//                           <Select
//                             options={columns}
//                             fieldNames={{
//                               value: 'name',
//                               label: 'name'
//                             }}
//                             mode="multiple"
//                           />
//                         </Form.Item>

//                         <Form.Item
//                           name="unique_key"
//                           label="Unique key"
//                           extra="A unique key must contain all of the column(s) of the shard key."
//                           rules={[
//                             {
//                               required: true,
//                               type: 'array',
//                               min: 1,
//                               message: Messages.RequiredField
//                             }
//                           ]}
//                         >
//                           <Select
//                             options={columns}
//                             fieldNames={{
//                               value: 'name',
//                               label: 'name'
//                             }}
//                             mode="multiple"
//                           />
//                         </Form.Item>

//                         <Form.Item
//                           name="timeseries_column"
//                           label="Time series column"
//                           extra="A datetime / timestamp column can be selected for activating time-series functions."
//                           rules={[{ required: false, type: 'string' }]}
//                         >
//                           <Select
//                             options={columns.filter((c: TableColumn) =>
//                               [
//                                 'datetime',
//                                 'timestamp',
//                                 'timestampOnCreate',
//                                 'timestampOnUpdate'
//                               ].includes(c.type)
//                             )}
//                             fieldNames={{
//                               value: 'name',
//                               label: 'name'
//                             }}
//                           />
//                         </Form.Item>

//                         {/*
//                             TODO: add a collection of indexes (hash | ) one day...
//                             <Form.Item name="hashIndexes" label="Hash indexes" rules={[{ required: false, type: 'array' }]}>
//                                 <Select
//                                     options={columns}
//                                     fieldNames={{
//                                         value: 'name',
//                                         label: 'name',
//                                     }}
//                                     mode='multiple'
//                                 />
//                             </Form.Item> */}
//                       </>
//                     )
//                   }}
//                 </Form.Item>
//               </Col>

//               <Col span={1}></Col>
//               <Col span={11}>
//                 <Form.Item
//                   // tooltip={{ icon: <FontAwesomeIcon icon={faQuestionCircle} />, title: "sdazeazeaez" }}
//                   labelCol={{ span: 24 }}
//                   wrapperCol={{ span: 24 }}
//                   name="columns"
//                   label="Columns"
//                   rules={[{ required: true, type: 'array', min: 1 }]}
//                 >
//                   <TableColumnsInput />
//                 </Form.Item>
//               </Col>
//             </Row>
//           </Form>
//         </Drawer>
//       )}
//     </>
//   )
// }

// export default CreateCustomTableButton

// type TableColumnsInputProps = {
//   value?: object[]
//   onChange?: (data: any) => void
// }

// const TableColumnsInput: FC<TableColumnsInputProps> = ({ value = [], onChange }) => {
//   const removeColumn = (index: number) => {
//     let columns = value.slice()
//     columns.splice(index, 1)
//     onChange?.(columns)
//   }

//   const movePosition = (fromIndex: number, toIndex: number) => {
//     const updatedColumns: any[] = [...value]

//     if (toIndex >= updatedColumns.length) {
//       var k = toIndex - updatedColumns.length + 1
//       while (k--) {
//         updatedColumns.push(undefined)
//       }
//     }
//     updatedColumns.splice(toIndex, 0, updatedColumns.splice(fromIndex, 1)[0])

//     onChange?.(updatedColumns)
//   }

//   return (
//     <div>
//       {value && value.length > 0 && (
//         <Table
//           size="middle"
//           bordered={false}
//           pagination={false}
//           rowKey="name"
//           showHeader={false}
//           className="margin-b-m"
//           columns={[
//             {
//               title: '',
//               key: 'name',
//               render: (x) => <b>{x.name}</b>
//             },
//             {
//               title: '',
//               key: 'type',
//               render: (x: TableColumn) => {
//                 let defaultValue = undefined
//                 if (x.type === 'varchar' && x.default_string)
//                   defaultValue = "'" + x.default_string + "'"
//                 if (x.type === 'boolean' && x.default_boolean === true) defaultValue = 'true'
//                 if (x.type === 'boolean' && x.default_boolean === false) defaultValue = 'false'
//                 if (x.type === 'timestamp' && x.default_timestamp)
//                   defaultValue = x.default_timestamp

//                 return (
//                   <>
//                     {x.type}
//                     {x.size && '(' + x.size + ')'}{' '}
//                     {x.is_required && (
//                       <Tag color="red">
//                         not null
//                       </Tag>
//                     )}{' '}
//                     {defaultValue && 'default ' + defaultValue} {x.extra && x.extra}
//                   </>
//                 )
//               }
//             },
//             {
//               title: '',
//               key: 'actions',
//               width: 130,
//               render: (_text, _record: any, index: number) => {
//                 return (
//                   <div className={GlobalCSS.text_right}>
//                     <Button.Group className="margin-r-m">
//                       {index !== 0 && (
//                         <Button
//                           type="dashed"
//                           size="small"
//                           onClick={movePosition.bind(null, index, index - 1)}
//                         >
//                           <FontAwesomeIcon icon={faAngleUp} />
//                         </Button>
//                       )}
//                       {index !== value.length - 1 && (
//                         <Button
//                           type="dashed"
//                           size="small"
//                           onClick={movePosition.bind(null, index, index + 1)}
//                         >
//                           <FontAwesomeIcon icon={faAngleDown} />
//                         </Button>
//                       )}
//                     </Button.Group>

//                     <Button type="dashed" size="small" onClick={removeColumn.bind(null, index)}>
//                       <FontAwesomeIcon icon={faTrashCan} />
//                     </Button>
//                   </div>
//                 )
//               }
//             }
//           ]}
//           dataSource={value}
//         />
//       )}

//       <AddColumnButton
//         onComplete={(newColumn: any) => {
//           let columns = value ? value.slice() : []
//           columns.push(newColumn)
//           onChange?.(columns)
//         }}
//       />
//     </div>
//   )
// }

// const AddColumnButton = ({ onComplete }: any) => {
//   const [form] = Form.useForm()
//   const [modalVisible, setModalVisible] = useState(false)

//   const onClicked = () => {
//     setModalVisible(true)
//   }

//   return (
//     <>
//       <Button type="primary" block  onClick={onClicked}>
//         Add column
//       </Button>

//       <Modal
//         visible={modalVisible}
//         title="Add column"
//         okText="Confirm"
//         cancelText="Cancel"
//         width={600}
//         onCancel={() => {
//           setModalVisible(false)
//         }}
//         onOk={() => {
//           form
//             .validateFields()
//             .then((values: any) => {
//               form.resetFields()
//               setModalVisible(false)
//               onComplete(values)
//             })
//             .catch(console.error)
//         }}
//       >
//         <Form
//           form={form}
//           name="form_add_column"
//           labelCol={{ span: 10 }}
//           wrapperCol={{ span: 14 }}
//           layout="horizontal"
//         >
//           <Form.Item
//             name="name"
//             label="Column name"
//             rules={[
//               {
//                 required: true,
//                 type: 'string',
//                 pattern: /^([a-z0-9_])+$/,
//                 message: Messages.InvalidTableColumName
//               }
//             ]}
//           >
//             <Input />
//           </Form.Item>

//           <Form.Item
//             name="type"
//             label="Data type"
//             rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
//           >
//             <Select
//               options={[
//                 { value: 'varchar', label: 'Text - less than 21,845 characters' },
//                 { value: 'longtext', label: 'Long text - more than 21,845 characters' },
//                 { value: 'boolean', label: 'True or False' },
//                 { value: 'number', label: 'Number' },
//                 { value: 'date', label: 'Date (YYYY-MM-DD)' },
//                 { value: 'datetime', label: 'Date & time (YYYY-MM-DD HH:mm:ss)' },
//                 { value: 'timestamp', label: 'Timestamp (secs)' },
//                 { value: 'json', label: 'JSON object' }
//                 // { value: 'timestampOnCreate', label: 'Timestamp on create' },
//                 // { value: 'timestampOnUpdate', label: 'Timestamp on update' },
//               ]}
//               onChange={(value) => {
//                 if (value === 'varchar') {
//                   form.setFieldsValue({ size: 50 })
//                 } else {
//                   form.setFieldsValue({ size: undefined })
//                 }
//               }}
//             />
//           </Form.Item>

//           <Form.Item noStyle shouldUpdate>
//             {(funcs) => {
//               if (funcs.getFieldValue('type') === 'varchar') {
//                 return (
//                   <Form.Item
//                     name="size"
//                     label="Text max characters"
//                     rules={[{ required: true, type: 'integer', min: 1, max: 21845 }]}
//                   >
//                     <InputNumber />
//                   </Form.Item>
//                 )
//               }
//             }}
//           </Form.Item>
//           <Form.Item noStyle shouldUpdate>
//             {(funcs) => {
//               if (funcs.getFieldValue('type')) {
//                 return (
//                   <Form.Item
//                     valuePropName="checked"
//                     name="is_required"
//                     label="Not Null?"
//                     rules={[{ required: false, type: 'boolean' }]}
//                   >
//                     <Switch />
//                   </Form.Item>
//                 )
//               }
//             }}
//           </Form.Item>

//           <Form.Item noStyle shouldUpdate>
//             {(funcs) => {
//               const type = funcs.getFieldValue('type')
//               if (!type) return

//               switch (type) {
//                 case 'varchar':
//                   return (
//                     <Form.Item
//                       name="default_string"
//                       label="Default value"
//                       rules={[{ required: false, type: 'string' }]}
//                     >
//                       <Input
//                         onChange={(e) => {
//                           // remove emojis
//                           if (e.target.value) {
//                             funcs.setFieldsValue({
//                               defaultText: e.target.value.replace(/\p{Emoji}/gu, '')
//                             })
//                           }
//                         }}
//                       />
//                     </Form.Item>
//                   )
//                 case 'boolean':
//                   return (
//                     <Form.Item
//                       name="default_boolean"
//                       label="Default value"
//                       rules={[{ required: false, type: 'boolean' }]}
//                     >
//                       <Select
//                         allowClear
//                         options={[
//                           { value: true, label: 'True' },
//                           { value: false, label: 'False' }
//                         ]}
//                       />
//                     </Form.Item>
//                   )
//                 case 'number':
//                   return (
//                     <Form.Item
//                       name="default_number"
//                       label="Default value"
//                       rules={[{ required: false, type: 'number' }]}
//                     >
//                       <InputNumber style={{ width: '100%' }} />
//                     </Form.Item>
//                   )
//                 case 'timestamp':
//                   return (
//                     <>
//                       <Form.Item
//                         name="default_timestamp"
//                         label="Default value"
//                         rules={[{ required: false, type: 'string' }]}
//                       >
//                         <Select
//                           options={[{ value: 'CURRENT_TIMESTAMP', label: 'CURRENT_TIMESTAMP' }]}
//                         />
//                       </Form.Item>

//                       <Form.Item
//                         name="extra"
//                         label="Extra"
//                         rules={[{ required: false, type: 'string' }]}
//                       >
//                         <Select
//                           options={[
//                             {
//                               value: 'ON UPDATE CURRENT_TIMESTAMP',
//                               label: 'ON UPDATE CURRENT_TIMESTAMP'
//                             }
//                           ]}
//                         />
//                       </Form.Item>
//                     </>
//                   )
//                 default:
//               }
//             }}
//           </Form.Item>

//           <Form.Item
//             name="description"
//             label="Description"
//             rules={[{ required: false, type: 'string' }]}
//           >
//             <Input />
//           </Form.Item>
//         </Form>
//       </Modal>
//     </>
//   )
// }
