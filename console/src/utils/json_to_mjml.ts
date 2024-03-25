const indentPad = (n: number) => Array(n + 1).join(' ')

const TAG_CONVERSION: any = {
  'mj-dev': 'mj-raw'
}

const lineAttributes = (attrs: any) =>
  Object.keys(attrs)
    .filter((key) => key !== 'passport')
    .map((key) => `${key}="${attrs[key]}"`)
    .sort()
    .join(' ')

export default function json2xml(node: any, indent = 0) {
  let { tagName } = node
  const { children, content, attributes } = node

  if (tagName in TAG_CONVERSION) {
    tagName = TAG_CONVERSION[tagName] // eslint-disable-line prefer-destructuring
  }

  const space = indentPad(indent)

  let attrs = (attributes && ` ${lineAttributes(attributes)}`) || ''

  if (!attrs.trim()) {
    attrs = ''
  }

  const inside =
    (content && `\n${space}  ${content}\n${space}`) ||
    (children &&
      `\n${children.map((c: any) => `${json2xml(c, indent + 2)}`).join('\n')}\n${space}`) ||
    ''

  return `${space}<${tagName}${attrs}>${inside}</${tagName}>`
}
