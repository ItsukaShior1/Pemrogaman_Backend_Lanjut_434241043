package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"api-student/app/model"
)

var (
	ErrPortfolioNotFound = errors.New("portofolio tidak ditemukan")
	ErrProjectNotFound   = errors.New("proyek tidak ditemukan")
	ErrProjectLimit      = errors.New("jumlah proyek sudah mencapai batas")
)

// MaxProjectsPerPortfolio membatasi pertumbuhan array projects. Satu dokumen
// MongoDB maksimal 16 MB dan array yang dibiarkan tumbuh tanpa batas akan
// menabrak batas itu suatu saat.
const MaxProjectsPerPortfolio = 20

type PortfolioRepository interface {
	EnsureIndexes(ctx context.Context) error
	FindByStudentID(ctx context.Context, studentID int) (model.Portfolio, error)
	Replace(ctx context.Context, p model.Portfolio) (model.Portfolio, error)
	AddProject(ctx context.Context, studentID int, project model.Project) (model.Portfolio, error)
	DeleteProject(ctx context.Context, studentID int, projectID string) error
	TechnologyStats(ctx context.Context, limit int) ([]model.TechnologyStat, error)
}

// portfolioDocument adalah bentuk portofolio SEBAGAIMANA TERSIMPAN di
// MongoDB. Struct ini milik repository; model hanya mengenal bentuk JSON.
type portfolioDocument struct {
	ID           bson.ObjectID         `bson:"_id,omitempty"`
	StudentID    int                   `bson:"student_id"`
	Headline     string                `bson:"headline"`
	Bio          string                `bson:"bio"`
	Projects     []projectDocument     `bson:"projects"`
	Certificates []certificateDocument `bson:"certificates"`
	CreatedAt    time.Time             `bson:"created_at"`
	UpdatedAt    time.Time             `bson:"updated_at"`
}

type projectDocument struct {
	ID           bson.ObjectID `bson:"_id"`
	Title        string        `bson:"title"`
	Technologies []string      `bson:"technologies"`
	URL          string        `bson:"url"`
	Year         int           `bson:"year"`
}

type certificateDocument struct {
	ID     bson.ObjectID `bson:"_id"`
	Name   string        `bson:"name"`
	Issuer string        `bson:"issuer"`
	Date   time.Time     `bson:"date"`
}

func (d portfolioDocument) toModel() model.Portfolio {
	projects := make([]model.Project, 0, len(d.Projects))
	for _, p := range d.Projects {
		tech := p.Technologies
		if tech == nil {
			tech = []string{}
		}
		projects = append(projects, model.Project{
			ID: p.ID.Hex(), Title: p.Title, Technologies: tech,
			URL: p.URL, Year: p.Year,
		})
	}

	certs := make([]model.Certificate, 0, len(d.Certificates))
	for _, c := range d.Certificates {
		certs = append(certs, model.Certificate{
			ID: c.ID.Hex(), Name: c.Name, Issuer: c.Issuer, Date: c.Date,
		})
	}

	return model.Portfolio{
		ID: d.ID.Hex(), StudentID: d.StudentID, Headline: d.Headline, Bio: d.Bio,
		Projects: projects, Certificates: certs,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

type portfolioMongoRepository struct{ coll *mongo.Collection }

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioMongoRepository{coll: db.Collection("student_portfolios")}
}

// EnsureIndexes dijalankan sekali saat aplikasi menyala.
//
// Index unik pada student_id adalah penjaga aturan "satu mahasiswa satu
// portofolio": kodenya bisa salah, index tidak.
func (r *portfolioMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{

			Keys:    bson.D{{Key: "student_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("portfolio_student_unique"),
		},
		{

			Keys:    bson.D{{Key: "projects.technologies", Value: 1}},
			Options: options.Index().SetName("portfolio_technologies"),
		},
		{

			Keys:    bson.D{{Key: "projects._id", Value: 1}},
			Options: options.Index().SetName("portfolio_project_ids"),
		},
	})
	if err != nil {
		return fmt.Errorf("membuat index student_portfolios: %w", err)
	}
	return nil
}

func (r *portfolioMongoRepository) FindByStudentID(
	ctx context.Context, studentID int,
) (model.Portfolio, error) {
	var doc portfolioDocument
	err := r.coll.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Portfolio{}, ErrPortfolioNotFound
		}
		return model.Portfolio{}, fmt.Errorf("mengambil portofolio: %w", err)
	}
	return doc.toModel(), nil
}

// Replace membuat portofolio bila belum ada, atau menggantinya seluruhnya
// bila sudah ada.
//
// Upsert dipakai agar "satu mahasiswa satu portofolio" tetap dijaga index
// unik student_id, bukan oleh pengecekan manual di Go.
func (r *portfolioMongoRepository) Replace(
	ctx context.Context, p model.Portfolio,
) (model.Portfolio, error) {
	t := now()

	projects := make([]projectDocument, 0, len(p.Projects))
	for _, pr := range p.Projects {
		oid, err := parseObjectID(pr.ID)
		if err != nil {
			oid = bson.NewObjectID()
		}
		tech := pr.Technologies
		if tech == nil {
			tech = []string{}
		}
		projects = append(projects, projectDocument{
			ID: oid, Title: pr.Title, Technologies: tech, URL: pr.URL, Year: pr.Year,
		})
	}

	certs := make([]certificateDocument, 0, len(p.Certificates))
	for _, c := range p.Certificates {
		oid, err := parseObjectID(c.ID)
		if err != nil {
			oid = bson.NewObjectID()
		}
		certs = append(certs, certificateDocument{
			ID: oid, Name: c.Name, Issuer: c.Issuer, Date: c.Date,
		})
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var doc portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx,
		bson.M{"student_id": p.StudentID},
		bson.M{
			"$set": bson.M{
				"headline":     p.Headline,
				"bio":          p.Bio,
				"projects":     projects,
				"certificates": certs,
				"updated_at":   t,
			},
			// created_at hanya diisi saat dokumen benar-benar baru dibuat.
			"$setOnInsert": bson.M{"created_at": t},
		},
		opts,
	).Decode(&doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return model.Portfolio{}, ErrDuplicate
		}
		return model.Portfolio{}, fmt.Errorf("menyimpan portofolio: %w", err)
	}
	return doc.toModel(), nil
}

// AddProject menambahkan satu proyek ke dalam array projects.
//
// Syarat "projects.19 tidak ada" berarti array belum berisi 20 elemen
// (indeks 0 sampai 19). Pemeriksaan batas dan penambahan terjadi dalam satu
// operasi atomik, sehingga dua permintaan serentak tidak dapat sama-sama
// lolos ketika tinggal satu tempat tersisa.
func (r *portfolioMongoRepository) AddProject(
	ctx context.Context, studentID int, project model.Project,
) (model.Portfolio, error) {
	tech := project.Technologies
	if tech == nil {
		tech = []string{}
	}

	doc := projectDocument{
		ID: bson.NewObjectID(), Title: project.Title,
		Technologies: tech, URL: project.URL, Year: project.Year,
	}
	limitField := "projects." + strconv.Itoa(MaxProjectsPerPortfolio-1)

	var updated portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx,
		bson.M{"student_id": studentID, limitField: bson.M{"$exists": false}},
		bson.M{
			"$push": bson.M{"projects": doc},
			"$set":  bson.M{"updated_at": now()},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err == nil {
		return updated.toModel(), nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, fmt.Errorf("menambah proyek: %w", err)
	}

	count, err := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("memeriksa portofolio: %w", err)
	}
	if count == 0 {
		return model.Portfolio{}, ErrPortfolioNotFound
	}
	return model.Portfolio{}, ErrProjectLimit
}

// DeleteProject menghapus satu proyek dari dalam array projects memakai $pull.
//
// $pull menghapus hanya elemen yang cocok, tanpa mengubah elemen lain.
func (r *portfolioMongoRepository) DeleteProject(
	ctx context.Context, studentID int, projectID string,
) error {
	oid, err := parseObjectID(projectID)
	if err != nil {
		return err
	}

	// Filter ikut mencocokkan proyek yang akan dihapus. Dengan begitu
	// MatchedCount membedakan "portofolionya ada tapi proyeknya tidak ada"
	// dari "portofolionya tidak ada".
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"student_id": studentID, "projects._id": oid},
		bson.M{
			"$pull": bson.M{"projects": bson.M{"_id": oid}},
			"$set":  bson.M{"updated_at": now()},
		},
	)
	if err != nil {
		return fmt.Errorf("menghapus proyek: %w", err)
	}
	if res.MatchedCount == 0 {
		count, err := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
		if err != nil {
			return fmt.Errorf("memeriksa portofolio: %w", err)
		}
		if count == 0 {
			return ErrPortfolioNotFound
		}
		return ErrProjectNotFound
	}
	return nil
}

// TechnologyStats menghitung teknologi paling banyak dipakai di seluruh
// portofolio.
//
// Dua kali $unwind: yang pertama memecah satu dokumen portofolio menjadi satu
// dokumen per proyek, yang kedua memecah satu proyek menjadi satu dokumen per
// teknologi. Baru setelah itu $group bisa menghitung.
func (r *portfolioMongoRepository) TechnologyStats(
	ctx context.Context, limit int,
) ([]model.TechnologyStat, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$projects"}},
		{{Key: "$unwind", Value: "$projects.technologies"}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$projects.technologies",
			"total": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "total", Value: -1},
			{Key: "_id", Value: 1},
		}}},
		{{Key: "$limit", Value: int64(limit)}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("menghitung statistik teknologi: %w", err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		Technology string `bson:"_id"`
		Total      int    `bson:"total"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("membaca statistik teknologi: %w", err)
	}

	result := make([]model.TechnologyStat, 0, len(rows))
	for _, r := range rows {
		result = append(result, model.TechnologyStat{
			Technology: r.Technology, Total: r.Total,
		})
	}
	return result, nil
}
