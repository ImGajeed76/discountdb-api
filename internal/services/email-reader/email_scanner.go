package email_reader

import "log"

func ScanNewEmails(gmail *GMailClient) error {
	emails, err := gmail.ListEmails()
	if err != nil {
		return err
	}

	log.Printf("Found %d new emails", len(emails))

	return nil
}
