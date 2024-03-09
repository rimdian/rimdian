const KeepURLPath = (url: string) => {
  if (url.indexOf('http') === -1) return url
  // parse url and keep path with qs
  const parser = document.createElement('a')

  parser.href = url
  const path = parser.pathname + parser.search + parser.hash
  parser.remove()

  return path
}

export default KeepURLPath
