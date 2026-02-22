package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type Book struct {
	ID_Buku int    `json:"id_buku"`
	Judul   string `json:"judul"`
	Author  string `json:"author"`
	Stock   int    `json:"stock"`
}

type History struct {
	ID           int    `json:"id"`
	ID_Buku      int    `json:"id_buku"`
	JudulBuku    string `json:"judul_buku"`
	BorrowerName string `json:"borrower_name"`
	BorrowDate   string `json:"borrow_date"`
	Status       string `json:"status"`
}

func main() {
	connStr := "user=postgres password=123456 dbname=bookingSystem sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Gagal buka koneksi:", err)
	}
	defer db.Close()
	fmt.Println("Sukses connect db PAK")

	http.HandleFunc("/api/books", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "GET" {
			rows, err := db.Query("SELECT id_buku, judul, author, stock FROM books ORDER BY id_buku DESC")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var books []Book
			for rows.Next() {
				var b Book
				if err := rows.Scan(&b.ID_Buku, &b.Judul, &b.Author, &b.Stock); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				books = append(books, b)
			}
			json.NewEncoder(w).Encode(books)
			return
		}

		if r.Method == "POST" {
			var newBook Book
			if err := json.NewDecoder(r.Body).Decode(&newBook); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			err := db.QueryRow("INSERT INTO books (judul, author, stock) VALUES ($1, $2, $3) RETURNING id_buku",
				newBook.Judul, newBook.Author, newBook.Stock).Scan(&newBook.ID_Buku)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(newBook)
			return
		}
	})

	http.HandleFunc("/api/book", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var request struct {
			ID_Buku      int    `json:"id_buku"`
			BorrowerName string `json:"borrower_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := db.Exec("UPDATE books SET stock = stock - 1 WHERE id_buku = $1 AND stock > 0", request.ID_Buku)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Gagal meminjam: Stok habis atau buku tidak ditemukan", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("INSERT INTO borrow_history (id_buku, borrower_name, status) VALUES ($1, $2, 'Dipinjam')", request.ID_Buku, request.BorrowerName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "Buku berhasil dipinjam!"})
	})

	http.HandleFunc("/api/return", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var request struct {
			ID_Buku int `json:"id_buku"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := db.Exec("UPDATE books SET stock = stock + 1 WHERE id_buku = $1", request.ID_Buku)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Gagal mengembalikan: buku tidak ditemukan", http.StatusBadRequest)
			return
		}

		updateHistoryQuery := `
			UPDATE borrow_history 
			SET status = 'Dikembalikan' 
			WHERE id = (
				SELECT id FROM borrow_history 
				WHERE id_buku = $1 AND status = 'Dipinjam' 
				ORDER BY borrow_date ASC LIMIT 1
			)
		`
		db.Exec(updateHistoryQuery, request.ID_Buku)

		json.NewEncoder(w).Encode(map[string]string{"message": "Buku berhasil dikembalikan!"})
	})

	http.HandleFunc("/api/book/delete", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var request struct {
			ID_Buku int `json:"id_buku"`
		}
		json.NewDecoder(r.Body).Decode(&request)

		_, err := db.Exec("DELETE FROM books WHERE id_buku = $1", request.ID_Buku)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "Buku berhasil dihapus!"})
	})

	http.HandleFunc("/api/book/update", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var b Book
		json.NewDecoder(r.Body).Decode(&b)

		_, err := db.Exec("UPDATE books SET judul = $1, author = $2, stock = $3 WHERE id_buku = $4",
			b.Judul, b.Author, b.Stock, b.ID_Buku)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "Buku berhasil diperbarui!"})
	})

	http.HandleFunc("/api/history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		query := `
			SELECT h.id, h.id_buku, b.judul, h.borrower_name, h.borrow_date, h.status 
			FROM borrow_history h 
			JOIN books b ON h.id_buku = b.id_buku 
			ORDER BY h.id DESC
		`
		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var histories []History
		for rows.Next() {
			var h History
			if err := rows.Scan(&h.ID, &h.ID_Buku, &h.JudulBuku, &h.BorrowerName, &h.BorrowDate, &h.Status); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			histories = append(histories, h)
		}
		json.NewEncoder(w).Encode(histories)
	})

	fmt.Println("Server jalan di http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}