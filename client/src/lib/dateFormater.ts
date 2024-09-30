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

  const diffSeconds = (now.getTime() - inputDate.getTime()) / 1000

  const unit = units.find((u) => diffSeconds >= u.seconds)!
  const value = Math.round(diffSeconds / unit.seconds)
  const timeAgo = `${value} ${unit.label} ${suffix}`

  return { formated: formattedDate, diff: timeAgo }
}
