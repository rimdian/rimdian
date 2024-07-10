import { css } from '@emotion/css'
import { faArrowRight } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
// import { Col, Row, Steps } from 'antd'
import Layout from 'components/common/layout'
// import UpsertLeadStagesButton from 'components/workspace/button_lead_stages'
import UpsertDomainButton from 'components/domain/button_upsert'
import { DataLogBatch, Organization, Workspace } from 'interfaces'
// import { useState } from 'react'
import { QueryObserverResult } from '@tanstack/react-query'
import CSS from 'utils/css'
import { CurrentWorkspaceCtxValue } from './context_current_workspace'

type Props = {
  organization: Organization
  workspaceCtx: CurrentWorkspaceCtxValue
  refreshWorkspace: () => Promise<QueryObserverResult<Workspace, unknown>>
  apiGET: (endpoint: string) => Promise<any>
  apiPOST: (endpoint: string, data: any) => Promise<any>
  collectorPOST: (sync: boolean, batch: DataLogBatch) => Promise<any>
  onComplete?: () => void
}

const RouteWorkspaceSetup = (props: Props) => {
  return (
    <Layout currentOrganization={props.organization} currentWorkspaceCtx={props.workspaceCtx}>
      <h1>Setup - {props.workspaceCtx.workspace.name}</h1>
      <DomainSetup {...props} onComplete={() => {}} />
    </Layout>
  )
}

const DomainSetup = (props: Props) => {
  if (props.workspaceCtx.workspace.domains.length === 0) {
    return (
      <div className={css([CSS.blockCTA, CSS.padding_v_l])}>
        <h2>Domains</h2>

        <p>
          In Rimdian, a Domain represents a way to interact with your business, it can be a web
          browser, a native app, a physical store....
        </p>
        <p>All user interactions (web sessions, orders, phone calls...) are recorded per domain.</p>
        <p>
          You will then be able to compare domains performances, filter reports per domain etc...
        </p>

        <div className={CSS.margin_t_l}>
          <UpsertDomainButton
            workspaceId={props.workspaceCtx.workspace.id}
            organizationId={props.organization.id}
            btnContent={
              <>
                Create a new domain &nbsp;
                <FontAwesomeIcon icon={faArrowRight} />
              </>
            }
            btnType="primary"
            btnSize="large"
            apiPOST={props.apiPOST}
            refreshWorkspace={props.refreshWorkspace}
            onComplete={props.onComplete}
          />
        </div>
      </div>
    )
  }

  return <></>
}
// const ConversionRuleSetup = (props: Props) => {
//   if (props.workspace.has_orders === false && props.workspace.has_leads === false) {
//     return (
//       <div className="block-cta padding-v-l">
//         <h2>Conversions settings</h2>

//         <p>
//           In Rimdian, you can track orders, a lead pipeline or subscriptions for recurring payments.
//         </p>
//         <p>
//           The Dashboards &amp; Analytics KPIs will automatically be customized according to your
//           configuration.
//         </p>

//         <div className="margin-t-l">
//           TODO
//           {/* <UpsertConversionRuleButton
//             workspaceId={props.workspace.id}
//             organizationId={props.organization.id}
//             btnContent={
//               <>
//                 Create a new conversion rule &nbsp;
//                 <FontAwesomeIcon icon={faArrowRight} />
//               </>
//             }
//             btnType="primary"
//             btnSize="large"
//             apiPOST={props.apiPOST}
//             refreshWorkspace={props.refreshWorkspace}
//             onComplete={props.onComplete}
//           /> */}
//         </div>
//       </div>
//     )
//   }
//   return <></>
// }

export default RouteWorkspaceSetup
