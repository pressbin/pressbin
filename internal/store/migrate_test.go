package store

import "testing"

func TestParseMigrationVersion(t *testing.T) {
	tests := []struct {
		file    string
		want    int
		wantErr bool
	}{
		{"001_initial.sql", 1, false},
		{"003_remove_dev_sample_posts.sql", 3, false},
		{"010_add_tags.sql", 10, false},
		{"bad.sql", 0, true},
		{"_nope.sql", 0, true},
	}
	for _, tc := range tests {
		got, err := parseMigrationVersion(tc.file)
		if tc.wantErr {
			if err == nil {
				t.Errorf("%s: expected error", tc.file)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: %v", tc.file, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%s: got %d, want %d", tc.file, got, tc.want)
		}
	}
}

func TestMigrate_idempotent(t *testing.T) {
	st := openTestStore(t)
	if err := st.migrate(); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
}
