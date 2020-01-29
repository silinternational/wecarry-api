package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"

	"github.com/silinternational/wecarry-api/domain"
)

type PostFile struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	PostID    int       `json:"post_id" db:"post_id"`
	Post      Post      `belongs_to:"posts"`
	FileID    int       `json:"file_id" db:"file_id"`
	File      File      `belongs_to:"files"`
}

// String can be helpful for serializing the model
func (p PostFile) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// PostFiles is merely for convenience and brevity
type PostFiles []PostFile

// String can be helpful for serializing the model
func (p PostFiles) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (p *PostFile) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (p *PostFile) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (p *PostFile) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Create stores the PostFile data as a new record in the database.
func (p *PostFile) Create() error {
	return create(p)
}

// AttachFile assigns a previously-stored File to this PostFile. Parameter `fileID` is the UUID
// of the file to attach.
func (p *PostFile) AttachFile(fileID string) (File, error) {
	if p.ID < 1 {
		return File{}, fmt.Errorf("invalid PostFile ID %d", p.ID)
	}

	var f File
	if err := f.FindByUUID(fileID); err != nil {
		err = fmt.Errorf("error finding post file with id %s ... %s", fileID, err)
		return f, err
	}

	oldID := p.FileID
	p.FileID = f.ID
	if err := DB.UpdateColumns(p, "file_id"); err != nil {
		return f, err
	}

	if err := f.SetLinked(); err != nil {
		domain.ErrLogger.Printf("error marking post file %d as linked, %s", f.ID, err)
	}

	oldFile := File{ID: oldID}
	if err := oldFile.ClearLinked(); err != nil {
		domain.ErrLogger.Printf("error marking post file %d as unlinked, %s", oldFile.ID, err)
	}

	return f, nil
}
