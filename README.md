# MongoQuerier

MongoQuerier is a thin generic facilitator pattern layer to simplify my MongoDB operations by providing an abstraction layer for CRUD operations on collections.

## Installation

To use MongoQuerier in your Go project, simply import it:

```go
import "github.com/your-username/mongoquerier"
```

Make sure you have Go modules enabled (go mod init) to manage dependencies.

## Usage
### Initialization

```go
type Product struct {
	ID       *primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `bson:"name"`
	Price    float64            `bson:"price"`
	Quantity int                `bson:"quantity"`
	// ... other fields
}

// Initialize a MongoQuerier for a specific model
querier := NewQuerier[Product](mongoAdapter, "your_collection_name")

// If your model has a composite ID
compositeQuerier := NewQuerierWithCompositeID[ModelWithCompositeID](mongoAdapter, "your_composite_collection")
```

### CRUD Operations
MongoQuerier provides methods for common CRUD operations:

* InsertOne: Insert a single document into the collection.
* InsertMany: Insert multiple documents into the collection.
* Find: Retrieve documents based on a filter.
* FindOne: Retrieve a single document based on a filter.
* UpdateOne: Update a single document based on a filter.
* UpdateMany: Update multiple documents based on a filter.
* ReplaceOne: Replace a single document based on a filter.
* DeleteOne: Delete a single document based on a filter.
* DeleteMany: Delete multiple documents based on a filter.
* CountDocuments: Count documents based on a filter.
* Distinct: Retrieve distinct values for a field based on a filter.

### Examples:
```go
// Insert a single document
newProduct := Product{Name: "Example Product", Price: 99.99, Quantity: 10}
insertedID, err := querier.InsertOne(context.Background(), newProduct)

// Find documents
filter := Product{Name: "Example Product"}
documents, err := querier.Find(
    context.Background(), 
    filter, 
    options.Find().SetSkip(0).SetLimit(10)) // Pagination for page 0 with size 10)

// Update a document
filter := Product{Name: "Example Product"}
update := Product{Quantity: 15} // Define only the fields you want to update
updatedDocument, err := querier.UpdateOne(context.Background(), filter, update)

// Delete a document
deleteFilter := Product{Name: "Example Product"}
deletedDocument, err := querier.DeleteOne(context.Background(), deleteFilter)
```

### Functionalities
Below is a summary of the project's functionalities and their implementation status:
| `_id`        | Functionality   | Implemented | M_based |
|--------------|-----------------|-------------|---------|
| `ObjectId()` | InsertOne       | ✅          | -       |
| `ObjectId()` | InsertMany      | ✅          | -       |
| `ObjectId()` | Find            | ✅          | ✅      |
| `ObjectId()` | FindOne         | ✅          | ✅      |
| `ObjectId()` | Aggregate       | ❌          | -       |
| `ObjectId()` | UpdateOne       | ✅          | ✅      |
| `ObjectId()` | UpdateMany      | ✅          | ✅      |
| `ObjectId()` | ReplaceOne      | ✅          | ✅      |
| `ObjectId()` | DeleteOne       | ✅          | ✅      |
| `ObjectId()` | DeleteMany      | ✅          | ✅      |
| `ObjectId()` | CountDocuments  | ✅          | ✅      |
| `ObjectId()` | Distinct        | ✅          | ✅      |
| `ObjectId()` | BulkWrite       | ❌          | -       |

## Contribution
Contributions to MongoQuerier are welcome! Feel free to open issues or pull requests for new features, enhancements, or bug fixes.


## License
This project is licensed under the MIT License - see the LICENSE file for details.
