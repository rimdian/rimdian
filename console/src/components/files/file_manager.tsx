import {
  Alert,
  Button,
  Col,
  Dropdown,
  Menu,
  Modal,
  Popover,
  Row,
  Space,
  Table,
  Tooltip,
  message
} from 'antd'
import { FileManagerProps, Item } from './interfaces'
import CSS from 'utils/css'
import Block from 'components/common/block'
import { ReactNode, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faFolder, faFolderOpen, faTrashAlt, faTrashCan } from '@fortawesome/free-regular-svg-icons'
import {
  faArrowUpFromBracket,
  faArrowUpRightFromSquare,
  faEllipsisVertical,
  faRefresh
} from '@fortawesome/free-solid-svg-icons'
import { css } from '@emotion/css'
import dayjs from 'dayjs'
import filesize from 'filesize'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import ButtonFilesSettings from './button_settings'
import { S3Client, ListBucketsCommand } from '@aws-sdk/client-s3'

const folderItem = css({
  margin: '8px',
  padding: '8px 0',
  cursor: 'pointer',
  borderRadius: '4px',
  '&:hover, &.selected': {
    backgroundColor: '#f0f4ff'
  }
})

const folderIcon = css({
  padding: '0px 8px',
  color: '#4e6cff'
})

const itemIcon = css({
  padding: '4px',
  lineHeight: '12px',
  borderRadius: '50%'
})

const actionIcon = css({
  padding: '4px 6px',
  color: 'inherit !important',
  '&:hover': {
    color: '#4e6cff',
    cursor: 'pointer'
  }
})

const filesContainer = css({
  position: 'relative',
  overflow: 'auto',
  paddingBottom: '40px'
})

const bottomToolbar = css({
  position: 'absolute',
  bottom: 0,
  left: 0,
  right: 0,
  padding: '16px 16px',
  textAlign: 'right'
})

export const FileManager = (props: FileManagerProps) => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [selectedPath, setSelectedPath] = useState(props.currentPath || '/')
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])
  const [items, setItems] = useState<Item[] | undefined>(undefined)
  const [isLoading, setIsLoading] = useState(false)
  const s3ClientRef = useRef<S3Client | undefined>(undefined)

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
      region: 'REGION'
    })
  }, [workspaceCtx.workspace.files_settings])

  const goToFolder = useCallback((path: string) => {
    console.log('go to folder', path)
    setSelectedPath(path)
  }, [])

  const onDelete = () => {
    console.log('delete')
  }

  const selectItem = (items: Item[]) => {
    console.log('selected items', items)
  }

  const deleteItem = (item: Item) => {
    console.log('delete item', item)
    return Promise.resolve()
  }

  const refresh = () => {
    console.log('refresh')
  }

  const toggleSelectionForItem = (item: Item) => {
    // ignore items not accepted
    if (!props.acceptItem(item)) return

    if (props.multiple) {
      let newKeys = [...selectedRowKeys]
      // remove if exists
      if (newKeys.includes(item.id)) {
        newKeys = selectedRowKeys.filter((k) => k !== item.id)
      } else {
        newKeys.push(item.id)
      }
      setSelectedRowKeys(newKeys)
      props.onSelect(items ? items.filter((x) => newKeys.includes(x.id)) : [])
    } else {
      setSelectedRowKeys([item.id])
      props.onSelect([item])
    }
  }

  return (
    <Block classNames={[filesContainer]} style={{ height: props.height }}>
      {workspaceCtx.workspace.files_settings.endpoint === '' && (
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
      )}
      {workspaceCtx.workspace.files_settings.endpoint !== '' && (
        <>
          <Table
            dataSource={items}
            loading={isLoading}
            pagination={false}
            // scroll={{ y: props.height - 80 }}
            size="small"
            rowKey="id"
            locale={{ emptyText: 'Folder is empty' }}
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
                      disabled: !props.acceptItem(record as Item)
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
                render: (item) => (
                  <div onClick={toggleSelectionForItem.bind(null, item)}>
                    {isImage(item.contentType) && (
                      <Popover
                        placement="right"
                        content={<img src={item.url} alt="" height="400" />}
                      >
                        <img src={item.url} alt="" height="30" />
                      </Popover>
                    )}
                  </div>
                )
              },
              {
                title: 'Name',
                key: 'name',
                render: (item) => {
                  return (
                    <div
                      className={CSS.font_size_xs}
                      onClick={toggleSelectionForItem.bind(null, item)}
                    >
                      {item.name}
                    </div>
                  )
                }
              },
              {
                title: 'Size',
                key: 'size',
                width: 100,
                render: (item) => {
                  return (
                    <div
                      className={CSS.font_size_xs}
                      onClick={toggleSelectionForItem.bind(null, item)}
                    >
                      {filesize(item.size, { round: 0 })}
                    </div>
                  )
                }
              },
              {
                title: 'Uploaded at',
                key: 'uploaded',
                width: 120,
                render: (item) => {
                  return (
                    <Tooltip title={dayjs.unix(item.uploadedAt).format('llll')}>
                      <div
                        className={CSS.font_size_xs}
                        onClick={toggleSelectionForItem.bind(null, item)}
                      >
                        {dayjs.unix(item.uploadedAt).format('ll')}
                      </div>
                    </Tooltip>
                  )
                }
              },
              {
                title: 'Last modified',
                key: 'lastModified',
                width: 120,
                render: (item) => {
                  return (
                    <Tooltip title={dayjs.unix(item.lastModifiedAt).format('llll')}>
                      <div
                        className={CSS.font_size_xs}
                        onClick={toggleSelectionForItem.bind(null, item)}
                      >
                        {dayjs.unix(item.lastModifiedAt).format('ll')}
                      </div>
                    </Tooltip>
                  )
                }
              },
              {
                title: (
                  <Tooltip title="Refresh the list">
                    <Button
                      size="small"
                      type="text"
                      onClick={() => refresh()}
                      icon={<FontAwesomeIcon icon={faRefresh} />}
                    />
                  </Tooltip>
                ),
                key: 'actions',
                width: 40,
                render: (item) => {
                  return (
                    <div style={{ textAlign: 'right' }}>
                      <ItemMenu item={item} deleteItem={deleteItem} />
                    </div>
                  )
                }
              }
            ]}
          />
        </>
      )}

      <div className={bottomToolbar}>
        <Space>
          <Button
            type="primary"
            ghost
            onClick={() => {
              console.log('create folder')
            }}
          >
            New folder
          </Button>
          <Button
            type="primary"
            onClick={() => {
              console.log('upload')
            }}
          >
            <FontAwesomeIcon icon={faArrowUpFromBracket} className={CSS.padding_r_s} />
            Upload
          </Button>
        </Space>
      </div>
    </Block>
  )
}

const isImage = (contentType: string): boolean => {
  return contentType.includes('image')
}

export interface ItemMenuProps {
  // fileProvider: FileProvider
  item: Item
  deleteItem: (node: Item) => Promise<any>
}

const ItemMenu = (props: ItemMenuProps) => {
  const [loading, setLoading] = useState(false)

  const onDelete = () => {
    Modal.confirm({
      title: 'Do you want to delete this item?',
      //   icon: <ExclamationCircleOutlined />,
      content:
        'Deleting "' + props.item.name + '" will remove it from the list but keep it online.',
      okButtonProps: { danger: true, loading: loading },
      cancelButtonProps: { loading: loading },
      okText: 'Delete',
      closable: !loading,
      onOk() {
        setLoading(true)
        props
          .deleteItem(props.item)
          .then(() => {
            setLoading(false)
          })
          .catch((e) => {
            setLoading(false)
            message.error(e)
          })
      }
    })
  }

  return (
    <span onClick={(e: any) => e.stopPropagation()}>
      <Dropdown
        className={itemIcon}
        trigger={['click']}
        overlay={
          <Menu>
            <Menu.Item icon={<FontAwesomeIcon icon={faArrowUpRightFromSquare} />}>
              <a href={props.item.url} target="_blank" rel="noreferrer" className={actionIcon}>
                Open in a window
              </a>
            </Menu.Item>
            <Menu.Item icon={<FontAwesomeIcon icon={faTrashCan} />} danger onClick={onDelete}>
              Delete
            </Menu.Item>
          </Menu>
        }
      >
        <FontAwesomeIcon icon={faEllipsisVertical} />
      </Dropdown>
    </span>
  )
}
