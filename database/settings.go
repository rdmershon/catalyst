package database

import (
	"context"
	"errors"

	"github.com/arangodb/go-driver"
	"github.com/gin-gonic/gin"

	"github.com/SecurityBrewery/catalyst/database/busdb"
	"github.com/SecurityBrewery/catalyst/generated/models"
)

func toUserDataResponse(key string, doc *models.UserData) *models.UserDataResponse {
	return &models.UserDataResponse{
		Email:      doc.Email,
		ID:         key,
		Image:      doc.Image,
		Name:       doc.Name,
		Timeformat: doc.Timeformat,
	}
}

func (db *Database) UserDataCreate(ctx context.Context, id string, userdata *models.UserData) error {
	if userdata == nil {
		return errors.New("requires setting")
	}
	if id == "" {
		return errors.New("requires username")
	}

	_, err := db.userdataCollection.CreateDocument(ctx, ctx, id, userdata)
	return err
}

func (db *Database) UserDataGetOrCreate(ctx *gin.Context, id string, newUserData *models.UserData) (*models.UserDataResponse, error) {
	setting, err := db.UserDataGet(ctx, id)
	if err != nil {
		return toUserDataResponse(id, newUserData), db.UserDataCreate(ctx, id, newUserData)
	}
	return setting, nil
}

func (db *Database) UserDataGet(ctx context.Context, id string) (*models.UserDataResponse, error) {
	var doc models.UserData
	meta, err := db.userdataCollection.ReadDocument(ctx, id, &doc)
	if err != nil {
		return nil, err
	}

	return toUserDataResponse(meta.Key, &doc), err
}

func (db *Database) UserDataList(ctx context.Context) ([]*models.UserDataResponse, error) {
	query := "FOR d IN @@collection SORT d.username ASC RETURN d"
	cursor, _, err := db.Query(ctx, query, map[string]interface{}{"@collection": UserDataCollectionName}, busdb.ReadOperation)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()
	var docs []*models.UserDataResponse
	for {
		var doc models.UserData
		meta, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}
		docs = append(docs, toUserDataResponse(meta.Key, &doc))
	}

	return docs, err
}

func (db *Database) UserDataUpdate(ctx context.Context, id string, userdata *models.UserData) (*models.UserDataResponse, error) {
	var doc models.UserData
	ctx = driver.WithReturnNew(ctx, &doc)

	meta, err := db.userdataCollection.ReplaceDocument(ctx, id, userdata)
	if err != nil {
		return nil, err
	}

	return toUserDataResponse(meta.Key, &doc), nil
}
