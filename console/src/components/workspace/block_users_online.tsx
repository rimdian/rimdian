import { Spin, Tooltip } from 'antd'
import FormatNumber from 'utils/format_number'
import { useEffect, useRef, useState } from 'react'
import CSS from 'utils/css'
import { useRimdianCube } from './context_cube'

export type UsersOnlineProps = {
  workspaceId: string
  timezone: string
  refreshKey: string
}

export const UsersOnline = (props: UsersOnlineProps) => {
  const { cubeApi } = useRimdianCube()
  const refreshKey = useRef('')
  const [loadingOnline, setLoadingOnline] = useState<boolean>(true)
  const [usersOnline, setUsersOnline] = useState<string | undefined>(undefined)
  const [loading24h, setLoading24h] = useState<boolean>(true)
  const [users24h, setUsers24h] = useState<string | undefined>(undefined)

  useEffect(() => {
    if (refreshKey.current === props.refreshKey) {
      return
    }

    refreshKey.current = props.refreshKey

    setLoadingOnline(true)
    setLoading24h(true)

    cubeApi
      .load({
        measures: ['Session.users_online'],
        dimensions: [],
        timezone: props.timezone,
        renewQuery: true
      })
      .then((resultSet) => {
        setUsersOnline(FormatNumber(resultSet?.tablePivot()[0]['Session.users_online'] as number))
        setLoadingOnline(false)
        // if (error) {
        //   setError(undefined)
        // }
      })
      .catch((_error) => {
        // setError(error.toString())
      })

    cubeApi
      .load({
        measures: ['Session.users_last_24h'],
        dimensions: [],
        timezone: props.timezone,
        renewQuery: true
      })
      .then((resultSet) => {
        setUsers24h(FormatNumber(resultSet?.tablePivot()[0]['Session.users_last_24h'] as number))
        setLoading24h(false)
        // if (error) {
        //   setErrorGraph(undefined)
        // }
      })
      .catch((_error) => {
        // setErrorGraph(error.toString())
      })
  }, [props.refreshKey, cubeApi, props.timezone, loadingOnline])

  return (
    <>
      <Tooltip title="In the last 5 minutes">
        <span className={CSS.font_size_m + ' ' + CSS.font_weight_semibold}>
          {loadingOnline ? <Spin size="small" className={CSS.margin_r_s} /> : usersOnline}
        </span>
        &nbsp;
        <span className={CSS.padding_r_l}>users online</span>
      </Tooltip>
      <span className={CSS.font_size_m + ' ' + CSS.font_weight_semibold}>
        {loading24h ? <Spin size="small" className={CSS.margin_r_s} /> : users24h}
      </span>
      &nbsp; users in the last 24h
    </>
  )
}
