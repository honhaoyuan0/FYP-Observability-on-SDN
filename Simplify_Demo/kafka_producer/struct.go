package main

type Flow struct {
	IngressPort int64 `json:"ingress_port"`
	EgressPort  int64 `json:"egress_port"`
	VlanID      int64 `json:"vlan_id"`
	Bytes       int64 `json:"bytes"`
}
