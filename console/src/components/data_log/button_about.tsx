import { Button, Drawer } from 'antd'
import { DataLog } from 'interfaces'
import Code from 'utils/prism'
import dayjs from 'dayjs'
import { useState } from 'react'
import CSS from 'utils/css'
import { BlockDataLog } from 'components/item_timeline/block_list'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faSearch } from '@fortawesome/free-solid-svg-icons'

interface ButtonDataLogDataProps {
  workspaceId: string
  apiGET: (endpoint: string) => Promise<any>
  dataImport: DataLog
  accountTimezone: string
}

const ButtonDataLogData = (props: ButtonDataLogDataProps) => {
  const [visible, setVisible] = useState(false)
  return (
    <>
      <Button size="small" type="text" onClick={() => setVisible(true)}>
        <FontAwesomeIcon icon={faSearch} />
      </Button>
      {visible && (
        <Drawer title="Data log" open={visible} width={900} onClose={() => setVisible(false)}>
          <>
            <p>
              <b>ID:</b> {props.dataImport.id}
            </p>
            <p>
              <b>Created at:</b> {props.dataImport.db_created_at}
            </p>
            <p>
              <b>Last update:</b> {dayjs(props.dataImport.db_updated_at).fromNow()} -{' '}
              {dayjs(props.dataImport.db_created_at).tz(props.accountTimezone).format('lll')} in{' '}
              {props.accountTimezone}
            </p>
            <p>
              <b>Context:</b>
            </p>
            <div className={CSS.font_size_xs + ' ' + CSS.margin_b_m}>
              <Code language="json">{JSON.stringify(props.dataImport.context, null, 2)}</Code>
            </div>
            {props.dataImport.origin !== 2 && (
              <>
                <p>
                  <b>Item:</b>
                </p>
                <div className={CSS.font_size_xs + ' ' + CSS.margin_b_m}>
                  <Code language="json">
                    {JSON.stringify(JSON.parse(props.dataImport.item), null, 2)}
                  </Code>
                </div>
              </>
            )}
            <p>
              <b>Operations done:</b>
            </p>
            <BlockDataLog
              workspaceId={props.workspaceId}
              apiGET={props.apiGET}
              parentDataLog={props.dataImport}
            />
          </>
        </Drawer>
      )}
    </>
  )
}

export default ButtonDataLogData
