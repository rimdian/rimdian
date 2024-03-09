import { CubeContext } from '@cubejs-client/react'
import { CurrentWorkspaceCtxValue } from './context_current_workspace'
import dayjs from 'dayjs'
import { useContext, useEffect, useRef, useState } from 'react'
import { Alert, Modal } from 'antd'
import { ResultSet } from '@cubejs-client/core'

interface LicenseWarningProps {
  workspaceCtx: CurrentWorkspaceCtxValue
}

export default function LicenseWarning(props: LicenseWarningProps) {
  const queryDone = useRef(false)
  const { cubejsApi } = useContext(CubeContext)
  const [dataQuotaFull, setDataQuota] = useState(false)

  // do a cubejs query to get the count of data_logs in the last 90 days
  useEffect(() => {
    if (queryDone.current || !cubejsApi) {
      return
    }

    queryDone.current = true

    const from = dayjs().subtract(90, 'day')

    cubejsApi
      .load({
        measures: ['Data_log.count'],
        filters: [
          {
            dimension: 'Data_log.event_at_trunc',
            operator: 'gte',
            values: [from.toISOString()]
          }
        ]
      })
      .then((res: ResultSet) => {
        const total = res.rawData()[0]['Data_log.count']
        if (total >= props.workspaceCtx.workspace.license_info.dlo90) {
          setDataQuota(true)
        }
      })
      .catch((err) => {
        console.error('error getting data_logs count', err)
      })
  }, [cubejsApi, props.workspaceCtx.workspace.license_info.dlo90])

  if (dataQuotaFull) {
    return (
      <Modal open={true} footer={null} centered closable={false}>
        <Alert
          message="Data quota reached"
          description={`You have reached your data quota of ${props.workspaceCtx.workspace.license_info.dlo90} data_logs over the last 90 days. Please upgrade your license to continue using Rimdian.`}
          type="warning"
          showIcon
        />
      </Modal>
    )
  }
  return null
}
