package main

import (
	"context"
	"fmt"
	_ "fmt"
	_ "github.com/go-kivik/kivik/v4/kiviktest/client"
	"github.com/go-kivik/kiviktest/v3/client"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-kivik/couchdb/v3"
	"github.com/go-kivik/kivik/v3"
)

var db *kivik.DB

// Struct representing a student
type Student struct {
	ID     string `json:"_id,omitempty"`  // CouchDB document ID (optional for insert)
	Rev    string `json:"_rev,omitempty"` // CouchDB revision ID (used for updates)
	Name   string `json:"name"`
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
	r.POST("/upload", uploadFileHandler)
	//r.GET("/files/:id/:filename", getFile)
	r.GET("/documents/filter", filterDocuments)
	r.PUT("/documents/:id", updateDocument)
	//r.DELETE("/delete", deleteDocument)
	http.HandleFunc("/delete", deleteDocumentHandler)
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

//	func uploadFileHandler(c *gin.Context) {
//		docID := c.PostForm("docID")
//		file, err := c.FormFile("file")
//		if err != nil {
//			c.JSON(http.StatusBadRequest, gin.H{"error": "File form parsing error: " + err.Error()})
//			return
//		}
//
//		doc := make(map[string]interface{})
//		err = db.Get(context.TODO(), docID).ScanDoc(&doc)
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get document: " + err.Error()})
//			return
//		}
//		rev, _ := doc["_rev"].(string)
//
//		openedFile, err := file.Open()
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "File opening error: " + err.Error()})
//			return
//		}
//		defer openedFile.Close()
//
//		_, err = db.PutAttachment(context.TODO(), docID, file.Filename, &kivik.Attachment{
//			Filename:    file.Filename,
//			Content:     openedFile,
//			ContentType: file.Header.Get("Content-Type"),
//		}, kivik.Options{"rev": rev})
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload attachment: " + err.Error()})
//			return
//		}
//
//		c.JSON(http.StatusOK, gin.H{"status": "File uploaded successfully"})
//	}
func uploadFileHandler(c *gin.Context) {
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

	// Put attachment to CouchDB
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

//
//// Upload file to CouchDB
//func uploadFile(c *gin.Context) {
//	docID := c.PostForm("docID")
//	file, err := c.FormFile("file")
//	//file, header, err := c.Request.FormFile("file")
//	if err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file"})
//		return
//	}
//	defer file.Close()
//
//	docID := c.PostForm("id")
//	rev := c.PostForm("rev")
//	rev, _ := doc["_rev"].(string)
//	openedFile, err := file.Open()
//	attachment := &driver.Attachment{
//		Filename:    header.Filename,
//		ContentType: header.Header.Get("Content-Type"),
//		Content:     file,
//	}
//
//	// Save the attachment to the specified document
//	_, err = db.PutAttachment(context.Background(), docID, attachment.Filename, attachment.rev, attachment.attachment)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload file: %v", err)})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
//}
//
//// getFile handles retrieving a file from a CouchDB document
//func getFile(c *gin.Context) {
//	docID := c.Param("id") // The ID of the document to get the file from
//	attachmentName := c.Query("attachment_name")
//
//	// Get the attachment from the document
//	att, err := db.GetAttachment(context.Background(), docID, attachmentName)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve attachment"})
//		return
//	}
//	defer att.Content.Close()
//
//	// Set the content type header to the attachment's content type
//	c.Writer.Header().Set("Content-Type", att.ContentType)
//	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+attachmentName)
//
//	// Write the content of the attachment to the response
//	_, err = c.Writer.Write(att.Content)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send attachment content"})
//		return
//	}
//}

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
	var updatedDoc map[string]interface{}
	if err := c.BindJSON(&updatedDoc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Retrieve the current document to get the latest revision ID
	doc := db.Get(c, id)
	err := db.Get(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	// Read the existing document to get the revision ID
	var existingDoc map[string]interface{}
	if err := doc.ScanDoc(&existingDoc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read document"})
		return
	}

	// Update the document with the new data
	updatedDoc["_rev"] = existingDoc["_rev"]

	rev, ok := existingDoc["_rev"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert revision ID to string"})
		return
	}
	updatedDoc["_rev"] = rev

	c.JSON(http.StatusOK, gin.H{"message": "Document updated successfully"})
}
func deleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// Get the document ID from the query parameters
	docID := r.URL.Query().Get("id")
	if docID == "" {
		http.Error(w, "Document ID is required", http.StatusBadRequest)
		return
	}

	// Delete the document from CouchDB
	err := client.Delete(db, docID)
	if err != nil {
		http.Error(w, "Error deleting document", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Document deleted successfully")
}
