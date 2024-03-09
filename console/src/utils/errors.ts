import { message } from 'antd'

export const HandleAxiosError = (e: any) => {
  message.error(e.response?.data?.message || e.message)
  console.error(e)
}

export const HandleCubeError = (e: any) => {
  message.error(e.response?.data?.message || e.message)
  console.error(e)
}
