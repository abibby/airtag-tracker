run:
	GOOS=linux GOARCH=amd64 go build -o airtag-tracker && docker compose up