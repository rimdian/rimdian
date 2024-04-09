// extract TLD from URL
const extractTLD = (url: string) => {
  const hostname = new URL(url).hostname
  return hostname.split('.').slice(-2).join('.')
}

export default extractTLD
