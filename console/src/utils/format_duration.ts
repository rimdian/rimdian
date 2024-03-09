const FormatDuration = (totalSecs: number, precision?: number) => {
  if (precision === undefined) {
    precision = 2
  }

  if (!totalSecs) {
    return '0s'
  }

  totalSecs = Math.round(totalSecs)

  let result, mins, secs, hours, days

  // in days
  if (totalSecs >= 86400) {
    days = Math.round(totalSecs / 86400)
    hours = Math.round((totalSecs - days * 86400) / 3600)
    mins = Math.round((totalSecs - days * 86400 - hours * 3600) / 60)
    secs = totalSecs - days * 86400 - hours * 3600 - mins * 60

    result = days + 'd'
    if (hours > 0 && precision >= 2) result += ' ' + hours + 'h'
    if (mins > 0 && precision >= 3) result += ' ' + mins + 'm'
    if (secs > 0 && precision >= 4) result += ' ' + secs + 's'
    return result
  } else if (totalSecs >= 3600) {
    // in hours
    hours = Math.round(totalSecs / 3600)
    mins = Math.round((totalSecs - hours * 3600) / 60)
    secs = totalSecs - hours * 3600 - mins * 60

    result = hours + 'h'
    if (mins > 0 && precision >= 2) result += ' ' + mins + 'm'
    if (secs > 0 && precision >= 3) result += ' ' + secs + 's'
    return result
  } else if (totalSecs >= 60) {
    // in mins
    mins = Math.round(totalSecs / 60)
    secs = totalSecs % 60
    result = mins + 'm'
    if (secs > 0) result += ' ' + secs + 's'
    return result
  } else {
    return totalSecs + 's'
  }
}

export default FormatDuration
