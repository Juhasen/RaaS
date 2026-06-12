package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"listing/handlers"
	"listing/models"
	"listing/repository"
	"listing/service"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestListingHandler_CRUD(t *testing.T) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Skip("Skipping handler test: MongoDB connection failed")
		return
	}
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	if err := client.Ping(ctx, nil); err != nil {
		t.Skip("Skipping handler test: MongoDB ping failed")
		return
	}

	dbName := "raas_test"
	db := client.Database(dbName)
	_ = db.Collection("listings").Drop(ctx)

	repo := repository.NewMongoRepository(client, dbName)
	ls := service.NewListingService(repo)
	as := service.NewAvailabilityService(repo)
	availH := handlers.NewAvailabilityHandler(as)
	h := handlers.NewListingHandler(ls, availH)

	e := echo.New()

	// 1. Test Create Listing
	payload := `{"host_id":"host-h","title":"Cozy Cabin","description":"A nice cabin","price_per_day":150.0,"location_id":"loc-h","location_label":"New York, NY"}`
	req := httptest.NewRequest(http.MethodPost, "/listings", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.CreateListing(c); err != nil {
		t.Fatalf("CreateListing failed: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rec.Code)
	}

	var created models.Listing
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("Failed to unmarshal created listing: %v", err)
	}
	if created.Title != "Cozy Cabin" || created.ID.IsZero() {
		t.Errorf("Created listing details mismatch: %+v", created)
	}

	// 2. Test Get Listing
	req = httptest.NewRequest(http.MethodGet, "/listings/"+created.ID.Hex(), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(created.ID.Hex())

	if err := h.GetListing(c); err != nil {
		t.Fatalf("GetListing failed: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var fetched models.Listing
	if err := json.Unmarshal(rec.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("Failed to unmarshal fetched listing: %v", err)
	}
	if fetched.ID.Hex() != created.ID.Hex() {
		t.Errorf("Fetched listing ID mismatch: %s vs %s", fetched.ID.Hex(), created.ID.Hex())
	}

	// 3. Test Update Listing
	updatePayload := `{"host_id":"host-h","title":"Super Cozy Cabin","description":"A very nice cabin","price_per_day":180.0,"location_id":"loc-h","location_label":"New York, NY"}`
	req = httptest.NewRequest(http.MethodPut, "/listings/"+created.ID.Hex(), strings.NewReader(updatePayload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(created.ID.Hex())

	if err := h.UpdateListing(c); err != nil {
		t.Fatalf("UpdateListing failed: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var updated models.Listing
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("Failed to unmarshal updated listing: %v", err)
	}
	if updated.Title != "Super Cozy Cabin" || updated.PricePerDay != 180.0 {
		t.Errorf("Updated listing details mismatch: %+v", updated)
	}

	// 4. Test Delete Listing
	req = httptest.NewRequest(http.MethodDelete, "/listings/"+created.ID.Hex(), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(created.ID.Hex())

	if err := h.DeleteListing(c); err != nil {
		t.Fatalf("DeleteListing failed: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", rec.Code)
	}

	// 5. Test Get Deleted Listing (should be 404)
	req = httptest.NewRequest(http.MethodGet, "/listings/"+created.ID.Hex(), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(created.ID.Hex())

	if err := h.GetListing(c); err != nil {
		t.Fatalf("GetListing failed: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for deleted listing, got %d", rec.Code)
	}
}
