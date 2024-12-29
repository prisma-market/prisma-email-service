package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/kihyun1998/prisma-market/prisma-email-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TemplateRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewTemplateRepository(mongoURI string) (*TemplateRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database("prisma_market")
	collection := db.Collection("email_templates")

	// 템플릿 이름에 대한 unique 인덱스 생성
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, err
	}

	return &TemplateRepository{
		db:         db,
		collection: collection,
	}, nil
}

// CreateTemplate 새로운 이메일 템플릿 생성
func (r *TemplateRepository) CreateTemplate(ctx context.Context, template *models.Template) error {
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, template)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("template with this name already exists")
		}
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		template.ID = oid
	}

	return nil
}

// GetTemplateByID ID로 템플릿 조회
func (r *TemplateRepository) GetTemplateByID(ctx context.Context, id primitive.ObjectID) (*models.Template, error) {
	var template models.Template
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&template)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// GetTemplateByName 이름으로 템플릿 조회
func (r *TemplateRepository) GetTemplateByName(ctx context.Context, name string) (*models.Template, error) {
	var template models.Template
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&template)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// UpdateTemplate 템플릿 업데이트
func (r *TemplateRepository) UpdateTemplate(ctx context.Context, id primitive.ObjectID, template *models.Template) error {
	template.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"subject":      template.Subject,
			"html_content": template.HTMLContent,
			"text_content": template.TextContent,
			"variables":    template.Variables,
			"updated_at":   template.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateByID(ctx, id, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("template not found")
	}

	return nil
}

// DeleteTemplate 템플릿 삭제
func (r *TemplateRepository) DeleteTemplate(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("template not found")
	}

	return nil
}

// ListTemplates 모든 템플릿 조회
func (r *TemplateRepository) ListTemplates(ctx context.Context) ([]*models.Template, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*models.Template
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

// Close MongoDB 연결 종료
func (r *TemplateRepository) Close(ctx context.Context) error {
	return r.db.Client().Disconnect(ctx)
}
