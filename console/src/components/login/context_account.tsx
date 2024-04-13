import { useState, useContext, createContext, useEffect, useCallback } from 'react'
import Axios from 'axios'
import {
  AccountLoginResult,
  AccountRefreshAccessTokenResult,
  Account,
  DataLogBatch
} from 'interfaces'
import { HandleAxiosError } from 'utils/errors'
import dayjs from 'dayjs'
import numbro from 'numbro'
const frFR = require('numbro/languages/fr-FR')
numbro.registerLanguage(frFR)

const AccountContext = createContext<AccountContextValue | null>(null)
const ADMIN_KEY = 'account'

export function useAccount(): AccountContextValue {
  const contextValue = useContext(AccountContext)
  if (!contextValue) {
    throw new Error('Missing AccountContextProvider in its parent.')
  }
  return contextValue
}

type Props = {
  children?: React.ReactNode
}

export interface AccountContextValue {
  account?: AccountLoginResult
  initializing: boolean
  login: (account: AccountLoginResult) => void
  logout: (accountToken: AccountLoginResult) => Promise<void>
  updateAccountProfile: (values: Account) => void
  apiGET: (endpoint: string) => Promise<any>
  apiPOST: (endpoint: string, data: any) => Promise<any>
  collectorPOST: (sync: boolean, batch: DataLogBatch) => Promise<any>
}

export const AuthProvider = (props: Props) => {
  const [account, setAccount] = useState<AccountLoginResult | undefined>(undefined)
  const [initializing, setInitializing] = useState(true)

  const login = useCallback((account: AccountLoginResult) => {
    // store data in localStorage
    window.localStorage.setItem(ADMIN_KEY, JSON.stringify(account))
    numbro.setLanguage(account.account.locale, 'en-US')
    setAccount(account)
  }, [])

  const logout = (accountToken: AccountLoginResult): Promise<void> => {
    return new Promise<void>((resolve, reject) => {
      Axios.post(
        window.Config.API_ENDPOINT + '/account.logout',
        {},
        {
          headers: { Authorization: 'Bearer ' + accountToken.refresh_token }
        }
      )
        .then((res) => {
          setAccount(undefined)
          window.localStorage.removeItem(ADMIN_KEY)
          resolve()
        })
        .catch((e) => {
          HandleAxiosError(e)
          reject(e)
        })
    })
  }

  const refreshAccessToken = useCallback(
    (accountToken: AccountLoginResult): Promise<void> => {
      return new Promise<void>((resolve, reject) => {
        Axios.post(
          window.Config.API_ENDPOINT + '/account.refreshAccessToken',
          {},
          {
            headers: { Authorization: 'Bearer ' + accountToken.refresh_token }
          }
        )
          .then((res) => {
            // console.log('res', res)
            const result = res.data as AccountRefreshAccessTokenResult

            accountToken.access_token = result.access_token
            accountToken.access_token_expires_at = result.access_token_expires_at

            login(accountToken)
            resolve()
          })
          .catch((e) => {
            HandleAxiosError(e)
            reject(e)
          })
      })
    },
    [login]
  )

  const updateAccountProfile = (values: Account) => {
    if (!account?.account) {
      return
    }

    const newAccountProfile = Object.assign({}, { ...account?.account }, values)

    const newAccountLogin = { ...account }
    newAccountLogin.account = newAccountProfile

    window.localStorage.setItem(ADMIN_KEY, JSON.stringify(newAccountLogin))
    setAccount(newAccountLogin)
  }

  // resume account from eventual stored token
  useEffect(() => {
    // retrieve existing token from local storage
    const storedToken = window.localStorage.getItem(ADMIN_KEY)
    // console.log('loading stored token', storedToken)
    if (!storedToken) {
      setInitializing(false)
      return
    }

    const accountToken: AccountLoginResult = JSON.parse(storedToken)

    // abort if refresh token expired
    if (dayjs(accountToken.refresh_token_expires_at).isBefore(dayjs())) {
      window.localStorage.removeItem(ADMIN_KEY)
      setInitializing(false)
      return
    }

    // check if access token is expiring in less than 10 mins
    const accessTokenExpiresIn = dayjs(accountToken.access_token_expires_at).diff(
      dayjs(),
      'minutes'
    )
    // console.log('accessTokenExpiresIn', accessTokenExpiresIn)

    if (accessTokenExpiresIn > 10) {
      login(accountToken)
      setInitializing(false)
      return
    }

    // refresh access token now
    refreshAccessToken(accountToken)
      .then(() => setInitializing(false))
      .catch((_e) => setInitializing(false))
  }, [refreshAccessToken, login])

  // automatically refresh existing access token very 10 mins
  useEffect((): any => {
    const intervalID = window.setInterval(() => {
      const storedToken = window.localStorage.getItem(ADMIN_KEY)
      if (storedToken) {
        const accountToken: AccountLoginResult = JSON.parse(storedToken)

        // refresh if refresh token is still valid
        if (dayjs(accountToken.refresh_token_expires_at).isAfter(dayjs())) {
          refreshAccessToken(accountToken)
        }
      }
    }, 1000 * 60 * 10) // every 10 mins

    return () => {
      window.clearInterval(intervalID)
    }
  }, [refreshAccessToken])

  const apiGET = (endpoint: string): Promise<any> => {
    return new Promise((resolve, reject) => {
      if (!account) {
        return reject('Account not logged in')
      }

      Axios.get(window.Config.API_ENDPOINT + endpoint, {
        headers: { Authorization: 'Bearer ' + account.access_token }
      })
        .then((res) => {
          resolve(res.data)
        })
        .catch((e) => {
          HandleAxiosError(e)
          reject(e)
        })
    })
  }

  const collectorPOST = (sync: boolean, dataImport: DataLogBatch): Promise<any> => {
    return new Promise((resolve, reject) => {
      if (!account) {
        return reject('Account not logged in')
      }

      Axios.post(window.Config.COLLECTOR_ENDPOINT + (sync ? '/sync' : 'data'), dataImport, {
        headers: { Authorization: 'Bearer ' + account.access_token }
      })
        .then((res) => {
          resolve(res.data)
        })
        .catch((e) => {
          HandleAxiosError(e)
          reject(e)
        })
    })
  }

  const apiPOST = (endpoint: string, data: any): Promise<any> => {
    return new Promise((resolve, reject) => {
      if (!account) {
        return reject('Account not logged in')
      }

      Axios.post(window.Config.API_ENDPOINT + endpoint, data, {
        headers: { Authorization: 'Bearer ' + account.access_token }
      })
        .then((res) => {
          resolve(res.data)
        })
        .catch((e) => {
          HandleAxiosError(e)
          reject(e)
        })
    })
  }

  return (
    <AccountContext.Provider
      value={{
        initializing,
        account,
        login,
        logout,
        updateAccountProfile,
        apiGET,
        apiPOST,
        collectorPOST
      }}
    >
      {props.children}
    </AccountContext.Provider>
  )
}
