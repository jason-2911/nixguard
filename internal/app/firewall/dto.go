package firewall

import "github.com/nixguard/nixguard/internal/domain/firewall"

// CreateRuleInput is the DTO for creating a firewall rule.
type CreateRuleInput struct {
	Interface   string             `json:"interface"`
	Direction   firewall.Direction `json:"direction" validate:"required,oneof=in out"`
	Action      firewall.Action    `json:"action" validate:"required,oneof=pass block reject"`
	Protocol    firewall.Protocol  `json:"protocol" validate:"required"`
	Source      firewall.Address   `json:"source" validate:"required"`
	Destination firewall.Address   `json:"destination" validate:"required"`
	Log         bool               `json:"log"`
	Description string             `json:"description"`
	Order       int                `json:"order"`
	IsFloating  bool               `json:"is_floating"`
	Interfaces  []string           `json:"interfaces"`
	Gateway     string             `json:"gateway"`
	Category    string             `json:"category"`
	StateType   string             `json:"state_type"`
	MaxStates   int                `json:"max_states"`
	Tag         string             `json:"tag"`
	Tagged      string             `json:"tagged"`
	Schedule    *firewall.Schedule `json:"schedule"`
}

// UpdateRuleInput is the DTO for updating a firewall rule (partial update).
type UpdateRuleInput struct {
	Interface   *string             `json:"interface,omitempty"`
	Direction   *firewall.Direction `json:"direction,omitempty"`
	Action      *firewall.Action    `json:"action,omitempty"`
	Protocol    *firewall.Protocol  `json:"protocol,omitempty"`
	Source      *firewall.Address   `json:"source,omitempty"`
	Destination *firewall.Address   `json:"destination,omitempty"`
	Enabled     *bool               `json:"enabled,omitempty"`
	Description *string             `json:"description,omitempty"`
	Log         *bool               `json:"log,omitempty"`
	Order       *int                `json:"order,omitempty"`
	IsFloating  *bool               `json:"is_floating,omitempty"`
	Interfaces  *[]string           `json:"interfaces,omitempty"`
	Gateway     *string             `json:"gateway,omitempty"`
	Category    *string             `json:"category,omitempty"`
	StateType   *string             `json:"state_type,omitempty"`
	MaxStates   *int                `json:"max_states,omitempty"`
	Tag         *string             `json:"tag,omitempty"`
	Tagged      *string             `json:"tagged,omitempty"`
	Schedule    *firewall.Schedule  `json:"schedule,omitempty"`
}

// CreateAliasInput is the DTO for creating a firewall alias.
type CreateAliasInput struct {
	Name        string             `json:"name" validate:"required,alphanum"`
	Type        firewall.AliasType `json:"type" validate:"required"`
	Description string             `json:"description"`
	Entries     []string           `json:"entries" validate:"required,min=1"`
	UpdateFreq  string             `json:"update_freq"`
}

// UpdateAliasInput is the DTO for updating a firewall alias.
type UpdateAliasInput struct {
	Name        *string             `json:"name,omitempty"`
	Type        *firewall.AliasType `json:"type,omitempty"`
	Description *string             `json:"description,omitempty"`
	Entries     *[]string           `json:"entries,omitempty"`
	UpdateFreq  *string             `json:"update_freq,omitempty"`
	Enabled     *bool               `json:"enabled,omitempty"`
}

// CreateNATInput is the DTO for creating a NAT rule.
type CreateNATInput struct {
	Type           firewall.NATType  `json:"type" validate:"required"`
	Interface      string            `json:"interface" validate:"required"`
	Protocol       firewall.Protocol `json:"protocol" validate:"required"`
	Source         firewall.Address  `json:"source"`
	Destination    firewall.Address  `json:"destination"`
	RedirectTarget string            `json:"redirect_target" validate:"required"`
	RedirectPort   string            `json:"redirect_port"`
	Description    string            `json:"description"`
	NATReflection  bool              `json:"nat_reflection"`
}

// UpdateNATInput is the DTO for updating an existing NAT rule.
type UpdateNATInput struct {
	Type           *firewall.NATType  `json:"type,omitempty"`
	Interface      *string            `json:"interface,omitempty"`
	Protocol       *firewall.Protocol `json:"protocol,omitempty"`
	Source         *firewall.Address  `json:"source,omitempty"`
	Destination    *firewall.Address  `json:"destination,omitempty"`
	RedirectTarget *string            `json:"redirect_target,omitempty"`
	RedirectPort   *string            `json:"redirect_port,omitempty"`
	Description    *string            `json:"description,omitempty"`
	NATReflection  *bool              `json:"nat_reflection,omitempty"`
	Enabled        *bool              `json:"enabled,omitempty"`
}

// TrafficCaptureInput controls live packet capture and pcap export.
type TrafficCaptureInput struct {
	Interface string `json:"interface"`
	SourceIP  string `json:"source_ip"`
	DestIP    string `json:"dest_ip"`
	Protocol  string `json:"protocol"`
	Count     int    `json:"count"`
	SnapLen   int    `json:"snap_len"`
}
