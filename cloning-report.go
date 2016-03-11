package main

import "time"

type diskProfile struct {
	Type               string `json:"type"`
	SerialNumber       string `json:"serial_number"`
	PhysicalSectorSize int32  `json:"physical_sector_size"`
	LogicalSectorSize  int32  `json:"logical_sector_size"`
	Capacity           int64  `json:"capacity"`
}

type hashes struct {
	MD5Hash    string `json:"md5"`
	SHA1Hash   string `json:"sha1"`
	SHA256Hash string `json:"sha256"`
	SHA512Hash string `json:"sha512"`
}

type cloningReport struct {
	SessionName          string      `json:"session_name"`
	StartTime            time.Time   `json:"start_time"`
	EndTime              time.Time   `json:"end_time"`
	DiskProfile          diskProfile `json:"disk_profile"`
	Hashes               hashes      `json:"hashes"`
	UnreadLogicalSectors []int64     `json:"unread_logical_sectors"`
}
