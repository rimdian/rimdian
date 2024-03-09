export default {}
// import { useState } from 'react'
// import { Button, message, Popconfirm, Tooltip } from 'antd'
// import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
// import { faTrashCan } from '@fortawesome/free-regular-svg-icons'
// import { SizeType } from 'antd/lib/config-provider/SizeContext'

// type Props = {
//   table: string
//   column: string
//   workspaceId: string
//   apiPOST: (endpoint: string, data: any) => Promise<any>
//   onComplete: () => void
//   btnSize?: SizeType
// }

// const DeleteCustomColumnsButton = (props: Props) => {
//   const [loading, setLoading] = useState(false)

//   const onConfirm = () => {
//     setLoading(true)

//     props
//       .apiPOST('/customDimension.delete', {
//         workspace_id: props.workspaceId,
//         table: props.table,
//         column: props.column
//       })
//       .then(() => {
//         setLoading(false)
//         message.success('This custom dimension has been deleted!')
//         props.onComplete()
//       })
//       .catch((_) => {
//         setLoading(false)
//       })
//   }

//   return (
//     <Popconfirm
//       okText="Delete custom dimension"
//       okButtonProps={{ danger: true }}
//       placement="topRight"
//       title="Would you like to delete this custom dimension with all its data?"
//       onConfirm={onConfirm}
//       disabled={loading}
//     >
//       <Tooltip title="Delete custom dimension" placement="bottom">
//         <Button type="default" size={props.btnSize} loading={loading}>
//           <FontAwesomeIcon icon={faTrashCan} />
//         </Button>
//       </Tooltip>
//     </Popconfirm>
//   )
// }

// export default DeleteCustomColumnsButton
