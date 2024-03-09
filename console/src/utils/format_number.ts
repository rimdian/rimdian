import numbro from 'numbro'
const FormatNumber = (value: number) => {
  if (value === null || value === undefined) return 'null'
  return numbro(value).format({
    average: false,
    mantissa: 2,
    optionalMantissa: true,
    thousandSeparated: true
  })
}

export default FormatNumber
