package main

type FileMetadata struct {
    FileID     int
    Size       int
    Name       string
    UploadTime time.Time
    Priority   int
    Storage    string
    Region     string
}
