import { useEffect, useRef, ReactNode } from 'react'
import Prism from 'prismjs'

type CodeProps = {
  language: 'javascript' | 'json' | 'css' | 'markup' | 'sql'
  children?: ReactNode
  style?: any
}

export default function Code(props: CodeProps) {
  const codeEl = useRef(null)

  useEffect(() => {
    if (codeEl.current) {
      Prism.highlightElement(codeEl.current)
      // Prism.highlightAll()
    }
  }, [])

  return (
    <pre>
      <code style={props.style} ref={codeEl} className={'language-' + props.language}>
        {props.children}
      </code>
    </pre>
  )
}
