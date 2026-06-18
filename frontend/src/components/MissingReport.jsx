import { useState } from 'react'
import { api } from '../api.js'

// Admin tool: list which cards still haven't filled in today's (still-open)
// matches, with a ready-to-copy message for the group.
export default function MissingReport() {
  const [report, setReport] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [copied, setCopied] = useState(false)

  async function run() {
    setLoading(true)
    setError('')
    setCopied(false)
    try {
      setReport(await api.adminMissingToday())
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  async function copy() {
    try {
      await navigator.clipboard.writeText(buildText(report))
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      setError('No se pudo copiar automáticamente; selecciona el texto y cópialo a mano.')
    }
  }

  return (
    <section className="admin-section">
      <h2>Validar marcadores de hoy</h2>
      <p className="tablas-intro">
        Lista los cartones que aún no han llenado los partidos de hoy (los que todavía
        no comienzan), para avisar al grupo.
      </p>
      <button className="primary-btn admin-save-btn" onClick={run} disabled={loading}>
        {loading ? 'Revisando…' : 'Validar partidos de hoy'}
      </button>

      {error && <p className="error">{error}</p>}

      {report && report.matchCount === 0 && (
        <p className="empty">No hay partidos de hoy pendientes por llenar.</p>
      )}

      {report && report.matchCount > 0 && report.cards.length === 0 && (
        <p className="missing-allgood">✅ ¡Todos llenaron los {report.matchCount} partido(s) de hoy!</p>
      )}

      {report && report.matchCount > 0 && report.cards.length > 0 && (
        <div className="missing-report">
          <div className="missing-head">
            <span>{report.cards.length} cartón(es) por completar · {report.matchCount} partido(s) hoy</span>
            <button className="fix-save" onClick={copy}>{copied ? 'Copiado ✓' : 'Copiar'}</button>
          </div>
          <pre className="missing-text">{buildText(report)}</pre>
        </div>
      )}
    </section>
  )
}

// buildText renders the WhatsApp-friendly message (also used for the on-screen
// block, so what you copy is exactly what you see).
function buildText(r) {
  if (!r || r.matchCount === 0) return 'No hay partidos de hoy pendientes por llenar.'
  const lines = []
  lines.push(`⚽ Faltan marcadores de hoy (${r.date})`)
  lines.push(`Partidos: ${r.matches.join(', ')}`)
  lines.push('')
  if (!r.cards || r.cards.length === 0) {
    lines.push('✅ ¡Todos llenaron los partidos de hoy! 🎉')
    return lines.join('\n')
  }
  lines.push('Por favor completen sus cartones:')
  for (const c of r.cards) {
    const all = c.missing.length === r.matchCount
    const detail = all ? 'faltan todos' : `faltan ${c.missing.join(', ')}`
    lines.push(`• ${c.playerName} · ${c.cardLabel}: ${detail}`)
  }
  return lines.join('\n')
}
