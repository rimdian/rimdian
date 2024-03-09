import { TableInformationSchema } from './schemas'
import CSS from 'utils/css'

type BlockDBDiagramProps = {
  data?: TableInformationSchema[]
  isLoading: boolean
  isFetching: boolean
}

const BlockDBDiagram = (_props: BlockDBDiagramProps) => {
  return <div className={CSS.margin_t_m}>Coming soon...</div>
}

export default BlockDBDiagram
