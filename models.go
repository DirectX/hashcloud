package main

import (
	"time"
)

type UploadRequestMeta struct {
	Owner     string    `json:"owner"`
	Signature string    `json:"signature"`
}

// Access Control List
// ETH address string -> role
type ACL = map[string]int

const (
	RoleOwner   = 1 // Can manage viewers and managers
	RoleManager = 2 // Can manage viewers
	RoleViewer  = 3 // Can only view documents
)

type FileMeta struct {
	IP          string    `json:"ip"`
	Timestamp   time.Time `json:"timestamp"`
	ACL         ACL       `json:"acl"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content-type"`
}
