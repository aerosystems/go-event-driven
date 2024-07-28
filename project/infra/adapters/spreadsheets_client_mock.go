package adapters

import (
	"context"
	"sync"
)

type SpreadsheetsClientMock struct {
	mock sync.Mutex

	AppendedRows map[string][][]string
}

func (scm *SpreadsheetsClientMock) AppendRow(_ context.Context, spreadsheetName string, row []string) error {
	scm.mock.Lock()
	defer scm.mock.Unlock()

	if scm.AppendedRows == nil {
		scm.AppendedRows = make(map[string][][]string)
	}

	scm.AppendedRows[spreadsheetName] = append(scm.AppendedRows[spreadsheetName], row)

	return nil
}
