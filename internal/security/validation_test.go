package security

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "user@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"valid email with plus", "user+tag@example.com", false},
		{"empty email", "", true},
		{"missing @", "userexample.com", true},
		{"missing domain", "user@", true},
		{"missing local part", "@example.com", true},
		{"invalid characters", "user name@example.com", true},
		{"too long", string(make([]byte, 300)) + "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"strong password", "MyP@ssw0rd123!", false},
		{"another strong", "C0mpl3x!Pass", false},
		{"too short", "Short1!", true},
		{"no uppercase", "mypassw0rd!", true},
		{"no lowercase", "MYPASSW0RD!", true},
		{"no number", "MyPassword!", true},
		{"no special", "MyPassword123", true},
		{"too long", string(make([]byte, 150)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContainsXSS(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"clean text", "Hello World", false},
		{"script tag", "<script>alert('xss')</script>", true},
		{"javascript protocol", "javascript:alert(1)", true},
		{"onerror attribute", "<img onerror='alert(1)'>", true},
		{"iframe tag", "<iframe src='evil.com'>", true},
		{"mixed case script", "<ScRiPt>alert(1)</ScRiPt>", true},
		{"eval function", "eval('alert(1)')", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.ContainsXSS(tt.input)
			if got != tt.want {
				t.Errorf("ContainsXSS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainsSQLi(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"clean text", "Hello World", false},
		{"single quote", "O'Brien", true},
		{"SQL comment", "admin' --", true},
		{"union select", "' UNION SELECT * FROM users --", true},
		{"drop table", "'; DROP TABLE users; --", true},
		{"semicolon", "test;", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.ContainsSQLi(tt.input)
			if got != tt.want {
				t.Errorf("ContainsSQLi() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"clean filename", "document.pdf", "document.pdf"},
		{"path traversal", "../../../etc/passwd", "etcpasswd"},
		{"null bytes", "file\x00.txt", "file.txt"},
		{"special chars", "file!@#$%.txt", "file_____.txt"},
		{"forward slash", "path/to/file.txt", "pathtofile.txt"},
		{"backslash", "path\\to\\file.txt", "pathtofile.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.SanitizeFilename(tt.filename)
			if got != tt.want {
				t.Errorf("SanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCommonPassword(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"common password", "password", true},
		{"common 123456", "123456", true},
		{"common qwerty", "qwerty", true},
		{"strong password", "MyStr0ng!P@ssw0rd", false},
		{"random strong", "X9#mK2$pL5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.IsCommonPassword(tt.password)
			if got != tt.want {
				t.Errorf("IsCommonPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeHTML(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain text", "Hello World", "Hello World"},
		{"script tags", "<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"ampersand", "Tom & Jerry", "Tom &amp; Jerry"},
		{"quotes", `He said "Hello"`, "He said &#34;Hello&#34;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.SanitizeHTML(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}
