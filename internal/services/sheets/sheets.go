package sheets

import (
	"fmt"

	"google.golang.org/api/sheets/v4"
)

// ReadRange reads raw 2D grid values from a spreadsheet range (e.g. "Sheet1!A1:E10").
func (s *Service) ReadRange(spreadsheetID, readRange string) ([][]interface{}, error) {
	resp, err := s.client.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("read sheet range %s: %w", readRange, err)
	}
	return resp.Values, nil
}

// AppendRow appends a row of values to the specified sheet/table range.
func (s *Service) AppendRow(spreadsheetID, appendRange string, row []interface{}) error {
	valRange := &sheets.ValueRange{Values: [][]interface{}{row}}
	call := s.client.Spreadsheets.Values.Append(spreadsheetID, appendRange, valRange).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS")

	if _, err := call.Do(); err != nil {
		return fmt.Errorf("append sheet row: %w", err)
	}
	return nil
}

// CreateSpreadsheet creates a new Google Spreadsheet with a title.
func (s *Service) CreateSpreadsheet(title string) (*sheets.Spreadsheet, error) {
	ss := &sheets.Spreadsheet{Properties: &sheets.SpreadsheetProperties{Title: title}}
	res, err := s.client.Spreadsheets.Create(ss).Do()
	if err != nil {
		return nil, fmt.Errorf("create spreadsheet: %w", err)
	}
	return res, nil
}

// AddSheet creates a new tab/sheet within an existing spreadsheet.
func (s *Service) AddSheet(spreadsheetID, title string) error {
	req := &sheets.Request{
		AddSheet: &sheets.AddSheetRequest{Properties: &sheets.SheetProperties{Title: title}},
	}
	batchReq := &sheets.BatchUpdateSpreadsheetRequest{Requests: []*sheets.Request{req}}
	if _, err := s.client.Spreadsheets.BatchUpdate(spreadsheetID, batchReq).Do(); err != nil {
		return fmt.Errorf("add sheet %s: %w", title, err)
	}
	return nil
}

// UpdateRange overwrites specific cells or a range with new values.
func (s *Service) UpdateRange(spreadsheetID, writeRange string, values [][]interface{}) error {
	valRange := &sheets.ValueRange{Values: values}
	call := s.client.Spreadsheets.Values.Update(spreadsheetID, writeRange, valRange).
		ValueInputOption("USER_ENTERED")
	if _, err := call.Do(); err != nil {
		return fmt.Errorf("update sheet range %s: %w", writeRange, err)
	}
	return nil
}
