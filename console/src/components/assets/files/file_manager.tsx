import {
  Alert,
  Button,
  Form,
  Input,
  Modal,
  Popconfirm,
  Popover,
  Space,
  Table,
  Tooltip,
  message
} from 'antd'
import { FileManagerProps, StorageObject } from './interfaces'
import CSS, { colorPrimary } from 'utils/css'
import Block from 'components/common/block'
import { ChangeEvent, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faCopy, faFolder, faTrashAlt, faTrashCan } from '@fortawesome/free-regular-svg-icons'
import {
  faArrowUpFromBracket,
  faArrowUpRightFromSquare,
  faCog,
  faRefresh
} from '@fortawesome/free-solid-svg-icons'
import { css } from '@emotion/css'
import dayjs from 'dayjs'
import filesize from 'filesize'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import ButtonFilesSettings from './button_settings'
import {
  S3Client,
  ListObjectsV2Command,
  ListObjectsV2CommandInput,
  PutObjectCommand,
  PutObjectCommandInput,
  DeleteObjectCommand,
  DeleteObjectCommandInput
} from '@aws-sdk/client-s3'
import GetContentType from 'utils/file_extension'

const folderRow = css({
  fontWeight: 'bold',
  cursor: 'pointer'
})

const filesContainer = css({
  position: 'relative',
  overflow: 'auto',
  paddingBottom: '40px'
})

export const FileManager = (props: FileManagerProps) => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [currentPath, setCurrentPath] = useState('')
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])
  const [items, setItems] = useState<StorageObject[] | undefined>(undefined)
  const [isLoading, setIsLoading] = useState(false)
  const [newFolderModalVisible, setNewFolderModalVisible] = useState(false)
  const [newFolderLoading, setNewFolderLoading] = useState(false)
  const s3ClientRef = useRef<S3Client | undefined>(undefined)
  const inputFileRef = useRef<HTMLInputElement>(null)
  const [isUploading, setIsUploading] = useState(false)
  const [form] = Form.useForm()

  const goToPath = (path: string) => {
    // reset selection on path change
    setSelectedRowKeys([])
    props.onSelect([])
    setCurrentPath(path)
  }

  const fetchObjects = useCallback(() => {
    if (!s3ClientRef.current) return

    setIsLoading(true)
    const input: ListObjectsV2CommandInput = {
      Bucket: workspaceCtx.workspace.files_settings.bucket
    }

    const command = new ListObjectsV2Command(input)
    s3ClientRef.current.send(command).then((response) => {
      // console.log('response', response)
      if (!response.Contents) {
        setItems([])
        setIsLoading(false)
        return
      }

      const newItems = response.Contents.map((x) => {
        const key = x.Key as string
        let endpoint = workspaceCtx.workspace.files_settings.endpoint

        if (workspaceCtx.workspace.files_settings.cdn_endpoint !== '') {
          endpoint = workspaceCtx.workspace.files_settings.cdn_endpoint
        }

        const isFolder = key.endsWith('/')
        let name =
          key
            .split('/')
            .filter((x) => x !== '')
            .pop() || ''

        if (!isFolder) {
          name = key.split('/').pop() || ''
        }

        // console.log('item', x)

        let itemPath = ''
        const pathParts = key.split('/')

        if (isFolder) {
          itemPath = pathParts.slice(0, pathParts.length - 2).join('/') + '/'
          // console.log('folder path', itemCurrentPath)
        } else {
          itemPath = pathParts.slice(0, pathParts.length - 1).join('/') + '/'
          // console.log('file path', itemCurrentPath)
        }

        if (itemPath === '/') itemPath = ''

        const item = {
          key: key,
          name: name,
          path: itemPath,
          is_folder: isFolder,
          last_modified: x.LastModified
        } as StorageObject

        if (!isFolder) {
          item.file_info = {
            size: x.Size as number,
            size_human: filesize(x.Size || 0, { round: 0 }),
            content_type: GetContentType(key),
            url: endpoint + '/' + key
          }
        }

        return item
      })

      // console.log('new items', newItems)
      setItems(newItems)
      setIsLoading(false)
    })
  }, [
    workspaceCtx.workspace.files_settings.bucket,
    workspaceCtx.workspace.files_settings.cdn_endpoint,
    workspaceCtx.workspace.files_settings.endpoint
  ])

  // init
  useEffect(() => {
    if (workspaceCtx.workspace.files_settings.endpoint === '') {
      return
    }
    if (s3ClientRef.current) return

    s3ClientRef.current = new S3Client({
      endpoint: workspaceCtx.workspace.files_settings.endpoint,
      credentials: {
        accessKeyId: workspaceCtx.workspace.files_settings.access_key,
        secretAccessKey: workspaceCtx.workspace.files_settings.secret_key
      },
      region: workspaceCtx.workspace.files_settings.region || 'us-east-1'
    })

    fetchObjects()
  }, [
    workspaceCtx.workspace.files_settings.endpoint,
    workspaceCtx.workspace.files_settings.access_key,
    workspaceCtx.workspace.files_settings.secret_key,
    workspaceCtx.workspace.files_settings.region,
    fetchObjects
  ])

  const deleteObject = (key: string, isFolder: boolean) => {
    if (!s3ClientRef.current) {
      message.error('S3 client is not initialized.')
      return
    }

    const s3Client = s3ClientRef.current

    const input: DeleteObjectCommandInput = {
      Bucket: workspaceCtx.workspace.files_settings.bucket,
      Key: key
    }

    s3Client
      .send(new DeleteObjectCommand(input))
      .then(() => {
        if (isFolder) {
          fetchObjects()
          message.success('Folder deleted successfully.')
          // go to previous path
          setCurrentPath(key.split('/').slice(0, -2).join('/') + '/')
        } else {
          message.success('File deleted successfully.')
        }
      })
      .catch((error) => {
        message.error('Failed to delete file: ' + error)
      })
  }

  const selectItem = (items: StorageObject[]) => {
    console.log('selected items', items)
  }

  const toggleSelectionForItem = (item: StorageObject) => {
    // ignore items not accepted
    if (!props.acceptItem(item)) return

    if (props.multiple) {
      let newKeys = [...selectedRowKeys]
      // remove if exists
      if (newKeys.includes(item.key)) {
        newKeys = selectedRowKeys.filter((k) => k !== item.key)
      } else {
        newKeys.push(item.key)
      }
      setSelectedRowKeys(newKeys)
      props.onSelect(items ? items.filter((x) => newKeys.includes(x.key)) : [])
    } else {
      setSelectedRowKeys([item.key])
      props.onSelect([item])
    }
  }

  const toggleNewFolderModal = () => {
    setNewFolderModalVisible(!newFolderModalVisible)
  }

  const onSubmitNewFolder = () => {
    if (!s3ClientRef.current) {
      message.error('S3 client is not initialized.')
      return
    }

    if (newFolderLoading) return

    const s3Client = s3ClientRef.current

    form.validateFields().then((values) => {
      setNewFolderLoading(true)

      // create folder in S3
      const folderName = values.name
      const key = currentPath === '' ? folderName + '/' : currentPath + folderName + '/'

      const input: ListObjectsV2CommandInput = {
        Bucket: workspaceCtx.workspace.files_settings.bucket,
        Prefix: key
      }

      s3Client
        .send(new ListObjectsV2Command(input))
        .then((response) => {
          // console.log('response', response)
          if (response.Contents && response.Contents.length > 0) {
            message.error('Folder already exists.')
            return
          }

          const input: PutObjectCommandInput = {
            Bucket: workspaceCtx.workspace.files_settings.bucket,
            Key: key,
            Body: ''
          }

          s3Client
            .send(new PutObjectCommand(input))
            .then(() => {
              message.success('Folder created successfully.')
              setNewFolderLoading(false)
              fetchObjects()
            })
            .catch((error) => {
              message.error('Failed to create folder: ' + error)
              setNewFolderLoading(false)
            })
        })
        .catch((error) => {
          message.error('Failed to create folder: ' + error)
          setNewFolderLoading(false)
        })

      form.resetFields()
      toggleNewFolderModal()
    })
  }

  const itemsAtPath = useMemo(() => {
    if (!items) return []
    return items
      .filter((x) => x.path === currentPath)
      .sort((a, b) => {
        // by folders first, then by last_modified
        if (a.is_folder && !b.is_folder) return -1
        if (!a.is_folder && b.is_folder) return 1
        if (a.last_modified > b.last_modified) return -1
        if (a.last_modified < b.last_modified) return 1
        return 0
      })
  }, [items, currentPath])

  const onFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return
    if (isUploading) return
    if (!s3ClientRef.current) return

    // console.log(e.target.files)

    for (var i = 0; i < e.target.files.length; i++) {
      setIsUploading(true)
      const file = e.target.files.item(i) as File

      s3ClientRef.current
        .send(
          new PutObjectCommand({
            Bucket: workspaceCtx.workspace.files_settings.bucket,
            Key: currentPath + file.name,
            Body: file,
            ContentType: file.type
          })
        )
        .then(() => {
          message.success('File' + file.name + ' uploaded successfully.')
          setIsUploading(false)
          fetchObjects()
        })
        .catch((error) => {
          message.error('Failed to upload file: ' + error)
          setIsUploading(false)
        })
    }
  }

  const onBrowseFiles = () => {
    if (inputFileRef.current) {
      inputFileRef.current.click()
    }
  }

  if (workspaceCtx.workspace.files_settings.endpoint === '') {
    return (
      <Alert
        className={CSS.margin_b_l}
        message={
          <>
            File storage is not configured.
            <ButtonFilesSettings>
              <Button type="link">Configure now</Button>
            </ButtonFilesSettings>
          </>
        }
        type="warning"
        showIcon
      />
    )
  }

  return (
    <Block classNames={[filesContainer]} style={{ height: props.height }}>
      {workspaceCtx.workspace.files_settings.endpoint !== '' && (
        <>
          <div className={CSS.padding_a_m} style={{ borderBottom: '1px solid rgba(0,0,0,0.1)' }}>
            <div className={CSS.pull_right}>
              <Space>
                {currentPath !== '' && (
                  <Tooltip title="Delete folder" placement="bottom">
                    <Popconfirm
                      placement="topRight"
                      title={
                        <>
                          Do you want to delete the <b>{currentPath}</b> folder with all its
                          content?
                        </>
                      }
                      onConfirm={() => deleteObject(currentPath, true)}
                      okText="Delete folder"
                      cancelText="Cancel"
                      okButtonProps={{
                        danger: true
                      }}
                    >
                      <Button
                        size="small"
                        type="text"
                        onClick={() => fetchObjects()}
                        icon={<FontAwesomeIcon icon={faTrashAlt} />}
                      />
                    </Popconfirm>
                  </Tooltip>
                )}
                <Tooltip title="Refresh the list">
                  <Button
                    size="small"
                    type="text"
                    onClick={() => fetchObjects()}
                    icon={<FontAwesomeIcon icon={faRefresh} />}
                  />
                </Tooltip>

                <ButtonFilesSettings>
                  <Tooltip title="Storage settings">
                    <Button type="text" size="small">
                      <FontAwesomeIcon icon={faCog} />
                    </Button>
                  </Tooltip>
                </ButtonFilesSettings>
                <span role="button" onClick={onBrowseFiles}>
                  <input
                    type="file"
                    ref={inputFileRef}
                    onChange={onFileChange}
                    hidden
                    accept={props.acceptFileType}
                    multiple={false}
                  />
                  <Button
                    type="primary"
                    size="small"
                    className={CSS.pull_right}
                    loading={isUploading}
                  >
                    <FontAwesomeIcon icon={faArrowUpFromBracket} className={CSS.padding_r_s} />
                    Upload
                  </Button>
                </span>
              </Space>
            </div>

            <Space>
              <div>
                <Button type="text" size="small" onClick={() => goToPath('')}>
                  {workspaceCtx.workspace.files_settings.bucket}
                </Button>
                {currentPath
                  .split('/')
                  .filter((x) => x !== '')
                  .map((part, index, array) => {
                    const isLast = index === array.length - 1
                    const fullPath = array.slice(0, index + 1).join('/') + '/'
                    return (
                      <>
                        /
                        <Button
                          disabled={isLast}
                          type="text"
                          size="small"
                          onClick={() => goToPath(fullPath)}
                        >
                          {part}
                        </Button>
                      </>
                    )
                  })}
              </div>
              <Button type="primary" size="small" ghost onClick={toggleNewFolderModal}>
                New folder
              </Button>
            </Space>
          </div>
          <Table
            dataSource={itemsAtPath}
            loading={isLoading}
            pagination={false}
            size="middle"
            // scroll={{ y: props.height - 80 }}
            rowKey="key"
            locale={{ emptyText: 'Folder is empty' }}
            rowClassName={(record: StorageObject) => {
              return record.is_folder ? folderRow : ''
            }}
            onRow={(record: StorageObject) => {
              return {
                onClick: () => {
                  if (record.is_folder) {
                    setCurrentPath(record.key)
                  }
                }
              }
            }}
            rowSelection={
              props.withSelection
                ? {
                    type: props.multiple ? 'checkbox' : 'radio',
                    selectedRowKeys: selectedRowKeys,
                    onChange: (selectedRowKeys: React.Key[], selectedRows: any[]) => {
                      // console.log(`selectedRowKeys: ${selectedRowKeys}`, 'selectedRows: ', selectedRows);
                      setSelectedRowKeys(selectedRowKeys)
                      selectItem(selectedRows)
                    },
                    getCheckboxProps: (record: any) => ({
                      disabled: !props.acceptItem(record as StorageObject)
                      // name: record.name,
                    })
                  }
                : undefined
            }
            columns={[
              {
                title: '',
                key: 'preview',
                width: 70,
                render: (item: StorageObject) => {
                  if (item.is_folder) {
                    return (
                      <div onClick={toggleSelectionForItem.bind(null, item)}>
                        <FontAwesomeIcon
                          icon={faFolder}
                          className={CSS.font_size_m}
                          style={{ color: colorPrimary }}
                        />
                      </div>
                    )
                  }
                  return (
                    <div onClick={toggleSelectionForItem.bind(null, item)}>
                      {item.file_info.content_type.includes('image') && (
                        <Popover
                          placement="right"
                          content={<img src={item.file_info.url} alt="" height="400" />}
                        >
                          <img src={item.file_info.url} alt="" height="30" />
                        </Popover>
                      )}
                    </div>
                  )
                }
              },
              {
                title: 'Name',
                key: 'name',
                render: (item: StorageObject) => {
                  return <div onClick={toggleSelectionForItem.bind(null, item)}>{item.name}</div>
                }
              },
              {
                title: 'Size',
                key: 'size',
                width: 100,
                render: (item: StorageObject) => {
                  return (
                    <div onClick={toggleSelectionForItem.bind(null, item)}>
                      {item.is_folder ? '-' : item.file_info.size_human}
                    </div>
                  )
                }
              },
              {
                title: 'Last modified',
                key: 'lastModified',
                width: 120,
                render: (item: StorageObject) => {
                  return (
                    <Tooltip title={dayjs(item.last_modified).format('llll')}>
                      <div onClick={toggleSelectionForItem.bind(null, item)}>
                        {dayjs(item.last_modified).format('ll')}
                      </div>
                    </Tooltip>
                  )
                }
              },
              {
                title: '',
                key: 'actions',
                width: 40,
                className: CSS.text_right,
                render: (item: StorageObject) => {
                  if (item.is_folder) return
                  return (
                    <Space>
                      <Tooltip title="Copy URL">
                        <Button
                          type="text"
                          size="small"
                          onClick={() => {
                            navigator.clipboard.writeText(item.file_info.url)
                            message.success('URL copied to clipboard.')
                          }}
                        >
                          <FontAwesomeIcon icon={faCopy} />
                        </Button>
                      </Tooltip>
                      <Tooltip title="Open in a window">
                        <a href={item.file_info.url} target="_blank" rel="noreferrer">
                          <Button type="text" size="small">
                            <FontAwesomeIcon icon={faArrowUpRightFromSquare} />
                          </Button>
                        </a>
                      </Tooltip>
                      <Popconfirm
                        title="Do you want to permanently delete this file from your storage?"
                        onConfirm={() => deleteObject(item.key, false)}
                        placement="topRight"
                        okText="Delete"
                        cancelText="Cancel"
                        okButtonProps={{
                          danger: true
                        }}
                      >
                        <Button type="text" size="small">
                          <FontAwesomeIcon icon={faTrashCan} />
                        </Button>
                      </Popconfirm>
                    </Space>
                  )
                }
              }
            ]}
          />
        </>
      )}
      {newFolderModalVisible && (
        <Modal
          title="Create new folder"
          open={newFolderModalVisible}
          onOk={onSubmitNewFolder}
          onCancel={toggleNewFolderModal}
          confirmLoading={newFolderLoading}
        >
          <Form form={form}>
            <Form.Item
              label="Folder name"
              name="name"
              rules={[
                {
                  required: true,
                  type: 'string',
                  validator(_rule, value, callback) {
                    // alphanumeric, lowercase, underscore, dash
                    if (!/^[a-z0-9_-]+$/.test(value)) {
                      callback(
                        'Only lowercase alphanumeric characters, underscore, and dash are allowed.'
                      )
                      return
                    }
                    callback()
                  }
                }
              ]}
            >
              <Input
                addonBefore={currentPath !== '/' ? currentPath : '/'}
                onChange={(e) => {
                  // trim spaces
                  form.setFieldsValue({ folderName: e.target.value.trim() })
                }}
              />
            </Form.Item>
          </Form>
        </Modal>
      )}
    </Block>
  )
}
