package main

import (
	"reflect"
	"testing"
)

func TestParseMigrations(t *testing.T) {
	tests := []struct {
		name    string
		files   []string
		want    []migration
		wantErr bool
	}{
		{
			name:  "sorts numerically, not alphabetically",
			files: []string{"v10__ten.sql", "v2__reminders.sql", "v1__init.sql"},
			want: []migration{
				{1, "v1__init.sql"},
				{2, "v2__reminders.sql"},
				{10, "v10__ten.sql"},
			},
		},
		{name: "no files", files: nil, want: nil},
		{name: "duplicate version", files: []string{"v2__a.sql", "v02__b.sql"}, wantErr: true},
		{name: "missing v prefix", files: []string{"1__init.sql"}, wantErr: true},
		{name: "missing separator", files: []string{"v1_init.sql"}, wantErr: true},
		{name: "missing name", files: []string{"v1__.sql"}, wantErr: true},
		{name: "non-numeric version", files: []string{"vx__init.sql"}, wantErr: true},
		{name: "version zero", files: []string{"v0__init.sql"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMigrations(tt.files)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseMigrations(%v) error = %v, wantErr %v", tt.files, err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseMigrations(%v) = %v, want %v", tt.files, got, tt.want)
			}
		})
	}
}
