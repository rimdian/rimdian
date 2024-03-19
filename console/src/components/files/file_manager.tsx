import { Col, Dropdown, Menu, Modal, Popover, Row, Table, Tooltip, message } from 'antd'
import { FileManagerProps, FoldersTree, Item } from './interfaces'
import CSS from 'utils/css'
import Block from 'components/common/block'
import { ReactNode, useCallback, useMemo, useState } from 'react'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faFolder, faFolderOpen, faTrashAlt, faTrashCan } from '@fortawesome/free-regular-svg-icons'
import { faArrowUpRightFromSquare, faEllipsisVertical } from '@fortawesome/free-solid-svg-icons'
import { css } from '@emotion/css'
import dayjs from 'dayjs'
import filesize from 'filesize'

const folderItem = css({
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

export const FileManager = (props: FileManagerProps) => {
  const [selectedPath, setSelectedPath] = useState(props.currentPath || '/')

  const goToFolder = useCallback((path: string) => {
    console.log('go to folder', path)
    setSelectedPath(path)
  }, [])

  const onDelete = () => {
    console.log('delete')
  }

  const renderTree = useCallback(
    (folder: FoldersTree, selectedPath: string, level: number): ReactNode => {
      const selected = folder.path === selectedPath

      return (
        <div key={folder.path}>
          <div
            onClick={goToFolder.bind(null, folder.path)}
            className={folderItem + (selected ? ' selected' : '')}
            style={{ paddingLeft: 16 * level + 'px' }}
          >
            {level > 0 && (
              <Dropdown
                trigger={['click']}
                overlay={
                  <Menu>
                    <Menu.Item
                      icon={<FontAwesomeIcon icon={faTrashAlt} />}
                      danger
                      onClick={onDelete}
                    >
                      Delete
                    </Menu.Item>
                  </Menu>
                }
              >
                <FontAwesomeIcon icon={faEllipsisVertical} />
              </Dropdown>
            )}
            <span className={folderIcon}>
              {selected ? (
                <FontAwesomeIcon icon={faFolderOpen} />
              ) : (
                <FontAwesomeIcon icon={faFolder} />
              )}
            </span>{' '}
            {folder.name}
          </div>

          {folder.children.map((sub) => renderTree(sub, selectedPath, level + 1))}
        </div>
      )
    },
    [goToFolder]
  )

  const folders = useMemo(() => {
    return renderTree(props.foldersTree, selectedPath, 0)
  }, [props.foldersTree, selectedPath, renderTree])

  const findCurrentFolder = useCallback((f: FoldersTree, path: string): FoldersTree | undefined => {
    if (f.path === path) {
      return f
    }

    return f.children.find((child) => {
      return findCurrentFolder(child, path)
    })
  }, [])

  const currentFolder = useMemo(() => {
    return findCurrentFolder(props.foldersTree, selectedPath)
  }, [props.foldersTree, selectedPath, findCurrentFolder])

  return (
    <div>
      <Row gutter={24}>
        <Col span={8}>
          <Block classNames={[CSS.padding_a_m]}>
            <div style={{ height: props.height, overflow: 'auto' }}>{folders}</div>
          </Block>
        </Col>
        <Col span={16}>
          <Block classNames={[CSS.padding_a_m]}>
            <div style={{ height: props.height - 40, overflow: 'auto' }}>
              <RenderFiles
                folder={currentFolder}
                acceptFileType={props.acceptFileType}
                acceptItem={props.acceptItem}
                onSelect={props.onSelect}
              />
            </div>
          </Block>
        </Col>
      </Row>
    </div>
  )
}

type RenderFilesProps = {
  folder?: FoldersTree
  withSelection?: boolean
  multiple?: boolean
  acceptFileType: string
  acceptItem: (item: Item) => boolean
  onSelect: (items: Item[]) => void
}

const isImage = (contentType: string): boolean => {
  return contentType.includes('image')
}

const RenderFiles = (props: RenderFilesProps) => {
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])

  if (!props.folder) {
    return <div>&larr; Please select a folder</div>
  }

  const items: Item[] = []

  const selectItem = (items: Item[]) => {
    console.log('selected items', items)
  }

  const deleteItem = (item: Item) => {
    console.log('delete item', item)
    return Promise.resolve()
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
      props.onSelect(items.filter((x) => newKeys.includes(x.id)))
    } else {
      setSelectedRowKeys([item.id])
      props.onSelect([item])
    }
  }

  const loading = true

  return (
    <div>
      <Table
        dataSource={items}
        loading={loading}
        pagination={false}
        // scroll={{ y: props.height - 80 }}
        size="small"
        rowKey="id"
        locale={{ emptyText: 'No files found' }}
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
                  <Popover placement="right" content={<img src={item.url} alt="" height="400" />}>
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
                <div className={CSS.font_size_xs} onClick={toggleSelectionForItem.bind(null, item)}>
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
                <div className={CSS.font_size_xs} onClick={toggleSelectionForItem.bind(null, item)}>
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
                    className="fm-table-content"
                    onClick={toggleSelectionForItem.bind(null, item)}
                  >
                    {dayjs.unix(item.lastModifiedAt).format('ll')}
                  </div>
                </Tooltip>
              )
            }
          },
          {
            title: '',
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
    </div>
  )
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
