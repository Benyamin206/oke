package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

// Middleware authorize yang membungkus handler untuk otorisasi
func authorize(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Meneruskan request ke handler berikutnya
		next.ServeHTTP(w, r)
	}
}

// Handler untuk mendapatkan semua jadwal
func getAllJadwal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		// Memeriksa apakah metode HTTP adalah GET
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Membuka koneksi ke database MySQL
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_jadwal")
	if err != nil {
		// Mengembalikan kesalahan jika koneksi gagal
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Menjalankan query untuk mendapatkan semua jadwal
	rows, err := db.Query("SELECT * FROM jadwals")
	if err != nil {
		// Mengembalikan kesalahan jika query gagal
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Menyimpan hasil query ke dalam slice of map
	var jadwals []map[string]string
	for rows.Next() {
		var id, kapal_id, nahkoda_id, rute_id, waktu_berangkat, stok, harga string
		if err := rows.Scan(&id, &kapal_id, &nahkoda_id, &rute_id, &waktu_berangkat, &stok, &harga); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Menambahkan hasil query ke dalam slice jadwals
		jadwal := map[string]string{
			"id":              id,
			"kapal_id":        kapal_id,
			"nahkoda_id":      nahkoda_id,
			"rute_id":         rute_id,
			"waktu_berangkat": waktu_berangkat,
			"stok":            stok,
			"harga":           harga,
		}
		jadwals = append(jadwals, jadwal)
	}
	// Mengirimkan response dalam format JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jadwals)
}

// Handler untuk mendapatkan jadwal berdasarkan ID
func getJadwalByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse parameter ID dari URL
	id := r.URL.Query().Get("id")
	if id == "" {
		// Mengembalikan kesalahan jika parameter ID tidak ada
		http.Error(w, "Parameter ID is required", http.StatusBadRequest)
		return
	}

	// Buka koneksi ke database
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_jadwal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query untuk mendapatkan jadwal berdasarkan ID
	row := db.QueryRow("SELECT * FROM jadwals WHERE id = ?", id)

	// Buat variabel untuk menyimpan data kapal
	var (
		jadwalID, kapal_id, nahkoda_id, rute_id, waktu_berangkat, stok, harga string
	)

	// Scan hasil query ke variabel kapal
	err = row.Scan(&jadwalID, &kapal_id, &nahkoda_id, &rute_id, &waktu_berangkat, &stok, &harga)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Jadwal not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Buat peta untuk mengirimkan sebagai respons JSON
	// Membuat peta untuk mengirimkan sebagai response JSON
	kapal := map[string]string{
		"id":              jadwalID,
		"kapal_id":        kapal_id,
		"nahkoda_id":      nahkoda_id,
		"rute_id":         rute_id,
		"waktu_berangkat": waktu_berangkat,
		"stok":            stok,
		"harga":           harga,
	}

	// Set header dan kirimkan response JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kapal)
}

// Handler untuk membuat jadwal baru
func createJadwal(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse data kapal dari body request
	var jadwal map[string]string
	err := json.NewDecoder(r.Body).Decode(&jadwal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Buka koneksi ke database
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_jadwal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Menjalankan query untuk insert data jadwal baru
	// Eksekusi query untuk insert data kapal baru
	result, err := db.Exec("INSERT INTO jadwals (kapal_id, nahkoda_id, rute_id, waktu_berangkat, stok, harga) VALUES (?, ?, ?, ?, ?, ?)", jadwal["kapal_id"], jadwal["nahkoda_id"], jadwal["rute_id"], jadwal["waktu_berangkat"], jadwal["stok"], jadwal["harga"])
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

func updateStok(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID  string `json:"id"`
		Qty string `json:"qty"`
	}

	// Parse JSON body
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.ID == "" || input.Qty == "" {
		http.Error(w, "Parameters ID and qty are required", http.StatusBadRequest)
		return
	}

	// Membuka koneksi ke database
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_jadwal")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Menjalankan query untuk memperbarui stok jadwal
	_, err = db.Exec("UPDATE jadwals SET stok = ? WHERE id = ?", input.Qty, input.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Mengirimkan response sukses
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Stok updated successfully"})
}

// Fungsi utama untuk menjalankan server
func main() {
	// Membuat multiplexer untuk menangani berbagai endpoint
	var mux = http.NewServeMux()

	// login

	// Route untuk mendapatkan semua pengguna
	// Menetapkan handler untuk berbagai endpoint dengan middleware authorize
	mux.HandleFunc("/get-all-jadwal", authorize(getAllJadwal))

	mux.HandleFunc("/create-jadwal", authorize(createJadwal)) // Menambahkan handler untuk create-kapal

	mux.HandleFunc("/get-jadwal-by-id", authorize(getJadwalByID))

	mux.HandleFunc("/update-stok", authorize(updateStok))

	fmt.Println("jadwal server running on port : 9012")

	// Jalankan server HTTP pada port 9010

	http.ListenAndServe(":9012", mux)
}
