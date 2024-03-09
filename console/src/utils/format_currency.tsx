import numbro from 'numbro'

type FormatCurrencyOptions = {
  className?: string
  light?: boolean
  currencyDisplay?: 'symbol' | 'code' | 'name'
  originalAmount?: number // used when the amount has been converted to another currency
  originalCurrency?: string // used when the amount has been converted to another currency
  classNameOriginal?: string
}

const FormatCurrency = (amount: number, currency: string, options?: FormatCurrencyOptions) => {
  let str

  // let amountFormatted = N(amount / 100).format(options?.light ? '0,0' : '0,0[.]00')
  let amountFormatted = numbro(amount / 100).format({
    average: false,
    optionalMantissa: true,
    mantissa: options?.light ? 0 : 2,
    thousandSeparated: true
  })

  switch (currency) {
    case 'ARS':
      str = '$' + amountFormatted
      break
    case 'AUD':
      str = '$' + amountFormatted
      break
    case 'BRL':
      str = 'R$' + amountFormatted
      break
    case 'CAD':
      str = '$' + amountFormatted
      break
    case 'COP':
      str = '$' + amountFormatted
      break
    case 'CZK':
      str = 'Kč' + amountFormatted
      break
    case 'DKK':
      str = 'kr' + amountFormatted
      break
    case 'EUR':
      str = amountFormatted + '€'
      break
    case 'HKD':
      str = '$' + amountFormatted
      break
    case 'IDR':
      str = 'Rp' + amountFormatted
      break
    case 'ILS':
      str = '₪' + amountFormatted
      break
    case 'JPY':
      str = '¥' + amountFormatted
      break
    case 'KRW':
      str = '₩' + amountFormatted
      break
    case 'MYR':
      str = 'RM' + amountFormatted
      break
    case 'MXN':
      str = '$' + amountFormatted
      break
    case 'NZD':
      str = '$' + amountFormatted
      break
    case 'PKR':
      str = '₨' + amountFormatted
      break
    case 'PHP':
      str = '₱' + amountFormatted
      break
    case 'SGD':
      str = '$' + amountFormatted
      break
    case 'ZAR':
      str = 'R' + amountFormatted
      break
    case 'TWD':
      str = 'NT$' + amountFormatted
      break
    case 'GBP':
      str = '£' + amountFormatted
      break
    case 'USD':
      str = '$' + amountFormatted
      break
    default:
      str = currency + amountFormatted
  }

  if (
    options?.originalCurrency &&
    options?.originalCurrency !== currency &&
    options?.originalAmount &&
    options?.originalAmount > 0
  ) {
    return (
      <span>
        <span className={options?.className || ''}>{str}</span>{' '}
        <span className={options?.classNameOriginal || 'size-12'}>
          ({FormatCurrency(options?.originalAmount, options?.originalCurrency)})
        </span>
      </span>
    )
  } else {
    return <span className={options?.className || ''}>{str}</span>
  }
}

export default FormatCurrency
