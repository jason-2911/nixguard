package nftables

import "testing"

func TestIsPrivilegeError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "operation not permitted",
			err:  errorString("command nft exited with 1: Operation not permitted"),
			want: true,
		},
		{
			name: "permission denied",
			err:  errorString("command tcpdump exited with 1: Permission denied"),
			want: true,
		},
		{
			name: "syntax error",
			err:  errorString("command nft exited with 1: Error: Set member cannot be prefix"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPrivilegeError(tt.err); got != tt.want {
				t.Fatalf("isPrivilegeError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseNFTJSONCounters_FiltersOtherTables(t *testing.T) {
	input := `{
  "nftables": [
    { "counter": { "name": "cnt_keep_me", "table": "nixguard", "packets": 12, "bytes": 1440 } },
    { "counter": { "name": "cnt_ignore_me", "table": "other", "packets": 99, "bytes": 9999 } }
  ]
}`

	stats := parseNFTJSONCounters(input)
	if len(stats) != 1 {
		t.Fatalf("expected 1 counter, got %d", len(stats))
	}

	stat, ok := stats["keep_me"]
	if !ok {
		t.Fatalf("expected keep_me counter")
	}
	if stat.Packets != 12 || stat.Bytes != 1440 {
		t.Fatalf("unexpected stat: %+v", stat)
	}
}

func TestParseNFTJSONCounters_FiltersForeignTables(t *testing.T) {
	input := `{
		"nftables": [
			{"counter": {"name":"cnt_rule1","table":"nixguard","packets":12,"bytes":345}},
			{"counter": {"name":"cnt_rule2","table":"other","packets":99,"bytes":1000}}
		]
	}`

	stats := parseNFTJSONCounters(input)
	if len(stats) != 1 {
		t.Fatalf("expected 1 counter, got %d", len(stats))
	}
	if stats["rule1"].Packets != 12 {
		t.Fatalf("expected rule1 packets 12, got %d", stats["rule1"].Packets)
	}
	if _, ok := stats["rule2"]; ok {
		t.Fatal("did not expect counters from non-nixguard table")
	}
}

type errorString string

func (e errorString) Error() string {
	return string(e)
}
