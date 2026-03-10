package lang

import (
	"testing"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
)

func TestIsCodeFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"main.go", true},
		{"pkg/handler.go", true},
		{"script.py", true},
		{"app/user_service.rb", true},
		{"src/App.java", true},
		{"lib/index.ts", true},
		{"api/main.js", true},
		{"service.go", true},
		{"main_test.go", true},
		{"test_foo.py", true},
		{"foo_test.rb", true},
		{"README.md", false},
		{"doc.txt", false},
		{"Dockerfile", false},
		{"file.xyz", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsCodeFile(tt.path); got != tt.want {
				t.Errorf("IsCodeFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"pkg_test.go", true},
		{"handler_test.go", true},
		{"test_user.py", true},
		{"user_test.py", true},
		{"test_services.py", true},
		{"foo_spec.rb", true},
		{"user_spec.rb", true},
		{"UserServiceTest.java", true},
		{"SomethingTest.java", true},
		{"api.test.ts", true},
		{"api.spec.ts", true},
		{"main.go", false},
		{"user.py", false},
		{"service.rb", false},
		{"README.md", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsTestFile(tt.path); got != tt.want {
				t.Errorf("IsTestFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestFilterCodeFiles(t *testing.T) {
	files := []domain.FileChange{
		{Filename: "main.go"},
		{Filename: "pkg/handler_test.go"},
		{Filename: "service.py"},
		{Filename: "test_service.py"},
		{Filename: "README.md"},
		{Filename: "user_spec.rb"},
		{Filename: "lib/helper.rb"},
	}
	got := FilterCodeFiles(files)
	var names []string
	for _, f := range got {
		names = append(names, f.Filename)
	}
	// Deve conter apenas código que não é arquivo de teste
	wantNames := []string{"main.go", "service.py", "lib/helper.rb"}
	if len(got) != len(wantNames) {
		t.Errorf("FilterCodeFiles len = %d, want %d; got %v", len(got), len(wantNames), names)
	}
	gotSet := make(map[string]bool)
	for _, n := range names {
		gotSet[n] = true
	}
	for _, w := range wantNames {
		if !gotSet[w] {
			t.Errorf("FilterCodeFiles missing %q; got %v", w, names)
		}
	}
}

func TestFileUnitName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"main.go", "main"},
		{"pkg/handler.go", "handler"},
		{"internal/user_service.py", "user_service"},
		{"deep/path/to/file.rb", "file"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := FileUnitName(tt.path)
			if got != tt.want {
				t.Errorf("FileUnitName(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestFileUnitName_Empty(t *testing.T) {
	if FileUnitName("") != "" {
		t.Error("FileUnitName(\"\") should return empty")
	}
}

func TestFilterCodeFiles_Empty(t *testing.T) {
	got := FilterCodeFiles(nil)
	if got != nil {
		t.Errorf("FilterCodeFiles(nil) should return nil, got %v", got)
	}
	got = FilterCodeFiles([]domain.FileChange{})
	if len(got) != 0 {
		t.Errorf("FilterCodeFiles([]) should return empty slice, got len %d", len(got))
	}
}
