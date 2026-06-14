// Package scoring implements the point system of the Polla Mundialista 2026,
// exactly as defined in reglamento_polla_mundial_2026.md.
//
// It is intentionally a pure, dependency-free package: every function is a
// deterministic transformation from a prediction plus an official result into
// a number of points. This keeps the money-affecting logic isolated, easy to
// audit, and fully unit-testable without a database or network.
//
// The reglamento defines three independent kinds of prediction, scored
// separately:
//
//   - Match score (marcador): sections 1 (group), 3 (knockout) and 4 (final).
//     See match.go.
//   - Bracket matchup (which two teams meet in a knockout slot): the
//     "acierta los equipos que se enfrentan" rows of sections 3 and 4.
//     See bracket.go.
//   - Group final positions: section 2. See positions.go.
//
// Where the reglamento is explicitly ambiguous (the document mandates
// confirming ambiguities with the organizer rather than assuming), the
// behaviour is documented inline and kept configurable.
package scoring
