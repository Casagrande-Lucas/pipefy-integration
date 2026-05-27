package valueobject

import "testing"

func TestNewOutboxEventStatus(t *testing.T) {
	t.Parallel()

	valid := []struct {
		input string
		want  OutboxEventStatus
	}{
		{"pending", OutboxStatusPending},
		{"processing", OutboxStatusProcessing},
		{"done", OutboxStatusDone},
		{"failed", OutboxStatusFailed},
	}

	for _, tc := range valid {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got, err := NewOutboxEventStatus(tc.input)
			if err != nil {
				t.Fatalf("NewOutboxEventStatus(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("NewOutboxEventStatus(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}

	invalid := []string{"", "PENDING", "queued", "done ", "unknown"}

	for _, v := range invalid {
		t.Run("invalid_"+v, func(t *testing.T) {
			t.Parallel()
			_, err := NewOutboxEventStatus(v)
			if err == nil {
				t.Errorf("NewOutboxEventStatus(%q) expected error; got nil", v)
			}
		})
	}
}

func TestOutboxEventStatusIsValid(t *testing.T) {
	t.Parallel()

	for _, s := range []OutboxEventStatus{OutboxStatusPending, OutboxStatusProcessing, OutboxStatusDone, OutboxStatusFailed} {
		if !s.IsValid() {
			t.Errorf("%q.IsValid() = false; want true", s)
		}
	}

	if OutboxEventStatus("unknown").IsValid() {
		t.Error(`OutboxEventStatus("unknown").IsValid() = true; want false`)
	}
}

func TestOutboxEventStatusString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		status OutboxEventStatus
		want   string
	}{
		{OutboxStatusPending, "pending"},
		{OutboxStatusProcessing, "processing"},
		{OutboxStatusDone, "done"},
		{OutboxStatusFailed, "failed"},
	}

	for _, tc := range cases {
		if tc.status.String() != tc.want {
			t.Errorf("%q.String() = %q; want %q", tc.status, tc.status.String(), tc.want)
		}
	}
}
