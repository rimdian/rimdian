import { Tag, Tooltip, Badge } from 'antd'
import { SubscriptionList } from 'interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { useSearchParams } from 'react-router-dom'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faChevronRight } from '@fortawesome/free-solid-svg-icons'
import numbro from 'numbro'
import CSS from 'utils/css'
import Block from 'components/common/block'
import { css } from '@emotion/css'
import ButtonUpsertSegment from 'components/segment/button_upsert'
import { Segment } from 'components/segment/interfaces'
import ButtonUpsertSubscriptionList from 'components/subscription_list/button_upsert'

const secondMenuCSS = {
  title: css({
    padding: CSS.M,
    fontWeight: 'bold'
  }),

  counter: css(
    {
      fontSize: '10px',
      paddingLeft: CSS.S
    },
    CSS.font_weight_semibold
  ),

  list: css({
    listStyleType: 'none',
    margin: 0,
    padding: '8px 0',

    '& > li': {
      position: 'relative',
      cursor: 'pointer',
      padding: CSS.S + ' ' + CSS.L,
      borderLeft: '2px solid rgba(0,0,0,0)',
      '.chevron': {
        position: 'absolute',
        top: 16,
        left: 7,
        fontSize: '10px',
        color: '#4e6cff',
        opacity: 0
      },
      '&:hover': {
        '.chevron': {
          opacity: 1
        }
      }
    },

    '& > li.active': {
      '.chevron': {
        opacity: 1
      }
    }
  })
}

interface BlockSidebarProps {
  segments: Segment[]
  currentSegment?: Segment
  currentList?: SubscriptionList
}

const BlockSidebarUsers = (props: BlockSidebarProps) => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [_searchParams, setSearchParams] = useSearchParams()

  return (
    <>
      <Block title="Segments" extra={<ButtonUpsertSegment />}>
        <ul className={secondMenuCSS.list}>
          {/* <li
            onClick={() => setSearchParams({})}
            className={props.currentSegment && props.currentSegment.id === '_all' ? 'active' : ''}
          >
            <FontAwesomeIcon className="chevron" icon={faChevronRight} />
            All users
            <span className={secondMenuCSS.counter}>
              {workspaceCtx.segmentsMap._all &&
                numbro(workspaceCtx.segmentsMap['_all'].users_count).format({
                  totalLength: 3,
                  trimMantissa: true
                })}
            </span>
          </li> */}
          {props.segments.map((segment: Segment) => {
            return (
              <li
                key={segment.id}
                onClick={() => setSearchParams({ segment_id: segment.id })}
                className={
                  props.currentSegment && props.currentSegment.id === segment.id ? 'active' : ''
                }
              >
                {segment.status === 'building' && (
                  <Tooltip className={CSS.pull_right} title="Building...">
                    <Badge status="processing" />
                  </Tooltip>
                )}
                {segment.status === 'active' && (
                  <Tooltip className={CSS.pull_right} title="Active">
                    <Badge status="success" />
                  </Tooltip>
                )}
                {segment.status === 'deleted' && (
                  <Tooltip className={CSS.pull_right} title="Deleted">
                    <Badge status="error" />
                  </Tooltip>
                )}

                <FontAwesomeIcon className="chevron" icon={faChevronRight} />
                <Tag color={segment.color}>{segment.name}</Tag>
                <span className={secondMenuCSS.counter}>
                  {workspaceCtx.segmentsMap.anonymous &&
                    numbro(segment.users_count).format({
                      totalLength: 3,
                      trimMantissa: true
                    })}
                </span>
              </li>
            )
          })}
        </ul>
      </Block>

      <Block title="Subscription lists" extra={<ButtonUpsertSubscriptionList />}>
        {workspaceCtx.subscriptionLists.length === 0 && (
          <div className={CSS.padding_a_m}>
            Subscription lists are used to send emails to your users. You can create a new list by
            clicking the button above.
          </div>
        )}
        <ul className={secondMenuCSS.list}>
          {workspaceCtx.subscriptionLists.map((list: SubscriptionList) => {
            return (
              <li
                key={list.id}
                onClick={() => setSearchParams({ list_id: list.id })}
                className={props.currentList?.id === list.id ? 'active' : ''}
              >
                <FontAwesomeIcon className="chevron" icon={faChevronRight} />
                <Tag color={list.color}>{list.name}</Tag>
                <span className={secondMenuCSS.counter}>
                  {numbro(list.active_users).format({
                    totalLength: 3,
                    trimMantissa: true
                  })}
                </span>
              </li>
            )
          })}
        </ul>
      </Block>
    </>
  )
}

export default BlockSidebarUsers
