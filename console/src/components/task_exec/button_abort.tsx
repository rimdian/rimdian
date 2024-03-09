import { faTimes } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { Button, message, Popconfirm } from 'antd'
import { TaskExec } from 'interfaces'
import { useState } from 'react'
import CSS from 'utils/css'

type ButtonAbortTaskProps = {
  taskExec: TaskExec
  onAbort?: () => void
  workspaceId: string
  apiPOST: (endpoint: string, data: any) => Promise<any>
}

const ButtonAbortTask = (props: ButtonAbortTaskProps) => {
  const [isAborting, setIsAborting] = useState(false)
  if (props.taskExec.status === 1 || props.taskExec.status === -2) return null

  const abortTask = () => {
    if (isAborting) return
    setIsAborting(true)

    props
      .apiPOST('/taskExec.abort', {
        id: props.taskExec.id,
        workspace_id: props.workspaceId
      })
      .then(() => {
        props.onAbort && props.onAbort()
      })
      .catch((e) => {
        setIsAborting(false)
        message.error('Error aborting task: ' + e)
      })
  }

  return (
    <Popconfirm title="Do you really want to abort this task?" onConfirm={abortTask}>
      <Button className={CSS.margin_r_xs} loading={isAborting} danger size="small">
        <FontAwesomeIcon icon={faTimes} />
      </Button>
    </Popconfirm>
  )
}
export default ButtonAbortTask
