package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/manyminds/api2go"
	"github.com/manyminds/api2go/jsonapi"
	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
)

type Note struct {
	ID       int64     `json:"-" db:"id,omitempty"`
	Content  string    `json:"content" db:"content"`
	Asset    *Asset    `json:"-" db:"-"`
	AssetID  string    `json:"-" db:"asset_id"`
	Created  time.Time `json:"created" db:"created"`
	Modified time.Time `json:"modified" db:"modified"`
}

func (n Note) GetName() string {
	return "notes"
}

func (n Note) GetID() string {
	return strconv.FormatInt(n.ID, 10)
}

func (n *Note) SetID(id string) error {
	if id == "" {
		return nil
	}

	i, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	n.ID = int64(i)
	return nil
}

func (n Note) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{{
		Type:         "assets",
		Name:         "asset",
		Relationship: jsonapi.ToOneRelationship,
	}}
}

func (n Note) GetReferencedIDs() []jsonapi.ReferenceID {
	return []jsonapi.ReferenceID{{
		ID:           base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(n.AssetID)),
		Type:         "assets",
		Name:         "asset",
		Relationship: jsonapi.ToOneRelationship,
	}}
}

func (n *Note) SetToOneReferenceID(name, ID string) error {
	if name != "asset" {
		return fmt.Errorf("No such relationship %s", name)
	}

	id, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(ID)

	if err != nil {
		return err
	}

	n.AssetID = string(id)
	return DB.Tx(DB.Context(), func(tx sqlbuilder.Tx) error {
		return tx.Collection("asset").Find(db.Cond{"id": string(id), "deleted": "FALSE"}).One(&n.Asset)
	})
}

func NewNoteResource() *NoteResource {
	return &NoteResource{100}
}

type NoteResource struct {
	PageSize int
}

func (n *NoteResource) FindAll(req api2go.Request) (api2go.Responder, error) {
	results := make([]Note, 0, n.PageSize)
	filters := ParseFilters(req.QueryParams)

	ID, ok := req.QueryParams["assetsID"]
	if ok {
		aid, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(ID[0])
		if err == nil {
			filters["asset_id"] = string(aid)
		}
	}

	err := DB.Collection("note").Find().And(filters).Limit(n.PageSize).OrderBy("modified").All(&results)

	if err != nil {
		return nil, err
	}

	return &Response{results, http.StatusOK}, nil
}

func (s *NoteResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	id, err := strconv.Atoi(ID)
	if err != nil {
		return nil, err
	}

	var note Note
	err = DB.Collection("note").Find(db.Cond{"id": id}).One(&note)

	if err != nil {
		return nil, err
	}

	return &Response{note, http.StatusOK}, nil
}

func (n *NoteResource) Create(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	note := obj.(Note)
	note.Created = time.Now().UTC()
	note.Modified = note.Created

	err := DB.Tx(DB.Context(), func(tx sqlbuilder.Tx) error {
		_, err := tx.Collection("note").Insert(note)
		return err
	})

	if err != nil {
		return nil, err
	}

	return &Response{note, http.StatusCreated}, nil
}
