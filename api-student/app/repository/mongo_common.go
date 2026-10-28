package repository

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrInvalidID = errors.New("format id tidak valid")

// now membulatkan waktu ke milidetik. Tipe date pada BSON hanya menyimpan
// sampai milidetik, sedangkan time.Time di Go menyimpan sampai nanodetik.
// Tanpa pembulatan, nilai yang dikirim ke client saat membuat dokumen akan
// berbeda dari nilai yang dibaca kembali dari database.
func now() time.Time {
	return time.Now().UTC().Truncate(time.Millisecond)
}

// parseObjectID mengubah id hex dari URL menjadi ObjectID. Kegagalan berarti
// format id tidak valid, bukan dokumen yang tidak ada.
func parseObjectID(id string) (bson.ObjectID, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return bson.ObjectID{}, ErrInvalidID
	}
	return oid, nil
}
