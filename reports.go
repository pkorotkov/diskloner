package main

import "time"

type diskProfile struct {
	PartitionTableType string `json:"partition_table_type"`
	Model              string `json:"model"`
	SerialNumber       string `json:"serial_number"`
	PhysicalSectorSize int    `json:"physical_sector_size"`
	LogicalSectorSize  int    `json:"logical_sector_size"`
	Capacity           int64  `json:"capacity"`
}

type hashes struct {
	MD5Hash    string `json:"md5"`
	SHA1Hash   string `json:"sha1"`
	SHA256Hash string `json:"sha256"`
	SHA512Hash string `json:"sha512"`
}

type CloningReport struct {
	Name                 string      `json:"name"`
	UUID                 string      `json:"uuid"`
	StartTime            time.Time   `json:"start_time"`
	EndTime              time.Time   `json:"end_time"`
	BlockDeviceType      string      `json:"block_device_type"`
	DiskProfile          diskProfile `json:"disk_profile"`
	Hashes               hashes      `json:"hashes"`
	UnreadLogicalSectors []int64     `json:"unread_logical_sectors"`
}
