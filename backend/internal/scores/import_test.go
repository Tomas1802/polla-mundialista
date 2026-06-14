package scores

import "testing"

func TestParseName(t *testing.T) {
	cases := []struct {
		file     string
		wantName string
		wantNo   int
	}{
		{"GRUPOS_Fase1_AlbertoRicardo1.csv", "Alberto Ricardo", 1},
		{"GRUPOS_Fase1_AlbertoRicardo3.csv", "Alberto Ricardo", 3},
		{"GRUPOS_Fase1_FreddyArevalo.csv", "Freddy Arevalo", 1},
		{"GRUPOS_Fase1_IsaMartinezNatsParra.csv", "Isa Martinez Nats Parra", 1},
		{"GRUPOS_Fase1_MariaPaulaBuitrago2.csv", "Maria Paula Buitrago", 2},
	}
	for _, c := range cases {
		name, no := parseName(c.file)
		if name != c.wantName || no != c.wantNo {
			t.Errorf("parseName(%q) = (%q, %d), want (%q, %d)", c.file, name, no, c.wantName, c.wantNo)
		}
	}
}
