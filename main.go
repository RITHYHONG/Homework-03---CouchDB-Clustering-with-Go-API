package main

import (
	"context"
	_ "fmt"
	_ "github.com/go-kivik/kivik/v4/kiviktest/client"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-kivik/couchdb/v3"
	"github.com/go-kivik/kivik/v3"
)

var db *kivik.DB

// Struct representing a student
type Student struct {
	ID     string `json:"_id,omitempty"`
	Rev    string `json:"_rev,omitempty"`
	Email  string `json:"email"`
	Course string `json:"course"`
	Age    int    `json:"age"`
	Gender string `json:"gender"`
}

func main() {
	// Initialize CouchDB connection
	client, err := kivik.New("couch", "http://Rithyhong:Couchdb2003@localhost:5984/")
	if err != nil {
		panic(err)
	}
	if err != nil {
		log.Fatalf("Failed to connect to CouchDB: %v", err)
	}
	db = client.DB(context.TODO(), "student")
	if err != nil {
		log.Fatalf("Failed to access database: %v", err)
	}

	r := gin.Default()

	// Routes
	r.POST("/documents", insertDocument)
	r.GET("/documents", readAllDocuments)
	r.GET("/documents/:id", getDocumentByID)
	r.POST("/upload", uploadFile)
	//r.GET("/files/:id/:filename", getFile)
	r.GET("/documents/filter", filterDocuments)
	r.PUT("/documents/:id", updateDocument)
	r.DELETE("/documents/:id", deleteDocument)
	r.Run(":8080") // Listen and serve on port 8080
}

// insertDocument handles the insertion of a new student document
func insertDocument(c *gin.Context) {
	var student Student
	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Insert the document into CouchDB
	docID, _, err := db.CreateDoc(context.Background(), student)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document inserted successfully", "document_id": docID})
}

func uploadFile(c *gin.Context) {
	docID := c.Param("docID")
	filename := c.Param("filename")
	rev := c.Query("rev")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer openedFile.Close()

	_, err = db.PutAttachment(context.TODO(), docID, filename, &kivik.Attachment{
		Filename:    filename,
		Content:     openedFile,
		ContentType: file.Header.Get("Content-Type"),
	}, kivik.Options{"rev": rev})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload attachment: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "File uploaded successfully"})
}

// Read all documents
func readAllDocuments(c *gin.Context) {
	rows, err := db.AllDocs(context.Background(), kivik.Options{"include_docs": true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve documents"})
		return
	}
	defer rows.Close() // Ensure rows are closed when done

	var docs []map[string]interface{}
	for rows.Next() {
		var doc map[string]interface{}
		if err := rows.ScanDoc(&doc); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse document"})
			return
		}
		docs = append(docs, doc)
	}

	c.JSON(http.StatusOK, docs)
}

// getDocumentByID handles the retrieval of a document by ID
func getDocumentByID(c *gin.Context) {
	id := c.Param("id")

	row := db.Get(context.Background(), id)
	if row.Err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Document not found"})
		return
	}

	var doc map[string]interface{}
	if err := row.ScanDoc(&doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode document"})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// Filter documents (example: filter by a specific key)
func filterDocuments(c *gin.Context) {
	name := c.Query("name")
	rows, err := db.Query(context.Background(), `function(doc) { if (doc.name === "`+name+`") { emit(doc._id, doc); } }`, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to filter documents"})
		return
	}
	defer rows.Close() // Ensure rows are closed when done

	var docs []map[string]interface{}
	for rows.Next() {
		var doc map[string]interface{}
		if err := rows.ScanDoc(&doc); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse document"})
			return
		}
		docs = append(docs, doc)
	}

	c.JSON(http.StatusOK, docs)
}

// Update an existing document
func updateDocument(c *gin.Context) {
	id := c.Param("id")
	rev := c.Query("rev")

	// Fetch existing document
	err := db.Get(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var updateData map[string]interface{}
	if err := c.BindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure revision is included in the update
	updateData["_rev"] = rev

	// Check for and handle potential type mismatches in the update data
	for key, value := range updateData {
		switch v := value.(type) {
		case float64: // JSON numbers are float64 by default
			// Optionally convert float64 to int if needed
			updateData[key] = int(v)
		case string, bool:
			// Handle other types if necessary
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data type"})
			return
		}
	}

	_, putErr := db.Put(context.Background(), id, updateData)
	if putErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": putErr})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Document updated"})
}

func deleteDocument(c *gin.Context) {
	id := c.Param("id")
	rev := c.Query("rev")
	_, err := db.Delete(context.Background(), id, rev)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Document deleted"})
}
