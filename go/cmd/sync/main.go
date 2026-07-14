package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchutil"
	"github.com/rezafahlevi/gotodo/internal/domain"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found")
	}

	// Connect to Postgres
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	// Connect to OpenSearch
	osURL := os.Getenv("OPENSEARCH_URL")
	if osURL == "" {
		osURL = "http://localhost:9200"
	}
	osClient, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{osURL},
	})
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	// Create Index with mapping and ngram analyzer
	mapping := `
	{
		"settings": {
			"analysis": {
				"analyzer": {
					"trigram_analyzer": {
						"type": "custom",
						"tokenizer": "ngram_tokenizer",
						"filter": ["lowercase"]
					}
				},
				"tokenizer": {
					"ngram_tokenizer": {
						"type": "ngram",
						"min_gram": 3,
						"max_gram": 3
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"id": { "type": "integer" },
				"title": { 
					"type": "text",
					"fields": {
						"ngram": {
							"type": "text",
							"analyzer": "trigram_analyzer"
						}
					}
				},
				"status": { "type": "integer" }
			}
		}
	}`

	res, err := osClient.Indices.Create(
		"todos",
		osClient.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	if err != nil {
		log.Fatalf("Error creating index request: %s", err)
	}
	defer res.Body.Close()
	
	if res.IsError() && res.StatusCode != 400 {
		log.Fatalf("Error creating index: %s", res.String())
	}

	// Setup Bulk Indexer
	indexer, err := opensearchutil.NewBulkIndexer(opensearchutil.BulkIndexerConfig{
		Index:      "todos",
		Client:     osClient,
		NumWorkers: 1,
		FlushBytes: 1 * 1024 * 1024,
	})
	if err != nil {
		log.Fatalf("Error creating the indexer: %s", err)
	}

	// Fetch and index records in batches
	log.Println("Starting data synchronization...")
	batchSize := 100000
	var lastID int = 0
	totalProcessed := 0
	startTime := time.Now()

	for {
		rows, err := dbPool.Query(ctx, "SELECT id, title, status FROM todos WHERE id > $1 ORDER BY id ASC LIMIT $2", lastID, batchSize)
		if err != nil {
			log.Fatalf("Error querying data: %s", err)
		}

		var todos []domain.Todo
		for rows.Next() {
			var t domain.Todo
			if err := rows.Scan(&t.ID, &t.Title, &t.Status); err != nil {
				log.Fatalf("Error scanning row: %s", err)
			}
			todos = append(todos, t)
			lastID = t.ID
		}
		rows.Close()

		if len(todos) == 0 {
			break
		}

		for _, t := range todos {
			docID := strconv.Itoa(t.ID)
			
			// Simple JSON generation
			titleEscaped := strings.ReplaceAll(t.Title, `"`, `\"`)
			docBody := fmt.Sprintf(`{"id":%d,"title":"%s","status":%d}`, t.ID, titleEscaped, t.Status)

			err := indexer.Add(ctx, opensearchutil.BulkIndexerItem{
				Action:     "index",
				DocumentID: docID,
				Body:       strings.NewReader(docBody),
				OnFailure: func(ctx context.Context, item opensearchutil.BulkIndexerItem, res opensearchutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						log.Printf("ERROR: %s", err)
					} else {
						log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
					}
				},
			})
			if err != nil {
				log.Fatalf("Unexpected error adding item to indexer: %s", err)
			}
		}

		totalProcessed += len(todos)
		log.Printf("Indexed %d records... (%.2f seconds)", totalProcessed, time.Since(startTime).Seconds())
	}

	if err := indexer.Close(ctx); err != nil {
		log.Fatalf("Unexpected error closing indexer: %s", err)
	}

	stats := indexer.Stats()
	log.Printf("Successfully synchronized %d documents in %.2f seconds", stats.NumAdded, time.Since(startTime).Seconds())
}
