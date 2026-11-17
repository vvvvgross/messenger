package main

import "testing"

func TestParseHTTPAddr(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "default address",
			input: "",
			want:  ":8080",
		},
		{
			name:  "all interfaces",
			input: ":8080",
			want:  ":8080",
		},
		{
			name:  "IPv4 address",
			input: "127.0.0.1:8080",
			want:  "127.0.0.1:8080",
		},
		{
			name:  "hostname",
			input: "localhost:8080",
			want:  "localhost:8080",
		},
		{
			name:  "IPv6 address",
			input: "[::1]:8080",
			want:  "[::1]:8080",
		},
		{
			name:  "minimum port",
			input: ":1",
			want:  ":1",
		},
		{
			name:  "maximum port",
			input: ":65535",
			want:  ":65535",
		},
		{
			name:    "missing separator",
			input:   "localhost",
			wantErr: true,
		},
		{
			name:    "missing port",
			input:   "localhost:",
			wantErr: true,
		},
		{
			name:    "zero port",
			input:   ":0",
			wantErr: true,
		},
		{
			name:    "port out of range",
			input:   ":65536",
			wantErr: true,
		},
		{
			name:    "negative port",
			input:   ":-1",
			wantErr: true,
		},
		{
			name:    "positive sign",
			input:   ":+8080",
			wantErr: true,
		},
		{
			name:    "named port",
			input:   "localhost:http",
			wantErr: true,
		},
		{
			name:    "fractional port",
			input:   ":80.5",
			wantErr: true,
		},
		{
			name:    "leading whitespace",
			input:   " localhost:8080",
			wantErr: true,
		},
		{
			name:    "trailing whitespace",
			input:   "localhost:8080 ",
			wantErr: true,
		},
		{
			name:    "embedded whitespace",
			input:   "local host:8080",
			wantErr: true,
		},
		{
			name:    "newline",
			input:   "localhost:8080\n",
			wantErr: true,
		},
		{
			name:    "Unicode whitespace",
			input:   "localhost:\u00a08080",
			wantErr: true,
		},
		{
			name:    "IPv6 without brackets",
			input:   "::1:8080",
			wantErr: true,
		},
		{
			name:    "URL instead of address",
			input:   "http://localhost:8080",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHTTPAddr(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				if got != "" {
					t.Errorf("address = %q on error, want empty string", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("address = %q, want %q", got, tt.want)
			}
		})
	}
}
