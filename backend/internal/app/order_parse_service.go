// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

// ParseOrdersService handles parse-time validation of submitted order text.
type ParseOrdersService struct {
	parser OrderParser
}

// NewParseOrdersService creates a ParseOrdersService with the given parser.
func NewParseOrdersService(parser OrderParser) *ParseOrdersService {
	return &ParseOrdersService{parser: parser}
}

// Parse parses raw order text and returns a ParseResult.
// Returns an error only for unexpected parser failures.
// Empty input is valid and returns a zero-value ParseResult.
func (s *ParseOrdersService) Parse(text string) (ParseResult, error) {
	if text == "" {
		return ParseResult{}, nil
	}

	orders, diagnostics, err := s.parser.Parse(text)
	if err != nil {
		return ParseResult{}, err
	}

	return ParseResult{
		Orders:      orders,
		Diagnostics: diagnostics,
	}, nil
}
