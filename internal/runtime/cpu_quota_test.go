package runtime

import "testing"

func TestWindowsCPUQuotaRateNormalizesAcrossLogicalCPUs(t *testing.T) {
	tests := []struct {
		name   string
		quota  int64
		cores  int64
		want   uint32
		failed bool
	}{
		{name: "one core", quota: 100000, cores: 1, want: 10000},
		{name: "half of eight cores", quota: 400000, cores: 8, want: 5000},
		{name: "small quota rounds to minimum", quota: 1, cores: 8, want: 1},
		{name: "above machine capacity", quota: 800001, cores: 8, failed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := windowsCPUQuotaRate(tt.quota, tt.cores)
			if tt.failed {
				if err == nil {
					t.Fatalf("expected capacity error, got rate %d", got)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("rate=%d err=%v, want %d", got, err, tt.want)
			}
		})
	}
}
