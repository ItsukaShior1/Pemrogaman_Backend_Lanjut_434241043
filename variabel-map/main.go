package main

import "fmt"

func main() {
	var nama string = "Sari"
	var nim int = 123456
	var ipk float64 = 3.85
	var aktif bool = true
	var hobi []string
	hobi = []string{"membaca", "coding", "makan"}

	fmt.Println("=== Variabel ===")
	fmt.Printf("nama  : %s   (tipe %T)\n", nama, nama)
	fmt.Printf("nim   : %d   (tipe %T)\n", nim, nim)
	fmt.Printf("ipk   : %.2f (tipe %T)\n", ipk, ipk)
	fmt.Printf("aktif : %v  (tipe %T)\n", aktif, aktif)
	fmt.Printf("hobi  : %v  (tipe %T)\n", hobi, hobi)

	mahasiswa := make(map[string]int)
	mahasiswa["Sari"] = 22001
	mahasiswa["Budi"] = 22002
	mahasiswa["Ani"] = 22003

	fmt.Println("\n=== Map setelah ditambah ===")
	fmt.Println(mahasiswa)

	if nimSari, ada := mahasiswa["Sari"]; ada {
		fmt.Printf("NIM Sari = %d (ditemukan)\n", nimSari)
	} else {
		fmt.Println("Sari tidak ditemukan")
	}

	if nimX, ada := mahasiswa["Xavier"]; ada {
		fmt.Printf("NIM Xavier = %d\n", nimX)
	} else {
		fmt.Println("Xavier belum punya NIM")
	}

	delete(mahasiswa, "Budi")
	fmt.Println("\n=== Map setelah Budi dihapus ===")
	fmt.Println(mahasiswa)

	fmt.Println("\n=== Telusuri seluruh isi map ===")
	for nama, nim := range mahasiswa {
		fmt.Printf("%s -> %d\n", nama, nim)
	}
}
