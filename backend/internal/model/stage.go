package model

// Stage constants as reported by football-data.org for a World Cup.
const (
	StageGroup        = "GROUP_STAGE"
	StageLast32       = "LAST_32" // dieciseisavos (48-team format)
	StageLast16       = "LAST_16" // octavos
	StageQuarter      = "QUARTER_FINALS"
	StageSemi         = "SEMI_FINALS"
	StageThirdPlace   = "THIRD_PLACE"
	StageFinal        = "FINAL"
)

// IsGroupStage reports whether the stage is the group phase.
func IsGroupStage(stage string) bool { return stage == StageGroup }

// IsFinal reports whether the stage is the final match.
func IsFinal(stage string) bool { return stage == StageFinal }

// IsKnockout reports whether the stage is a knockout round scored by section 3
// of the reglamento (everything past the group stage except the final, which
// has its own higher point tiers).
func IsKnockout(stage string) bool {
	switch stage {
	case StageLast32, StageLast16, StageQuarter, StageSemi, StageThirdPlace:
		return true
	default:
		return false
	}
}
