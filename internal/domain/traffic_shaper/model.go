// Package traffic_shaper contains the domain model for QoS and traffic shaping.
// Maps to OPNsense: Firewall > Traffic Shaper, Pipes, Queues.
package traffic_shaper

// Pipe defines a bandwidth limit (dummynet pipe equivalent).
type Pipe struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Bandwidth   int    `json:"bandwidth" db:"bandwidth"`     // Kbit/s
	BurstSize   int    `json:"burst_size" db:"burst_size"`   // KB
	Delay       int    `json:"delay" db:"delay"`             // ms
	PacketLoss  int    `json:"packet_loss" db:"packet_loss"` // percent
	Scheduler   string `json:"scheduler" db:"scheduler"`     // fifo, wf2q+, fq_codel
	Mask        string `json:"mask" db:"mask"`               // src-ip, dst-ip, none
	Enabled     bool   `json:"enabled" db:"enabled"`
	Description string `json:"description" db:"description"`
}

// Queue is a weighted sub-queue within a pipe.
type Queue struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	PipeID      string `json:"pipe_id" db:"pipe_id"`
	Weight       int    `json:"weight" db:"weight"` // 1-100
	Priority    int    `json:"priority" db:"priority"` // 0-7
	QueueSize   int    `json:"queue_size" db:"queue_size"` // slots
	Mask        string `json:"mask" db:"mask"`
	Enabled     bool   `json:"enabled" db:"enabled"`
	Description string `json:"description" db:"description"`
}

// ShaperRule assigns traffic to a pipe/queue.
type ShaperRule struct {
	ID          string `json:"id" db:"id"`
	Sequence    int    `json:"sequence" db:"sequence"`
	Interface   string `json:"interface" db:"interface_name"`
	Direction   string `json:"direction" db:"direction"` // in, out
	Protocol    string `json:"protocol" db:"protocol"`
	Source      string `json:"source" db:"source"`
	Destination string `json:"destination" db:"destination"`
	TargetPipe  string `json:"target_pipe" db:"target_pipe"`
	TargetQueue string `json:"target_queue" db:"target_queue"`
	DSCPMark    string `json:"dscp_mark" db:"dscp_mark"` // set DSCP value
	Enabled     bool   `json:"enabled" db:"enabled"`
	Description string `json:"description" db:"description"`
}

// ShaperConfig defines the global traffic shaping strategy per interface.
type ShaperConfig struct {
	ID            string `json:"id" db:"id"`
	Interface     string `json:"interface" db:"interface_name"`
	UploadBW      int    `json:"upload_bw" db:"upload_bw"`     // Kbit/s
	DownloadBW    int    `json:"download_bw" db:"download_bw"` // Kbit/s
	Scheduler     string `json:"scheduler" db:"scheduler"`     // htb, hfsc, cake, fq_codel
	Enabled       bool   `json:"enabled" db:"enabled"`
}
