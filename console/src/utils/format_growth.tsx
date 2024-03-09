import { css } from '@emotion/css'
import numbro from 'numbro'
import CSS from './css'

const growthCss = css({
  fontSize: 11,
  paddingLeft: CSS.S
})

const FormatGrowth = (
  value: number | undefined,
  previousValue: number | undefined,
  goodIsBad?: boolean
) => {
  if (!value || !previousValue) {
    return ''
  }

  const growth = (value - previousValue) / previousValue
  let growthSymbol = ''
  let good, bad

  if (growth > 0) {
    growthSymbol = '▴'
    good = true
  } else if (growth < 0) {
    growthSymbol = '▾'
    bad = true
  }

  // inverse the growth colors
  if (goodIsBad) {
    good = !good
    bad = !bad
  }

  if (growth === 0) {
    return ''
  }

  return (
    <span className={css([growthCss, good && CSS.text_green, bad && CSS.text_orange])}>
      {growthSymbol}
      {numbro(growth).format({ output: 'percent', mantissa: 2 })}
    </span>
  )
}

export default FormatGrowth
