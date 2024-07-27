package adapters

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/spreadsheets"
	"net/http"
	"sync"
)

type SpreadsheetsClient struct {
	clients *clients.Clients
}

func NewSpreadsheetsClient(clients *clients.Clients) *SpreadsheetsClient {
	return &SpreadsheetsClient{
		clients: clients,
	}
}

func (c SpreadsheetsClient) AppendRow(ctx context.Context, spreadsheetName string, row []string) error {
	request := spreadsheets.PostSheetsSheetRowsJSONRequestBody{
		Columns: row,
	}

	fmt.Println("spreadsheetName:" + spreadsheetName)

	sheetsResp, err := c.clients.Spreadsheets.PostSheetsSheetRowsWithResponse(ctx, spreadsheetName, request)
	if err != nil {
		return err
	}
	if sheetsResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", sheetsResp.StatusCode())
	}

	return nil
}

type SpreadsheetsClientMock struct {
	mock sync.Mutex

	SpreadsheetNames []string
}

func (scm *SpreadsheetsClientMock) AppendRow(_ context.Context, spreadsheetName string, _ []string) error {
	scm.mock.Lock()
	defer scm.mock.Unlock()

	scm.SpreadsheetNames = append(scm.SpreadsheetNames, spreadsheetName)
	return nil
}
