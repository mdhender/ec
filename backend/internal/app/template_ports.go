// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

import "github.com/mdhender/ec/internal/domain"

// TemplateStore reads setup templates from the data directory.
// Template files are read-only; this interface has no write methods.
type TemplateStore interface {
	// ReadHomeworldTemplate reads homeworld-template.json from dataPath.
	ReadHomeworldTemplate(dataPath string) (domain.HomeworldTemplate, error)
	// ReadColonyTemplate reads colony-template.json from dataPath.
	ReadColonyTemplate(dataPath string) (domain.ColonyTemplate, error)
}
