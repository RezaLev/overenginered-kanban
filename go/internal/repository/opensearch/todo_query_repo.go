package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strconv"

	os "github.com/opensearch-project/opensearch-go/v2"
	"github.com/rezafahlevi/gotodo/internal/domain"
)

type todoQueryRepository struct {
	client *os.Client
}

func NewTodoQueryRepository(client *os.Client) domain.TodoQueryRepository {
	return &todoQueryRepository{client: client}
}

func (r *todoQueryRepository) FetchAll(ctx context.Context, searchQuery string, statusFilter *int, page int, limit int) ([]domain.Todo, int, error) {
	offset := (page - 1) * limit

	var mustClauses []interface{}

	if searchQuery != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"title": map[string]interface{}{
					"query": searchQuery,
					"fuzziness": "AUTO",
				},
			},
		})
	}

	if statusFilter != nil {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"status": *statusFilter,
			},
		})
	}

	queryBody := map[string]interface{}{
		"from": offset,
		"size": limit,
		"sort": []interface{}{
			map[string]interface{}{"status": map[string]string{"order": "asc"}},
			map[string]interface{}{"id": map[string]string{"order": "desc"}},
		},
		"track_total_hits": true,
	}

	if len(mustClauses) > 0 {
		queryBody["query"] = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		}
	} else {
		queryBody["query"] = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(queryBody); err != nil {
		return nil, 0, err
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("todos"),
		r.client.Search.WithBody(&buf),
		r.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, errors.New("error from opensearch: " + res.String())
	}

	var osResp struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.Todo `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&osResp); err != nil {
		return nil, 0, err
	}

	var todos []domain.Todo
	for _, hit := range osResp.Hits.Hits {
		todos = append(todos, hit.Source)
	}

	return todos, osResp.Hits.Total.Value, nil
}

func (r *todoQueryRepository) GetFacets(ctx context.Context, searchQuery string) (domain.Facet, error) {
	var mustClauses []interface{}

	if searchQuery != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"title": map[string]interface{}{
					"query": searchQuery,
					"fuzziness": "AUTO",
				},
			},
		})
	}

	queryBody := map[string]interface{}{
		"size": 0,
		"aggs": map[string]interface{}{
			"status_counts": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "status",
					"size": 10,
				},
			},
		},
	}

	if len(mustClauses) > 0 {
		queryBody["query"] = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		}
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(queryBody); err != nil {
		return nil, err
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("todos"),
		r.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New("error from opensearch: " + res.String())
	}

	var osResp struct {
		Aggregations struct {
			StatusCounts struct {
				Buckets []struct {
					Key      int `json:"key"`
					DocCount int `json:"doc_count"`
				} `json:"buckets"`
			} `json:"status_counts"`
		} `json:"aggregations"`
	}

	if err := json.NewDecoder(res.Body).Decode(&osResp); err != nil {
		return nil, err
	}

	facets := make(domain.Facet)
	for i := domain.StatusOpen; i <= domain.StatusCanceled; i++ {
		facets[i] = 0
	}
	
	for _, bucket := range osResp.Aggregations.StatusCounts.Buckets {
		facets[bucket.Key] = bucket.DocCount
	}

	return facets, nil
}

func (r *todoQueryRepository) GetByID(ctx context.Context, id int) (domain.Todo, error) {
	docID := strconv.Itoa(id)
	res, err := r.client.Get(
		"todos",
		docID,
		r.client.Get.WithContext(ctx),
	)
	if err != nil {
		return domain.Todo{}, err
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return domain.Todo{}, errors.New("todo not found")
		}
		return domain.Todo{}, errors.New("error from opensearch: " + res.String())
	}

	var osResp struct {
		Source domain.Todo `json:"_source"`
	}
	if err := json.NewDecoder(res.Body).Decode(&osResp); err != nil {
		return domain.Todo{}, err
	}

	return osResp.Source, nil
}
