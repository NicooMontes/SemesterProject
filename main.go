package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Metadatos para enviar al frontend
type FileMeta struct {
	FileID     int64     `json:"file_id"`
	Name       string    `json:"name"`
	Hash       string    `json:"hash"`
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploadedAt"`
}

var db *sql.DB

func main() {

	initDB()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend.html")
	})

	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/files", handleListFiles)
	http.HandleFunc("/download/", handleDownload)
	http.HandleFunc("/delete/", handleDelete)

	fmt.Println("Running on https://localhost:8443")
	http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", nil)
}

func initDB() {
	var err error

	dsn := "root:CEGroup1@tcp(127.0.0.1:3306)/cloud_storage_app?parseTime=true"

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to MySQL")
}

// ======================================
// POST /upload
// ======================================
func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file content", http.StatusInternalServerError)
		return
	}

	// ====== CALCULAR HASH ======
	hash := sha256.Sum256(fileBytes)
	hashHex := hex.EncodeToString(hash[:])

	// ====== GUARDAR ARCHIVO EN /uploads ======
	uploadPath := "uploads/" + header.Filename

	err = os.WriteFile(uploadPath, fileBytes, 0644)
	if err != nil {
		log.Println("ERROR writing file:", err)
		http.Error(w, "Failed to save file locally", http.StatusInternalServerError)
		return
	}

	// ====== COMPROBAR SI EXISTE EN LA BBDD ======
	var existingHash string
	err = db.QueryRow(`
        SELECT hash FROM files 
        WHERE filename = ? AND user_id = ?
    `, header.Filename, 1).Scan(&existingHash)

	if err != nil {
		if err == sql.ErrNoRows {
			// ===== INSERT NUEVO =====
			_, err = db.Exec(`
	            INSERT INTO files (user_id, size, filename, hash, version)
	            VALUES (?, ?, ?, ?, 1)
	        `, 1, len(fileBytes), header.Filename, hashHex)

			if err != nil {
				log.Println("DB INSERT ERROR:", err)
				http.Error(w, "Failed to insert metadata", 500)
				return
			}

		} else {
			// Error real de MySQL
			log.Println("DB SELECT ERROR:", err)
			http.Error(w, "Database error", 500)
			return
		}

	} else {
		// ===== UPDATE =====
		_, err = db.Exec(`
	        UPDATE files
	        SET size = ?, hash = ?, update_date = NOW(), version = version + 1
	        WHERE filename = ? AND user_id = ?
	    `, len(fileBytes), hashHex, header.Filename, 1)

		if err != nil {
			log.Println("DB UPDATE ERROR:", err)
			http.Error(w, "Failed to update metadata", 500)
			return
		}
	}

	// ===== RESPUESTA =====
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message":  "File uploaded successfully",
		"filename": header.Filename,
		"hash":     hashHex,
	})
}

// ======================================
// GET /files
// ======================================
func handleListFiles(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query(`
       SELECT file_id, size, filename, upload_date
       FROM files
       WHERE user_id = 1
       ORDER BY upload_date DESC
    `)

	if err != nil {
		log.Println("DB ERROR in /files:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []FileMeta

	for rows.Next() {
		var m FileMeta
		var uploadTime time.Time

		// IMPORTANTE: filename va en el medio
		if err := rows.Scan(&m.FileID, &m.Size, &m.Name, &uploadTime); err != nil {
			log.Println("SCAN ERROR:", err)
			continue
		}

		m.UploadedAt = uploadTime
		list = append(list, m)
	}

	log.Println("FILES RETURNED:", list)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// ======================================
// GET /download/{file_id}
// ======================================
func handleDownload(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/download/")
	if id == "" {
		http.Error(w, "Missing file ID", http.StatusBadRequest)
		return
	}

	var filename string
	err := db.QueryRow(`
        SELECT filename FROM files WHERE file_id = ?
    `, id).Scan(&filename)

	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	filePath := "uploads/" + filename
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(fileBytes)
}

// ======================================
// DELETE /delete/{file_id}
// ======================================
func handleDelete(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/delete/")
	if id == "" {
		http.Error(w, "Missing file ID", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`DELETE FROM files WHERE file_id = ?`, id)

	if err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "File deleted successfully",
	})
}
