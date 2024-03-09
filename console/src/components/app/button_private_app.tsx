import { useState } from 'react'
import { Button, message, Drawer, Space, Upload, Form, Input } from 'antd'
import { App, AppManifest, TableColumn } from 'interfaces'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import { useNavigate } from 'react-router-dom'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faInbox, faPlus } from '@fortawesome/free-solid-svg-icons'
import BlockAboutApp from './block_about'
import CSS, { backgroundColorBase } from 'utils/css'
import { css } from '@emotion/css'
import customApp from 'images/custom_app.png'
import Ajv from 'ajv'
import Messages from 'utils/formMessages'

// {
// 	Name:            "id",
// 	Type:            ColumnTypeVarchar,
// 	Size:            Int64Ptr(64),
// 	Description:     StringPtr("ID (sha1 of external_id)"),
// 	IsRequired:      true,
// 	HideInAnalytics: true,
// },
// {
// 	Name:            "external_id",
// 	Type:            ColumnTypeVarchar,
// 	Size:            Int64Ptr(256),
// 	Description:     StringPtr("External ID"),
// 	IsRequired:      true,
// 	HideInAnalytics: true,
// },
// {
// 	Name:        "created_at",
// 	Type:        ColumnTypeDatetime,
// 	Description: StringPtr("Created at"),
// 	IsRequired:  true,
// },
// {
// 	Name:            "updated_at",
// 	Type:            ColumnTypeDatetime,
// 	Description:     StringPtr("Field not persisted in DB. Used to merge the most recent fields"),
// 	IsRequired:      false,
// 	HideInAnalytics: true,
// },
// {
// 	Name:            "user_id",
// 	Type:            ColumnTypeVarchar,
// 	Size:            Int64Ptr(64),
// 	Description:     StringPtr("User ID"),
// 	IsRequired:      true,
// 	HideInAnalytics: true,
// },
// {
// 	Name:            "merged_from_user_id",
// 	Type:            ColumnTypeVarchar,
// 	Size:            Int64Ptr(64),
// 	IsRequired:      false,
// 	Description:     StringPtr("Merged from user ID"),
// 	HideInAnalytics: true,
// },
// {
// 	Name:            "fields_timestamp",
// 	Type:            ColumnTypeJSON,
// 	IsRequired:      true,
// 	Description:     StringPtr("Fields timestamp"),
// 	HideInAnalytics: true,
// },
// {
// 	Name:             "db_created_at",
// 	Type:             ColumnTypeTimestamp,
// 	Size:             Int64Ptr(6), // microsecond
// 	Description:      StringPtr("DB created at"),
// 	IsRequired:       true,
// 	DefaultTimestamp: StringPtr("CURRENT_TIMESTAMP(6)"),
// 	HideInAnalytics:  true,
// },
// {
// 	Name:             "db_updated_at",
// 	Type:             ColumnTypeTimestamp,
// 	Description:      StringPtr("DB updated at"),
// 	IsRequired:       true,
// 	DefaultTimestamp: StringPtr("CURRENT_TIMESTAMP"),
// 	ExtraDefinition:  StringPtr("ON UPDATE CURRENT_TIMESTAMP"),
// 	HideInAnalytics:  true,
// },

const reservedColumns = {
  id: {
    name: 'id',
    type: 'varchar',
    size: 64,
    is_required: true,
    description: 'ID (sha1 of external_id)',
    hide_in_analytics: true
  },
  external_id: {
    name: 'external_id',
    type: 'varchar',
    size: 256,
    is_required: true,
    description: 'External ID',
    hide_in_analytics: true
  },
  created_at: {
    name: 'created_at',
    type: 'datetime',
    is_required: true,
    description: 'Created at'
  },
  user_id: {
    name: 'user_id',
    type: 'varchar',
    size: 64,
    is_required: true,
    description: 'User ID',
    hide_in_analytics: true
  },
  merged_from_user_id: {
    name: 'merged_from_user_id',
    type: 'varchar',
    size: 64,
    is_required: false,
    description: 'Merged from user ID',
    hide_in_analytics: true
  },
  fields_timestamp: {
    name: 'fields_timestamp',
    type: 'json',
    is_required: true,
    description: 'Fields timestamp',
    hide_in_analytics: true
  },
  db_created_at: {
    name: 'db_created_at',
    type: 'timestamp',
    size: 6,
    is_required: true,
    default_timestamp: 'CURRENT_TIMESTAMP(6)',
    description: 'DB created at',
    hide_in_analytics: true
  },
  db_updated_at: {
    name: 'db_updated_at',
    type: 'timestamp',
    is_required: true,
    default_timestamp: 'CURRENT_TIMESTAMP',
    extra_definition: 'ON UPDATE CURRENT_TIMESTAMP',
    description: 'DB updated at',
    hide_in_analytics: true
  }
} as any

const tableColumnsSchema = {
  type: 'array',
  items: {
    type: 'object',
    properties: {
      name: { type: 'string' },
      type: { type: 'string' },
      size: { type: 'number', nullable: true },
      is_required: { type: 'boolean' },
      description: { type: 'string', nullable: true },
      default_boolean: { type: 'boolean', nullable: true },
      default_number: { type: 'number', nullable: true },
      default_text: { type: 'string', nullable: true },
      default_date: { type: 'string', nullable: true },
      default_datetime: { type: 'string', nullable: true },
      default_timestamp: { type: 'string', nullable: true },
      default_json: { type: 'string', nullable: true },
      extra_definition: { type: 'string', nullable: true },
      hide_in_analytics: { type: 'boolean' }
    },
    required: ['name', 'type', 'is_required']
  }
}

const appManifestSchema = {
  type: 'object',
  properties: {
    id: { type: 'string' },
    name: { type: 'string' },
    homepage: { type: 'string' },
    author: { type: 'string' },
    icon_url: { type: 'string' },
    short_description: { type: 'string' },
    description: { type: 'string' },
    version: { type: 'string' },
    ui_endpoint: { type: 'string' },
    webhook_endpoint: { type: 'string' },
    app_tables: {
      type: 'array',
      nullable: true,
      items: {
        type: 'object',
        properties: {
          name: { type: 'string' },
          storage_type: { type: 'string' },
          description: { type: 'string' },
          columns: tableColumnsSchema,
          shard_key: {
            type: 'array',
            items: {
              type: 'string'
            }
          },
          unique_key: {
            type: 'array',
            items: {
              type: 'string'
            }
          },
          sort_key: {
            type: 'array',
            items: {
              type: 'string'
            }
          },
          timeseries_column: { type: 'string', nullable: true },
          joins: {
            type: 'array',
            nullable: true,
            items: {
              type: 'object',
              properties: {
                external_table: { type: 'string' },
                relationship: {
                  type: 'string',
                  enum: ['one_to_one', 'one_to_many', 'many_to_one']
                },
                local_column: { type: 'string' },
                external_column: { type: 'string' }
              },
              required: ['external_table', 'relationship', 'local_column', 'external_column']
            }
          },
          indexes: {
            type: 'array',
            nullable: true,
            items: {
              type: 'object',
              properties: {
                name: { type: 'string' },
                columns: {
                  type: 'array',
                  items: {
                    type: 'string'
                  }
                }
              },
              required: ['name', 'columns']
            }
          }
        },
        required: ['name', 'description', 'columns', 'shard_key', 'unique_key', 'sort_key']
      }
    },
    extra_columns: {
      type: 'array',
      nullable: true,
      items: {
        type: 'object',
        properties: {
          kind: { type: 'string' }, // = table name (user, pageview...)
          columns: tableColumnsSchema
        },
        required: ['kind', 'columns']
      }
    },
    data_hooks: {
      type: 'array',
      nullable: true,
      items: {
        type: 'object',
        properties: {
          id: { type: 'string' },
          name: { type: 'string' },
          on: {
            type: 'string',
            enum: ['on_validation', 'on_success']
          },
          kind: {
            type: 'array',
            items: { type: 'string' }
          },
          action: {
            type: 'array',
            items: { type: 'string' }
          }
        },
        required: ['id', 'name', 'on', 'kind', 'action']
      }
    },
    tasks: {
      type: 'array',
      nullable: true,
      items: {
        type: 'object',
        properties: {
          id: { type: 'string' },
          name: { type: 'string' },
          on_multiple_exec: {
            type: 'string',
            enum: ['allow', 'discard_new', 'retry_later', 'abort_existing']
          },
          is_cron: { type: 'boolean' },
          minutes_interval: { type: 'number' }
        },
        required: ['id', 'name', 'on_multiple_exec']
      }
    },
    sql_queries: {
      type: 'array',
      nullable: true,
      items: {
        type: 'object',
        properties: {
          id: { type: 'string' },
          type: { type: 'string', enum: ['select'] },
          name: { type: 'string' },
          description: { type: 'string' },
          query: { type: 'string' },
          test_args: {
            type: 'array',
            nullable: true,
            items: {
              type: ['string', 'number', 'boolean']
            }
          }
        },
        required: ['id', 'type', 'name', 'description', 'query']
      }
    }
  },
  required: [
    'id',
    'name',
    'homepage',
    'author',
    'icon_url',
    'short_description',
    'description',
    'version',
    'ui_endpoint',
    'webhook_endpoint'
  ],
  additionalProperties: false
}

type PrivateAppButtonProps = {
  workspaceCtx: CurrentWorkspaceCtxValue
}

const PrivateAppButton = (props: PrivateAppButtonProps) => {
  const [drawerVisible, setDrawerVisible] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const navigate = useNavigate()
  const [manifest, setManifest] = useState<AppManifest | null>(null)
  const [schemaErrors, setSchemaErrors] = useState<any[]>([])
  const [form] = Form.useForm()

  const closeDrawer = () => {
    setManifest(null)
    setSchemaErrors([])
    setDrawerVisible(false)
  }

  const install = () => {
    if (isLoading) return
    if (!manifest) return

    form.validateFields().then((values: any) => {
      setIsLoading(true)

      props.workspaceCtx
        .apiPOST('/app.install', {
          workspace_id: props.workspaceCtx.workspace.id,
          manifest: manifest,
          secret_key: values.secret_key
        })
        .then((app: App) => {
          // reload apps list
          props.workspaceCtx
            .refetchApps()
            .then(() => {
              props.workspaceCtx
                .refreshWorkspace()
                .then(() => {
                  setIsLoading(false)
                  message.success(app.name + ' has been installed.')
                  // redirect to app
                  navigate(
                    '/orgs/' +
                      props.workspaceCtx.organization.id +
                      '/workspaces/' +
                      props.workspaceCtx.workspace.id +
                      '/apps/' +
                      app.id
                  )
                })
                .catch(() => {})
            })
            .catch(() => {})
        })
        .catch((e) => {
          setIsLoading(false)
        })
    })
  }

  // console.log('initialValues', initialValues);

  const onManifestUpload = (file: any) => {
    // console.log('beforeUpload', file)

    // read the file content
    const reader = new FileReader()
    reader.readAsText(file)
    reader.onload = (e) => {
      let data = null
      try {
        data = JSON.parse(e.target?.result as string) as AppManifest
      } catch (e: any) {
        console.log(e)
        message.error('Could not parse manifest.json file.')
        return false
      }

      // replace reserved columns
      if (data.app_tables) {
        data.app_tables.forEach((table) => {
          let hasUser = false
          table.columns.forEach((column: TableColumn) => {
            if (column.name === 'user_id') {
              hasUser = true
            }

            if (column.name in reservedColumns) {
              column = Object.assign(column, reservedColumns[column.name])
            }
          })

          // make sure the merged_from_user_id column is added
          if (hasUser) {
            let hasMergedFromUserId = false
            table.columns.forEach((column: TableColumn) => {
              if (column.name === 'merged_from_user_id') {
                hasMergedFromUserId = true
              }
            })

            if (!hasMergedFromUserId) {
              table.columns.push(reservedColumns.merged_from_user_id)
            }
          }
        })
      }

      const ajv = new Ajv() // options can be passed, e.g. {allErrors: true}
      const validate = ajv.compile(appManifestSchema)
      // console.log('manifest', data)

      const valid = validate(data)
      if (!valid) {
        console.log(validate.errors)
        setSchemaErrors(validate.errors || [])
        message.error('Invalid manifest.json file.')
      } else {
        // id should start with 'appx_'
        if (!data.id.startsWith('appx_')) {
          message.error('Invalid app id. It should start with "appx_".')
          return false
        }
        setManifest(data)
        setSchemaErrors([])
      }
    }
    // abort upload
    return false
  }

  return (
    <>
      <Button
        type="primary"
        loading={isLoading}
        onClick={() => setDrawerVisible(true)}
        icon={<FontAwesomeIcon icon={faPlus} className={CSS.padding_r_xs} />}
      >
        Create
      </Button>
      {drawerVisible && (
        <Drawer
          title={
            <>
              <img
                src={customApp}
                className={css(CSS.appIcon, CSS.margin_r_m)}
                style={{ height: 30 }}
                alt=""
              />
              Create private app
            </>
          }
          width={1024}
          open={true}
          onClose={closeDrawer}
          extra={
            <Space>
              <Button key="a" ghost type="primary" loading={isLoading} onClick={closeDrawer}>
                Cancel
              </Button>
              <Button
                key="b"
                loading={isLoading}
                disabled={!manifest}
                onClick={install}
                type="primary"
              >
                Create
              </Button>
            </Space>
          }
          headerStyle={{ backgroundColor: backgroundColorBase }}
          bodyStyle={{ backgroundColor: backgroundColorBase }}
        >
          <>
            <Form
              form={form}
              labelCol={{ span: 6 }}
              wrapperCol={{ span: 14 }}
              layout="horizontal"
              // className={CSS.margin_a_m + ' ' + CSS.margin_b_xl}
            >
              <Form.Item
                name="secret_key"
                label="App secret key"
                // tooltip="This secret key will be used to verify webhooks signatures."
                rules={[{ required: true, message: Messages.RequiredField }]}
              >
                <Input placeholder="The secret key is used to authenticate webhooks signatures." />
              </Form.Item>
            </Form>
            {!manifest && (
              <>
                <div style={{ height: 300 }}>
                  <Upload.Dragger
                    accept=".json"
                    showUploadList={false}
                    beforeUpload={onManifestUpload}
                  >
                    <p className={CSS.font_size_xxl}>
                      <FontAwesomeIcon icon={faInbox} />
                    </p>
                    <p className="ant-upload-text">
                      Click or drag a <b>manifest.json</b> file to this area
                    </p>
                  </Upload.Dragger>
                </div>
                {schemaErrors.length > 0 && (
                  <div className={CSS.margin_t_m}>
                    <ul>
                      {schemaErrors.map((error, index) => (
                        <li key={index} className={CSS.text_red}>
                          {error.instancePath} - {error.keyword} {error.message}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </>
            )}
            {manifest && <BlockAboutApp manifest={manifest} />}
          </>
        </Drawer>
      )}
    </>
  )
}

export default PrivateAppButton
