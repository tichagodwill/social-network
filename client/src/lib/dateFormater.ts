const units = [
  { label: 'years', seconds: 31536000 },
  { label: 'months', seconds: 2592000 },
  { label: 'days', seconds: 86400 },
  { label: 'hours', seconds: 3600 },
  { label: 'minutes', seconds: 60 },
  { label: 'seconds', seconds: 1 },
]

export function getFormattedDate(inputDate: Date) {
  const formattedDate = inputDate.toDateString() // Format: "Tue Sep 24 2024"

  // diff
  let suffix = 'ago'
  let now = new Date()

  if (now < inputDate) {
    const tmp = now
    suffix = 'ahead'

    now = inputDate
    inputDate = tmp
  }

  const diffSeconds = (now.getTime() - inputDate.getTime()) / 1000 + 1

  const unit = units.find((u) => diffSeconds >= u.seconds)!
  const value = Math.round(diffSeconds / unit.seconds)
  const timeAgo = `${value} ${unit.label} ${suffix}`

  return { formated: formattedDate, diff: timeAgo }
}

export function getLastDate(input: Date) {
  const now = new Date()
  const diff = (now.getTime() - input.getTime()) / 1000 + 1

  if (diff < 3)
    return 'now'

  if (diff < 60) // 32s
    return `${diff}s`

  if (diff < 3600) { // <1hour: 32min
    const minutes = Math.floor(diff / 60)
    return `${minutes}s`
  }

  if (diff < 86400) { // <1day: 12:33pm
    return input.toLocaleTimeString('en-US', { hour: 'numeric', minute: 'numeric', hour12: true })
  }

  if (diff < 604800) { // <1week: sun
    return input.toLocaleDateString('en-US', { weekday: 'short' })
  }

  if (diff < 31104000) { // <1year: 12 jan
    return input.toLocaleDateString('en-US', { day: 'numeric', month: 'short' })
  }

  // >1year: 2 year
  const years = Math.round(diff / 31104000)
  return `${years} year`
}
