import cubejs, { CubeApi, LoadMethodOptions, Query, SqlQuery } from '@cubejs-client/core'
import { Workspace } from 'interfaces'
import { useContext, createContext, useRef } from 'react'
import { queue } from 'async'

const RimdianCubeContext = createContext<RimdianCubeContextValue | null>(null)

export function useRimdianCube(): RimdianCubeContextValue {
  const contextValue = useContext(RimdianCubeContext)
  if (!contextValue) {
    throw new Error('Missing RimdianCubeContextProvider in its parent.')
  }
  return contextValue
}

type Props = {
  workspace: Workspace
  children?: React.ReactNode
}

export interface RimdianCubeContextValue {
  cubeApi: CubeApi
}

export const RimdianCubeProvider = (props: Props) => {
  const cubeApi = cubejs(props.workspace.cubejs_token || '', {
    apiUrl: window.Config.CUBEJS_ENDPOINT + '/cubejs-api/v1'
  }) as any

  // wrap the cubeApi.load method in to a queue
  // to limit the number of concurrent requests to the cubejs server
  const loadMethod = cubeApi.load.bind(cubeApi)
  const sqlMethod = cubeApi.sql.bind(cubeApi)

  // https://caolan.github.io/async/v3/docs.html#queue
  const concurency = 6
  const cubeQueue = queue((task: any, callback: any) => {
    if (task.kind === 'load') {
      loadMethod(task.query, task.options)
        .then((res: any) => {
          callback(null, res)
        })
        .catch((err: any) => {
          callback(err)
        })
    }

    if (task.kind === 'sql') {
      sqlMethod(task.query)
        .then((res: any) => {
          callback(null, res)
        })
        .catch((err: any) => {
          callback(err)
        })
    }
  }, concurency)

  cubeApi.load = (query: Query, options?: LoadMethodOptions) => {
    return new Promise((resolve, reject) => {
      cubeQueue.push({ query, options, kind: 'load' }, (err: any, res: any) => {
        if (err) {
          reject(err)
        }

        resolve(res)
      })
    })
  }

  cubeApi.sql = (query: SqlQuery) => {
    return new Promise((resolve, reject) => {
      cubeQueue.push({ query, kind: 'sql' }, (err: any, res: any) => {
        if (err) {
          reject(err)
        }

        resolve(res)
      })
    })
  }

  const cubeApiRef = useRef<CubeApi>(cubeApi)

  return (
    <RimdianCubeContext.Provider
      value={{
        cubeApi: cubeApiRef.current
      }}
    >
      {props.children}
    </RimdianCubeContext.Provider>
  )
}
