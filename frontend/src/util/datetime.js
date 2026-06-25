// Format a UTC timestamp in the visitor's own timezone and in Spanish.
const dateFmt = new Intl.DateTimeFormat('es', {
  weekday: 'long',
  day: 'numeric',
  month: 'long',
})
const timeFmt = new Intl.DateTimeFormat('es', {
  hour: '2-digit',
  minute: '2-digit',
})
const shortDateFmt = new Intl.DateTimeFormat('es', {
  day: 'numeric',
  month: 'short',
})

export function formatDate(utc) {
  return capitalize(dateFmt.format(new Date(utc)))
}

// Compact date like "14 jun" for dense lists.
export function formatShortDate(utc) {
  return shortDateFmt.format(new Date(utc)).replace('.', '')
}

export function formatTime(utc) {
  return timeFmt.format(new Date(utc))
}

function capitalize(s) {
  return s.charAt(0).toUpperCase() + s.slice(1)
}
