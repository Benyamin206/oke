package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func authorize(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// authorization := r.Header.Get("Authorization")

		// if authorization != "user_tertentu" {
		// 	http.Error(w, "Unauthorized (Anda Tidak Punya Akses)", http.StatusUnauthorized)
		// 	return
		// }
		next.ServeHTTP(w, r)
	}
}

func getAllRute(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_rute")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, lokasi_berangkat, lokasi_tujuan FROM rutes")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var rutes []map[string]string
	for rows.Next() {
		var id, lokasi_berangkat, lokasi_tujuan string
		if err := rows.Scan(&id, &lokasi_berangkat, &lokasi_tujuan); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		rute := map[string]string{
			"id":               id,
			"lokasi_berangkat": lokasi_berangkat,
			"lokasi_tujuan":    lokasi_tujuan,
		}
		rutes = append(rutes, rute)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rutes)
}

func getRuteByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Parse parameter ID dari URL
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Parameter ID is required", http.StatusBadRequest)
		return
	}

	// Buka koneksi ke database
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_rute")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query untuk mendapatkan kapal berdasarkan ID
	row := db.QueryRow("SELECT id, lokasi_berangkat, lokasi_tujuan FROM rutes WHERE id = ?", id)

	// Buat variabel untuk menyimpan data kapal
	var (
		ruteID, lokasi_berangkat, lokasi_tujuan string
	)

	// Scan hasil query ke variabel rute
	err = row.Scan(&ruteID, &lokasi_berangkat, &lokasi_tujuan)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "rute not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Buat peta untuk mengirimkan sebagai respons JSON
	rute := map[string]string{
		"id":               ruteID,
		"lokasi_berangkat": lokasi_berangkat,
		"lokasi_tujuan":    lokasi_tujuan,
	}

	// Set header dan kirimkan response JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rute)
}

func createRute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse data kapal dari body request
	var rute map[string]string
	err := json.NewDecoder(r.Body).Decode(&rute)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Buka koneksi ke database
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_rute")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Eksekusi query untuk insert data kapal baru
	result, err := db.Exec("INSERT INTO rutes (lokasi_berangkat, lokasi_tujuan) VALUES (?, ?)", rute["lokasi_berangkat"], rute["lokasi_tujuan"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ambil ID kapal yang baru saja dibuat
	id, _ := result.LastInsertId()

	// Kirim respons dengan ID kapal yang baru saja dibuat
	w.Header().Set("Content-Type", "application/json")
	response := map[string]int64{"id": id}
	json.NewEncoder(w).Encode(response)
}

func updateRute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID              string `json:"id"`
		LokasiBerangkat string `json:"lokasi_berangkat"`
		LokasiTujuan    string `json:"lokasi_tujuan"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_rute")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "UPDATE rutes SET lokasi_berangkat = ?, lokasi_tujuan = ? WHERE id = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(input.LokasiBerangkat, input.LokasiTujuan, input.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Rute updated successfully"))
}

func main() {
	var mux = http.NewServeMux()

	// login
	// Route untuk mendapatkan semua pengguna
	mux.HandleFunc("/get-all-rute", authorize(getAllRute))
	mux.HandleFunc("/create-rute", authorize(createRute)) // Menambahkan handler untuk create-kapal

	// Route untuk mendapatkan kapal berdasarkan ID
	mux.HandleFunc("/get-rute-by-id", authorize(getRuteByID))

	mux.HandleFunc("/update-rute", authorize(updateRute))

	fmt.Println("rute server running on port : 9006")

	// Jalankan server HTTP pada port 9004
	http.ListenAndServe(":9006", mux)
}
