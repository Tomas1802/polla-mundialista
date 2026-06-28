// Collapsible rules summary, taken from reglamento_polla_mundial_2026.md.
export default function Reglas() {
  return (
    <details className="reglas">
      <summary>📖 Reglas y puntos</summary>
      <div className="reglas-body">
        <p>Solo ingresas <strong>marcadores</strong>. Con ellos calculamos tu tabla de posiciones y tu puesto.
          Puedes editar un marcador hasta que el partido empieza.</p>

        <h4>Fase de grupos</h4>
        <ul>
          <li><b>7</b> — marcador exacto</li>
          <li><b>5</b> — marcador correcto pero con los equipos al revés</li>
          <li><b>4</b> — aciertas el equipo ganador (no el marcador)</li>
          <li><b>3</b> — aciertas que es empate (no el marcador)</li>
          <li><b>1</b> — pronosticaste, pero fallaste el resultado</li>
          <li><b>0</b> — marcador vacío o incompleto</li>
        </ul>

        <h4>Eliminatorias (cuenta el tiempo extra)</h4>
        <ul>
          <li><b>7</b> — marcador y ganador correctos</li>
          <li><b>5</b> — marcador correcto, ganador equivocado</li>
          <li><b>3</b> — ganador correcto, marcador equivocado</li>
          <li><b>1</b> — nada correcto · <b>0</b> — vacío</li>
        </ul>

        <h4>Final</h4>
        <ul>
          <li><b>10</b> / <b>8</b> / <b>6</b> / <b>1</b> — igual que eliminatorias, con más puntos</li>
        </ul>

        <h4>Tablas de grupo</h4>
        <ul>
          <li><b>7</b> — orden 1-2-3-4 exacto</li>
          <li><b>4</b> — 1º y 2º en orden correcto</li>
          <li><b>3</b> — los 3 clasificados correctos (en cualquier orden)</li>
          <li><b>2</b> — al menos un clasificado en su posición correcta</li>
          <li><b>1</b> — ninguna de las anteriores</li>
        </ul>

        <p className="reglas-note">En caso de empate en el ranking, desempata quien tenga más marcadores exactos.</p>
      </div>
    </details>
  )
}
