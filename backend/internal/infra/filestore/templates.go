// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package filestore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mdhender/ec/internal/app"
	"github.com/mdhender/ec/internal/domain"
)

var _ app.TemplateStore = (*Store)(nil)

// ReadHomeworldTemplate reads homeworld-template.json from dataPath.
func (s *Store) ReadHomeworldTemplate(dataPath string) (domain.HomeworldTemplate, error) {
	path := filepath.Join(dataPath, "homeworld-template.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.HomeworldTemplate{}, fmt.Errorf("reading homeworld template: %w", err)
	}
	var tmpl domain.HomeworldTemplate
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return domain.HomeworldTemplate{}, fmt.Errorf("parsing homeworld template: %w", err)
	}
	return tmpl, nil
}

// ReadColonyTemplate reads colony-template.json from dataPath.
func (s *Store) ReadColonyTemplate(dataPath string) (domain.ColonyTemplate, error) {
	path := filepath.Join(dataPath, "colony-template.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.ColonyTemplate{}, fmt.Errorf("reading colony template: %w", err)
	}
	var tmpl domain.ColonyTemplate
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return domain.ColonyTemplate{}, fmt.Errorf("parsing colony template: %w", err)
	}
	return tmpl, nil
}
