import numbro from 'numbro'
const FormatPercent = (value: number) => {
  return numbro(value).format({
    average: false,
    output: 'percent',
    mantissa: 2,
    optionalMantissa: true
  })
}

export default FormatPercent
