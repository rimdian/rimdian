import { Button, Drawer, Collapse } from 'antd'
import { useState } from 'react'
import CSS from 'utils/css'
import Code from 'utils/prism'
const { Panel } = Collapse

export interface ExecutedSQL {
  name: string
  sql: string
  args: any[]
}

export interface ButtonSQLExecutedProps {
  queries: ExecutedSQL[]
}

export const ButtonSQLExecuted = (props: ButtonSQLExecutedProps) => {
  const [visible, setVisible] = useState(false)

  return (
    <>
      <Button size="small" onClick={() => setVisible(true)}>
        View SQL
      </Button>
      {visible && (
        <Drawer
          title="SQL Executed"
          width="90%"
          placement="right"
          closable={true}
          onClose={() => setVisible(false)}
          open={true}
        >
          <div style={{ height: '100%', overflow: 'auto' }}>
            <Collapse accordion>
              {props.queries.map((query, index) => {
                return (
                  <Panel header={query.name} key={index}>
                    <div className={CSS.font_size_xs}>
                      <Code language="sql">{query.sql}</Code>
                    </div>
                    {query.args.length > 0 && (
                      <div className={CSS.margin_t_s}>
                        <b>Arguments:</b>
                      </div>
                    )}
                    {query.args.map((arg, index) => {
                      return <div key={index}>{arg}</div>
                    })}
                  </Panel>
                )
              })}
            </Collapse>
          </div>
        </Drawer>
      )}
    </>
  )
}
