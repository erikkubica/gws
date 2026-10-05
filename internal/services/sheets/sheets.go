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
	valRange := &sheets.ValueRange{
		Values: [][]interface{}{row},
	}
	call := s.client.Spreadsheets.Values.Append(spreadsheetID, appendRange, valRange).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS")

	if _, err := call.Do(); err != nil {
		return fmt.Errorf("append sheet row: %w", err)
	}
	return nil
}
