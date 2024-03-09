import { useState } from 'react'
import { Button, Popconfirm, Tooltip, Modal } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { SizeType } from 'antd/lib/config-provider/SizeContext'
import { faArrowRotateRight } from '@fortawesome/free-solid-svg-icons'
import Code from 'utils/prism'
import CSS from 'utils/css'

type Props = {
  dataLogId: string
  workspaceId: string
  apiPOST: (endpoint: string, data: any) => Promise<any>
  onComplete: () => void
  btnSize?: SizeType
  className?: string
}

const ReprocessDataLogButton = (props: Props) => {
  const [loading, setLoading] = useState(false)

  const onConfirm = () => {
    setLoading(true)

    props
      .apiPOST('/dataLog.reprocessOne', {
        workspace_id: props.workspaceId,
        id: props.dataLogId
      })
      .then((data) => {
        setLoading(false)
        Modal.info({
          title: 'Data log results',
          // content is a div that show the JSON data
          content: (
            <div className={CSS.font_size_xs} style={{ maxHeight: 650, overflow: 'auto' }}>
              <Code language="json">{JSON.stringify(data, null, 2)}</Code>
            </div>
          )
        })
        // console.log(data)
        props.onComplete()
      })
      .catch((_) => {
        setLoading(false)
      })
  }

  return (
    <Popconfirm
      okText="Reprocess"
      placement="topRight"
      title="Would you like to reprocess this data import?"
      onConfirm={onConfirm}
      disabled={loading}
    >
      <Tooltip title="Reprocess this data import" placement="bottom">
        <Button type="default" className={props.className} size={props.btnSize} loading={loading}>
          <FontAwesomeIcon icon={faArrowRotateRight} />
        </Button>
      </Tooltip>
    </Popconfirm>
  )
}

export default ReprocessDataLogButton
