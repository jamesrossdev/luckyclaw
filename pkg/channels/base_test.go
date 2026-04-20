package channels

import "testing"

func TestBaseChannelIsAllowed(t *testing.T) {
	tests := []struct {
		name      string
		allowList []string
		senderID  string
		want      bool
	}{
		{
			name:      "empty allowlist allows all",
			allowList: nil,
			senderID:  "anyone",
			want:      true,
		},
		{
			name:      "sender ID matches bare numeric allowlist",
			allowList: []string{"123456"},
			senderID:  "123456",
			want:      true,
		},
		{
			name:      "compound sender matches bare numeric allowlist",
			allowList: []string{"123456"},
			senderID:  "123456|alice@s.whatsapp.net",
			want:      true,
		},
		{
			name:      "compound sender matches right-side identity allowlist",
			allowList: []string{"254112457495"},
			senderID:  "71717598818505|254112457495@s.whatsapp.net",
			want:      true,
		},
		{
			name:      "compound sender matches left-side identity allowlist",
			allowList: []string{"71717598818505"},
			senderID:  "71717598818505|254112457495@s.whatsapp.net",
			want:      true,
		},
		{
			name:      "sender ID matches allowlist with domain suffix",
			allowList: []string{"123456@s.whatsapp.net"},
			senderID:  "123456",
			want:      true,
		},
		{
			name:      "numeric sender matches legacy compound allowlist → denied (no legacy support)",
			allowList: []string{"123456|alice"},
			senderID:  "123456",
			want:      false,
		},
		{
			name:      "username-style allowlist → denied (no legacy support)",
			allowList: []string{"@alice"},
			senderID:  "123456|alice",
			want:      false,
		},
		{
			name:      "telegram-style compound sender does not match username allowlist",
			allowList: []string{"alice"},
			senderID:  "123456|alice",
			want:      false,
		},
		{
			name:      "non matching sender is denied",
			allowList: []string{"123456"},
			senderID:  "654321|bob",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := NewBaseChannel("test", nil, nil, tt.allowList)
			if got := ch.IsAllowed(tt.senderID); got != tt.want {
				t.Fatalf("IsAllowed(%q) = %v, want %v", tt.senderID, got, tt.want)
			}
		})
	}
}
