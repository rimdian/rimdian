const FormatConversionRole = (role: number) => {
  if (role === 0) return 'Alone'
  if (role === 1) return 'Initiator'
  if (role === 2) return 'Assisting'
  if (role === 3) return 'Closer'
  return 'Unknown'
}

export default FormatConversionRole
