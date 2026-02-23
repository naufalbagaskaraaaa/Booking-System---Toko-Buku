import { useState, useEffect } from 'react'

const BASE_URL = 'https://bukusukamaju.user.cloudjkt01.com'

function App() {
  const [books, setBooks] = useState([])
  const [histories, setHistories] = useState([])
  const [newJudul, setNewJudul] = useState("")
  const [newAuthor, setNewAuthor] = useState("")
  const [newStock, setNewStock] = useState(1)
  const [isEditMode, setIsEditMode] = useState(false)
  const [editID, setEditID] = useState(null)
  const [searchTerm, setSearchTerm] = useState("")
  const [activePage, setActivePage] = useState("books")

  const fetchBooks = () => {
    fetch(`${BASE_URL}/api/books`)
      .then(async response => response.json())
      .then(data => setBooks(data === null ? [] : data))
      .catch(error => console.error("Error fetch:", error))
  }

  const fetchHistories = () => {
    fetch(`${BASE_URL}/api/history`)
      .then(async response => response.json())
      .then(data => setHistories(data === null ? [] : data))
      .catch(error => console.error("Error fetch history:", error))
  }

  useEffect(() => {
    fetchBooks()
    fetchHistories()
  }, [])

  const handleAddOrUpdateBook = (e) => {
    e.preventDefault()

    if (!newJudul || !newAuthor || newStock < 1) {
      alert("Mohon isi semua data dengan benar!")
      return
    }

    if (isEditMode) {
      const updatedData = {
        id_buku: editID,
        judul: newJudul,
        author: newAuthor,
        stock: parseInt(newStock)
      }

      fetch(`${BASE_URL}/api/book/update`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updatedData)
      })
      .then(async response => {
        if (!response.ok) throw new Error(await response.text())
        return response.json()
      })
      .then(() => {
        alert("Buku berhasil diperbarui!")
        resetForm()
        fetchBooks()
      })
      .catch(error => alert(`Error Update: ${error.message}`))
    } else {
      const newBookData = {
        judul: newJudul,
        author: newAuthor,
        stock: parseInt(newStock)
      }

      fetch(`${BASE_URL}/api/books`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newBookData)
      })
      .then(async response => {
        if (!response.ok) throw new Error(await response.text())
        return response.json()
      })
      .then(() => {
        alert("Buku baru berhasil ditambahkan!")
        resetForm()
        fetchBooks()
      })
      .catch(error => alert(`Error Add: ${error.message}`))
    }
  }

  const resetForm = () => {
    setIsEditMode(false)
    setEditID(null)
    setNewJudul("")
    setNewAuthor("")
    setNewStock(1)
  }

  const startEdit = (book) => {
    setIsEditMode(true)
    setEditID(book.id_buku)
    setNewJudul(book.judul)
    setNewAuthor(book.author)
    setNewStock(book.stock)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const handleDelete = (ID_Buku) => {
    if (!window.confirm("Yakin mau hapus buku ini?")) return

    fetch(`${BASE_URL}/api/book/delete`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id_buku: ID_Buku })
    })
    .then(async response => {
        if (!response.ok) throw new Error(await response.text())
        alert("Buku dihapus!")
        fetchBooks()
    })
    .catch(error => alert(`Error Delete: ${error.message}`))
  }

  const handleBooking = (ID_Buku, Judul) => {
    const borrower = prompt(`Masukkan nama peminjam untuk buku "${Judul}":`)
    if (!borrower) {
      alert("Peminjaman dibatalkan karena nama tidak diisi.")
      return
    }

    fetch(`${BASE_URL}/api/book`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id_buku: ID_Buku, borrower_name: borrower })
    })
    .then(async (response) => {
      if (!response.ok) {
        const errData = await response.text()
        throw new Error(errData)
      }
      return response.json()
    })
    .then(() => {
      alert(`Sukses: ${Judul} berhasil dipinjam oleh ${borrower}!`)
      fetchBooks()
      fetchHistories()
    })
    .catch(error => {
      alert(`Gagal: ${error.message}`)
    })
  }

  const handleReturn = (ID_Buku, Judul) => {
    fetch(`${BASE_URL}/api/return`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id_buku: ID_Buku })
    })
    .then(async (response) => {
      if (!response.ok) {
        const errData = await response.text()
        throw new Error(errData)
      }
      return response.json()
    })
    .then(() => {
      alert(`Berhasil: ${Judul} sudah dikembalikan ke perpustakaan!`)
      fetchBooks()
      fetchHistories()
    })
    .catch(error => {
      alert(`Gagal: ${error.message}`)
    })
  }

  const isBorrowed = (id_buku) => {
    return histories.some(h => h.id_buku === id_buku && h.status === 'Dipinjam')
  }

  const filteredBooks = books.filter(book => 
    book.judul.toLowerCase().includes(searchTerm.toLowerCase()) || 
    book.author.toLowerCase().includes(searchTerm.toLowerCase())
  )

  return (
    <div style={{ fontFamily: "'Inter', sans-serif", minHeight: "100vh", backgroundColor: "#1f2937", color: "#f9fafb" }}>

      <nav style={{ backgroundColor: "#111827", padding: "0 40px", display: "flex", alignItems: "center", gap: "8px", borderBottom: "1px solid #374151", position: "sticky", top: 0, zIndex: 100 }}>
        <span style={{ fontWeight: "700", fontSize: "1rem", marginRight: "32px", color: "#f9fafb" }}>Reservasi Buku</span>

        <button
          onClick={() => setActivePage("books")}
          style={{ padding: "18px 20px", background: "none", border: "none", cursor: "pointer", fontSize: "0.9rem", fontWeight: "600", color: activePage === "books" ? "#10b981" : "#9ca3af", borderBottom: activePage === "books" ? "2px solid #10b981" : "2px solid transparent" }}
        >
          Koleksi Buku
        </button>

        <button
          onClick={() => setActivePage("history")}
          style={{ padding: "18px 20px", background: "none", border: "none", cursor: "pointer", fontSize: "0.9rem", fontWeight: "600", color: activePage === "history" ? "#10b981" : "#9ca3af", borderBottom: activePage === "history" ? "2px solid #10b981" : "2px solid transparent" }}
        >
          Riwayat ({histories.length})
        </button>
      </nav>

      <div style={{ padding: "40px 20px" }}>
        <div style={{ maxWidth: "750px", margin: "auto" }}>

          {activePage === "books" && (
            <>
              <div style={{ textAlign: "center", marginBottom: "30px" }}>
                <h1 style={{ fontSize: "2.5rem", margin: "0 0 10px 0", fontWeight: "800", letterSpacing: "-1px" }}>Reservasi Buku</h1>
                <p style={{ color: "#9ca3af", margin: 0 }}>BookingSystem - Toko Buku Suka Maju</p>
              </div>

              <div style={{ backgroundColor: "#ffffff", padding: "25px", borderRadius: "12px", marginBottom: "30px", boxShadow: "0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)", color: "#111827" }}>
                <h3 style={{ margin: "0 0 15px 0", fontSize: "1.2rem", color: "#374151" }}>
                  {isEditMode ? "✏️ Edit Buku" : "+ Tambah Koleksi Baru"}
                </h3>
                <form onSubmit={handleAddOrUpdateBook} style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
                  <input type="text" placeholder="Judul Buku (Cth: Bumi Manusia)" value={newJudul} onChange={(e) => setNewJudul(e.target.value)} style={{ padding: "12px", borderRadius: "8px", border: "1px solid #d1d5db", backgroundColor: "#f9fafb", color: "#111827", fontSize: "1rem" }} />
                  <input type="text" placeholder="Nama Penulis" value={newAuthor} onChange={(e) => setNewAuthor(e.target.value)} style={{ padding: "12px", borderRadius: "8px", border: "1px solid #d1d5db", backgroundColor: "#f9fafb", color: "#111827", fontSize: "1rem" }} />
                  <input type="number" placeholder="Jumlah Stok" min="1" value={newStock} onChange={(e) => setNewStock(e.target.value)} style={{ padding: "12px", borderRadius: "8px", border: "1px solid #d1d5db", backgroundColor: "#f9fafb", color: "#111827", fontSize: "1rem" }} />
                  <div style={{ display: "flex", gap: "10px", marginTop: "5px" }}>
                    <button type="submit" style={{ flex: 1, padding: "12px", backgroundColor: isEditMode ? "#6366f1" : "#10b981", color: "white", border: "none", borderRadius: "8px", fontSize: "1rem", fontWeight: "600", cursor: "pointer" }}>
                      {isEditMode ? "Update Buku" : "Simpan Buku"}
                    </button>
                    {isEditMode && (
                      <button type="button" onClick={resetForm} style={{ padding: "12px", backgroundColor: "#ef4444", color: "white", border: "none", borderRadius: "8px", fontSize: "1rem", fontWeight: "600", cursor: "pointer" }}>
                        Batal
                      </button>
                    )}
                  </div>
                </form>
              </div>

              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", borderBottom: "1px solid #374151", paddingBottom: "10px", marginBottom: "20px" }}>
                <h3 style={{ margin: 0, color: "#f3f4f6" }}>Daftar Buku Tersedia</h3>
                <input type="text" placeholder="🔍 Cari judul atau penulis..." value={searchTerm} onChange={(e) => setSearchTerm(e.target.value)} style={{ padding: "8px 15px", borderRadius: "20px", border: "none", backgroundColor: "#374151", color: "#f9fafb", outline: "none", fontSize: "0.9rem", width: "250px" }} />
              </div>

              {filteredBooks.length === 0 ? (
                <div style={{ textAlign: "center", padding: "40px", backgroundColor: "#374151", borderRadius: "12px" }}>
                  <p style={{ margin: 0, color: "#9ca3af" }}>
                    {books.length === 0 ? "Belum ada koleksi buku di database." : "Buku yang dicari tidak ditemukan."}
                  </p>
                </div>
              ) : (
                <ul style={{ listStyleType: "none", padding: 0, margin: 0 }}>
                  {filteredBooks.map((book) => (
                    <li key={book.id_buku} style={{ backgroundColor: "#ffffff", color: "#111827", padding: "20px", marginBottom: "15px", borderRadius: "12px", display: "flex", justifyContent: "space-between", alignItems: "center", boxShadow: "0 1px 3px 0 rgba(0, 0, 0, 0.1)", flexWrap: "wrap", gap: "15px" }}>
                      <div>
                        <h3 style={{ margin: "0 0 6px 0", fontSize: "1.2rem", fontWeight: "700", color: "#111827" }}>{book.judul}</h3>
                        <p style={{ margin: 0, color: "#6b7280", fontSize: "0.95rem" }}>Penulis: {book.author}</p>
                        <div style={{ marginTop: "10px", display: "inline-block", padding: "4px 10px", borderRadius: "9999px", fontSize: "0.85rem", fontWeight: "600", backgroundColor: book.stock > 0 ? "#d1fae5" : "#fee2e2", color: book.stock > 0 ? "#065f46" : "#991b1b" }}>
                          {book.stock > 0 ? `Sisa Stok: ${book.stock}` : "Stok Habis"}
                        </div>
                      </div>
                      <div style={{ display: "flex", flexDirection: "column", gap: "10px", alignItems: "flex-end" }}>
                        <div style={{ display: "flex", gap: "10px" }}>
                          <button onClick={() => startEdit(book)} style={{ padding: "6px 12px", backgroundColor: "#fbbf24", color: "#000", border: "none", borderRadius: "6px", cursor: "pointer", fontWeight: "600", fontSize: "0.85rem" }}>✏️ Edit</button>
                          <button onClick={() => handleDelete(book.id_buku)} style={{ padding: "6px 12px", backgroundColor: "#ef4444", color: "#fff", border: "none", borderRadius: "6px", cursor: "pointer", fontWeight: "600", fontSize: "0.85rem" }}>🗑️ Hapus</button>
                        </div>
                        <div style={{ display: "flex", gap: "10px" }}>
                          <button
                            onClick={() => handleReturn(book.id_buku, book.judul)}
                            disabled={!isBorrowed(book.id_buku)}
                            style={{ padding: "10px 15px", backgroundColor: isBorrowed(book.id_buku) ? "#f59e0b" : "#e5e7eb", color: isBorrowed(book.id_buku) ? "#ffffff" : "#9ca3af", border: "none", borderRadius: "8px", cursor: isBorrowed(book.id_buku) ? "pointer" : "not-allowed", fontWeight: "600", fontSize: "0.95rem" }}
                          >
                            Kembalikan
                          </button>
                          <button
                            onClick={() => handleBooking(book.id_buku, book.judul)}
                            disabled={book.stock === 0}
                            style={{ padding: "10px 20px", backgroundColor: book.stock > 0 ? "#3b82f6" : "#e5e7eb", color: book.stock > 0 ? "#ffffff" : "#9ca3af", border: "none", borderRadius: "8px", cursor: book.stock > 0 ? "pointer" : "not-allowed", fontWeight: "600", fontSize: "0.95rem" }}
                          >
                            {book.stock > 0 ? "Pinjam" : "Kosong"}
                          </button>
                        </div>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </>
          )}

          {activePage === "history" && (
            <div style={{ backgroundColor: "#ffffff", padding: "20px", borderRadius: "12px", color: "#111827" }}>
              <h3 style={{ margin: "0 0 15px 0", borderBottom: "1px solid #e5e7eb", paddingBottom: "10px" }}>🕒 Riwayat Transaksi</h3>

              {histories.length === 0 ? (
                <p style={{ color: "#6b7280", textAlign: "center", margin: "20px 0" }}>Belum ada transaksi.</p>
              ) : (
                <div style={{ overflowX: "auto" }}>
                  <table style={{ width: "100%", borderCollapse: "collapse", textAlign: "left", fontSize: "0.9rem" }}>
                    <thead>
                      <tr style={{ backgroundColor: "#f3f4f6", borderBottom: "2px solid #d1d5db" }}>
                        <th style={{ padding: "10px" }}>Peminjam</th>
                        <th style={{ padding: "10px" }}>Buku</th>
                        <th style={{ padding: "10px" }}>Tanggal</th>
                        <th style={{ padding: "10px" }}>Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      {histories.map(h => (
                        <tr key={h.id} style={{ borderBottom: "1px solid #e5e7eb" }}>
                          <td style={{ padding: "10px", fontWeight: "600" }}>{h.borrower_name}</td>
                          <td style={{ padding: "10px" }}>{h.judul_buku}</td>
                          <td style={{ padding: "10px", color: "#6b7280" }}>{new Date(h.borrow_date).toLocaleString('id-ID')}</td>
                          <td style={{ padding: "10px" }}>
                            <span style={{ padding: "4px 8px", borderRadius: "4px", backgroundColor: h.status === 'Dipinjam' ? '#fef3c7' : '#d1fae5', color: h.status === 'Dipinjam' ? '#92400e' : '#065f46', fontWeight: "bold", fontSize: "0.8rem" }}>
                              {h.status}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}

        </div>
      </div>
    </div>
  )
}

export default App