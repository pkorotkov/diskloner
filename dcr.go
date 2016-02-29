package main

import "time"

type DiskCloningReport struct {
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	BenchPath    string    `json:"bench_path"`
	Type         string    `json:"type"`
	SerialNumber string    `json:"serial_number"`
	FullCapacity int64     `json:"full_capacity"`
	MD5Hash      string    `json:"md5_hash"`
	SHA1Hash     string    `json:"sha1_hash"`
	SHA256Hash   string    `json:"sha256_hash"`
	SHA512Hash   string    `json:"sha512_hash"`
}
