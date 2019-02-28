package main

import (
	"time"
)

type Response struct {
	OK           bool   `json:"ok"`
	ErrorMessage string `json:"errorMessage"`
}

// Access Control List
// ETH address string -> role
type ACL = map[string]int

const (
	RoleNone    = 0 // Can't do anything
	RoleOwner   = 1 // Can manage viewers and managers
	RoleManager = 2 // Can manage viewers
	RoleViewer  = 3 // Can only view documents
)

// Internal data structure for FileMeta
type FileMeta struct {
	Hash        string    `json:"hash"`
	IP          string    `json:"ip"`
	Timestamp   time.Time `json:"timestamp"`
	ACL         ACL       `json:"acl"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"contentType"`
	ContentSize int64     `json:"contentSize"`
}

// Public version of FileMeta for end-user usage
type FileMetaPublic struct {
	Hash        string    `json:"hash"`
	Timestamp   time.Time `json:"timestamp"`
	ACL         ACL       `json:"acl"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"contentType"`
	ContentSize int64     `json:"contentSize"`
}
