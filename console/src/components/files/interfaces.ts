export interface FileInfo {
  size: number
  size_human: string
  content_type: string // guessed from the file extension
  url: string
}

export interface StorageObject {
  key: string
  name: string
  is_folder: boolean
  path: string
  last_modified: Date
  file_info: FileInfo
}

export interface FileManagerProps {
  currentPath?: string
  itemFilters?: ItemFilter[]
  onError: (error: any) => void
  onSelect: (items: StorageObject[]) => void
  height: number
  acceptFileType: string
  acceptItem: (item: StorageObject) => boolean
  withSelection?: boolean
  multiple?: boolean
}

export interface FilesSettings {
  endpoint: string
  access_key: string
  encrypted_secret_key: string
  secret_key: string
  bucket: string
  region: string
  cdn_endpoint: string
}

export interface ItemFilter {
  key: string // item key
  value: any
  operator: string // contains equals greaterThan lessThan
}
