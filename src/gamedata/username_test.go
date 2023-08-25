package gamedata

import "testing"

func TestCleanUsername(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"", ""},

		{"ok", "ok"},
		{"1", "1"},

		{"!ok!", "ok"},
		{"~hel!lo~", "hello"},
		{"バカ senpai", "senpai"},
		{"the バカ senpai", "the senpai"},
		{"Cool-Name", "Cool_Name"},
		{"გამარჯობა Giorgi", "Giorgi"},
		{"привет Ivan", "Ivan"},
		{"привет?Ivan", "Ivan"},
		{"-Stealth-", "_Stealth_"},
		{"🔥fire mage🔥", "fire mage"},
		{"evil😈guy", "evil guy"},
		{"evil😈😈guy", "evil guy"},
		{"😈evil😈😈😈guy😈", "evil guy"},
		{"#yes#no", "yesno"},
	}

	for _, test := range tests {
		have := CleanUsername(test.s)
		if have != test.want {
			t.Fatalf("clean(%q):\nhave %q\nwant %q", test.s, have, test.want)
		}
	}
}
