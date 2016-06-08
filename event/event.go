package event

import (
	"encoding/json"

	"google.golang.org/cloud/datastore"

	"github.com/builderscon/octav/octav/tools"
	"github.com/pkg/errors"
)

func NewCoreAttrs(name string) *CoreAttrs {
	c := &CoreAttrs{}
	c.SetID(tools.UUID())
	c.SetName(name)
	return c
}

func (c CoreAttrs) ID() string {
	return c.id
}

func (c CoreAttrs) Name() string {
	return c.name
}

func (c *CoreAttrs) SetID(id string) {
	c.id = id
}

func (c *CoreAttrs) SetName(name string) {
	c.name = name
}

func (c *CoreAttrs) UnmarshalJSON(buf []byte) error {
	m := struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{}

	if err := json.Unmarshal(buf, &m); err != nil {
		return errors.Wrap(err, "failed to unmarshal CoreAttrs")
	}

	c.SetID(m.ID)
	c.SetName(m.Name)
	return nil
}

func (c *CoreAttrs) MarshalJSON() ([]byte, error) {
	m := struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{}
	m.ID = c.id
	m.Name = c.name

	buf, err := json.Marshal(&m)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal CoreAttrs")
	}

	return buf, nil
}

func (c *CoreAttrs) Save() ([]datastore.Property, error) {
	return []datastore.Property{
		{
			Name:  "ID",
			Value: c.id,
		},
		{
			Name:  "Name",
			Value: c.name,
		},
	}, nil
}

func (c *CoreAttrs) Load(ps []datastore.Property) error {
	for _, p := range ps {
		switch p.Name {
		case "ID":
			c.id = p.Value.(string)
		case "Name":
			c.name = p.Value.(string)
		}
	}
	return nil
}
