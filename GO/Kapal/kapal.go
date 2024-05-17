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
		next.ServeHTTP(w, r)
	}
}

func getAllKapal(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_kapal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, nama, deskripsi, pemilik_kapal_id FROM kapals")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var kapals []map[string]string
	for rows.Next() {
		var id, nama, deskripsi, pemilik_kapal_id string
		if err := rows.Scan(&id, &nama, &deskripsi, &pemilik_kapal_id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		kapal := map[string]string{
			"id":               id,
			"nama":             nama,
			"deskripsi":        deskripsi,
			"pemilik_kapal_id": pemilik_kapal_id,
		}
		kapals = append(kapals, kapal)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kapals)
}

func getKapalByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// Parse parameter ID dari URL
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Parameter ID is required", http.StatusBadRequest)
		return
	}

	// Buka koneksi ke database
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_kapal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query untuk mendapatkan kapal berdasarkan ID
	row := db.QueryRow("SELECT id, nama, deskripsi, pemilik_kapal_id FROM kapals WHERE id = ?", id)

	// Buat variabel untuk menyimpan data kapal
	var (
		kapalID, nama, deskripsi, pemilikKapalID string
	)

	// Scan hasil query ke variabel kapal
	err = row.Scan(&kapalID, &nama, &deskripsi, &pemilikKapalID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Kapal not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Buat peta untuk mengirimkan sebagai respons JSON
	kapal := map[string]string{
		"id":               kapalID,
		"nama":             nama,
		"deskripsi":        deskripsi,
		"pemilik_kapal_id": pemilikKapalID,
	}

	// Set header dan kirimkan response JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kapal)
}

func createKapal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse data kapal dari body request
	var kapal map[string]string
	err := json.NewDecoder(r.Body).Decode(&kapal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Buka koneksi ke database
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_kapal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Eksekusi query untuk insert data kapal baru
	result, err := db.Exec("INSERT INTO kapals (nama, deskripsi, pemilik_kapal_id) VALUES (?, ?, ?)", kapal["nama"], kapal["deskripsi"], kapal["pemilik_kapal_id"])
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

func updateKapal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID        string `json:"id"`
		Nama      string `json:"nama"`
		Deskripsi string `json:"deskripsi"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_kapal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "UPDATE kapals SET nama = ?, deskripsi = ? WHERE id = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(input.Nama, input.Deskripsi, input.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Kapal updated successfully"))
}

func getKapalsByPemilikKapalId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse parameter pemilik_kapal_id dari URL
	pemilikKapalID := r.URL.Query().Get("pemilik_kapal_id")
	if pemilikKapalID == "" {
		http.Error(w, "Parameter pemilik_kapal_id is required", http.StatusBadRequest)
		return
	}

	// Buka koneksi ke database
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_kapal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query untuk mendapatkan kapal berdasarkan pemilik_kapal_id
	rows, err := db.Query("SELECT id, nama, deskripsi, pemilik_kapal_id FROM kapals WHERE pemilik_kapal_id = ?", pemilikKapalID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Buat slice untuk menyimpan data kapal
	var kapals []map[string]string
	for rows.Next() {
		var (
			kapalID, nama, deskripsi, pemilikKapalID string
		)
		if err := rows.Scan(&kapalID, &nama, &deskripsi, &pemilikKapalID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		kapal := map[string]string{
			"id":               kapalID,
			"nama":             nama,
			"deskripsi":        deskripsi,
			"pemilik_kapal_id": pemilikKapalID,
		}
		kapals = append(kapals, kapal)
	}

	// Set header dan kirimkan response JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kapals)
}

func main() {
	var mux = http.NewServeMux()

	// login
	// Route untuk mendapatkan semua pengguna
	mux.HandleFunc("/get-all-kapal", authorize(getAllKapal))
	mux.HandleFunc("/create-kapal", authorize(createKapal)) // Menambahkan handler untuk create-kapal

	// Route untuk mendapatkan kapal berdasarkan ID
	mux.HandleFunc("/get-kapal-by-id", authorize(getKapalByID))

	mux.HandleFunc("/update-kapal", authorize(updateKapal))

	mux.HandleFunc("/get-kapals-by-pemilik-kapal-id", getKapalsByPemilikKapalId)

	fmt.Println("kapal server running on port : 9010")

	// Jalankan server HTTP pada port 9010
	http.ListenAndServe(":9010", mux)
}
