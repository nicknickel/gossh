package connection

import "testing"

func TestItem_FinalAddr(t *testing.T) {
	tests := []struct {
		name     string
		item     Item
		expected string
	}{
		{
			name:     "no user or address",
			item:     Item{Name: "host", Conn: Connection{}},
			expected: "host",
		},
		{
			name:     "with user and address",
			item:     Item{Name: "host", Conn: Connection{User: "user", Address: "addr"}},
			expected: "user@addr",
		},
		{
			name:     "with user no address",
			item:     Item{Name: "host", Conn: Connection{User: "user"}},
			expected: "user@host",
		},
		{
			name:     "with address no user",
			item:     Item{Name: "host", Conn: Connection{Address: "addr"}},
			expected: "addr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.FinalAddr(); got != tt.expected {
				t.Errorf("FinalAddr() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestItem_FilterValue(t *testing.T) {
	tests := []struct {
		name     string
		item     Item
		expected string
	}{
		{
			name:     "basic",
			item:     Item{Name: "host", Conn: Connection{Address: "addr", User: "user", Description: "desc"}},
			expected: "host addr user desc",
		},
		{
			name:     "empty fields",
			item:     Item{Name: "host", Conn: Connection{}},
			expected: "host   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.FilterValue(); got != tt.expected {
				t.Errorf("FilterValue() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestItem_Title(t *testing.T) {
	item := Item{Name: "host"}
	if got := item.Title(); got != "host" {
		t.Errorf("Title() = %v, want %v", got, "host")
	}
}

func TestItem_Description(t *testing.T) {
	tests := []struct {
		name     string
		item     Item
		expected string
	}{
		{
			name:     "with description",
			item:     Item{Name: "host", Conn: Connection{Address: "addr", Description: "desc"}},
			expected: "addr -> desc",
		},
		{
			name:     "no description",
			item:     Item{Name: "host", Conn: Connection{}},
			expected: "host -> ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.Description(); got != tt.expected {
				t.Errorf("Description() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestItem_WindowName(t *testing.T) {
	tests := []struct {
		name     string
		item     Item
		expected string
	}{
		{
			name:     "with name no address or user",
			item:     Item{Name: "host", Conn: Connection{}},
			expected: "host",
		},
		{
			name:     "with name and address no user",
			item:     Item{Name: "host", Conn: Connection{Address: "addr"}},
			expected: "host (addr)",
		},
		{
			name:     "with name and user no address",
			item:     Item{Name: "host", Conn: Connection{User: "user"}},
			expected: "user@host",
		},
		{
			name:     "with name and address and user",
			item:     Item{Name: "host", Conn: Connection{User: "user", Address: "addr"}},
			expected: "user@host (addr)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.WindowName(); got != tt.expected {
				t.Errorf("WindowName() = %v, want %v", got, tt.expected)
			}
		})
	}
}
