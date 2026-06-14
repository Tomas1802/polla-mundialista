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

export function formatDate(utc) {
  return capitalize(dateFmt.format(new Date(utc)))
}

export function formatTime(utc) {
  return timeFmt.format(new Date(utc))
}

function capitalize(s) {
  return s.charAt(0).toUpperCase() + s.slice(1)
}
