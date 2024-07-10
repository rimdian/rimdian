import { useEffect, useMemo, useState } from 'react'
import { Button, Steps } from 'antd'
import { App } from 'interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Block from 'components/common/block'
import CSS from 'utils/css'

interface AppUpgradingProps {
  app: App
}

const BlockAppUpgrading = (props: AppUpgradingProps) => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [stateVisible, setStateVisible] = useState(false)

  const upgradeTask = useMemo(() => {
    return workspaceCtx.runningTasks.find((x) => x.task_id === 'system_upgrade_app')
  }, [workspaceCtx.runningTasks])

  const currentStep = useMemo(() => {
    if (!upgradeTask) return 0
    switch (upgradeTask.state.workers[0].stage) {
      case 'app_tables':
        return 1
      case 'extra_columns':
        return 2
      case 'finalize':
        return 3
      default:
        return 0
    }
  }, [upgradeTask])

  useEffect(() => {
    // refetch apps if no upgrade task is running
    if (!upgradeTask) {
      workspaceCtx.refetchApps()
    }
  }, [upgradeTask, workspaceCtx])

  return (
    <div className={CSS.container + ' ' + CSS.margin_t_xxl}>
      <Block
        title="Upgrading status"
        style={{ width: 600 }}
        extra={
          <Button type="primary" ghost size="small" onClick={() => setStateVisible(!stateVisible)}>
            View state
          </Button>
        }
      >
        <Steps
          direction="vertical"
          className={CSS.padding_a_xl}
          current={currentStep}
          items={[
            {
              title: 'Validation',
              description: 'Validating the new app manifest'
            },
            {
              title: 'App tables',
              description: 'Creating & updating app custom tables'
            },
            {
              title: 'Extra columns',
              description: 'Adding or removing app extra columns on native tables'
            },
            {
              title: 'Finalize',
              description: 'Update tasks, data hooks, and replace the app manifest'
            }
          ]}
        />

        {stateVisible && (
          <pre className={CSS.padding_a_l}>{JSON.stringify(upgradeTask, null, 2)}</pre>
        )}
      </Block>
    </div>
  )
}

export default BlockAppUpgrading
