package service

import (
	"context"
	"tickets/models"
)

type SpreadsheetsClient interface {
	AppendRow(ctx context.Context, spreadsheetName string, row []string) error
}

type ReceiptsClient interface {
	IssueReceipt(ctx context.Context, request models.IssueReceiptRequest) error
}
