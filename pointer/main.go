package main

import "fmt"

func swap(a, b *int) {
	temp := *a
	*a = *b
	*b = temp
}

func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

func passByValue(x int) {
	x = 100
}

func passByPointer(x *int) {
	*x = 100
}

func main() {
	fmt.Println("=== swap via pointer ===")
	x, y := 10, 20
	fmt.Printf("sebelum : x=%d, y=%d\n", x, y)
	swap(&x, &y)
	fmt.Printf("sesudah : x=%d, y=%d\n", x, y)

	fmt.Println("\n=== updateSlice via pointer ===")
	buah := []string{"apel", "jeruk"}
	fmt.Printf("sebelum : %v\n", buah)
	updateSlice(&buah, "mangga")
	fmt.Printf("sesudah : %v\n", buah)

	fmt.Println("\n=== pass by value vs pass by pointer ===")
	nilai := 42
	fmt.Printf("sebelum         : nilai=%d\n", nilai)

	passByValue(nilai)
	fmt.Printf("setelah passByValue  : nilai=%d  (tetap, karena yang dikirim salinan)\n", nilai)

	passByPointer(&nilai)
	fmt.Printf("setelah passByPointer: nilai=%d  (berubah, karena yang dikirim alamat)\n", nilai)
}
