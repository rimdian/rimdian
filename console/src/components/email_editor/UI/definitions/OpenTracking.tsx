import { Alert, Space } from 'antd'
import { BlockDefinitionInterface, BlockRenderSettingsProps } from '../../Block'
import { BlockEditorRendererProps } from '../../BlockEditorRenderer'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faEye } from '@fortawesome/free-solid-svg-icons'

const OpenTrackingBlockDefinition: BlockDefinitionInterface = {
  name: 'Open tracking',
  kind: 'openTracking',
  containsDraggables: false,
  isDraggable: true,
  draggableIntoGroup: 'column',
  isDeletable: true,
  defaultData: {},
  menuSettings: {},

  RenderSettings: (props: BlockRenderSettingsProps) => {
    return (
      <div className="rmdeditor-padding-h-l">
        <Alert
          type="info"
          showIcon
          message="An invisible tracking pixel will be added to the email. When the email is opened, the pixel will be loaded and the open event will be recorded."
        />
      </div>
    )
  },

  renderEditor: (props: BlockEditorRendererProps) => {
    return (
      <div
        style={{
          width: '100%',
          backgroundColor: '#f7f7f7',
          padding: '12px',
          border: '1px solid #e8e8e8',
          borderRadius: '4px'
        }}
      >
        OPEN_TRACKING_PIXEL
      </div>
    )
  },

  renderMenu: (_blockDefinition: BlockDefinitionInterface) => {
    return (
      <div className="rmdeditor-ui-block">
        <Space size="middle">
          <FontAwesomeIcon icon={faEye} />
          Open tracking
        </Space>
      </div>
    )
  }
}

export default OpenTrackingBlockDefinition
