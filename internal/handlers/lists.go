package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/shindakun/attodo/internal/models"
	"github.com/shindakun/attodo/internal/session"
	"github.com/shindakun/bskyoauth"
)

const ListCollection = "app.attodo.list"

type ListHandler struct {
	client *bskyoauth.Client
}

func NewListHandler(client *bskyoauth.Client) *ListHandler {
	return &ListHandler{client: client}
}

// HandleLists handles list CRUD operations
func (h *ListHandler) HandleLists(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleListLists(w, r)
	case http.MethodPost:
		h.handleCreateList(w, r)
	case http.MethodPut:
		h.handleUpdateList(w, r)
	case http.MethodDelete:
		h.handleDeleteList(w, r)
	case http.MethodPatch:
		h.handleManageTasks(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleListDetail shows a specific list with its tasks (authenticated)
func (h *ListHandler) HandleListDetail(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract rkey from URL path (e.g., /app/lists/view/abc123)
	path := strings.TrimPrefix(r.URL.Path, "/app/lists/view/")
	rkey := strings.TrimSuffix(path, "/")

	if rkey == "" {
		http.Error(w, "List ID required", http.StatusBadRequest)
		return
	}

	// Get the list
	log.Printf("Fetching list with rkey: %s", rkey)
	var record map[string]interface{}
	var err error
	sess, err = h.WithRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		record, err = h.getRecord(r.Context(), s, rkey)
		return err
	})

	if err != nil {
		log.Printf("Failed to get list rkey=%s: %v", rkey, err)
		http.Error(w, fmt.Sprintf("List not found: %v", err), http.StatusNotFound)
		return
	}

	log.Printf("Successfully fetched list record: %+v", record)

	// Parse list
	list := parseListRecord(record)
	list.RKey = rkey
	list.URI = fmt.Sprintf("at://%s/%s/%s", sess.DID, ListCollection, rkey)

	// Resolve DID to handle for public sharing URL
	dir := identity.DefaultDirectory()
	atid, err := syntax.ParseAtIdentifier(sess.DID)
	if err == nil {
		ident, err := dir.Lookup(r.Context(), *atid)
		if err == nil {
			list.OwnerHandle = ident.Handle.String()
		}
	}

	// Resolve tasks from URIs
	if len(list.TaskURIs) > 0 {
		tasks, err := h.resolveTasksFromURIs(r.Context(), sess, list.TaskURIs)
		if err != nil {
			log.Printf("Failed to resolve tasks for list %s: %v", rkey, err)
			// Continue anyway, just with empty tasks
		} else {
			list.Tasks = tasks
		}
	}

	// Update session
	cookie, _ := r.Cookie("session_id")
	if cookie != nil {
		h.client.UpdateSession(cookie.Value, sess)
	}

	// Render list detail view
	w.Header().Set("Content-Type", "text/html")
	Render(w, "list-detail.html", list)
}

// HandlePublicListView shows a public read-only view of a list
func (h *ListHandler) HandlePublicListView(w http.ResponseWriter, r *http.Request) {
	// Extract handle and rkey from URL path (e.g., /list/@handle.bsky.social/abc123)
	path := strings.TrimPrefix(r.URL.Path, "/list/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		http.Error(w, "Invalid list URL format. Expected: /list/@handle/rkey", http.StatusBadRequest)
		return
	}

	handle := strings.TrimPrefix(parts[0], "@")
	rkey := parts[1]

	if handle == "" || rkey == "" {
		http.Error(w, "Handle and list ID required", http.StatusBadRequest)
		return
	}

	// Resolve the handle to a DID
	dir := identity.DefaultDirectory()
	atid, err := syntax.ParseAtIdentifier(handle)
	if err != nil {
		log.Printf("Failed to parse handle %s: %v", handle, err)
		http.Error(w, "Invalid handle", http.StatusBadRequest)
		return
	}

	ident, err := dir.Lookup(r.Context(), *atid)
	if err != nil {
		log.Printf("Failed to resolve handle %s: %v", handle, err)
		http.Error(w, "Handle not found", http.StatusNotFound)
		return
	}

	did := ident.DID.String()
	pds := ident.PDSEndpoint()

	log.Printf("Fetching public list: handle=%s, did=%s, rkey=%s", handle, did, rkey)

	// Fetch the list record publicly
	record, err := h.getPublicRecord(r.Context(), pds, did, rkey)
	if err != nil {
		log.Printf("Failed to get public list: %v", err)
		http.Error(w, "List not found or not public", http.StatusNotFound)
		return
	}

	// Parse list
	list := parseListRecord(record)
	list.RKey = rkey
	list.URI = fmt.Sprintf("at://%s/%s/%s", did, ListCollection, rkey)
	list.OwnerHandle = handle

	// Resolve tasks from URIs (public fetch)
	if len(list.TaskURIs) > 0 {
		tasks, err := h.resolvePublicTasksFromURIs(r.Context(), pds, did, list.TaskURIs)
		if err != nil {
			log.Printf("Failed to resolve public tasks for list %s: %v", rkey, err)
			// Continue anyway, just with empty tasks
		} else {
			list.Tasks = tasks
		}
	}

	// Render public list view
	w.Header().Set("Content-Type", "text/html")
	Render(w, "public-list.html", list)
}

// handleCreateList creates a new list
func (h *ListHandler) handleCreateList(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Create new list
	now := time.Now().UTC()
	list := &models.TaskList{
		Name:        name,
		Description: r.FormValue("description"),
		TaskURIs:    []string{}, // Empty initially
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Build record
	record := buildListRecord(list)

	// Create the record with retry logic
	var output *atproto.RepoCreateRecord_Output
	var err error
	sess, err = h.WithRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		output, err = h.client.CreateRecord(r.Context(), s, ListCollection, record)
		return err
	})

	if err != nil {
		log.Printf("Failed to create list after retries: %v", err)
		http.Error(w, "Failed to create list", http.StatusInternalServerError)
		return
	}

	// Update session
	cookie, _ := r.Cookie("session_id")
	if cookie != nil {
		h.client.UpdateSession(cookie.Value, sess)
	}

	// Extract RKey from URI
	list.RKey = extractRKey(output.Uri)
	list.URI = output.Uri

	log.Printf("List created: %s (%s)", list.Name, list.RKey)

	// Return the list partial for HTMX
	w.Header().Set("Content-Type", "text/html")
	Render(w, "list-item.html", list)
}

// handleListLists retrieves all lists
func (h *ListHandler) handleListLists(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get lists from repository
	var lists []*models.TaskList
	var err error
	sess, err = h.WithRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		lists, err = h.ListRecords(r.Context(), s)
		return err
	})

	if err != nil {
		log.Printf("Failed to list lists: %v", err)
		http.Error(w, "Failed to list lists", http.StatusInternalServerError)
		return
	}

	// Update session
	cookie, _ := r.Cookie("session_id")
	if cookie != nil {
		h.client.UpdateSession(cookie.Value, sess)
	}

	// Return HTML partials for HTMX
	w.Header().Set("Content-Type", "text/html")
	for _, list := range lists {
		Render(w, "list-item.html", list)
	}
}

// handleUpdateList updates an existing list
func (h *ListHandler) handleUpdateList(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	rkey := r.FormValue("rkey")
	if rkey == "" {
		http.Error(w, "rkey is required", http.StatusBadRequest)
		return
	}

	// Get current list
	var record map[string]interface{}
	var err error
	sess, err = h.WithRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		record, err = h.getRecord(r.Context(), s, rkey)
		return err
	})

	if err != nil {
		log.Printf("Failed to get list: %v", err)
		http.Error(w, "Failed to get list", http.StatusInternalServerError)
		return
	}

	// Parse existing record
	list := parseListRecord(record)
	list.RKey = rkey
	// Build URI from DID and collection
	list.URI = fmt.Sprintf("at://%s/%s/%s", sess.DID, ListCollection, rkey)

	// Update fields
	if name := r.FormValue("name"); name != "" {
		list.Name = name
	}
	list.Description = r.FormValue("description")
	list.UpdatedAt = time.Now().UTC()

	// Handle task URI updates if provided
	if taskURIsJSON := r.FormValue("taskUris"); taskURIsJSON != "" {
		var taskURIs []string
		if err := json.Unmarshal([]byte(taskURIsJSON), &taskURIs); err == nil {
			list.TaskURIs = taskURIs
		}
	}

	// Build record and update
	updatedRecord := buildListRecord(list)
	sess, err = h.WithRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		return h.updateRecord(r.Context(), s, rkey, updatedRecord)
	})

	if err != nil {
		log.Printf("Failed to update list: %v", err)
		http.Error(w, "Failed to update list", http.StatusInternalServerError)
		return
	}

	// Update session
	cookie, _ := r.Cookie("session_id")
	if cookie != nil {
		h.client.UpdateSession(cookie.Value, sess)
	}

	log.Printf("List updated: %s", rkey)

	// Return updated list partial
	w.Header().Set("Content-Type", "text/html")
	Render(w, "list-item.html", list)
}

// handleManageTasks adds or removes tasks from a list
func (h *ListHandler) handleManageTasks(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	rkey := r.FormValue("rkey")
	taskURI := r.FormValue("taskUri")
	action := r.FormValue("action") // "add" or "remove"

	if rkey == "" || taskURI == "" || action == "" {
		http.Error(w, "rkey, taskUri, and action are required", http.StatusBadRequest)
		return
	}

	// Get current list
	var record map[string]interface{}
	var err error
	sess, err = h.WithRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		record, err = h.getRecord(r.Context(), s, rkey)
		return err
	})

	if err != nil {
		log.Printf("Failed to get list: %v", err)
		http.Error(w, "Failed to get list", http.StatusInternalServerError)
		return
	}

	// Parse list
	list := parseListRecord(record)
	list.RKey = rkey
	list.URI = fmt.Sprintf("at://%s/%s/%s", sess.DID, ListCollection, rkey)

	// Modify task URIs based on action
	switch action {
	case "add":
		// Check if task is already in the list
		found := false
		for _, uri := range list.TaskURIs {
			if uri == taskURI {
				found = true
				break
			}
		}
		if !found {
			list.TaskURIs = append(list.TaskURIs, taskURI)
		}
	case "remove":
		// Remove task from list
		newURIs := make([]string, 0, len(list.TaskURIs))
		for _, uri := range list.TaskURIs {
			if uri != taskURI {
				newURIs = append(newURIs, uri)
			}
		}
		list.TaskURIs = newURIs
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	// Update timestamp
	list.UpdatedAt = time.Now().UTC()

	// Build record and update
	updatedRecord := buildListRecord(list)
	sess, err = h.WithRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		return h.updateRecord(r.Context(), s, rkey, updatedRecord)
	})

	if err != nil {
		log.Printf("Failed to update list tasks: %v", err)
		http.Error(w, "Failed to update list", http.StatusInternalServerError)
		return
	}

	// Update session
	cookie, _ := r.Cookie("session_id")
	if cookie != nil {
		h.client.UpdateSession(cookie.Value, sess)
	}

	log.Printf("Task %s %sd to/from list %s", taskURI, action, rkey)

	// Extract task rkey from URI (e.g., at://did:plc:xxx/app.attodo.task/rkey)
	taskRKey := extractRKey(taskURI)

	// Fetch the updated task to return it with its new list associations
	var taskRecord map[string]interface{}
	sess, err = h.WithRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		var fetchErr error
		taskRecord, fetchErr = h.getTaskRecord(r.Context(), s, taskRKey)
		return fetchErr
	})

	if err != nil {
		log.Printf("Failed to fetch updated task %s: %v", taskRKey, err)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
		return
	}

	// Parse task
	task := parseTaskRecord(taskRecord)
	task.RKey = taskRKey
	task.URI = taskURI

	// Get all lists to populate the Lists field for this task
	var allLists []*models.TaskList
	sess, err = h.WithRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		var fetchErr error
		allLists, fetchErr = h.ListRecords(r.Context(), s)
		return fetchErr
	})

	if err == nil {
		// Find lists that contain this task
		taskLists := make([]*models.TaskList, 0)
		for _, l := range allLists {
			for _, uri := range l.TaskURIs {
				if uri == taskURI {
					taskLists = append(taskLists, l)
					break
				}
			}
		}
		task.Lists = taskLists
	}

	// Update session (already updated from last WithRetry call above)
	if cookie != nil {
		h.client.UpdateSession(cookie.Value, sess)
	}

	// Return updated task partial for HTMX to swap
	w.Header().Set("Content-Type", "text/html")
	Render(w, "task-item.html", task)
}

// handleDeleteList deletes a list
func (h *ListHandler) handleDeleteList(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rkey := r.URL.Query().Get("rkey")
	if rkey == "" {
		http.Error(w, "rkey is required", http.StatusBadRequest)
		return
	}

	// Delete with retry logic
	var err error
	sess, err = h.WithRetry(r.Context(), sess, func(s *bskyoauth.Session) error {
		return h.client.DeleteRecord(r.Context(), s, ListCollection, rkey)
	})

	if err != nil {
		log.Printf("Failed to delete list after retries: %v", err)
		http.Error(w, "Failed to delete list", http.StatusInternalServerError)
		return
	}

	// Update session
	cookie, _ := r.Cookie("session_id")
	if cookie != nil {
		h.client.UpdateSession(cookie.Value, sess)
	}

	log.Printf("List deleted: %s", rkey)
	w.WriteHeader(http.StatusOK)
}

// listRecords fetches all lists from the repository
// ListRecords fetches all list records for the given session (public for cross-handler access)
func (h *ListHandler) ListRecords(ctx context.Context, sess *bskyoauth.Session) ([]*models.TaskList, error) {
	// Build the XRPC URL
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.listRecords?repo=%s&collection=%s",
		sess.PDS, sess.DID, ListCollection)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add authorization header
	req.Header.Set("Authorization", "Bearer "+sess.AccessToken)

	// Make request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("XRPC ERROR %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result struct {
		Records []struct {
			URI   string                 `json:"uri"`
			Value map[string]interface{} `json:"value"`
		} `json:"records"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert records to TaskList objects
	lists := make([]*models.TaskList, 0, len(result.Records))
	for _, record := range result.Records {
		list := parseListRecord(record.Value)
		list.URI = record.URI
		list.RKey = extractRKey(record.URI)
		lists = append(lists, list)
	}

	return lists, nil
}

// getRecord retrieves a single record using com.atproto.repo.getRecord
func (h *ListHandler) getRecord(ctx context.Context, sess *bskyoauth.Session, rkey string) (map[string]interface{}, error) {
	// Build the XRPC URL
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?repo=%s&collection=%s&rkey=%s",
		sess.PDS, sess.DID, ListCollection, rkey)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+sess.AccessToken)

	// Create HTTP client with DPoP transport
	transport := bskyoauth.NewDPoPTransport(http.DefaultTransport, sess.DPoPKey, sess.AccessToken, sess.DPoPNonce)
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Update nonce if present
	if dpopTransport, ok := transport.(bskyoauth.DPoPTransport); ok {
		sess.DPoPNonce = dpopTransport.GetNonce()
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("XRPC ERROR %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var result struct {
		URI   string                 `json:"uri"`
		CID   string                 `json:"cid"`
		Value map[string]interface{} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Value, nil
}

// updateRecord updates a record using com.atproto.repo.putRecord
func (h *ListHandler) updateRecord(ctx context.Context, sess *bskyoauth.Session, rkey string, record map[string]interface{}) error {
	log.Printf("updateRecord: DID=%s, Collection=%s, RKey=%s", sess.DID, ListCollection, rkey)

	// Resolve the actual PDS endpoint for this user
	pdsHost, err := h.resolvePDSEndpoint(ctx, sess.DID)
	if err != nil {
		return fmt.Errorf("failed to resolve PDS endpoint: %w", err)
	}
	log.Printf("updateRecord: Resolved PDS=%s", pdsHost)

	// Add $type field to the record if not present
	if _, exists := record["$type"]; !exists {
		record["$type"] = ListCollection
	}

	// Build the request body
	body := map[string]interface{}{
		"repo":       sess.DID,
		"collection": ListCollection,
		"rkey":       rkey,
		"record":     record,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the request to the resolved PDS endpoint
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.putRecord", pdsHost)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// Create DPoP transport for authentication
	dpopTransport := bskyoauth.NewDPoPTransport(
		http.DefaultTransport,
		sess.DPoPKey,
		sess.AccessToken,
		sess.DPoPNonce,
	)

	httpClient := &http.Client{
		Transport: dpopTransport,
		Timeout:   10 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("updateRecord: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
		return fmt.Errorf("XRPC ERROR %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var output atproto.RepoPutRecord_Output
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("updateRecord: Success! URI=%s", output.Uri)
	return nil
}

// resolvePDSEndpoint resolves the PDS endpoint for a given DID
func (h *ListHandler) resolvePDSEndpoint(ctx context.Context, did string) (string, error) {
	dir := identity.DefaultDirectory()
	atid, err := syntax.ParseAtIdentifier(did)
	if err != nil {
		return "", err
	}

	ident, err := dir.Lookup(ctx, *atid)
	if err != nil {
		return "", err
	}

	return ident.PDSEndpoint(), nil
}

// resolveTasksFromURIs fetches task records from their AT URIs
func (h *ListHandler) resolveTasksFromURIs(ctx context.Context, sess *bskyoauth.Session, taskURIs []string) ([]*models.Task, error) {
	tasks := make([]*models.Task, 0, len(taskURIs))

	for _, uri := range taskURIs {
		// Parse the URI to extract collection and rkey
		// Format: at://did:plc:xxx/app.attodo.task/rkey
		parts := strings.Split(uri, "/")
		if len(parts) < 4 {
			log.Printf("Invalid task URI format: %s", uri)
			continue
		}

		collection := parts[len(parts)-2]
		rkey := parts[len(parts)-1]

		// Only fetch if it's a task collection
		if collection != "app.attodo.task" {
			log.Printf("Skipping non-task URI: %s", uri)
			continue
		}

		// Fetch the task record
		taskRecord, err := h.getTaskRecord(ctx, sess, rkey)
		if err != nil {
			log.Printf("Failed to fetch task %s: %v", rkey, err)
			continue
		}

		// Parse task
		task := parseTaskRecord(taskRecord)
		task.RKey = rkey
		task.URI = uri

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// getTaskRecord retrieves a single task record using com.atproto.repo.getRecord
func (h *ListHandler) getTaskRecord(ctx context.Context, sess *bskyoauth.Session, rkey string) (map[string]interface{}, error) {
	// Build the XRPC URL for tasks
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?repo=%s&collection=%s&rkey=%s",
		sess.PDS, sess.DID, "app.attodo.task", rkey)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+sess.AccessToken)

	// Create HTTP client with DPoP transport
	transport := bskyoauth.NewDPoPTransport(http.DefaultTransport, sess.DPoPKey, sess.AccessToken, sess.DPoPNonce)
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Update nonce if present
	if dpopTransport, ok := transport.(bskyoauth.DPoPTransport); ok {
		sess.DPoPNonce = dpopTransport.GetNonce()
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("XRPC ERROR %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var result struct {
		URI   string                 `json:"uri"`
		CID   string                 `json:"cid"`
		Value map[string]interface{} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Value, nil
}

// parseTaskRecord parses a task record from AT Protocol
func parseTaskRecord(record map[string]interface{}) *models.Task {
	task := &models.Task{}

	if title, ok := record["title"].(string); ok {
		task.Title = title
	}
	if desc, ok := record["description"].(string); ok {
		task.Description = desc
	}
	if completed, ok := record["completed"].(bool); ok {
		task.Completed = completed
	}
	if createdAt, ok := record["createdAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			task.CreatedAt = t
		}
	}
	if completedAt, ok := record["completedAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, completedAt); err == nil {
			task.CompletedAt = &t
		}
	}
	// Parse due date if present
	if dueDate, ok := record["dueDate"].(string); ok {
		if t, err := time.Parse(time.RFC3339, dueDate); err == nil {
			task.DueDate = &t
		}
	}
	// Parse tags
	if tags, ok := record["tags"].([]interface{}); ok {
		task.Tags = make([]string, 0, len(tags))
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				task.Tags = append(task.Tags, tagStr)
			}
		}
	}

	return task
}

// withRetry handles token refresh and retries
// WithRetry executes an operation with automatic token refresh on errors (public for cross-handler access)
func (h *ListHandler) WithRetry(ctx context.Context, sess *bskyoauth.Session, fn func(*bskyoauth.Session) error) (*bskyoauth.Session, error) {
	const maxRetries = 3

	for i := 0; i < maxRetries; i++ {
		err := fn(sess)
		if err == nil {
			return sess, nil
		}

		// Check if it's a token expiration error
		if strings.Contains(err.Error(), "400") || strings.Contains(err.Error(), "401") {
			log.Printf("Token may be expired, attempting refresh (attempt %d/%d)", i+1, maxRetries)

			// Try to refresh the token
			newSess, refreshErr := h.client.RefreshToken(ctx, sess)
			if refreshErr != nil {
				log.Printf("Failed to refresh token: %v", refreshErr)
				return sess, err // Return original error
			}

			sess = newSess
			continue
		}

		// Not a token error, return immediately
		return sess, err
	}

	return sess, fmt.Errorf("max retries exceeded")
}

// Helper functions for record building/parsing

func buildListRecord(list *models.TaskList) map[string]interface{} {
	return map[string]interface{}{
		"$type":       ListCollection,
		"name":        list.Name,
		"description": list.Description,
		"taskUris":    list.TaskURIs,
		"createdAt":   list.CreatedAt.Format(time.RFC3339),
		"updatedAt":   list.UpdatedAt.Format(time.RFC3339),
	}
}

func parseListRecord(value map[string]interface{}) *models.TaskList {
	list := &models.TaskList{}

	if name, ok := value["name"].(string); ok {
		list.Name = name
	}
	if desc, ok := value["description"].(string); ok {
		list.Description = desc
	}
	if taskURIs, ok := value["taskUris"].([]interface{}); ok {
		list.TaskURIs = make([]string, 0, len(taskURIs))
		for _, uri := range taskURIs {
			if uriStr, ok := uri.(string); ok {
				list.TaskURIs = append(list.TaskURIs, uriStr)
			}
		}
	}
	if createdAt, ok := value["createdAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			list.CreatedAt = t
		}
	}
	if updatedAt, ok := value["updatedAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			list.UpdatedAt = t
		}
	}

	return list
}

// getPublicRecord retrieves a list record publicly (no authentication)
func (h *ListHandler) getPublicRecord(ctx context.Context, pds, did, rkey string) (map[string]interface{}, error) {
	// Build the XRPC URL
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?repo=%s&collection=%s&rkey=%s",
		pds, did, ListCollection, rkey)

	// Create request (no authentication for public access)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("XRPC ERROR %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var result struct {
		URI   string                 `json:"uri"`
		CID   string                 `json:"cid"`
		Value map[string]interface{} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Value, nil
}

// resolvePublicTasksFromURIs fetches task records publicly (no authentication)
func (h *ListHandler) resolvePublicTasksFromURIs(ctx context.Context, pds, did string, taskURIs []string) ([]*models.Task, error) {
	tasks := make([]*models.Task, 0, len(taskURIs))

	for _, uri := range taskURIs {
		// Parse the URI to extract rkey
		// Format: at://did:plc:xxx/app.attodo.task/rkey
		parts := strings.Split(uri, "/")
		if len(parts) < 4 {
			log.Printf("Invalid task URI format: %s", uri)
			continue
		}

		collection := parts[len(parts)-2]
		rkey := parts[len(parts)-1]

		// Only fetch if it's a task collection
		if collection != "app.attodo.task" {
			log.Printf("Skipping non-task URI: %s", uri)
			continue
		}

		// Fetch the task record publicly
		taskRecord, err := h.getPublicTaskRecord(ctx, pds, did, rkey)
		if err != nil {
			log.Printf("Failed to fetch public task %s: %v", rkey, err)
			continue
		}

		// Parse task
		task := parseTaskRecord(taskRecord)
		task.RKey = rkey
		task.URI = uri

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// getPublicTaskRecord retrieves a single task record publicly (no authentication)
func (h *ListHandler) getPublicTaskRecord(ctx context.Context, pds, did, rkey string) (map[string]interface{}, error) {
	// Build the XRPC URL for tasks
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?repo=%s&collection=%s&rkey=%s",
		pds, did, "app.attodo.task", rkey)

	// Create request (no authentication for public access)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("XRPC ERROR %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var result struct {
		URI   string                 `json:"uri"`
		CID   string                 `json:"cid"`
		Value map[string]interface{} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Value, nil
}
