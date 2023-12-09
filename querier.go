package mongoquerier

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var (
	ErrUnsupportedCollection  = errors.New("unsupported collection")
	ErrCollectionNameMismatch = errors.New("collection name mismatch")
	ErrFailedToCastInsertedID = errors.New("failed to type cast inserted ID to primitive.ObjectID or composite ID")
)

type Querier[Model any, IDModel any] struct {
	*MongoAdapter
	collection    *mongo.Collection
	IsIDComposite bool
}

func NewQuerier[Model any](madp *MongoAdapter, collectionName string) *Querier[Model, primitive.ObjectID] {
	collection := madp.GetCollection(collectionName)
	return &Querier[Model, primitive.ObjectID]{
		MongoAdapter: madp,
		collection:   collection,
	}
}

type IDContainer[IDModel any] struct {
	ID IDModel `json:"_id,omitempty"`
}

func NewQuerierWithCompositeID[Model any, IDModel any](madp *MongoAdapter, collectionName string) *Querier[Model, IDModel] {
	collection := madp.GetCollection(collectionName)
	return &Querier[Model, IDModel]{
		MongoAdapter:  madp,
		collection:    collection,
		IsIDComposite: true,
	}
}

func (q *Querier[Model, IDModel]) InsertOne(ctx context.Context, document Model, opts ...*options.InsertOneOptions) (insertedID IDModel, err error) {
	res, err := q.collection.InsertOne(ctx, document, opts...)
	if err != nil {
		return
	}

	insertedID, ok := res.InsertedID.(IDModel)
	if !ok {
		if q.IsIDComposite == true {
			var idContainer IDContainer[IDModel]
			idContainer, err = CastStruct[Model, IDContainer[IDModel]](document)
			insertedID = idContainer.ID
			if err != nil {
				return
			}
		} else {
			q.MongoAdapter.Error("Unable to cast InsertedID into ObjectID", zap.Error(err))
			err = ErrFailedToCastInsertedID
			return
		}
	}

	q.MongoAdapter.Debug(
		"Created a document",
		zap.String("collection_name", q.collection.Name()),
		zap.Any("_id", insertedID),
	)
	return
}

func (q *Querier[Model, IDModel]) InsertMany(ctx context.Context, documents []Model, opts ...*options.InsertManyOptions) ([]IDModel, error) {
	var insertedIDs []IDModel

	// Prepare a slice to store the inserted IDs.
	// Loop through the documents and perform bulk insertion.
	var insertModels []interface{}
	for _, doc := range documents {
		insertModels = append(insertModels, doc)
	}

	res, err := q.collection.InsertMany(ctx, insertModels, opts...)
	if err != nil {
		return nil, err
	}

	// Retrieve the inserted IDs from the result.
	for _, id := range res.InsertedIDs {
		insertedID, ok := id.(IDModel)
		if !ok {
			return nil, ErrFailedToCastInsertedID
		}
		insertedIDs = append(insertedIDs, insertedID)
	}

	q.MongoAdapter.Debug(
		"Inserted multiple documents",
		zap.String("collection_name", q.collection.Name()),
		zap.Int("documents_count", len(insertedIDs)),
	)

	return insertedIDs, nil
}

func (q *Querier[Model, IDModel]) Find(ctx context.Context, filter Model, opts ...*options.FindOptions) (documents []*Model, err error) {
	filterM, err := StructToM(filter)
	if err != nil {
		return
	}

	cursor, err := q.collection.Find(ctx, filterM, opts...)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var document Model
		if err = cursor.Decode(&document); err != nil {
			return
		}

		documents = append(documents, &document)
	}

	if err = cursor.Err(); err != nil {
		return
	}

	q.MongoAdapter.Debug(
		"Found all documents",
		zap.String("collection_name", q.collection.Name()),
		zap.Int("documents_count", len(documents)),
	)
	return
}

func (q *Querier[Model, IDModel]) FindByM(ctx context.Context, filter primitive.M, opts ...*options.FindOptions) (documents []*Model, err error) {
	cursor, err := q.collection.Find(ctx, filter, opts...)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var document Model
		if err = cursor.Decode(&document); err != nil {
			return
		}

		documents = append(documents, &document)
	}

	if err = cursor.Err(); err != nil {
		return
	}

	q.MongoAdapter.Debug(
		"Found all documents",
		zap.String("collection_name", q.collection.Name()),
		zap.Int("documents_count", len(documents)),
	)
	return
}

func (q *Querier[Model, IDModel]) FindOne(ctx context.Context, filter Model, opts ...*options.FindOneOptions) (document *Model, err error) {
	filterM, err := StructToM(filter)
	if err != nil {
		return
	}

	err = q.collection.FindOne(context.Background(), filterM, opts...).Decode(&document)
	if err != nil {
		return
	}

	q.MongoAdapter.Debug(
		"Found one document",
		zap.String("collection_name", q.collection.Name()),
		zap.Any("document", document),
	)
	return
}

func (q *Querier[Model, IDModel]) FindOneByM(ctx context.Context, filter primitive.M, opts ...*options.FindOneOptions) (document *Model, err error) {
	err = q.collection.FindOne(context.Background(), filter, opts...).Decode(&document)
	if err != nil {
		return
	}

	q.MongoAdapter.Debug(
		"Found one document",
		zap.String("collection_name", q.collection.Name()),
		zap.Any("document", document),
	)
	return
}

func (q *Querier[Model, IDModel]) UpdateOne(ctx context.Context, filter Model, update Model, opts ...*options.FindOneAndUpdateOptions) (document *Model, err error) {
	filterM, err := StructToM(filter)
	if err != nil {
		return
	}

	updateM, err := StructToM(update)
	if err != nil {
		return
	}
	updateM = bson.M{"$set": updateM}

	// opts = append(opts, options.FindOneAndUpdate().SetReturnDocument(options.After))
	err = q.collection.FindOneAndUpdate(
		ctx,
		filterM,
		updateM,
		opts...,
	).Decode(&document)
	if err != nil {
		return
	}

	q.MongoAdapter.Debug(
		"Updated one document",
		zap.String("collection_name", q.collection.Name()),
		zap.Any("document", document),
	)
	return
}

func (q *Querier[Model, IDModel]) UpdateOneByM(ctx context.Context, filter primitive.M, update Model, opts ...*options.FindOneAndUpdateOptions) (*Model, error) {
	// Convert the update model to primitive.M for use in the update operation.
	updateM, err := StructToM(update)
	if err != nil {
		return nil, err
	}
	updateM = bson.M{"$set": updateM}

	// opts = append(opts, options.FindOneAndUpdate().SetReturnDocument(options.After))
	var updatedDocument Model
	err = q.collection.FindOneAndUpdate(ctx, filter, updateM, opts...).Decode(&updatedDocument)
	if err != nil {
		return nil, err
	}

	q.MongoAdapter.Debug(
		"Updated one document by filter",
		zap.String("collection_name", q.collection.Name()),
		zap.Any("filter", filter),
		zap.Any("update", updateM),
		zap.Any("updated_document", updatedDocument),
	)

	return &updatedDocument, nil
}

func (q *Querier[Model, IDModel]) UpdateMany(ctx context.Context, filter Model, update Model, opts ...*options.UpdateOptions) ([]*Model, error) {
	// Convert filter and update models to primitive.M for use in the update operation.
	filterM, err := StructToM(filter)
	if err != nil {
		return nil, err
	}

	updateM, err := StructToM(update)
	if err != nil {
		return nil, err
	}
	updateM = bson.M{"$set": updateM}

	// Perform the update operation on multiple documents.
	result, err := q.collection.UpdateMany(ctx, filterM, updateM, opts...)
	if err != nil {
		return nil, err
	}

	q.MongoAdapter.Debug(
		"Updated multiple documents by filter",
		zap.String("collection_name", q.collection.Name()),
		zap.Any("filter", filterM),
		zap.Any("update", updateM),
		zap.Int("documents_modified", int(result.ModifiedCount)),
	)

	// Optionally, you can return some information about the updated documents if needed.
	// Here, we'll return nil to indicate success without specific document details.
	return nil, nil
}

func (q *Querier[Model, IDModel]) UpdateManyByM(ctx context.Context, filter primitive.M, update Model, opts ...*options.UpdateOptions) ([]*Model, error) {
	// Convert the update model to primitive.M for use in the update operation.
	updateM, err := StructToM(update)
	if err != nil {
		return nil, err
	}
	updateM = bson.M{"$set": updateM}

	// Perform the update operation on multiple documents based on the filter.
	// options := options.Update().SetUpsert(false)
	result, err := q.collection.UpdateMany(ctx, filter, updateM, opts...)
	if err != nil {
		return nil, err
	}

	q.MongoAdapter.Debug(
		"Updated multiple documents by filter (primitive.M)",
		zap.String("collection_name", q.collection.Name()),
		zap.Any("filter", filter),
		zap.Any("update", updateM),
		zap.Int("documents_modified", int(result.ModifiedCount)),
	)

	// Optionally, you can return some information about the updated documents if needed.
	// Here, we'll return nil to indicate success without specific document details.
	return nil, nil
}

func (q *Querier[Model, IDModel]) ReplaceOne(ctx context.Context, filter Model, replacement Model, opts ...*options.FindOneAndReplaceOptions) (*Model, error) {
	// Convert filter and replacement models to primitive.M for use in the replace operation.
	filterM, err := StructToM(filter)
	if err != nil {
		return nil, err
	}

	replacementM, err := StructToM(replacement)
	if err != nil {
		return nil, err
	}

	// Perform the replace operation on a single document.
	// options := options.Replace().SetUpsert(false)
	var replacedDocument Model
	err = q.collection.FindOneAndReplace(ctx, filterM, replacementM, opts...).Decode(&replacedDocument)
	if err != nil {
		return nil, err
	}

	q.MongoAdapter.Debug(
		"Replaced one document by filter",
		zap.String("collection_name", q.collection.Name()),
		zap.Any("filter", filterM),
		zap.Any("replacement", replacementM),
		zap.Any("replaced_document", replacedDocument),
	)

	return &replacedDocument, nil
}

func (q *Querier[Model, IDModel]) ReplaceOneByM(ctx context.Context, filter primitive.M, replacement Model, opts ...*options.FindOneAndReplaceOptions) (*Model, error) {
	// Convert the replacement model to primitive.M for use in the replace operation.
	replacementM, err := StructToM(replacement)
	if err != nil {
		return nil, err
	}

	// Perform the replace operation on a single document based on the filter.
	// options := options.Replace().SetUpsert(false)
	var replacedDocument Model
	err = q.collection.FindOneAndReplace(ctx, filter, replacementM, opts...).Decode(&replacedDocument)
	if err != nil {
		return nil, err
	}

	q.MongoAdapter.Debug(
		"Replaced one document by filter (primitive.M)",
		zap.String("collection_name", q.collection.Name()),
		zap.Any("filter", filter),
		zap.Any("replacement", replacementM),
		zap.Any("replaced_document", replacedDocument),
	)

	return &replacedDocument, nil
}

func (q *Querier[Model, IDModel]) DeleteOne(ctx context.Context, filter Model, opts ...*options.FindOneAndDeleteOptions) (document *Model, err error) {
	filterM, err := StructToM(filter)
	if err != nil {
		return
	}

	err = q.collection.FindOneAndDelete(
		ctx,
		filterM,
		opts...,
	).Decode(&document)
	if err != nil {
		return
	}

	q.MongoAdapter.Debug(
		"Deleted one document",
		zap.String("collection_name", q.collection.Name()),
		zap.Any("document", document),
	)
	return
}

func (q *Querier[Model, IDModel]) DeleteOneByM(ctx context.Context, filter primitive.M, opts ...*options.FindOneAndDeleteOptions) (*Model, error) {
	// Perform the delete operation on a single document based on the filter.
	var deletedDocument Model
	err := q.collection.FindOneAndDelete(ctx, filter, opts...).Decode(&deletedDocument)
	if err != nil {
		return nil, err
	}

	q.MongoAdapter.Debug(
		"Deleted one document by filter (primitive.M)",
		zap.String("collection_name", q.collection.Name()),
		zap.Any("filter", filter),
		zap.Any("deleted_document", deletedDocument),
	)

	return &deletedDocument, nil
}

func (q *Querier[Model, IDModel]) DeleteMany(ctx context.Context, filter Model, opts ...*options.DeleteOptions) (int64, error) {
	// Convert the filter model to primitive.M for use in the delete operation.
	filterM, err := StructToM(filter)
	if err != nil {
		return 0, err
	}

	// Perform the delete operation on multiple documents based on the filter.
	result, err := q.collection.DeleteMany(ctx, filterM, opts...)
	if err != nil {
		return 0, err
	}

	q.MongoAdapter.Debug(
		"Deleted multiple documents by filter",
		zap.String("collection_name", q.collection.Name()),
		zap.Any("filter", filterM),
		zap.Int64("documents_deleted", result.DeletedCount),
	)

	return result.DeletedCount, nil
}

func (q *Querier[Model, IDModel]) DeleteManyByM(ctx context.Context, filter primitive.M, opts ...*options.DeleteOptions) (int64, error) {
	// Perform the delete operation on multiple documents based on the filter.
	result, err := q.collection.DeleteMany(ctx, filter, opts...)
	if err != nil {
		return 0, err
	}

	q.MongoAdapter.Debug(
		"Deleted multiple documents by filter (primitive.M)",
		zap.String("collection_name", q.collection.Name()),
		zap.Any("filter", filter),
		zap.Int64("documents_deleted", result.DeletedCount),
	)

	return result.DeletedCount, nil
}

func (q *Querier[Model, IDModel]) CountDocuments(ctx context.Context, filter Model, opts ...*options.CountOptions) (int64, error) {
	// Convert the filter model to primitive.M for use in the count operation.
	filterM, err := StructToM(filter)
	if err != nil {
		return 0, err
	}

	// Perform the count operation on documents based on the filter.
	count, err := q.collection.CountDocuments(ctx, filterM, opts...)
	if err != nil {
		return 0, err
	}

	q.MongoAdapter.Debug(
		"Counted documents by filter",
		zap.String("collection_name", q.collection.Name()),
		zap.Any("filter", filterM),
		zap.Int64("documents_count", count),
	)

	return count, nil
}

func (q *Querier[Model, IDModel]) CountDocumentsByM(ctx context.Context, filter primitive.M, opts ...*options.CountOptions) (int64, error) {
	// Perform the count operation on documents based on the filter.
	count, err := q.collection.CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, err
	}

	q.MongoAdapter.Debug(
		"Counted documents by filter (primitive.M)",
		zap.String("collection_name", q.collection.Name()),
		zap.Any("filter", filter),
		zap.Int64("documents_count", count),
	)

	return count, nil
}

func (q *Querier[Model, IDModel]) Distinct(ctx context.Context, fieldName string, filter Model, opts ...*options.DistinctOptions) ([]interface{}, error) {
	// Convert the filter model to primitive.M for use in the distinct operation.
	filterM, err := StructToM(filter)
	if err != nil {
		return nil, err
	}

	// Perform the distinct operation on the specified field based on the filter.
	distinctValues, err := q.collection.Distinct(ctx, fieldName, filterM, opts...)
	if err != nil {
		return nil, err
	}

	q.MongoAdapter.Debug(
		"Retrieved distinct values for field",
		zap.String("collection_name", q.collection.Name()),
		zap.String("field_name", fieldName),
		zap.Any("filter", filterM),
		zap.Any("distinct_values", distinctValues),
	)

	return distinctValues, nil
}

func (q *Querier[Model, IDModel]) DistinctByM(ctx context.Context, fieldName string, filter primitive.M, opts ...*options.DistinctOptions) ([]interface{}, error) {
	// Perform the distinct operation on the specified field based on the filter.
	distinctValues, err := q.collection.Distinct(ctx, fieldName, filter, opts...)
	if err != nil {
		return nil, err
	}

	q.MongoAdapter.Debug(
		"Retrieved distinct values for field (primitive.M)",
		zap.String("collection_name", q.collection.Name()),
		zap.String("field_name", fieldName),
		zap.Any("filter", filter),
		zap.Any("distinct_values", distinctValues),
	)

	return distinctValues, nil
}

func (q *Querier[Model, IDModel]) DeleteCollection(ctx context.Context, collectionName string) error {
	if collectionName == q.collection.Name() {
		return q.collection.Drop(ctx)
	} else {
		return ErrCollectionNameMismatch
	}
}
