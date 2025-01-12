package database

import (
	"context"
	"errors"

	"github.com/arangodb/go-driver"

	"github.com/SecurityBrewery/catalyst/database/busdb"
	"github.com/SecurityBrewery/catalyst/generated/models"
)

func toAutomation(doc *models.AutomationForm) interface{} {
	return &models.Automation{
		Image:  doc.Image,
		Script: doc.Script,
		Schema: doc.Schema,
		Type:   doc.Type,
	}
}

func toAutomationResponse(id string, doc models.Automation) *models.AutomationResponse {
	return &models.AutomationResponse{
		ID:     id,
		Image:  doc.Image,
		Script: doc.Script,
		Schema: doc.Schema,
		Type:   doc.Type,
	}
}

func (db *Database) AutomationCreate(ctx context.Context, automation *models.AutomationForm) (*models.AutomationResponse, error) {
	if automation == nil {
		return nil, errors.New("requires automation")
	}
	if automation.ID == "" {
		return nil, errors.New("requires automation ID")
	}

	var doc models.Automation
	newctx := driver.WithReturnNew(ctx, &doc)

	meta, err := db.automationCollection.CreateDocument(ctx, newctx, automation.ID, toAutomation(automation))
	if err != nil {
		return nil, err
	}

	return toAutomationResponse(meta.Key, doc), nil
}

func (db *Database) AutomationGet(ctx context.Context, id string) (*models.AutomationResponse, error) {
	var doc models.Automation
	meta, err := db.automationCollection.ReadDocument(ctx, id, &doc)
	if err != nil {
		return nil, err
	}

	return toAutomationResponse(meta.Key, doc), nil
}

func (db *Database) AutomationUpdate(ctx context.Context, id string, automation *models.AutomationForm) (*models.AutomationResponse, error) {
	var doc models.Automation
	ctx = driver.WithReturnNew(ctx, &doc)

	meta, err := db.automationCollection.ReplaceDocument(ctx, id, toAutomation(automation))
	if err != nil {
		return nil, err
	}

	return toAutomationResponse(meta.Key, doc), nil
}

func (db *Database) AutomationDelete(ctx context.Context, id string) error {
	_, err := db.automationCollection.RemoveDocument(ctx, id)
	return err
}

func (db *Database) AutomationList(ctx context.Context) ([]*models.AutomationResponse, error) {
	query := "FOR d IN @@collection SORT d._key ASC RETURN UNSET(d, 'script')"
	cursor, _, err := db.Query(ctx, query, map[string]interface{}{"@collection": AutomationCollectionName}, busdb.ReadOperation)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()
	var docs []*models.AutomationResponse
	for {
		var doc models.Automation
		meta, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}

		docs = append(docs, toAutomationResponse(meta.Key, doc))
	}

	return docs, err
}
