package main

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/manyminds/api2go"
	"github.com/manyminds/api2go/jsonapi"
	"github.com/mattn/go-sqlite3"
	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
)

type Asset struct {
	ID       string    `json:"-"`
	Uri      string    `json:"uri" db:"id,omitempty"`
	Name     string    `json:"name" db:"name"`
	Created  time.Time `json:"created" db:"created"`
	Modified time.Time `json:"modified" db:"modified"`
	Deleted  bool      `json:"-" db:"deleted,omitempty"`
	Notes    []*Note   `json:"-"`
	NoteIDs  []string  `json:"-"`
}

func (a Asset) GetName() string {
	return "assets"
}

func (a Asset) GetID() string {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(a.Uri))
}

func (a *Asset) SetID(id string) error {
	a.ID = id
	return nil
}

func (a Asset) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{{
		Type: "notes",
		Name: "notes",
	}}
}

func (a Asset) GetReferencedIDs() []jsonapi.ReferenceID {
	return []jsonapi.ReferenceID{}
	// notes := make([]Note, 0)
	// err := DB.Collection("note").Find(db.Cond{"asset_id": a.Uri}).OrderBy("created").All(&notes)

	// if err != nil {
	// 	return []jsonapi.ReferenceID{}
	// }

	// results := make([]jsonapi.ReferenceID, len(notes))
	// for i, note := range notes {
	// 	results[i] = jsonapi.ReferenceID{
	// 		ID:           strconv.FormatInt(note.ID, 10),
	// 		Type:         "notes",
	// 		Name:         "notes",
	// 		Relationship: jsonapi.ToManyRelationship,
	// 	}
	// }

	// return results
}

func (a Asset) GetReferencedStructs() []jsonapi.MarshalIdentifier {
	result := []jsonapi.MarshalIdentifier{}
	return result
}

func (a *Asset) SetToManyReferenceIDs(name string, IDs []string) error {
	a.NoteIDs = make([]string, len(IDs))

	for i, ID := range IDs {
		a.NoteIDs[i] = ID
	}

	return nil
}

func NewAssetResource() *AssetResource {
	return &AssetResource{
		100,
		base64.URLEncoding.WithPadding(base64.NoPadding),
	}
}

type AssetResource struct {
	PageSize int
	encoder  *base64.Encoding
}

func (a AssetResource) FindAll(req api2go.Request) (api2go.Responder, error) {
	results := make([]Asset, 0, a.PageSize)
	filters := ParseFilters(req.QueryParams)

	err := DB.Tx(nil, func(tx sqlbuilder.Tx) error {
		return tx.Collection("asset").Find(db.Cond{"deleted": "FALSE"}).And(filters).Limit(a.PageSize).OrderBy("modified").All(&results)
	})

	if err != nil {
		return nil, err
	}

	return &Response{results, http.StatusOK}, nil
}

func (a AssetResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	var empty Asset

	id, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(ID)

	if err != nil {
		panic(err)
	}

	err = DB.Collection("asset").Find(db.Cond{"id": string(id)}).One(&empty)

	if err != nil {
		if err == db.ErrNoMoreRows {
			return nil, api2go.NewHTTPError(err, "Not Found", http.StatusNotFound)
		}
		return nil, err
	}

	if empty.Deleted {
		return nil, api2go.NewHTTPError(nil, "Gone", http.StatusGone)
	}

	return &Response{empty, http.StatusOK}, nil
}

func (a AssetResource) Create(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	asset := obj.(Asset)
	asset.Created = time.Now().UTC()
	asset.Modified = asset.Created

	err := DB.Tx(DB.Context(), func(tx sqlbuilder.Tx) error {
		_, err := tx.Collection("asset").Insert(asset)
		return err
	})

	if err != nil {
		sqliteError, ok := err.(sqlite3.Error)

		if ok && sqliteError.Code == sqlite3.ErrConstraint {
			ae := api2go.NewHTTPError(sqliteError, "", http.StatusConflict)
			ae.Errors = []api2go.Error{{
				Title:  "Conflict",
				Status: "409",
				Detail: sqliteError.Error(),
			}}
			return nil, ae
		}

		return nil, err
	}

	return &Response{asset, http.StatusCreated}, nil
}

func (a AssetResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	return nil, api2go.NewHTTPError(nil, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (a AssetResource) Delete(ID string, r api2go.Request) (api2go.Responder, error) {
	response, err := a.FindOne(ID, r)

	if err != nil {
		return nil, err
	}

	asset := response.Result().(Asset)

	err = DB.Tx(DB.Context(), func(tx sqlbuilder.Tx) error {
		return tx.Collection("asset").Find(db.Cond{"id": asset.Uri}).Update(map[string]bool{"deleted": true})
	})

	if err != nil {
		return nil, err
	}

	return &Response{nil, http.StatusNoContent}, nil
}
