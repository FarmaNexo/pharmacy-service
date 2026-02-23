// internal/application/commands/upload_authorization_document_command.go
package commands

import "io"

type UploadAuthorizationDocumentCommand struct {
	PharmacyID  string
	Reader      io.Reader
	Filename    string
	ContentType string
	Size        int64
}

func (c UploadAuthorizationDocumentCommand) GetName() string {
	return "UploadAuthorizationDocumentCommand"
}
