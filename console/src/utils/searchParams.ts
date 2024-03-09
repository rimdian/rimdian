import { forEach } from 'lodash'

export const cleanParams = (params: object): object => {
  const cleaned: any = {}
  forEach(params, (val, key) => {
    if (val !== undefined) {
      cleaned[key] = val
    }
  })
  return cleaned
}

export const paramsToQueryString = (params: object): string => {
  return new URLSearchParams(cleanParams(params) as any).toString()
}
